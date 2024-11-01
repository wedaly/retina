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
