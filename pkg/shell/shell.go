package shell

func RunInPod(podName, podNamespace string) error {
	// TODO: ephemeral pod in pod netns.
	// TODO: how to get the image name/tag?
	return nil
}

func RunInNode(nodeName string) error {
	// TODO: ephemeral pod in node.
	// TODO: how to get the image name/tag?
	return nil
}
