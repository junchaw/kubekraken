package executor

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type TaskResult struct {
	TaskItem *Target `json:"taskItem" yaml:"taskItem"`

	Err    string `json:"err,omitempty" yaml:"err,omitempty"`
	Stdout string `json:"stdout,omitempty" yaml:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty" yaml:"stderr,omitempty"`

	HasErr    bool `json:"hasErr,omitempty" yaml:"hasErr,omitempty"`
	HasStdout bool `json:"hasStdout,omitempty" yaml:"hasStdout,omitempty"`
	HasStderr bool `json:"hasStderr,omitempty" yaml:"hasStderr,omitempty"`

	NeedToPrintErr      bool `json:"needToPrintErr,omitempty" yaml:"needToPrintErr,omitempty"`
	NeedToPrintStdout   bool `json:"needToPrintStdout,omitempty" yaml:"needToPrintStdout,omitempty"`
	NeedToPrintStderr   bool `json:"needToPrintStderr,omitempty" yaml:"needToPrintStderr,omitempty"`
	NeedToPrintAnything bool `json:"needToPrintAnything,omitempty" yaml:"needToPrintAnything,omitempty"`
}

func (r *TaskResult) ToJSON() ([]byte, error) {
	if !r.NeedToPrintAnything {
		return []byte(""), nil
	}

	rCopy := *r

	if !rCopy.NeedToPrintErr {
		rCopy.Err = ""
	}
	if !rCopy.NeedToPrintStdout {
		rCopy.Stdout = ""
	}
	if !rCopy.NeedToPrintStderr {
		rCopy.Stderr = ""
	}

	jsonContent, err := json.Marshal(rCopy)
	if err != nil {
		return nil, err
	}
	return jsonContent, nil
}

func (r *TaskResult) ToYAML() ([]byte, error) {
	if !r.NeedToPrintAnything {
		return []byte(""), nil
	}

	rCopy := *r

	if !rCopy.NeedToPrintErr {
		rCopy.Err = ""
	}
	if !rCopy.NeedToPrintStdout {
		rCopy.Stdout = ""
	}
	if !rCopy.NeedToPrintStderr {
		rCopy.Stderr = ""
	}

	yamlContent, err := yaml.Marshal(rCopy)
	if err != nil {
		return nil, err
	}
	return yamlContent, nil
}

func (r *TaskResult) ToYAMLInMultiDoc() ([]byte, error) {
	yamlContent, err := r.ToYAML()
	if err != nil {
		return nil, err
	}

	if len(yamlContent) == 0 {
		return []byte(""), nil
	}

	return fmt.Appendf(nil, "---\n%s", string(yamlContent)), nil
}

func (r *TaskResult) ToText(totalCount int) string {
	output := ""

	if r.NeedToPrintAnything {
		output += fmt.Sprintf("\n---\nTASK START: %s (%d/%d)\n", r.TaskItem.ID, r.TaskItem.Index, totalCount)
	}

	if r.NeedToPrintErr {
		output += fmt.Sprintf("\nERROR: %v\n", r.Err)
	}

	if r.NeedToPrintStderr {
		output += fmt.Sprintf("\nSTDERR:\n%s\n", strings.TrimSpace(string(r.Stderr)))
	}

	if r.NeedToPrintStdout {
		output += fmt.Sprintf("\nSTDOUT:\n%s\n", strings.TrimSpace(string(r.Stdout)))
	}

	if r.NeedToPrintAnything {
		output += fmt.Sprintf("\nTASK END: %s (%d/%d)\n", r.TaskItem.ID, r.TaskItem.Index, totalCount)
	}

	return output
}
