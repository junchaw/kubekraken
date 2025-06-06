# KubeKraken

<div align="center">
  <img src="docs/logo.png" alt="kubekraken logo" width="200">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/junchaw/kubekraken)](https://goreportcard.com/report/github.com/junchaw/kubekraken)
[![License](https://img.shields.io/github/license/junchaw/kubekraken?color=blue)](https://github.com/junchaw/kubekraken/blob/main/LICENSE)
[![Releases](https://img.shields.io/github/v/release/junchaw/kubekraken)](https://github.com/junchaw/kubekraken/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/junchaw/kubekraken.svg)](https://hub.docker.com/r/junchaw/kubekraken/)

Kubekraken is a powerful CLI tool that unleashes multiple kubectl commands in parallel—tame your clusters with ease.

```shell
kubekraken k -- get nodes
```

<div align="center">
  <img src="docs/screenshot.png" alt="kubekraken screenshot" width="100%">
</div>

## Installation

#### # With Homebrew

```shell
brew tap junchaw/awesome
brew install kubekraken
kubekraken -h
```

#### # With Docker

```shell
docker run junchaw/kubekraken -h
```

#### # Download from release page

First, download tar file from the [release page](https://github.com/junchaw/kubekraken/releases).

After downloading the tar file, extract it, then put `kubekraken` in your `PATH`.

#### # Build from source

```shell
git clone https://github.com/junchaw/kubekraken.git
cd kubekraken && make build
./bin/kubekraken -h
```

## Usage

```shell
kubekraken -h

# Run kubectl commands in parallel, by default will use KUBECONFIG environment variable, and all context in the kubeconfig file
kubekraken kubectl version

# We have a alias for kubectl, and you need to put "--" before the kubectl command if the args contains any flags,
# so that kubekraken can distinguish the kubectl args from the kubekraken args.
kubekraken k -- get pods -n kube-system -l k8s-app=kube-proxy

# You can use --kubeconfig-files to specify the kubeconfig files to use, it could be files or directories,
# if there is any directory, kubekraken will find all the kubeconfig files in the directory,
# --kubeconfig-filter can be used with directory to filter the kubeconfig files, but it will not filter files specified in --kubeconfig-files.
kubekraken --kubeconfig-files ~/.kube --kubeconfig-filter ".*.yaml$" -- rollout restart -n kube-system deployment/coredns

# You can use --context-filter to filter the contexts to run the commands on, it will filter the contexts in the kubeconfig files,
# or use --use-current-context to use the current context in the kubeconfig file, in this case --context-filter will be ignored.
kubekraken --kubeconfig-files ./kubeconfigs --context-filter "(us-west-1|us-west-2)" -- get nodes us-west-1-node-abc
kubekraken --kubeconfig-files ./kubeconfigs --use-current-context -- get nodes us-west-2-node-abc

# You can use --output-file to save the output to a file, it will save the output to the file,
# or use --output-dir to save the output to a directory, each context will have separate output files.
kubekraken --kubeconfig-files ./kubeconfigs --output-file ./tmp/output.txt -- get nodes us-west-2-node-abc
kubekraken --kubeconfig-files ./kubeconfigs --output-dir ./tmp/output -- get nodes us-west-2-node-abc
```

Other flags:

```shell
Run command to multiple clusters

Usage:
  kraken [command]

Available Commands:
  completion    Generate the autocompletion script for the specified shell
  help          Help about any command
  kubectl       Run kubectl commands
  list-contexts List available Kubernetes contexts

Flags:
      --context-exclude string      Regex exclude filter for context names (e.g. dev-.*)
      --context-filter string       Regex filter for context names (e.g. prd-.*)
  -h, --help                        help for kraken
      --kubeconfig-exclude string   Regex exclude filter for kubeconfig files, used with kubeconfig from directory, will not filter items specified in --kubeconfig-files (e.g. dev-.*\.yaml)
      --kubeconfig-files strings    Kubeconfig files, item could be directory or file, in case of directory, all files in the directory will be used, see --kubeconfig-filter (default [/Users/junchawu/.kube/config])
      --kubeconfig-filter string    Regex filter for kubeconfig files, used with kubeconfig from directory, will not filter items specified in --kubeconfig-files (e.g. prd-.*\.yaml)
      --no-stderr                   Do not print kubectl stderr
      --no-stdout                   Do not print kubectl stdout
      --output-conditions string    Output condition for the results, see document for more details
      --output-dir string           Output directory for the results, kubekraken will save stdout/stderr/error to files under this directory
      --output-file string          Output file for the results, kubekraken will save stdout/stderr/error to this file
      --output-format string        Output format for the results (text, json) (default "text")
      --use-current-context         Only use the current context from the kubeconfig file, this can be used with --context-filter and --context-exclude
      --workers int                 Number of workers to run concurrently (default 99)

Use "kraken [command] --help" for more information about a command.
```

#### Output conditions

Output conditions are used to filter output, it's useful when you want to focus on specific output, e.g. pod is crashing.

Output conditions has format like `operator1:value1,operator2:value2`.

###### Operator: contains

Filter output that contains specific string:

```shell
kubekraken --output-conditions "contains:ImagePullBackOff" k -- get pods
```

###### Operator: not-contains

Filter output that does not contain specific string:

```shell
kubekraken --output-conditions "not-contains:Running" k -- get pods
```
