package executor

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/junchaw/kubekraken/pkg/utils"
)

func (r *Run) processOneResult(taskItem *RunTarget) *RunResult {
	var args []string
	args = append(args, "--kubeconfig", taskItem.Kubeconfig)
	args = append(args, "--context", taskItem.Context)
	args = append(args, r.Options.Args...)
	stdout, stderr, err := utils.Exec("kubectl", args...)
	if err != nil {
		return &RunResult{
			TaskItem: taskItem,
			Err:      err,
			Stdout:   stdout,
			Stderr:   stderr,
		}
	}

	return &RunResult{
		TaskItem: taskItem,
		Err:      nil,
		Stdout:   stdout,
		Stderr:   stderr,
	}
}

func (r *Run) processOne(taskItem *RunTarget) {
	result := r.processOneResult(taskItem)

	// Lock is used to avoid race condition when writing to stdout/stderr and files
	r.Lock.Lock()
	defer r.Lock.Unlock()

	if result.HasErrorOrWarning() || !r.Options.NoStdout {
		fmt.Println()
		fmt.Println()
		fmt.Println(utils.Style.Dim.Render("---"))
		fmt.Println(utils.Style.Text.Render(fmt.Sprintf("TASK START: %s (%d/%d)", taskItem.ID, taskItem.Index, len(r.Options.Targets))))

		if result.Err != nil {
			fmt.Println(utils.Style.Warning.Render("ERROR:"))
			fmt.Println(utils.Style.Warning.Render(result.Err.Error()))
		}

		if len(result.Stderr) > 0 {
			fmt.Println(utils.Style.Warning.Render("STDERR:"))
			fmt.Println(utils.Style.Warning.Render(strings.TrimSpace(string(result.Stderr))))
		}
	}

	if r.Options.OutputFile != "" {
		output := fmt.Appendf(nil, "\n---\nTASK START: %s (%d/%d)\n", taskItem.ID, taskItem.Index, len(r.Options.Targets))

		if result.Err != nil {
			output = append(output, fmt.Appendf(nil, "\nERROR: %v\n", result.Err)...)
		}

		if len(result.Stderr) > 0 {
			output = fmt.Appendf(output, "\nSTDERR:\n%s\n", strings.TrimSpace(string(result.Stderr)))
		}

		if len(result.Stdout) > 0 {
			output = fmt.Appendf(output, "\nSTDOUT:\n%s\n", strings.TrimSpace(string(result.Stdout)))
		}

		output = fmt.Appendf(output, "\nTASK END: %s (%d/%d)\n---\n", taskItem.ID, taskItem.Index, len(r.Options.Targets))

		f, err := os.OpenFile(r.Options.OutputFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			r.Logger.Fatalf("failed to open output file: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(output); err != nil {
			r.Logger.Fatalf("failed to append result to file: %v", err)
		}
	}
	if r.Options.OutputDir != "" {
		errFile := path.Join(r.Options.OutputDir, result.TaskItem.ID+".err.txt")
		stdoutFile := path.Join(r.Options.OutputDir, result.TaskItem.ID+".stdout.txt")
		stderrFile := path.Join(r.Options.OutputDir, result.TaskItem.ID+".stderr.txt")

		if result.Err != nil {
			if err := os.WriteFile(errFile, []byte(result.Err.Error()), 0600); err != nil {
				r.Logger.Fatalf("failed to write err to file: %v", err)
			}
		}

		if err := os.WriteFile(stdoutFile, result.Stdout, 0600); err != nil {
			r.Logger.Fatalf("failed to write stdout to file: %v", err)
		}

		if err := os.WriteFile(stderrFile, result.Stderr, 0600); err != nil {
			r.Logger.Fatalf("failed to write stderr to file: %v", err)
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
			taskItem.Index = r.Counter
			r.Lock.Unlock()
			r.processOne(taskItem)
		}
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

	errCount := 0
	warnCount := 0
	summaryLines := []utils.StyleText{
		{Text: ""},
		{Text: ""},
		{Text: "---", Style: &utils.Style.Dim},
		{Text: "SUMMARY:", Style: &utils.Style.Text},
	}
	for _, result := range r.Results {
		if result.Err != nil {
			errCount++
			stderrStub := ""
			if len(result.Stderr) > 0 {
				stderrStub = ", stderr:"
			}
			summaryLines = append(summaryLines, utils.StyleText{
				Text:  fmt.Sprintf("- %s: error: %v%s", result.TaskItem.ID, result.Err.Error(), stderrStub),
				Style: &utils.Style.Warning,
			})
			if len(result.Stderr) > 0 {
				summaryLines = append(summaryLines,
					utils.StyleText{
						Text:  strings.TrimSpace(string(result.Stderr)),
						Style: &utils.Style.Warning,
					})
			}
			summaryLines = append(summaryLines, utils.StyleText{Text: ""})
		} else if len(result.Stderr) > 0 {
			warnCount++
			summaryLines = append(summaryLines,
				utils.StyleText{
					Text:  fmt.Sprintf("- %s: stderr:", result.TaskItem.ID),
					Style: &utils.Style.Warning,
				},
				utils.StyleText{
					Text:  strings.TrimSpace(string(result.Stderr)),
					Style: &utils.Style.Warning,
				},
				utils.StyleText{Text: ""})
		}
	}

	summaryLines = append(summaryLines, utils.StyleText{
		Text: fmt.Sprintf("%d successful (%d with warnings), %d error, %d total",
			len(r.Results)-errCount, warnCount, errCount, len(r.Results)),
		Style: &utils.Style.Text,
	})

	for _, line := range summaryLines {
		fmt.Println(line.Render())
	}
	if r.Options.OutputFile != "" {
		f, err := os.OpenFile(r.Options.OutputFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open output file: %v", err)
		}
		defer f.Close()
		lines := make([]string, len(summaryLines))
		for i, line := range summaryLines {
			lines[i] = line.GetText()
		}
		if _, err := f.Write([]byte(strings.Join(lines, "\n"))); err != nil {
			return fmt.Errorf("failed to write summary to output file: %v", err)
		}
		fmt.Printf("%s\n", utils.Style.Success.Render(fmt.Sprintf("Results are saved to file %s", r.Options.OutputFile)))
	}

	if r.Options.OutputDir != "" {
		f, err := os.OpenFile(path.Join(r.Options.OutputDir, "summary.txt"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open summary file: %v", err)
		}
		defer f.Close()
		lines := make([]string, len(summaryLines))
		for i, line := range summaryLines {
			lines[i] = line.GetText()
		}
		if _, err := f.Write([]byte(strings.Join(lines, "\n"))); err != nil {
			return fmt.Errorf("failed to write summary to file: %v", err)
		}
		fmt.Printf("%s\n", utils.Style.Success.Render(fmt.Sprintf("Results are saved to directory %s", r.Options.OutputDir)))
	}

	fmt.Printf("%s\n", utils.Style.Dim.Render("---"))

	if errCount > 0 {
		return errors.New("not all clusters were processed successfully")
	}

	return nil
}
