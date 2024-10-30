package shell

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func RunInPod(configFlags *genericclioptions.ConfigFlags, podName string) error {
	fmt.Printf("TODO: ns=%s, pod=%s\n", *configFlags.Namespace, podName)
	// TODO: ephemeral pod in pod netns.
	// TODO: how to get the image name/tag?
	return nil
}

func RunInNode(configFlags *genericclioptions.ConfigFlags, nodeName string) error {
	// TODO: ephemeral pod in node.
	// TODO: how to get the image name/tag?
	fmt.Printf("TODO: node=%s\n", nodeName)
	return nil
}
