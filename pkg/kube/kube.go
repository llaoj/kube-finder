package kube

import (
	"github.com/llaoj/kube-finder/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func Client() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", config.Get().Kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(kubeconfig)
}
