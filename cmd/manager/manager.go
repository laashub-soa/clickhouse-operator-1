package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/mackwong/clickhouse-operator/pkg/apis"
	"github.com/mackwong/clickhouse-operator/pkg/controller"
	"github.com/mackwong/clickhouse-operator/pkg/controller/clickhousecluster"
	"github.com/mackwong/clickhouse-operator/version"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)

func printVersion() {
	logrus.Infof("Operator Version: %s", version.Version)
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "kube-config",
			Usage: "specify the kube config path to be used",
			Value: "",
		},
	}
}

func Run(ctx *cli.Context) error {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	printVersion()

	kubeConfigPath := ctx.String("kube-config")

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatal(err, "Failed to get watch namespace")
		return err
	}

	// Get a config to talk to the apiserver

	cfg, err := getConfig(kubeConfigPath)
	if err != nil {
		logrus.Fatal(err, "")
		return err
	}

	// Become the leader before proceeding
	err = leader.Become(context.Background(), "clickhouse-operator-lock")
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	logrus.Info("Registering Components.")

	// Setup Scheme for all resources
	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		logrus.Fatal(err)
		return err
	}

	// Setup all Controllers
	if err = controller.AddToManager(mgr); err != nil {
		logrus.Fatal(err)
		return err
	}

	// Create the prometheus-operator ServiceMonitor resources
	if err = createServiceMonitor(cfg, os.Getenv("NAMESPACE")); err != nil {
		logrus.Warningf("Could not create ServiceMonitor object, error: %s", err.Error())
	}

	logrus.Info("Starting the Cmd.")

	// Start the Cmd
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		logrus.Fatal(err, "Manager exited non-zero")
		return err
	}
	return nil
}

func getConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath == "" {
		return rest.InClusterConfig()
	}
	config, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
}

// CreateServiceMonitor will automatically create the prometheus-operator ServiceMonitor resources
func createServiceMonitor(config *rest.Config, ns string) error {
	//Check if ServiceMonitor is registered in the cluster
	dc := discovery.NewDiscoveryClientForConfigOrDie(config)
	apiVersion := "monitoring.coreos.com/v1"
	kind := "ServiceMonitor"
	if ok, err := k8sutil.ResourceExists(dc, apiVersion, kind); err != nil {
		return err
	} else if !ok {
		return errors.New("cannot find ServiceMonitor registered in the cluster")
	}

	boolTrue := true
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	deployment, err := clientSet.AppsV1().Deployments(ns).Get(os.Getenv("DEPLOYMENT_NAME"), metav1.GetOptions{})
	if err != nil {
		return err
	}

	sm := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-clickhouse",
			Namespace: ns,
			Labels: map[string]string{
				"component":  "clickhouse",
				"prometheus": "kube-prometheus",
				"release":    "prometheus",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "v1",
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               "Deployment",
					Name:               deployment.Name,
					UID:                deployment.UID,
				},
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					clickhousecluster.CreateByLabelKey: clickhousecluster.OperatorLabelKey,
				},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "exporter",
					Interval: "15s",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{Any: true},
			PodTargetLabels:   []string{"instance_name"},
		},
	}

	mClient := monclientv1.NewForConfigOrDie(config)
	_, err = mClient.ServiceMonitors(ns).Create(sm)
	return err
}
