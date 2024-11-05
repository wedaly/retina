package shell

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

func ephemeralContainerForPodDebug(config Config) v1.EphemeralContainer {
	return v1.EphemeralContainer{
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
}

func hostNetworkPodForNodeDebug(config Config, debugPodNamespace string, nodeName string, os string) *v1.Pod {
	if os == "windows" {
		return windowsNodeDebugPod(config, debugPodNamespace, nodeName)
	} else {
		return linuxNodeDebugPod(config, debugPodNamespace, nodeName)
	}
}

func linuxNodeDebugPod(config Config, debugPodNamespace string, nodeName string) *v1.Pod {
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

	return pod
}

func windowsNodeDebugPod(config Config, debugPodNamespace string, nodeName string) *v1.Pod {
	// TODO
	return nil
}
