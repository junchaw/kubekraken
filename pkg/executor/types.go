package executor

import (
	"fmt"
	"strings"
	"sync"

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
		ID:         strings.ReplaceAll(fmt.Sprintf("%s@%s", strings.TrimPrefix(kubeconfig, "/"), context), "/", "--"),
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
