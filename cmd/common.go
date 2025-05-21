package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/junchaw/kubekraken/pkg/executor"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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

func ParseKubeconfigFile(
	logger *logrus.Logger,
	kubeconfigFile string,
	useCurrentContext bool,
	contextFilterRegex *regexp.Regexp,
	contextExcludeRegex *regexp.Regexp,
) (targets []executor.Target, err error) {
	logger.Infof("Parsing kubeconfig file %s", kubeconfigFile)

	data, err := os.ReadFile(kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig file %s: %v", kubeconfigFile, err)
	}

	var config KubeconfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig file %s: %v", kubeconfigFile, err)
	}

	for _, ctx := range config.Contexts {
		// If useCurrentContext is true, only include the current context,
		// note that filter and exclude are still applied.
		if useCurrentContext && ctx.Name != config.CurrentContext {
			logger.Debugf("Skipping non-current context %s in kubeconfig file %s", ctx.Name, kubeconfigFile)
			continue
		}

		if contextFilterRegex != nil && !contextFilterRegex.MatchString(ctx.Name) {
			logger.Debugf("Skipping context %s in kubeconfig file %s", ctx.Name, kubeconfigFile)
			continue
		}

		if contextExcludeRegex != nil && contextExcludeRegex.MatchString(ctx.Name) {
			logger.Debugf("Excluded context %s in kubeconfig file %s", ctx.Name, kubeconfigFile)
			continue
		}

		logger.Infof("Found context matching filter in kubeconfig file %s: %s", kubeconfigFile, ctx.Name)
		targets = append(targets, executor.NewTarget(kubeconfigFile, ctx.Name))
	}

	return targets, nil
}

func ParseKubeconfigFileOrDir(
	logger *logrus.Logger,
	kubeconfigFileOrDir string,
	kubeconfigFilterRegex *regexp.Regexp,
	kubeconfigExcludeRegex *regexp.Regexp,
	useCurrentContext bool,
	contextFilterRegex *regexp.Regexp,
	contextExcludeRegex *regexp.Regexp,
) (targets []executor.Target, err error) {
	logger.Infof("Parsing kubeconfig file or directory %s", kubeconfigFileOrDir)

	if strings.HasPrefix(kubeconfigFileOrDir, "~") {
		kubeconfigFileOrDir = strings.Replace(kubeconfigFileOrDir, "~", os.Getenv("HOME"), 1)
	}

	if _, err := os.Stat(kubeconfigFileOrDir); os.IsNotExist(err) {
		logger.Warnf("kubeconfig file %s does not exist", kubeconfigFileOrDir)
		return []executor.Target{}, nil
	}

	info, err := os.Stat(kubeconfigFileOrDir)
	if err != nil {
		logger.Warnf("failed to stat kubeconfig file %s: %v", kubeconfigFileOrDir, err)
		return []executor.Target{}, nil
	}

	if info.IsDir() {
		logger.Infof("Parsing kubeconfig directory %s", kubeconfigFileOrDir)

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
			if kubeconfigExcludeRegex != nil && kubeconfigExcludeRegex.MatchString(filepath.Join(kubeconfigFileOrDir, file.Name())) {
				continue
			}

			targetsInFile, err := ParseKubeconfigFile(
				logger,
				filepath.Join(kubeconfigFileOrDir, file.Name()),
				useCurrentContext,
				contextFilterRegex,
				contextExcludeRegex,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kubeconfig file %s: %v", filepath.Join(kubeconfigFileOrDir, file.Name()), err)
			}
			targets = append(targets, targetsInFile...)
		}
		return targets, nil
	}

	// is file
	contextsInFile, err := ParseKubeconfigFile(
		logger,
		kubeconfigFileOrDir,
		useCurrentContext,
		contextFilterRegex,
		contextExcludeRegex,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig file %s: %v", kubeconfigFileOrDir, err)
	}
	return contextsInFile, nil
}
