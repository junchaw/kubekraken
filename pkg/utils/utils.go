package utils

import (
	"bytes"
	"os/exec"
	"strings"
)

func Exec(name string, arg ...string) ([]byte, []byte, error) {
	var stdout = bytes.Buffer{}
	var stderr = bytes.Buffer{}
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func ExecWithStdin(stdin string, name string, arg ...string) ([]byte, []byte, error) {
	var stdout = bytes.Buffer{}
	var stderr = bytes.Buffer{}
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(stdin)
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}
