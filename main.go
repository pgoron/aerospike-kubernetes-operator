/*
Copyright 2021.

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
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	aerospikev1alpha1 "github.com/aerospike/aerospike-kubernetes-operator/api/v1alpha1"
	"github.com/aerospike/aerospike-kubernetes-operator/controllers/aerospikecluster"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	// "github.com/aerospike/aerospike-kubernetes-operator/api"
	"github.com/aerospike/aerospike-management-lib/asconfig"

	// ctrAdmission "github.com/aerospike/aerospike-kubernetes-operator/controllers/admission"
	"github.com/aerospike/aerospike-kubernetes-operator/controllers/configschema"

	log "github.com/inconshreveable/log15"

	// "github.com/operator-framework/operator-sdk/pkg/leader"

	//"github.com/operator-framework/operator-sdk/pkg/restmapper"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8Runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = k8sruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(aerospikev1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686

	// Webhook cert directory path
	certDir = "/tmp/cert"
)

const (
	logLevelEnvVar   = "LOG_LEVEL"
	syncPeriodEnvVar = "SYNC_PERIOD_SECOND"
)

var mgrGlobal ctrl.Manager

var SchemeGroupVersion = schema.GroupVersion{Group: aerospikev1alpha1.GroupVersion.Group, Version: aerospikev1alpha1.GroupVersion.Version}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLogger()

	printVersion()

	watchNs, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// ctx := context.TODO()
	// // Become the leader before proceeding
	// err = leader.Become(ctx, "aerospike-kubernetes-operator-lock")
	// if err != nil {
	// 	setupLog.Error(err, "Failed to become leader")
	// 	os.Exit(1)
	// }

	scheme := k8Runtime.NewScheme()
	SchemeBuilder := k8Runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		setupLog.Error(err, "Failed to add scheme")
		os.Exit(1)
	}
	err = corev1.AddToScheme(scheme)
	if err != nil {
		setupLog.Error(err, "Failed to add scheme")
		os.Exit(1)
	}
	err = appsv1.AddToScheme(scheme)
	if err != nil {
		setupLog.Error(err, "Failed to add scheme")
		os.Exit(1)
	}
	err = storagev1.AddToScheme(scheme)
	if err != nil {
		setupLog.Error(err, "Failed to add scheme")
		os.Exit(1)
	}
	err = admissionregistrationv1.AddToScheme(scheme)
	if err != nil {
		setupLog.Error(err, "Failed to add scheme")
		os.Exit(1)
	}
	if err := aerospikev1alpha1.AddToScheme(scheme); err != nil {
		setupLog.Error(err, "Failed to add schemes")
		os.Exit(1)
	}

	d := getSyncPeriod()
	setupLog.Info("Set sync period", "period", d)

	// clientBuilder := manager.NewClientBuilder().WithUncached(&aerospikev1alpha1.AerospikeCluster{})
	// clientBuilder.Build = newClient
	// Create a new Cmd to provide shared dependencies and start components
	options := ctrl.Options{
		Scheme: scheme,
		// MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: metricsAddr,
		// NewClient:          newClient,
		ClientBuilder: &newClientBuilder{},
		// SyncPeriod:             d,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "96242fdf.aerospike.com",
		HealthProbeBindAddress: probeAddr,
		Port:                   9443,
	}

	// Add support for multiple namespaces given in WATCH_NAMESPACE (e.g. ns1,ns2)
	// For more Info: https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
	if strings.Contains(watchNs, ",") {
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(watchNs, ","))
	} else {
		options.Namespace = watchNs
	}

	mgrGlobal, err = ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "Failed to create manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	// if err := aerospikev1alpha1.AddToScheme(mgrGlobal.GetScheme()); err != nil {
	// 	setupLog.Error(err, "Failed to add schemes")
	// 	os.Exit(1)
	// }
	setupLog.Info("Registering Components")

	setupLog.Info("Init aerospike-server config schemas")
	asconfig.InitFromMap(configschema.SchemaMap)

	if err := (&aerospikecluster.AerospikeClusterReconciler{
		Client: mgrGlobal.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AerospikeCluster"),
		Scheme: mgrGlobal.GetScheme(),
	}).SetupWithManager(mgrGlobal); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AerospikeCluster")
		os.Exit(1)
	}
	if err = (&aerospikev1alpha1.AerospikeCluster{}).SetupWebhookWithManager(mgrGlobal); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "AerospikeCluster")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgrGlobal.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgrGlobal.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	setupLog.Info("starting manager")
	if err := mgrGlobal.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func addKnownTypes(scheme *k8Runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&aerospikev1alpha1.AerospikeCluster{},
	)
	k8v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func printVersion() {
	// setupLog.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	setupLog.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	// setupLog.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func getLogLevel() log.Lvl {
	logLevel, found := os.LookupEnv(logLevelEnvVar)
	if !found {
		return log.LvlInfo
	}
	level, err := log.LvlFromString(logLevel)
	if err != nil {
		return log.LvlInfo
	}
	return level
}

// levelFilterHandler filters log messages based on the current log level.
func levelFilterHandler(h log.Handler, logLevel log.Lvl) log.Handler {
	return log.FilterHandler(func(r *log.Record) (pass bool) {
		return r.Lvl <= logLevel
	}, h)
}

// setupLogger sets up the setupLog from the config.
func setupLogger() {
	handler := log.Root().GetHandler()
	// caller handler
	handler = log.CallerFileHandler(handler)

	handler = levelFilterHandler(handler, getLogLevel())

	log.Root().SetHandler(handler)
}

func getSyncPeriod() *time.Duration {
	sync, found := os.LookupEnv(syncPeriodEnvVar)
	if !found {
		return nil
	}
	syncPeriod, err := strconv.Atoi(sync)
	if err != nil || syncPeriod == 0 {
		return nil
	}
	d := time.Duration(syncPeriod) * time.Second
	return &d
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}

type newClientBuilder struct{}

func (n *newClientBuilder) WithUncached(objs ...crclient.Object) manager.ClientBuilder {
	// n.uncached = append(n.uncached, objs...)
	return n
}

func (n *newClientBuilder) Build(cache cache.Cache, config *rest.Config, options crclient.Options) (crclient.Client, error) {
	// Create the Client for Write operations.
	return crclient.New(config, options)
}

// newClient creates the default caching client
// this will read/write directly from api-server
func newClient(cache cache.Cache, config *rest.Config, options crclient.Options) (crclient.Client, error) {
	// Create the Client for Write operations.
	return crclient.New(config, options)
}
