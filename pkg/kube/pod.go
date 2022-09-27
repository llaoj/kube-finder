package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod(ctx context.Context, namespaceName, podName string) (pod *corev1.Pod, err error) {
	kubeClient, err := Client()
	if err != nil {
		return nil, err
	}

	return kubeClient.CoreV1().Pods(namespaceName).Get(ctx, podName, metav1.GetOptions{})
}
