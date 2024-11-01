package cmd

import (
	"fmt"
	"os"

	"github.com/microsoft/retina/pkg/shell"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	configFlags             *genericclioptions.ConfigFlags
	matchVersionFlags       *cmdutil.MatchVersionFlags
	retinaShellImageRepo    string
	retinaShellImageVersion string
	mountHostFilesystem     bool
)

const defaultRetinaShellImageRepo = "mcr.microsoft.com/containernetworking/retina-shell" // TODO: This doesn't exist yet

var shellCmd = &cobra.Command{
	Use:   "shell (NODE | TYPE[[.VERSION].GROUP]/NAME)",
	Short: "Start a shell in a node or pod",
	Long: templates.LongDesc(`
	Start a shell with networking tools in a node or pod for adhoc debugging.

	* For nodes, this creates a pod on the node in the root network namespace.
	* For pods, this creates an ephemeral container inside the pod's network namespace.

	You can override the default image used for the shell container with either
	CLI flags (--retina-shell-image-repo and --retina-shell-image-version) or
	environment variables (RETINA_SHELL_IMAGE_REPO and RETINA_SHELL_IMAGE_VERSION).
	CLI flags take precedence over env vars.
`),

	Example: templates.Examples(`
		# start a shell in a node
		kubectl retina shell node0001

		# start a shell in a node, with debug pod in kube-system namespace
		kubectl retina shell -n kube-system node0001

		# start a shell in a node, mounting the host filesystem to /host
		kubectl retina shell node001 --mount-host-filesystem

		# start a shell in a pod
		kubectl retina shell -n kube-system pod/coredns-d459997b4-7cpzx
`),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, explicitNamespace, err := matchVersionFlags.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return err
		}

		r := resource.NewBuilder(configFlags).
			WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
			FilenameParam(explicitNamespace, &resource.FilenameOptions{}).
			NamespaceParam(namespace).DefaultNamespace().ResourceNames("nodes", args[0]).
			Do()
		if err := r.Err(); err != nil {
			return err
		}

		restConfig, err := matchVersionFlags.ToRESTConfig()
		if err != nil {
			return err
		}

		config := shell.Config{
			RestConfig:          restConfig,
			RetinaShellImage:    fmt.Sprintf("%s:%s", retinaShellImageRepo, retinaShellImageVersion),
			MountHostFilesystem: mountHostFilesystem,
		}

		return r.Visit(func(info *resource.Info, err error) error {
			switch obj := info.Object.(type) {
			case *v1.Node:
				podDebugNamespace := namespace
				nodeName := obj.Name
				return shell.RunInNode(config, nodeName, podDebugNamespace)
			case *v1.Pod:
				return shell.RunInPod(config, obj.Namespace, obj.Name)
			default:
				gvk := obj.GetObjectKind().GroupVersionKind()
				return fmt.Errorf("unsupported resource %s/%s", gvk.GroupVersion(), gvk.Kind)
			}
		})
	},
}

func init() {
	Retina.AddCommand(shellCmd)
	shellCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Avoid printing full usage message if the command exits with an error.
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		// Allow setting image repo and version via environment variables (CLI flags still take precedence).
		if !cmd.Flags().Changed("retina-shell-image-repo") {
			if envRepo := os.Getenv("RETINA_SHELL_IMAGE_REPO"); envRepo != "" {
				retinaShellImageRepo = envRepo
			}
		}
		if !cmd.Flags().Changed("retina-shell-image-version") {
			if envVersion := os.Getenv("RETINA_SHELL_IMAGE_VERSION"); envVersion != "" {
				retinaShellImageVersion = envVersion
			}
		}
	}
	shellCmd.Flags().StringVar(&retinaShellImageRepo, "retina-shell-image-repo", defaultRetinaShellImageRepo, "The container registry repository for the image to use for the shell container")
	shellCmd.Flags().StringVar(&retinaShellImageVersion, "retina-shell-image-version", Version, "The version (tag) of the image to use for the shell container")
	shellCmd.Flags().BoolVar(&mountHostFilesystem, "mount-host-filesystem", false, "Mount the host filesystem to /host and add capability SYS_CHROOT. Applies only to nodes, not pods.")
	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(shellCmd.PersistentFlags())
	matchVersionFlags = cmdutil.NewMatchVersionFlags(configFlags)
	matchVersionFlags.AddFlags(shellCmd.PersistentFlags())
}
