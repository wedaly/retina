package shell

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

const (
	imageRepo    = "widalytest.azurecr.io/wedaly/retina/retina-shell"
	imageVersion = "v0.0.16-122-g94ca3aa"
)

func RunInPod(configFlags *genericclioptions.ConfigFlags, podName string) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	namespace := *configFlags.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// TODO: and open connection with stdin/stdout?
	// TODO: what happens if you run this twice...?
	ephemeralContainer := v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:  "retina-shell",
			Image: fmt.Sprintf("%s:%s", imageRepo, imageVersion),
			Stdin: true,
			TTY:   true,
			SecurityContext: &v1.SecurityContext{
				Capabilities: &v1.Capabilities{
					Add: []v1.Capability{"NET_ADMIN", "NET_RAW"},
				},
			},
		},
	}

	ctx := context.Background()
	pod, err := clientset.CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ephemeralContainer)

	_, err = clientset.CoreV1().
		Pods(namespace).
		UpdateEphemeralContainers(ctx, podName, pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Ephemeral container added to pod %s\n", podName)
	return nil
}

func RunInNode(configFlags *genericclioptions.ConfigFlags, nodeName string) error {
	// TODO: ephemeral pod in node.
	// TODO: how to get the image name/tag?
	fmt.Printf("TODO: node=%s\n", nodeName)
	return nil
}
