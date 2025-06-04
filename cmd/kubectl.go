package cmd

import (
	"strings"

	"github.com/junchaw/kubekraken/pkg/executor"
	"github.com/spf13/cobra"
)

func NewKubectlCmd(opts *KrakenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubectl",
		Aliases: []string{"k"},
		Short:   "Run kubectl commands",
		Run: func(cmd *cobra.Command, args []string) {
			outputConditions := []executor.OutputCondition{}
			if opts.OutputConditions != "" {
				for condition := range strings.SplitSeq(opts.OutputConditions, ",") {
					parts := strings.SplitN(condition, ":", 2)
					outputConditions = append(outputConditions, executor.OutputCondition{
						Operator: parts[0],
						Value:    parts[1],
					})
				}
			}
			kr := executor.NewRun(&executor.RunOptions{
				Targets:          opts.Targets,
				Args:             args,
				Workers:          opts.Workers,
				OutputDir:        opts.OutputDir,
				OutputFile:       opts.OutputFile,
				OutputFormat:     opts.OutputFormat,
				PrintStdout:      !opts.NoStdout,
				PrintStderr:      !opts.NoStderr,
				OutputConditions: outputConditions,
				Logger:           logger,
			})
			if err := kr.Run(); err != nil {
				logger.Fatalf("failed to run kubectl: %v", err)
			}
		},
	}

	return cmd
}
