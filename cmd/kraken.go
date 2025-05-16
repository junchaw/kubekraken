package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var opts KrakenOptions

var logger = logrus.New()

func NewKrakenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kraken",
		Short: "Run command to multiple clusters",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.KubeconfigFilter != "" {
				re, err := regexp.Compile(opts.KubeconfigFilter)
				if err != nil {
					return fmt.Errorf("failed to compile kubeconfig filter: %v", err)
				}
				opts.KubeconfigFilterRegex = re
			}

			if opts.ContextFilter != "" {
				re, err := regexp.Compile(opts.ContextFilter)
				if err != nil {
					return fmt.Errorf("failed to compile context filter: %v", err)
				}
				opts.ContextFilterRegex = re
			}

			opts.Contexts = []Context{}

			for _, kubeconfigFile := range opts.KubeconfigFiles {
				if opts.KubeconfigFilterRegex != nil && !opts.KubeconfigFilterRegex.MatchString(kubeconfigFile) {
					continue
				}

				contexts, err := ParseKubeconfigFileOrDir(kubeconfigFile, opts.KubeconfigFilterRegex, opts.UseCurrentContext, opts.ContextFilterRegex)
				if err != nil {
					return fmt.Errorf("failed to parse kubeconfig file or directory %s: %v", kubeconfigFile, err)
				}
				opts.Contexts = append(opts.Contexts, contexts...)
			}

			return nil
		},
	}

	// Add subcommands
	cmd.AddCommand(NewListContextsCmd(&opts))

	// Add flags
	cmd.PersistentFlags().StringSliceVar(&opts.KubeconfigFiles, "kubeconfig-files", []string{os.Getenv("KUBECONFIG")}, "Kubeconfig files, item could be directory or file, in case of directory, all files in the directory will be used, see --kubeconfig-filter")
	cmd.PersistentFlags().StringVar(&opts.KubeconfigFilter, "kubeconfig-filter", "", "Regex filter for kubeconfig files, used with kubeconfig from directory (e.g. .*\\.yaml)")
	cmd.PersistentFlags().BoolVar(&opts.UseCurrentContext, "use-current-context", false, "Only use the current context from the kubeconfig file, if set, --kubeconfig-filter will be ignored")
	cmd.PersistentFlags().StringVar(&opts.ContextFilter, "context-filter", "", "Regex filter for context names (e.g. prd-.*), see --use-current-context if you want to use the default context")

	return cmd
}
