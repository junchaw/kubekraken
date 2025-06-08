package executor

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/junchaw/kubekraken/pkg/utils"
)

func (r *Run) processOneResult(taskItem *Target) *TaskResult {
	var args []string
	args = append(args, "--kubeconfig", taskItem.Kubeconfig)
	args = append(args, "--context", taskItem.Context)
	args = append(args, r.Options.Args...)
	stdoutBytes, stderrBytes, kubectlErr := utils.Exec("kubectl", args...)

	errString := ""
	if kubectlErr != nil {
		errString = kubectlErr.Error()
	}
	stdout := string(stdoutBytes)
	stderr := string(stderrBytes)

	hasStdout := len(stdout) > 0
	hasStderr := len(stderr) > 0
	hasErr := kubectlErr != nil

	needToPrintStdout := hasErr || r.Options.PrintStdout
	if !hasErr { // if there is error, we always print stdout, regardless of output condition
		for _, outputCondition := range r.Options.OutputConditions {
			if outputCondition.Operator == OutputConditionOperatorContains {
				if !strings.Contains(stdout, outputCondition.Value) {
					needToPrintStdout = false
					break
				}
			} else if outputCondition.Operator == OutputConditionOperatorNotContains {
				if strings.Contains(stdout, outputCondition.Value) {
					needToPrintStdout = false
					break
				}
			}
		}
	}

	needToPrintStderr := hasErr || (r.Options.PrintStderr && hasStderr)

	needToPrintErr := hasErr

	return &TaskResult{
		TaskItem: taskItem,

		Err:    errString,
		Stdout: stdout,
		Stderr: stderr,

		HasErr:    hasErr,
		HasStdout: hasStdout,
		HasStderr: hasStderr,

		NeedToPrintErr:      needToPrintErr,
		NeedToPrintStdout:   needToPrintStdout,
		NeedToPrintStderr:   needToPrintStderr,
		NeedToPrintAnything: needToPrintErr || needToPrintStdout || needToPrintStderr,
	}
}

func (r *Run) processOne(taskItem *Target) {
	result := r.processOneResult(taskItem)

	// Lock is used to avoid race condition when writing to stdout/stderr and files
	r.Lock.Lock()
	defer r.Lock.Unlock()

	if result.NeedToPrintAnything {
		fmt.Println()
		fmt.Println()
		fmt.Println(utils.Style.Dim.Render("---"))
		fmt.Println(utils.Style.Text.Render(fmt.Sprintf("TASK START: %s (%d/%d)", taskItem.ID, taskItem.Index, len(r.Options.Targets))))
	}

	if result.NeedToPrintErr {
		fmt.Println(utils.Style.Warning.Render("ERROR:"))
		fmt.Println(utils.Style.Warning.Render(result.Err))
	}

	// if there is an error, print stderr for troubleshooting
	if result.NeedToPrintStderr {
		fmt.Println(utils.Style.Warning.Render("STDERR:"))
		fmt.Println(utils.Style.Warning.Render(strings.TrimSpace(string(result.Stderr))))
	}

	// JSON doesn't support multi documents, need to write after merging all results
	if r.Options.OutputFile != "" && r.Options.OutputFormat != "json" {
		var output string
		if r.Options.OutputFormat == "yaml" || r.Options.OutputFormat == "yml" {
			yamlContent, err := result.ToYAMLInMultiDoc()
			if err != nil {
				r.Logger.Fatalf("failed to marshal result to yaml: %v", err)
			}
			output = string(yamlContent)
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

		if result.NeedToPrintErr {
			if err := utils.PutFileWithFormat(errFile, result.Err, r.Options.OutputFormat, func() string {
				return result.Err
			}); err != nil {
				r.Logger.Fatalf("failed to write err to file: %v", err)
			}
		}

		if result.NeedToPrintStdout {
			if err := utils.PutFileWithFormat(stdoutFile, result.Stdout, r.Options.OutputFormat, func() string {
				return result.Stdout
			}); err != nil {
				r.Logger.Fatalf("failed to write stdout to file: %v", err)
			}
		}

		if result.NeedToPrintStderr {
			if err := utils.PutFileWithFormat(stderrFile, result.Stderr, r.Options.OutputFormat, func() string {
				return result.Stderr
			}); err != nil {
				r.Logger.Fatalf("failed to write stderr to file: %v", err)
			}
		}
	}

	if result.NeedToPrintStdout {
		fmt.Println(utils.Style.Info.Render("STDOUT:"))
		fmt.Println(utils.Style.Info.Render(strings.TrimSpace(string(result.Stdout))))
	}

	if result.NeedToPrintAnything {
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
