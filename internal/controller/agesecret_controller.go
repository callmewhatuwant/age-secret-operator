/*
Copyright 2025.
Licensed under the Apache License, Version 2.0 (the "License");
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	age "filippo.io/age"
	"filippo.io/age/armor"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/callmewhatuwant/age-secret-operator/api/v1alpha1"
)

// AgeSecretReconciler reconciles AgeSecret resources.
type AgeSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Configurable via CLI flags (see cmd/main.go)
	KeyNamespace string // default: "age-system"
	KeyLabelKey  string // default: "app"
	KeyLabelVal  string // default: "age-key"
}

// +kubebuilder:rbac:groups=security.age.io,resources=agesecrets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=security.age.io,resources=agesecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.age.io,resources=agesecrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch

func (r *AgeSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("agesecret", req.NamespacedName)

	// 1. Load the AgeSecret resource.
	var cr securityv1alpha1.AgeSecret
	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cr.Status.ObservedGeneration == cr.Generation {
		var existing corev1.Secret
		secretKey := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
		if err := r.Get(ctx, secretKey, &existing); err == nil {
			logger.Info("already reconciled: skipping (observed generation matches and secret exists)")
			return ctrl.Result{}, nil
		}
	}

	// 2. List available AGE key Secrets (by namespace and label).
	keyList := &corev1.SecretList{}
	if err := r.List(ctx, keyList,
		client.InNamespace(r.KeyNamespace),
		client.MatchingLabels{r.KeyLabelKey: r.KeyLabelVal},
	); err != nil {
		logger.Error(err, "failed to list key secrets", "namespace", r.KeyNamespace)
		return ctrl.Result{}, err
	}
	if len(keyList.Items) == 0 {
		logger.Info("no AGE keys found, will retry", "namespace", r.KeyNamespace)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// 3. Decrypt each field in spec.encryptedData.
	plain := map[string][]byte{}
	for field, enc := range cr.Spec.EncryptedData {
		b, keyUsed, derr := decryptWithAge(ctx, enc, keyList.Items, cr.Spec.Recipients)
		if derr != nil {
			logger.Error(derr, "failed to decrypt", "field", field)
			return ctrl.Result{}, fmt.Errorf("decrypt %s: %w", field, derr)
		}
		logger.Info("decrypted field", "field", field, "keySecret", keyUsed)
		plain[field] = b
	}

	// 4. Create or update the target Secret (same name as the CR).
	secretName := cr.Name
	secretKey := types.NamespacedName{Name: secretName, Namespace: cr.Namespace}
	var secret corev1.Secret

	err := r.Get(ctx, secretKey, &secret)
	if apierrors.IsNotFound(err) {
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: cr.Namespace,
			},
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	for k, v := range plain {
		secret.Data[k] = v
	}
	if t := cr.Spec.Template.Type; t != "" {
		secret.Type = corev1.SecretType(t)
	} else {
		secret.Type = corev1.SecretTypeOpaque
	}

	if err := controllerutil.SetControllerReference(&cr, &secret, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if apierrors.IsNotFound(err) {
		if err := r.Create(ctx, &secret); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.Update(ctx, &secret); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 5. Update status — ignore NotFound, keep logs clean.
	cr.Status.ObservedGeneration = cr.Generation
	cr.Status.SecretName = secretName

	if uerr := r.Status().Update(ctx, &cr); uerr != nil {
		if apierrors.IsNotFound(uerr) {
			// CR was deleted before status update — ignore silently.
			return ctrl.Result{}, nil
		}
		logger.V(1).Info("non-fatal: failed to update status", "error", uerr)
	}

	logger.Info("reconciliation completed", "secret", secretKey.String())
	return ctrl.Result{}, nil
}

func (r *AgeSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.AgeSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// decryptWithAge decrypts the provided armored AGE content using available key Secrets.
func decryptWithAge(ctx context.Context, armored string, keySecrets []corev1.Secret, recipients []string) ([]byte, string, error) {
	logger := log.FromContext(ctx)
	for _, ks := range keySecrets {
		name := ks.GetName()

		privBytes, ok := ks.Data["private"]
		if !ok {
			logger.V(1).Info("missing 'private' field in key secret", "secret", name)
			continue
		}
		priv := strings.TrimSpace(string(privBytes))

		id, err := age.ParseX25519Identity(priv)
		if err != nil {
			logger.V(1).Info("failed to parse private identity", "secret", name, "err", err)
			continue
		}

		trimmed := strings.TrimLeft(armored, " \t\r\n")
		var src io.Reader = strings.NewReader(armored)
		if strings.HasPrefix(trimmed, "-----BEGIN AGE ENCRYPTED FILE-----") {
			src = armor.NewReader(strings.NewReader(armored))
		}

		decR, err := age.Decrypt(src, id)
		if err != nil {
			logger.V(1).Info("decryption failed with this key", "secret", name, "err", err, "recipients_hint", recipients)
			continue
		}
		plain, err := io.ReadAll(decR)
		if err != nil {
			logger.V(1).Info("failed to read decrypted data", "secret", name, "err", err)
			continue
		}
		return plain, name, nil
	}
	return nil, "", errors.New("failed to decrypt with any available key")
}
