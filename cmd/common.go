package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Context struct {
	Kubeconfig string
	Context    string
}

// KubeconfigFile represents the structure of a kubeconfig file
type KubeconfigFile struct {
	CurrentContext string `yaml:"current-context"`
	Contexts       []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster   string `yaml:"cluster"`
			User      string `yaml:"user"`
			Namespace string `yaml:"namespace,omitempty"`
		} `yaml:"context"`
	} `yaml:"contexts"`
}

type KrakenOptions struct {
	KubeconfigFiles   []string
	KubeconfigFilter  string
	UseCurrentContext bool
	ContextFilter     string

	// KubeconfigFilterRegex is the regex filter for kubeconfig files, parsed after reading arguments and before running commands
	KubeconfigFilterRegex *regexp.Regexp

	// ContextFilterRegex is the regex filter for context names, parsed after reading arguments and before running commands
	ContextFilterRegex *regexp.Regexp

	// Contexts is a list of contexts, parsed after reading arguments and before running commands
	Contexts []Context
}

func ParseKubeconfigFile(kubeconfigFile string, useCurrentContext bool, contextFilterRegex *regexp.Regexp) (contexts []Context, err error) {
	data, err := os.ReadFile(kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig file %s: %v", kubeconfigFile, err)
	}

	var config KubeconfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig file %s: %v", kubeconfigFile, err)
	}

	for _, ctx := range config.Contexts {
		// If useCurrentContext is true, only include the current context
		if useCurrentContext {
			if ctx.Name == config.CurrentContext {
				contexts = append(contexts, Context{
					Kubeconfig: kubeconfigFile,
					Context:    ctx.Name,
				})
			}
		} else {
			// if useCurrentContext is false

			// If contextFilterRegex is not provided, include all contexts
			if contextFilterRegex == nil || contextFilterRegex.MatchString(ctx.Name) {
				contexts = append(contexts, Context{
					Kubeconfig: kubeconfigFile,
					Context:    ctx.Name,
				})
			}
		}
	}

	return contexts, nil
}

func ParseKubeconfigFileOrDir(kubeconfigFileOrDir string, kubeconfigFilterRegex *regexp.Regexp, useCurrentContext bool, contextFilterRegex *regexp.Regexp) (contexts []Context, err error) {
	if strings.HasPrefix(kubeconfigFileOrDir, "~") {
		kubeconfigFileOrDir = strings.Replace(kubeconfigFileOrDir, "~", os.Getenv("HOME"), 1)
	}

	if _, err := os.Stat(kubeconfigFileOrDir); os.IsNotExist(err) {
		logger.Warnf("kubeconfig file %s does not exist", kubeconfigFileOrDir)
		return []Context{}, nil
	}

	info, err := os.Stat(kubeconfigFileOrDir)
	if err != nil {
		logger.Warnf("failed to stat kubeconfig file %s: %v", kubeconfigFileOrDir, err)
		return []Context{}, nil
	}

	if info.IsDir() {
		files, err := os.ReadDir(kubeconfigFileOrDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %v", kubeconfigFileOrDir, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue // skip sub directories
			}
			if kubeconfigFilterRegex != nil && !kubeconfigFilterRegex.MatchString(filepath.Join(kubeconfigFileOrDir, file.Name())) {
				continue
			}

			contextsInFile, err := ParseKubeconfigFile(filepath.Join(kubeconfigFileOrDir, file.Name()), useCurrentContext, contextFilterRegex)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kubeconfig file %s: %v", filepath.Join(kubeconfigFileOrDir, file.Name()), err)
			}
			contexts = append(contexts, contextsInFile...)
		}
		return contexts, nil
	}

	// is file
	contextsInFile, err := ParseKubeconfigFile(kubeconfigFileOrDir, useCurrentContext, contextFilterRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig file %s: %v", kubeconfigFileOrDir, err)
	}
	return contextsInFile, nil
}
