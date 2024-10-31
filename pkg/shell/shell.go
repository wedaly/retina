package shell

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/attach"
	"k8s.io/kubectl/pkg/cmd/exec"
)

const retinaShellCmd = "bash"

type Config struct {
	RestConfig          *rest.Config
	RetinaShellImage    string
	MountHostFilesystem bool
}

func RunInPod(config Config, podNamespace string, podName string) error {
	clientset, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return err
	}

	ephemeralContainer := v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:  fmt.Sprintf("retina-shell-%s", utilrand.String(5)),
			Image: config.RetinaShellImage,
			Stdin: true,
			TTY:   true,
			SecurityContext: &v1.SecurityContext{
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
					Add:  []v1.Capability{"NET_ADMIN", "NET_RAW"},
				},
			},
		},
	}

	ctx := context.Background()
	pod, err := clientset.CoreV1().
		Pods(podNamespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

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

func RunInNode(config Config, nodeName string, debugPodNamespace string) error {
	clientset, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return err
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("retina-shell-%s", utilrand.String(5)),
			Namespace: debugPodNamespace,
		},
		Spec: v1.PodSpec{
			NodeName:      nodeName,
			RestartPolicy: v1.RestartPolicyNever,
			Tolerations:   []v1.Toleration{{Operator: v1.TolerationOpExists}},
			HostNetwork:   true,
			Containers: []v1.Container{
				{
					Name:  "retina-shell",
					Image: config.RetinaShellImage,
					Stdin: true,
					TTY:   true,
					SecurityContext: &v1.SecurityContext{
						Capabilities: &v1.Capabilities{
							Drop: []v1.Capability{"ALL"},
							Add:  []v1.Capability{"NET_ADMIN", "NET_RAW"},
						},
					},
				},
			},
		},
	}

	if config.MountHostFilesystem {
		pod.Spec.Volumes = []v1.Volume{
			{
				Name: "host-filesystem",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/",
					},
				},
			},
		}
		pod.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      "host-filesystem",
				MountPath: "/host",
			},
		}

		pod.Spec.Containers[0].SecurityContext.Capabilities.Add = append(pod.Spec.Containers[0].SecurityContext.Capabilities.Add, "SYS_CHROOT")
	}

	// Create the pod
	ctx := context.Background()
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
		Attach:      &attach.DefaultRemoteAttach{},
		AttachFunc:  attach.DefaultAttachFunc,
		Pod:         pod,
		CommandName: retinaShellCmd,
	}

	return attachOpts.Run()
}

func waitForContainerRunning(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName, containerName string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		pod, err := clientset.CoreV1().
			Pods(namespace).
			Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
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
			return fmt.Errorf("timed out waiting for container %s to start", containerName)
		case <-time.After(1 * time.Second):
		}
	}
}
