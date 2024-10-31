package shell

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/attach"
	"k8s.io/kubectl/pkg/cmd/exec"
)

const (
	imageRepo    = "widalytest.azurecr.io/wedaly/retina/retina-shell"
	imageVersion = "v0.0.16-122-g94ca3aa"
)

func RunInPod(restConfig *rest.Config, configFlags *genericclioptions.ConfigFlags, podName string) error {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	namespace := *configFlags.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// TODO: and open connection with stdin/stdout?
	ephemeralContainer := v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:  fmt.Sprintf("retina-shell-%s", utilrand.String(5)),
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

	// TODO: poll for container ready status

	attachOpts := &attach.AttachOptions{
		Config: restConfig,
		StreamOptions: exec.StreamOptions{
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: ephemeralContainer.Name,
			IOStreams: genericiooptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stderr,
			},
			Stdin: true,
			TTY:   true,
		},
		Attach:      &attach.DefaultRemoteAttach{},
		AttachFunc:  attach.DefaultAttachFunc,
		Pod:         pod,
		CommandName: "bash", // TODO: const
	}

	return attachOpts.Run()
}

func RunInNode(restConfig *rest.Config, configFlags *genericclioptions.ConfigFlags, nodeName string) error {
	// TODO: ephemeral pod in node.
	// TODO: how to get the image name/tag?
	fmt.Printf("TODO: node=%s\n", nodeName)
	return nil
}
