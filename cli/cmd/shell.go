package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var namespace string

var shellCmd = &cobra.Command{
	Use:   "shell [target]",
	Short: "Start a shell in a node or pod",
	Long:  "Start a shell with networking tools in a node or pod for adhoc debugging.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		fmt.Printf("Starting shell session with target: %s\n", target)
		if namespace != "" {
			fmt.Printf("Using namespace: %s\n", namespace)
		}
	},
}

func init() {
	Retina.AddCommand(shellCmd)
	shellCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace for the shell session")
}
