package shell

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/attach"
	"k8s.io/kubectl/pkg/cmd/exec"
)

func attachToShell(restConfig *rest.Config, namespace string, podName string, containerName string, pod *v1.Pod) error {
	attachOpts := &attach.AttachOptions{
		Config: restConfig,
		StreamOptions: exec.StreamOptions{
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: containerName,
			IOStreams: genericiooptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stderr,
			},
			Stdin: true,
			TTY:   true,
			Quiet: true,
		},
		Attach:     &attach.DefaultRemoteAttach{},
		AttachFunc: attach.DefaultAttachFunc,
		Pod:        pod,
	}

	return attachOpts.Run()
}

func waitForContainerRunning(ctx context.Context, timeout time.Duration, clientset *kubernetes.Clientset, namespace, podName, containerName string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		pod, err := clientset.CoreV1().
			Pods(namespace).
			Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return waitTimeoutError(timeout, containerName)
			}
			return err
		}

		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == containerName && status.State.Running != nil {
				return nil
			}
		}
		for _, status := range pod.Status.EphemeralContainerStatuses {
			if status.Name == containerName && status.State.Running != nil {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return waitTimeoutError(timeout, containerName)
		case <-time.After(1 * time.Second):
		}
	}
}

func waitTimeoutError(timeout time.Duration, containerName string) error {
	return fmt.Errorf("timed out after %s waiting for container %s to start. The timeout can be increased by setting --timeout", timeout, containerName)
}
