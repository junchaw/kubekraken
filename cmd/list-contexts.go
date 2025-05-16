package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewListContextsCmd(opts *KrakenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-contexts",
		Short: "List available Kubernetes contexts",
		Run: func(cmd *cobra.Command, args []string) {
			for i, context := range opts.Targets {
				fmt.Printf("[%d/%d] %s@%s\n", i+1, len(opts.Targets), context.Kubeconfig, context.Context)
			}
		},
	}

	return cmd
}
