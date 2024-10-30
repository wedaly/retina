package shell

import "k8s.io/cli-runtime/pkg/genericclioptions"

func RunInPod(configFlags *genericclioptions.ConfigFlags, podName, podNamespace string) error {
	// TODO: ephemeral pod in pod netns.
	// TODO: how to get the image name/tag?
	return nil
}

func RunInNode(configFlags *genericclioptions.ConfigFlags, nodeName string) error {
	// TODO: ephemeral pod in node.
	// TODO: how to get the image name/tag?
	return nil
}
