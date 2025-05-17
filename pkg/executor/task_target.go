package executor

import (
	"fmt"
	"strings"
)

type Target struct {
	ID         string `json:"id" yaml:"id"`
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"`
	Context    string `json:"context" yaml:"context"`

	// Index is the index of the target during execution, will be set during execution
	Index int `json:"-" yaml:"-"`
}

func NewTarget(kubeconfig, context string) Target {
	return Target{
		ID:         strings.ReplaceAll(fmt.Sprintf("%s@%s", strings.TrimPrefix(kubeconfig, "/"), context), "/", "--"),
		Kubeconfig: kubeconfig,
		Context:    context,
	}
}
