package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewKrakenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kraken",
		Short: "Run command to multiple clusters",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello, World!")
		},
	}

	return cmd
}
