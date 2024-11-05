package shell

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getOperatingSystemForNode(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) (string, error) {
	node, err := clientset.CoreV1().
		Nodes().
		Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return node.Labels["kubernetes.io/os"], nil
}
