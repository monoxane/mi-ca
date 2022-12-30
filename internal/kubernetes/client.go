package kubernetes

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"mi-ca/internal/env"
	"mi-ca/internal/logging"
)

var (
	Client            *kubernetes.Clientset
	deploymentsClient typedappsv1.DeploymentInterface
	config            *rest.Config

	log logging.Logger
)

func Connect() error {
	log = logging.Log.With().Str("module", "kubernetes").Logger()

	if env.APP_MODE == "PROD" {
		log.Info().Msg("connecting to kubernetes api via injected tokens")
		c, err := rest.InClusterConfig()
		if err != nil {
			return err
		}
		config = c
	} else {
		log.Info().Msg("connecting to kubernetes api via local system kubeconfig")
		var kubeconfig *string

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String(
				"kubeconfig",
				filepath.Join(home, ".kube", "config"),
				"(optional) absolute path to the kubeconfig file",
			)
		} else {
			kubeconfig = flag.String(
				"kubeconfig",
				"",
				"absolute path to the kubeconfig file",
			)
		}
		flag.Parse()

		c, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return err
		}
		config = c
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	Client = clientset
	deploymentsClient = Client.AppsV1().Deployments(env.KUBERNETES_NAMESPACE)
	log.Info().Msg("kubernetes connection active")

	return nil
}
