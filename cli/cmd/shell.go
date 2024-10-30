package cmd

import (
	"fmt"
	"strings"

	"github.com/microsoft/retina/pkg/shell"
	"github.com/spf13/cobra"
)

var namespace string

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

		targetType, targetName := targetParts[0], targetParts[1]
		if targetType == "pod" || targetType == "pods" {
			return shell.RunInPod(targetName, namespace)
		} else if targetType == "node" || targetType == "nodes" {
			return shell.RunInNode(targetName)
		} else {
			return fmt.Errorf("target type must be either pods or nodes")
		}
	},
}

func init() {
	Retina.AddCommand(shellCmd)
	shellCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace for the shell session")
}
