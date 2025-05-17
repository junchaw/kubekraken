package executor

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/junchaw/kubekraken/pkg/utils"
	"gopkg.in/yaml.v2"
)

func (r *Run) processOneResult(taskItem *Target) *TaskResult {
	var args []string
	args = append(args, "--kubeconfig", taskItem.Kubeconfig)
	args = append(args, "--context", taskItem.Context)
	args = append(args, r.Options.Args...)
	stdout, stderr, err := utils.Exec("kubectl", args...)
	if err != nil {
		return &TaskResult{
			TaskItem: taskItem,
			Err:      err.Error(),
			Stdout:   string(stdout),
			Stderr:   string(stderr),
		}
	}

	return &TaskResult{
		TaskItem: taskItem,
		Err:      "",
		Stdout:   string(stdout),
		Stderr:   string(stderr),
	}
}

func (r *Run) processOne(taskItem *Target) {
	result := r.processOneResult(taskItem)

	// Lock is used to avoid race condition when writing to stdout/stderr and files
	r.Lock.Lock()
	defer r.Lock.Unlock()

	if result.HasErrorOrWarning() || !r.Options.NoStdout {
		fmt.Println()
		fmt.Println()
		fmt.Println(utils.Style.Dim.Render("---"))
		fmt.Println(utils.Style.Text.Render(fmt.Sprintf("TASK START: %s (%d/%d)", taskItem.ID, taskItem.Index, len(r.Options.Targets))))

		if result.Err != "" {
			fmt.Println(utils.Style.Warning.Render("ERROR:"))
			fmt.Println(utils.Style.Warning.Render(result.Err))
		}

		if len(result.Stderr) > 0 {
			fmt.Println(utils.Style.Warning.Render("STDERR:"))
			fmt.Println(utils.Style.Warning.Render(strings.TrimSpace(string(result.Stderr))))
		}
	}

	// JSON doesn't support multi documents, need to write after merging all results
	if r.Options.OutputFile != "" && r.Options.OutputFormat != "json" {
		var output string
		if r.Options.OutputFormat == "yaml" || r.Options.OutputFormat == "yml" {
			yamlContent, err := yaml.Marshal(result)
			if err != nil {
				r.Logger.Fatalf("failed to marshal result to yaml: %v", err)
			}
			output = "---\n" + string(yamlContent) + "\n"
		} else {
			output = result.ToText(len(r.Options.Targets))
		}

		f, err := os.OpenFile(r.Options.OutputFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			r.Logger.Fatalf("failed to open output file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write([]byte(output)); err != nil {
			r.Logger.Fatalf("failed to append result to file: %v", err)
		}
	}

	if r.Options.OutputDir != "" {
		ext := utils.FileExt(r.Options.OutputFormat)
		errFile := path.Join(r.Options.OutputDir, result.TaskItem.ID+".err"+ext)
		stdoutFile := path.Join(r.Options.OutputDir, result.TaskItem.ID+".stdout"+ext)
		stderrFile := path.Join(r.Options.OutputDir, result.TaskItem.ID+".stderr"+ext)

		if result.Err != "" {
			if err := utils.PutFileWithFormat(errFile, result.Err, r.Options.OutputFormat, func() string {
				return result.Err
			}); err != nil {
				r.Logger.Fatalf("failed to write err to file: %v", err)
			}
		}

		if len(result.Stdout) > 0 {
			if err := utils.PutFileWithFormat(stdoutFile, result.Stdout, r.Options.OutputFormat, func() string {
				return result.Stdout
			}); err != nil {
				r.Logger.Fatalf("failed to write stdout to file: %v", err)
			}
		}

		if len(result.Stderr) > 0 {
			if err := utils.PutFileWithFormat(stderrFile, result.Stderr, r.Options.OutputFormat, func() string {
				return result.Stderr
			}); err != nil {
				r.Logger.Fatalf("failed to write stderr to file: %v", err)
			}
		}
	}

	if !r.Options.NoStdout {
		if len(result.Stdout) > 0 {
			fmt.Println(utils.Style.Info.Render("STDOUT:"))
			fmt.Println(utils.Style.Info.Render(strings.TrimSpace(string(result.Stdout))))
		}
	}

	if result.HasErrorOrWarning() || !r.Options.NoStdout {
		fmt.Println(utils.Style.Text.Render(fmt.Sprintf("TASK END: %s (%d/%d)", taskItem.ID, taskItem.Index, len(r.Options.Targets))))
		fmt.Println(utils.Style.Dim.Render("---"))
	}

	r.Results[taskItem.ID] = *result
}

func (r *Run) startWorker(stopCh <-chan struct{}) {
	r.Wg.Add(1)
	defer r.Wg.Done()

	for {
		select {
		case <-stopCh:
			return
		case taskItem := <-r.NextTarget:
			r.Lock.Lock()
			r.Counter++
			r.Lock.Unlock()

			taskItem.Index = r.Counter
			r.processOne(taskItem)
		}
	}
}
