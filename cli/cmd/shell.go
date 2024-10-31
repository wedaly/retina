package cmd

import (
	"fmt"

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
	configFlags       *genericclioptions.ConfigFlags
	matchVersionFlags *cmdutil.MatchVersionFlags
)

var shellCmd = &cobra.Command{
	Use:   "shell (NODE | TYPE[[.VERSION].GROUP]/NAME)",
	Short: "Start a shell in a node or pod",
	Long:  templates.LongDesc("Start a shell with networking tools in a node or pod for adhoc debugging."),
	Example: templates.Examples(`
		# start a shell in a node
		kubectl retina shell node0001

		# start a shell in a node, with debug pod in kube-system namespace
		kubectl retina shell -n kube-system node0001

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

		return r.Visit(func(info *resource.Info, err error) error {
			switch obj := info.Object.(type) {
			case *v1.Node:
				return shell.RunInNode(restConfig, configFlags, obj.Name)
			case *v1.Pod:
				return shell.RunInPod(restConfig, configFlags, obj.Name)
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
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	}
	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(shellCmd.PersistentFlags())
	matchVersionFlags = cmdutil.NewMatchVersionFlags(configFlags)
	matchVersionFlags.AddFlags(shellCmd.PersistentFlags())
}
