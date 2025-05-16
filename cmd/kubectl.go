package cmd

import (
	"github.com/junchaw/kubekraken/pkg/executor"
	"github.com/spf13/cobra"
)

func NewKubectlCmd(opts *KrakenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubectl",
		Aliases: []string{"k"},
		Short:   "Run kubectl commands",
		Run: func(cmd *cobra.Command, args []string) {
			kr := executor.NewRun(&executor.RunOptions{
				Targets:    opts.Targets,
				Args:       args,
				Workers:    opts.Workers,
				OutputDir:  opts.OutputDir,
				OutputFile: opts.OutputFile,
				NoStdout:   opts.NoStdout,
				Logger:     logger,
			})
			if err := kr.Run(); err != nil {
				logger.Fatalf("failed to run kubectl: %v", err)
			}
		},
	}

	return cmd
}
