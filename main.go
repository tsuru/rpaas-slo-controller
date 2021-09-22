package main

import (
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/tsuru/rpaas-operator/api/v1alpha1"
	"github.com/tsuru/rpaas-slo-controller/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
)

var (
	enableLeaderElection = kingpin.Flag(
		"enable-leader-election",
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.").
		Bool()

	metricsAddr = kingpin.Flag(
		"metrics-addr", "The address the metric endpoint binds to.").
		Envar("METRICS_URL").
		Default(":8080").
		String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: *metricsAddr,
		Port:               9443,
		LeaderElection:     *enableLeaderElection,
		LeaderElectionID:   "65e201d7.tsuru.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.RpaasInstanceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("RpaasInstanceReconciler"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RpaasInstance")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
