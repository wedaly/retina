package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var namespace string

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell [target]",
	Short: "Start a shell session",
	Long:  `Start a shell session with the specified target.`,
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

	// Define the namespace flag
	shellCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace for the shell session")
}
