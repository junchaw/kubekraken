package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewListContextsCmd(opts *KrakenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-contexts",
		Short: "List available Kubernetes contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, context := range opts.Contexts {
				fmt.Printf("%s - %s\n", context.Kubeconfig, context.Context)
			}
			return nil
		},
	}

	return cmd
}
