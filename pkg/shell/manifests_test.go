package shell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestEphemeralContainerForPodDebug(t *testing.T) {
	ec := ephemeralContainerForPodDebug(Config{RetinaShellImage: "retina-shell:v0.0.1"})
	assert.True(t, strings.HasPrefix(ec.Name, "retina-shell-"), "Ephemeral container name does not start with the expected prefix")
	assert.Equal(t, "retina-shell:v0.0.1", ec.Image)
	assert.Equal(t, []v1.Capability{"ALL"}, ec.SecurityContext.Capabilities.Drop)
	assert.Equal(t, []v1.Capability{"NET_ADMIN", "NET_RAW"}, ec.SecurityContext.Capabilities.Add)
}

func TestHostNetworkPodForNodeDebug(t *testing.T) {
	config := Config{RetinaShellImage: "retina-shell:v0.0.1"}
	pod := hostNetworkPodForNodeDebug(config, "kube-system", "node0001")
	assert.True(t, strings.HasPrefix(pod.Name, "retina-shell-"), "Pod name does not start with the expected prefix")
	assert.Equal(t, "kube-system", pod.Namespace)
	assert.Equal(t, "node0001", pod.Spec.NodeName)
	assert.Equal(t, v1.RestartPolicyNever, pod.Spec.RestartPolicy)
	assert.Equal(t, []v1.Toleration{{Operator: v1.TolerationOpExists}}, pod.Spec.Tolerations)
	assert.True(t, pod.Spec.HostNetwork, "Pod does not have host network enabled")
	assert.Len(t, pod.Spec.Containers, 1)
	assert.Equal(t, "retina-shell:v0.0.1", pod.Spec.Containers[0].Image)
	assert.Equal(t, []v1.Capability{"ALL"}, pod.Spec.Containers[0].SecurityContext.Capabilities.Drop)
	assert.Equal(t, []v1.Capability{"NET_ADMIN", "NET_RAW"}, pod.Spec.Containers[0].SecurityContext.Capabilities.Add)
	assert.Equal(t, 0, len(pod.Spec.Volumes))
	assert.Equal(t, 0, len(pod.Spec.Containers[0].VolumeMounts))
}

func TestHostNetworkPodForNodeDebugWithMountHostFilesystem(t *testing.T) {
	config := Config{
		RetinaShellImage:    "retina-shell:v0.0.1",
		MountHostFilesystem: true,
	}
	pod := hostNetworkPodForNodeDebug(config, "kube-system", "node0001")
	assert.True(t, strings.HasPrefix(pod.Name, "retina-shell-"), "Pod name does not start with the expected prefix")
	assert.Equal(t, "kube-system", pod.Namespace)
	assert.Equal(t, "node0001", pod.Spec.NodeName)
	assert.Equal(t, v1.RestartPolicyNever, pod.Spec.RestartPolicy)
	assert.Equal(t, []v1.Toleration{{Operator: v1.TolerationOpExists}}, pod.Spec.Tolerations)
	assert.True(t, pod.Spec.HostNetwork, "Pod does not have host network enabled")
	assert.Len(t, pod.Spec.Containers, 1)
	assert.Equal(t, "retina-shell:v0.0.1", pod.Spec.Containers[0].Image)
	assert.Equal(t, []v1.Capability{"ALL"}, pod.Spec.Containers[0].SecurityContext.Capabilities.Drop)
	assert.Equal(t, []v1.Capability{"NET_ADMIN", "NET_RAW", "SYS_CHROOT"}, pod.Spec.Containers[0].SecurityContext.Capabilities.Add)
	assert.Equal(t, 1, len(pod.Spec.Volumes))
	assert.Equal(t, "host-filesystem", pod.Spec.Volumes[0].Name)
	assert.Equal(t, 1, len(pod.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, "host-filesystem", pod.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/host", pod.Spec.Containers[0].VolumeMounts[0].MountPath)
}
