package cmd

import (
	"os"
	"regexp"

	"github.com/junchaw/kubekraken/pkg/executor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var opts KrakenOptions

var logger = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Level:     logrus.WarnLevel,
}

func init() {
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logger.Level = logrus.WarnLevel
	} else {
		logger.Level = level
	}
}

type KrakenOptions struct {
	KubeconfigFiles   []string
	KubeconfigFilter  string
	KubeconfigExclude string
	UseCurrentContext bool
	ContextFilter     string
	ContextExclude    string

	Workers int

	OutputDir        string
	OutputFile       string
	OutputFormat     string
	NoStdout         bool
	NoStderr         bool
	OutputConditions string

	// KubeconfigFilterRegex is the regex filter for kubeconfig files, parsed after reading arguments and before running commands
	KubeconfigFilterRegex *regexp.Regexp

	// KubeconfigExcludeRegex is the regex exclude filter for kubeconfig files, parsed after reading arguments and before running commands
	KubeconfigExcludeRegex *regexp.Regexp

	// ContextFilterRegex is the regex filter for context names, parsed after reading arguments and before running commands
	ContextFilterRegex *regexp.Regexp

	// ContextExcludeRegex is the regex exclude filter for context names, parsed after reading arguments and before running commands
	ContextExcludeRegex *regexp.Regexp

	// Targets is a list of contexts, parsed after reading arguments and before running commands
	Targets []executor.Target
}

func NewKrakenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kraken",
		Short: "Run command to multiple clusters",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opts.KubeconfigFilter != "" {
				re, err := regexp.Compile(opts.KubeconfigFilter)
				if err != nil {
					logger.Fatalf("failed to compile kubeconfig filter: %v", err)
				}
				opts.KubeconfigFilterRegex = re
			}

			if opts.KubeconfigExclude != "" {
				re, err := regexp.Compile(opts.KubeconfigExclude)
				if err != nil {
					logger.Fatalf("failed to compile kubeconfig exclude: %v", err)
				}
				opts.KubeconfigExcludeRegex = re
			}

			if opts.ContextFilter != "" {
				re, err := regexp.Compile(opts.ContextFilter)
				if err != nil {
					logger.Fatalf("failed to compile context filter: %v", err)
				}
				opts.ContextFilterRegex = re
			}

			if opts.ContextExclude != "" {
				re, err := regexp.Compile(opts.ContextExclude)
				if err != nil {
					logger.Fatalf("failed to compile context exclude: %v", err)
				}
				opts.ContextExcludeRegex = re
			}

			opts.Targets = []executor.Target{}

			for _, kubeconfigFileOrDir := range opts.KubeconfigFiles {
				// Note that kubeconfig filter does not apply here, but applied to files under the directory
				targets, err := ParseKubeconfigFileOrDir(
					logger,
					kubeconfigFileOrDir,
					opts.KubeconfigFilterRegex,
					opts.KubeconfigExcludeRegex,
					opts.UseCurrentContext,
					opts.ContextFilterRegex,
					opts.ContextExcludeRegex,
				)
				if err != nil {
					logger.Fatalf("failed to parse kubeconfig file or directory %s: %v", kubeconfigFileOrDir, err)
				}
				opts.Targets = append(opts.Targets, targets...)
			}
		},
	}

	// Add flags
	cmd.PersistentFlags().StringSliceVar(&opts.KubeconfigFiles, "kubeconfig-files", []string{os.Getenv("KUBECONFIG")}, "Kubeconfig files, item could be directory or file, in case of directory, all files in the directory will be used, see --kubeconfig-filter")
	cmd.PersistentFlags().StringVar(&opts.KubeconfigFilter, "kubeconfig-filter", "", "Regex filter for kubeconfig files, used with kubeconfig from directory, will not filter items specified in --kubeconfig-files (e.g. prd-.*\\.yaml)")
	cmd.PersistentFlags().StringVar(&opts.KubeconfigExclude, "kubeconfig-exclude", "", "Regex exclude filter for kubeconfig files, used with kubeconfig from directory, will not filter items specified in --kubeconfig-files (e.g. dev-.*\\.yaml)")

	cmd.PersistentFlags().BoolVar(&opts.UseCurrentContext, "use-current-context", false, "Only use the current context from the kubeconfig file, this can be used with --context-filter and --context-exclude")
	cmd.PersistentFlags().StringVar(&opts.ContextFilter, "context-filter", "", "Regex filter for context names (e.g. prd-.*)")
	cmd.PersistentFlags().StringVar(&opts.ContextExclude, "context-exclude", "", "Regex exclude filter for context names (e.g. dev-.*)")

	cmd.PersistentFlags().IntVar(&opts.Workers, "workers", 99, "Number of workers to run concurrently")

	cmd.PersistentFlags().StringVar(&opts.OutputDir, "output-dir", "", "Output directory for the results, kubekraken will save stdout/stderr/error to files under this directory")
	cmd.PersistentFlags().StringVar(&opts.OutputFile, "output-file", "", "Output file for the results, kubekraken will save stdout/stderr/error to this file")
	cmd.PersistentFlags().StringVar(&opts.OutputFormat, "output-format", "text", "Output format for the results (text, json)")
	cmd.PersistentFlags().BoolVar(&opts.NoStdout, "no-stdout", false, "Do not print kubectl stdout")
	cmd.PersistentFlags().BoolVar(&opts.NoStderr, "no-stderr", false, "Do not print kubectl stderr")
	cmd.PersistentFlags().StringVar(&opts.OutputConditions, "output-conditions", "", "Output conditions for the results, see document for more details")

	// Add subcommands
	cmd.AddCommand(NewListContextsCmd(&opts))
	cmd.AddCommand(NewKubectlCmd(&opts))

	return cmd
}
