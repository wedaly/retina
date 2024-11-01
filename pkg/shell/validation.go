package shell

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func validateOperatingSystemSupportedForPod(ctx context.Context, clientset *kubernetes.Clientset, podNamespace, podName string) error {
	pod, err := clientset.CoreV1().
		Pods(podNamespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	nodeName := pod.Spec.NodeName
	return validateOperatingSystemSupportedForNode(ctx, clientset, nodeName)
}

func validateOperatingSystemSupportedForNode(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) error {
	node, err := clientset.CoreV1().
		Nodes().
		Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	osLabel := node.Labels["kubernetes.io/os"]
	if osLabel != "linux" { // Only Linux supported for now.
		return fmt.Errorf("node %s has unsupported operating system %s", nodeName, osLabel)
	}

	return nil
}
