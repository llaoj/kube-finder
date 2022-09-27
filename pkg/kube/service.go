package kube

import (
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Service(ctx context.Context, namespaceName, serviceName string) (*v1.Service, error) {
	clientset, err := Client()
	if err != nil {
		return nil, err
	}

	return clientset.CoreV1().Services(namespaceName).Get(ctx, serviceName, metav1.GetOptions{})
}

func ServiceLabelSelector(ctx context.Context, namespaceName, serviceName string) (string, error) {
	service, err := Service(ctx, namespaceName, serviceName)
	if err != nil {
		return "", err
	}
	labelSelector := ""
	for key, value := range service.Spec.Selector {
		labelSelector += key + "=" + value + ","
	}

	return strings.TrimRight(labelSelector, ","), nil
}
