package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/junchaw/kubekraken/pkg/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type RunOptions struct {
	Targets []Target

	Args []string

	Workers int

	OutputDir    string
	OutputFile   string
	OutputFormat string
	NoStdout     bool

	Logger *logrus.Logger
}

type Run struct {
	Options *RunOptions

	Wg sync.WaitGroup

	// Lock is used to avoid race condition when writing to stdout/stderr and files
	Lock sync.Mutex

	NextTarget chan *Target

	Counter int

	Results map[string]TaskResult

	Logger *logrus.Logger
}

func NewRun(opts *RunOptions) *Run {
	return &Run{
		Options:    opts,
		Wg:         sync.WaitGroup{},
		Lock:       sync.Mutex{},
		NextTarget: make(chan *Target),
		Results:    make(map[string]TaskResult),
		Logger:     opts.Logger,
	}
}

func (r *Run) Run() error {
	if r.Options.OutputFile != "" {
		outputDir := path.Dir(r.Options.OutputFile)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		// Create empty file if it doesn't exist
		if _, err := os.Stat(r.Options.OutputFile); os.IsNotExist(err) {
			if _, err := os.Create(r.Options.OutputFile); err != nil {
				return fmt.Errorf("failed to create output file: %v", err)
			}
		} else {
			// Truncate existing file
			if err := os.Truncate(r.Options.OutputFile, 0); err != nil {
				return fmt.Errorf("failed to truncate output file: %v", err)
			}
		}
		r.Logger.Infof("output file: %s", r.Options.OutputFile)
	}

	if r.Options.OutputDir != "" {
		// Empty the directory if it exists
		if err := os.RemoveAll(r.Options.OutputDir); err != nil {
			return fmt.Errorf("failed to remove output directory: %v", err)
		}
		if err := os.MkdirAll(r.Options.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
		r.Logger.Infof("output directory: %s", r.Options.OutputDir)
	}

	stopCh := make(chan struct{})
	for range r.Options.Workers {
		go r.startWorker(stopCh)
	}

	for _, target := range r.Options.Targets {
		r.NextTarget <- &target
	}

	close(stopCh)

	r.Logger.Infof("waiting for workers to exit")
	r.Wg.Wait()

	summary := RunSummary{
		Errors:       []TaskResult{},
		ErrorCount:   0,
		Warnings:     []TaskResult{},
		WarningCount: 0,
		TotalCount:   len(r.Results),
	}
	for _, result := range r.Results {
		if result.Err != "" {
			summary.ErrorCount++
			summary.Errors = append(summary.Errors, result)
		} else if len(result.Stderr) > 0 {
			summary.WarningCount++
			summary.Warnings = append(summary.Warnings, result)
		}
	}

	for _, result := range summary.ToStyledText() {
		fmt.Println(result.Render())
	}
	if r.Options.OutputFile != "" {
		// JSON doesn't support multi documents, need to write after merging all results
		if r.Options.OutputFormat == "json" {
			jsonContent, err := json.MarshalIndent(map[string]any{
				"results": r.Results,
				"summary": summary,
			}, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal summary to json: %v", err)
			}
			if err := os.WriteFile(r.Options.OutputFile, jsonContent, 0600); err != nil {
				return fmt.Errorf("failed to write summary to file: %v", err)
			}
		} else {
			f, err := os.OpenFile(r.Options.OutputFile, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("failed to open output file: %v", err)
			}
			defer f.Close()
			outputContent := ""
			if r.Options.OutputFormat == "yaml" {
				yamlContent, err := yaml.Marshal(summary)
				if err != nil {
					return fmt.Errorf("failed to marshal summary to yaml: %v", err)
				}
				outputContent = string(yamlContent)
			} else {
				outputContent = summary.ToText()
			}
			if _, err := f.Write([]byte("---\n" + outputContent)); err != nil {
				return fmt.Errorf("failed to write summary to file: %v", err)
			}
		}
		fmt.Printf("%s\n", utils.Style.Success.Render(fmt.Sprintf("Results are saved to file %s", r.Options.OutputFile)))
	}

	if r.Options.OutputDir != "" {
		summaryFile := path.Join(r.Options.OutputDir, "summary"+utils.FileExt(r.Options.OutputFormat))
		err := utils.PutFileWithFormat(summaryFile, summary, r.Options.OutputFormat, func() string {
			return summary.ToText()
		})
		if err != nil {
			return fmt.Errorf("failed to save summary to file: %v", err)
		}
		fmt.Printf("%s\n", utils.Style.Success.Render(fmt.Sprintf("Results are saved to directory %s", r.Options.OutputDir)))
	}

	fmt.Printf("%s\n", utils.Style.Dim.Render("---"))

	if summary.ErrorCount > 0 {
		return errors.New("not all clusters were processed successfully")
	}

	return nil
}
