/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"crypto/tls"
	"flag"
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	// webhooks
	// "sigs.k8s.io/controller-runtime/pkg/webhook"

	securityv1alpha1 "github.com/callmewhatuwant/age-secret-operator/api/v1alpha1"
	"github.com/callmewhatuwant/age-secret-operator/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(securityv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// nolint:gocyclo
func main() {
	var (
		metricsAddr                                      string
		metricsCertPath, metricsCertName, metricsCertKey string
		secureMetrics                                    bool
		metricsAuth                                      bool
		enableHTTP2                                      bool
		probeAddr                                        string
		enableLeaderElection                             bool
		leaderNS                                         string
		tlsOpts                                          []func(*tls.Config)
		keyNS, keyLabelKey, keyLabelVal                  string
		// webhooks
		// webhookCertPath, webhookCertName, webhookCertKey string

	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true, "Serve metrics over HTTPS.")
	flag.BoolVar(&metricsAuth, "metrics-auth", false,
		"Enable authentication/authorization on metrics endpoint")

	// webhooks
	// flag.StringVar(&webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	// flag.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	// flag.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")

	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderNS, "leader-election-namespace", "",
		"Namespace for the leader election Lease (defaults to POD_NAMESPACE or age-system).")

	flag.StringVar(&keyNS, "key-namespace", "age-secrets", "Namespace containing AGE key Secrets.")
	flag.StringVar(&keyLabelKey, "key-label-key", "app", "Label key for AGE key Secrets.")
	flag.StringVar(&keyLabelVal, "key-label-val", "age-key", "Label value for AGE key Secrets.")

	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")

	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Resolve leader election namespace:
	// explicit flag > POD_NAMESPACE > age-system
	if leaderNS == "" {
		if podNS := os.Getenv("POD_NAMESPACE"); podNS != "" {
			leaderNS = podNS
		} else {
			leaderNS = "age-system"
		}
	}

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	//// Initial webhook TLS options
	// webhookTLSOpts := tlsOpts
	// webhookServerOptions := webhook.Options{
	// 	TLSOpts: webhookTLSOpts,
	// }
	//
	// if len(webhookCertPath) > 0 {
	// 	setupLog.Info(
	// 		"Initializing webhook certificate watcher using provided certificates",
	// 		"webhook-cert-path", webhookCertPath,
	// 		"webhook-cert-name", webhookCertName,
	// 		"webhook-cert-key", webhookCertKey,
	// 	)
	//
	// 	webhookServerOptions.CertDir = webhookCertPath
	// 	webhookServerOptions.CertName = webhookCertName
	// 	webhookServerOptions.KeyName = webhookCertKey
	// }
	//
	// webhookServer := webhook.NewServer(webhookServerOptions)

	// if len(webhookCertPath) > 0 {
	// 	setupLog.Info(
	// 		"Initializing webhook certificate watcher using provided certificates",
	// 		"webhook-cert-path", webhookCertPath,
	// 		"webhook-cert-name", webhookCertName,
	// 		"webhook-cert-key", webhookCertKey,
	// 	)
	//
	// 	webhookServerOptions.CertDir = webhookCertPath
	// 	webhookServerOptions.CertName = webhookCertName
	// 	webhookServerOptions.KeyName = webhookCertKey
	// }
	//
	// webhookServer := webhook.NewServer(webhookServerOptions)

	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics && metricsAuth {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/prometheus/kustomization.yaml for TLS certification.
	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		metricsServerOptions.CertDir = metricsCertPath
		metricsServerOptions.CertName = metricsCertName
		metricsServerOptions.KeyName = metricsCertKey
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsServerOptions,
		// WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,

		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "age-secret-operator.age.io",
		LeaderElectionNamespace: leaderNS,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	split := strings.Split(keyNS, ",")
	namespaces := append([]string{"age-secrets"}, split...)

	if err := (&controller.AgeSecretReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		KeyNamespace: namespaces,
		KeyLabelKey:  keyLabelKey,
		KeyLabelVal:  keyLabelVal,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AgeSecret")
		os.Exit(1)
	}

	// nolint:goconst
	// if os.Getenv("ENABLE_WEBHOOKS") != "false" {
	//     if err := webhookv1alpha1.SetupServiceLevelObjectiveWebhookWithManager(mgr); err != nil {
	//         setupLog.Error(err, "unable to create webhook", "webhook", "ServiceLevelObjective")
	//         os.Exit(1)
	//     }
	// }
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "leaderElection", enableLeaderElection, "leaderNamespace", leaderNS)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
