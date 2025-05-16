package executor

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/junchaw/kubekraken/pkg/utils"
	"github.com/sirupsen/logrus"
)

type RunResult struct {
	TaskItem *RunTarget

	Err    error
	Stdout []byte
	Stderr []byte
}

func (r *RunResult) HasErrorOrWarning() bool {
	return r.Err != nil || len(r.Stderr) > 0
}

type RunTarget struct {
	ID         string
	Kubeconfig string
	Context    string

	// Index is the index of the target during execution, will be set during execution
	Index int
}

func NewTarget(kubeconfig, context string) RunTarget {
	return RunTarget{
		ID:         fmt.Sprintf("%s@%s", kubeconfig, context),
		Kubeconfig: kubeconfig,
		Context:    context,
	}
}

type RunOptions struct {
	Targets []RunTarget

	Args []string

	Workers int

	OutputDir  string
	OutputFile string
	NoStdout   bool

	Logger *logrus.Logger
}

type Run struct {
	Options *RunOptions

	Wg sync.WaitGroup

	// Lock is used to avoid race condition when writing to stdout/stderr and files
	Lock sync.Mutex

	NextTarget chan *RunTarget

	Counter int

	Results map[string]RunResult

	Logger *logrus.Logger
}

func NewRun(opts *RunOptions) *Run {
	return &Run{
		Options:    opts,
		Wg:         sync.WaitGroup{},
		Lock:       sync.Mutex{},
		NextTarget: make(chan *RunTarget),
		Results:    make(map[string]RunResult),
		Logger:     opts.Logger,
	}
}

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

	writeToFile := (r.Options.OutputFile != "" || r.Options.OutputDir != "")
	writeToStdout := !writeToFile && !r.Options.NoStdout

	if (writeToFile && result.HasErrorOrWarning()) || writeToStdout {
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
	} else if r.Options.OutputDir != "" {
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
	} else {
		// no output target specified, print to stdout
		if !r.Options.NoStdout && len(result.Stdout) > 0 {
			fmt.Println(utils.Style.Info.Render("STDOUT:"))
			fmt.Println(utils.Style.Info.Render(strings.TrimSpace(string(result.Stdout))))
		}
	}

	if (writeToFile && result.HasErrorOrWarning()) || writeToStdout {
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
			r.Logger.Fatalf("failed to create output directory: %v", err)
		}

		// Create empty file if it doesn't exist
		if _, err := os.Stat(r.Options.OutputFile); os.IsNotExist(err) {
			if _, err := os.Create(r.Options.OutputFile); err != nil {
				r.Logger.Fatalf("failed to create output file: %v", err)
			}
		} else {
			// Truncate existing file
			if err := os.Truncate(r.Options.OutputFile, 0); err != nil {
				r.Logger.Fatalf("failed to truncate output file: %v", err)
			}
		}
		r.Logger.Infof("output file: %s", r.Options.OutputFile)
	}

	if r.Options.OutputDir != "" {
		if err := os.MkdirAll(r.Options.OutputDir, 0755); err != nil {
			r.Logger.Fatalf("failed to create output directory: %v", err)
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
	fmt.Println()
	fmt.Println()
	fmt.Println(utils.Style.Dim.Render("---"))
	fmt.Println(utils.Style.Text.Render("SUMMARY:"))
	for _, result := range r.Results {
		if result.Err != nil {
			errCount++
			stderrStub := ""
			if len(result.Stderr) > 0 {
				stderrStub = ", stderr:"
			}
			fmt.Println(utils.Style.Warning.Render(fmt.Sprintf("- %s: error: %v%s", result.TaskItem.ID, result.Err.Error(), stderrStub)))
			if len(result.Stderr) > 0 {
				fmt.Println(utils.Style.Warning.Render(strings.TrimSpace(string(result.Stderr))))
			}
			fmt.Println()
		} else if len(result.Stderr) > 0 {
			warnCount++
			fmt.Println(utils.Style.Warning.Render(fmt.Sprintf("- %s: stderr:", result.TaskItem.ID)))
			fmt.Println(utils.Style.Warning.Render(strings.TrimSpace(string(result.Stderr))))
			fmt.Println()
		}
	}

	summaryText := fmt.Sprintf("%d successful (%d with warnings), %d error, %d total",
		len(r.Results)-errCount, warnCount, errCount, len(r.Results))
	fmt.Printf("\n%s\n", utils.Style.Text.Render(summaryText))

	if r.Options.OutputDir != "" {
		fmt.Printf("%s\n", utils.Style.Success.Render(fmt.Sprintf("all results are saved to %s", r.Options.OutputDir)))
	}

	if r.Options.OutputFile != "" {
		fmt.Printf("%s\n", utils.Style.Success.Render(fmt.Sprintf("all results are saved to %s", r.Options.OutputFile)))
	}

	fmt.Printf("%s\n", utils.Style.Dim.Render("---"))

	if errCount > 0 {
		return errors.New("not all clusters were processed successfully")
	}

	return nil
}
