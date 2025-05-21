package executor

import (
	"fmt"
	"strings"
)

type TaskResult struct {
	TaskItem *Target `json:"taskItem" yaml:"taskItem"`

	Err    string `json:"err" yaml:"err"`
	Stdout string `json:"stdout" yaml:"stdout"`
	Stderr string `json:"stderr" yaml:"stderr"`
}

func (r *TaskResult) ToText(totalCount int) string {
	output := fmt.Sprintf("\n---\nTASK START: %s (%d/%d)\n", r.TaskItem.ID, r.TaskItem.Index, totalCount)

	if r.Err != "" {
		output += fmt.Sprintf("\nERROR: %v\n", r.Err)
	}

	if len(r.Stderr) > 0 {
		output += fmt.Sprintf("\nSTDERR:\n%s\n", strings.TrimSpace(string(r.Stderr)))
	}

	if len(r.Stdout) > 0 {
		output += fmt.Sprintf("\nSTDOUT:\n%s\n", strings.TrimSpace(string(r.Stdout)))
	}

	output += fmt.Sprintf("\nTASK END: %s (%d/%d)\n", r.TaskItem.ID, r.TaskItem.Index, totalCount)

	return output
}
