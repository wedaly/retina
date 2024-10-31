package shell

import (
	"context"
	"fmt"
	"os"
	"time"

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
	// TODO: move to config and/or make overridable? or based on CLI retina version
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

	ephemeralContainer := v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:  fmt.Sprintf("retina-shell-%s", utilrand.String(5)),
			Image: fmt.Sprintf("%s:%s", imageRepo, imageVersion),
			Stdin: true,
			TTY:   true,
			SecurityContext: &v1.SecurityContext{
				Capabilities: &v1.Capabilities{
					Drop: []v1.Capability{"ALL"},
					Add:  []v1.Capability{"NET_ADMIN", "NET_RAW"},
				},
			},
		},
		// TODO: what command is it running? how does it stay up?
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

	return attachToShell(restConfig, namespace, podName, ephemeralContainer.Name, pod)
}

func RunInNode(restConfig *rest.Config, configFlags *genericclioptions.ConfigFlags, nodeName string) error {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	namespace := *configFlags.Namespace
	if namespace == "" {
		namespace = "default"
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("retina-shell-%s", utilrand.String(5)),
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			NodeName:      nodeName,
			RestartPolicy: v1.RestartPolicyNever,
			Tolerations:   []v1.Toleration{{Operator: v1.TolerationOpExists}},
			HostNetwork:   true,
			// TODO: HostPID? HostIPC?
			Containers: []v1.Container{
				{
					Name:  "retina-shell",
					Image: fmt.Sprintf("%s:%s", imageRepo, imageVersion),
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

	// TODO: optional host volume

	// Create the pod
	ctx := context.Background()
	_, err = clientset.CoreV1().
		Pods(namespace).
		Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// TODO: wait for contianer running
	time.Sleep(10 * time.Second)

	// TODO: delete on exit...
	return attachToShell(restConfig, namespace, pod.Name, pod.Spec.Containers[0].Name, pod)
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
		CommandName: "bash", // TODO: const
	}

	return attachOpts.Run()
}
