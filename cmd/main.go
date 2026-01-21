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
		metricsAddr          string
		secureMetrics        bool
		probeAddr            string
		enableLeaderElection bool
		leaderNS             string

		// key secret discovery
		keyNS, keyLabelKey, keyLabelVal string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8443", "Metrics bind address.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true, "Serve metrics over HTTPS.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.StringVar(&leaderNS, "leader-election-namespace", "",
		"Namespace for the leader election Lease (defaults to POD_NAMESPACE or age-system).")
	flag.StringVar(&keyNS, "key-namespace", "age-secrets", "Namespace containing AGE key Secrets.")
	flag.StringVar(&keyLabelKey, "key-label-key", "app", "Label key for AGE key Secrets.")
	flag.StringVar(&keyLabelVal, "key-label-val", "age-key", "Label value for AGE key Secrets.")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Resolve leader election namespace:
	// explicit flag > POD_NAMESPACE > age-system
	if leaderNS == "" {
		if podNS := os.Getenv("POD_NAMESPACE"); podNS != "" {
			leaderNS = podNS
		} else {
			leaderNS = "age-system"
		}
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	metricsOpts := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
	}

	if secureMetrics {
		metricsOpts.FilterProvider =
			filters.WithAuthenticationAndAuthorization
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsOpts,
		HealthProbeBindAddress: probeAddr,

		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "age-secret-operator.age.io",
		LeaderElectionNamespace: leaderNS,
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
