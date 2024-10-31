package cmd

import (
	"fmt"
	"strings"

	"github.com/microsoft/retina/pkg/shell"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	configFlags       *genericclioptions.ConfigFlags
	matchVersionFlags *cmdutil.MatchVersionFlags
)

var shellCmd = &cobra.Command{
	Use:   "shell [target]",
	Short: "Start a shell in a node or pod",
	Long:  "Start a shell with networking tools in a node or pod for adhoc debugging.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		targetParts := strings.SplitN(target, "/", 2)
		if len(targetParts) != 2 {
			return fmt.Errorf("target must be either pods/<pod> or nodes/<node>")
		}

		restConfig, err := matchVersionFlags.ToRESTConfig()
		if err != nil {
			return err
		}

		targetType, targetName := targetParts[0], targetParts[1]
		if targetType == "pod" || targetType == "pods" {
			return shell.RunInPod(restConfig, configFlags, targetName)
		} else if targetType == "node" || targetType == "nodes" {
			return shell.RunInNode(restConfig, configFlags, targetName)
		} else {
			return fmt.Errorf("target type must be either pods or nodes")
		}
	},
}

func init() {
	// TODO: suppress error output
	Retina.AddCommand(shellCmd)
	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(shellCmd.PersistentFlags())
	matchVersionFlags = cmdutil.NewMatchVersionFlags(configFlags)
	matchVersionFlags.AddFlags(shellCmd.PersistentFlags())
}
