package shell

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config is the configuration for starting a shell in a node or pod.
type Config struct {
	RestConfig          *rest.Config
	RetinaShellImage    string
	MountHostFilesystem bool // Applies only to nodes, not pods.
}

// RunInPod starts an interactive shell in a pod by creating and attaching to an ephemeral container.
func RunInPod(config Config, podNamespace string, podName string) error {
	ctx := context.Background()

	clientset, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return err
	}

	pod, err := clientset.CoreV1().
		Pods(podNamespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := validateOperatingSystemSupportedForNode(ctx, clientset, pod.Spec.NodeName); err != nil {
		return err
	}

	ephemeralContainer := ephemeralContainerForPodDebug(config)
	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ephemeralContainer)

	_, err = clientset.CoreV1().
		Pods(podNamespace).
		UpdateEphemeralContainers(ctx, podName, pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	if err := waitForContainerRunning(ctx, clientset, podNamespace, podName, ephemeralContainer.Name); err != nil {
		return err
	}

	return attachToShell(config.RestConfig, podNamespace, podName, ephemeralContainer.Name, pod)
}

// RunInNode starts an interactive shell on a node by creating a HostNetwork pod and attaching to it.
func RunInNode(config Config, nodeName string, debugPodNamespace string) error {
	ctx := context.Background()

	clientset, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return err
	}

	if err := validateOperatingSystemSupportedForNode(ctx, clientset, nodeName); err != nil {
		return err
	}

	pod := hostNetworkPodForNodeDebug(config, debugPodNamespace, nodeName)

	_, err = clientset.CoreV1().
		Pods(debugPodNamespace).
		Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	defer func() {
		// Best-effort cleanup.
		err := clientset.CoreV1().
			Pods(debugPodNamespace).
			Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete pod %s: %v\n", pod.Name, err)
		}
	}()

	if err := waitForContainerRunning(ctx, clientset, debugPodNamespace, pod.Name, pod.Spec.Containers[0].Name); err != nil {
		return err
	}

	return attachToShell(config.RestConfig, debugPodNamespace, pod.Name, pod.Spec.Containers[0].Name, pod)
}
