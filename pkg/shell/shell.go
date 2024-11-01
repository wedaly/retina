package shell

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	RestConfig          *rest.Config
	RetinaShellImage    string
	MountHostFilesystem bool
}

func RunInPod(config Config, podNamespace string, podName string) error {
	ctx := context.Background()

	clientset, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return err
	}

	if err := validateOperatingSystemSupportedForPod(ctx, clientset, podNamespace, podName); err != nil {
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
	ctx := context.Background()

	clientset, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return err
	}

	if err := validateOperatingSystemSupportedForNode(ctx, clientset, nodeName); err != nil {
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
