package cmds

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/shawn-hurley/osb-broker-k8s-lib/middleware"
	clientset "k8s.io/client-go/kubernetes"
	clientrest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mackwong/clickhouse-operator/pkg/broker"
	"gitlab.bj.sensetime.com/service-providers/osb-broker-lib/pkg/metrics"
	"gitlab.bj.sensetime.com/service-providers/osb-broker-lib/pkg/rest"
	"gitlab.bj.sensetime.com/service-providers/osb-broker-lib/pkg/server"
)

var Version string

var options struct {
	broker.Options

	Port                 int
	Insecure             bool
	TLSCert              string
	TLSKey               string
	TLSCertFile          string
	TLSKeyFile           string
	AuthenticateK8SToken bool
	KubeConfig           string
}

func BrokerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "port",
			Usage: "option to specify the port for broker to listen on",
			Value: 8443,
		},
		&cli.BoolFlag{
			Name:  "insecure",
			Usage: "use --insecure to use HTTP vs HTTPS.",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "authenticate-k8s-token",
			Usage: "option to specify if the broker should validate the bearer auth token with kubernetes",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "tls-cert-file",
			Usage: "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "tls-private-key-file",
			Usage: "File containing the default x509 private key matching --tls-cert-file.",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "tlsCert",
			Usage: "base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used.",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "tlsKey",
			Usage: "base-64 encoded PEM block to use as the private key matching the TLS certificate.",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "kube-config",
			Usage: "specify the kube config path to be used",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "catalogPath",
			Usage: "The path to the catalog",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "service-config",
			Usage: "specify the brokers config path to be used",
			Value: "/etc/broker/clickhouse.yaml",
		},
	}
}

func BrokerRun(ctx *cli.Context) error {
	logrus.Infof("Version is: %s", Version)

	options.Port = ctx.Int("port")
	options.Insecure = ctx.Bool("insecure")
	options.AuthenticateK8SToken = ctx.Bool("authenticate-k8s-token")
	options.TLSCertFile = ctx.String("tls-cert-file")
	options.TLSKeyFile = ctx.String("tls-private-key-file")
	options.TLSCert = ctx.String("tlsCert")
	options.TLSKey = ctx.String("tlsKey")
	options.KubeConfig = ctx.String("kube-config")
	options.CatalogPath = ctx.String("catalogPath")
	options.ServiceConfigPath = ctx.String("service-config")

	var err error
	if err = run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		logrus.Error(err)
	}
	return err
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	if (options.TLSCert != "" || options.TLSKey != "") &&
		(options.TLSCert == "" || options.TLSKey == "") {
		logrus.Info("To use TLS with specified cert or key data, both --tlsCert and --tlsKey must be used")
		return nil
	}

	addr := ":" + strconv.Itoa(options.Port)

	businessLogic, err := broker.NewCHCBrokerLogic(options.KubeConfig, options.Options)
	if err != nil {
		return err
	}

	err = businessLogic.Recovery()
	if err != nil {
		return err
	}

	// Prom. metrics
	reg := prom.NewRegistry()
	osbMetrics := metrics.New()
	reg.MustRegister(osbMetrics)

	api, err := rest.NewAPISurface(businessLogic, osbMetrics)
	if err != nil {
		return err
	}

	s := server.New(api, reg)
	if options.AuthenticateK8SToken {
		// get k8s client
		k8sClient, err := getKubernetesClient(options.KubeConfig)
		if err != nil {
			return err
		}
		// Create a User Info Authorizer.
		authz := middleware.SARUserInfoAuthorizer{
			SAR: k8sClient.AuthorizationV1().SubjectAccessReviews(),
		}
		// create TokenReviewMiddleware
		tr := middleware.TokenReviewMiddleware{
			TokenReview: k8sClient.AuthenticationV1().TokenReviews(),
			Authorizer:  authz,
		}
		// Use TokenReviewMiddleware.
		s.Router.Use(tr.Middleware)
	}

	logrus.Info("Starting broker!")

	if options.Insecure {
		err = s.Run(ctx, addr)
	} else {
		if options.TLSCert != "" && options.TLSKey != "" {
			logrus.Info("Starting secure broker with TLS cert and key data")
			err = s.RunTLS(ctx, addr, options.TLSCert, options.TLSKey)
		} else {
			if options.TLSCertFile == "" || options.TLSKeyFile == "" {
				logrus.Error("unable to run securely without TLS Certificate and Key. Please review options and if running with TLS, specify --tls-cert-file and --tls-private-key-file or --tlsCert and --tlsKey.")
				return nil
			}
			logrus.Info("Starting secure broker with file based TLS cert and key")
			err = s.RunTLSWithTLSFiles(ctx, addr, options.TLSCertFile, options.TLSKeyFile)
		}
	}
	return err
}

func getKubernetesClient(kubeConfigPath string) (clientset.Interface, error) {
	var clientConfig *clientrest.Config
	var err error
	if kubeConfigPath == "" {
		clientConfig, err = clientrest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeConfigPath)
		if err != nil {
			return nil, err
		}

		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	return clientset.NewForConfig(clientConfig)
}

func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			logrus.Info("Received SIGTERM, exiting gracefully...")
			f()
			os.Exit(0)
		case <-ctx.Done():
			os.Exit(0)
		}
	}
}
