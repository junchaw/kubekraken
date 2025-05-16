# KubeKraken

<div align="center">
  <img src="docs/logo.png" alt="kubekraken logo" width="200">
</div>

[![Docker Pulls](https://img.shields.io/docker/pulls/junchaw/kubekraken.svg)](https://hub.docker.com/r/junchaw/kubekraken/)
[![Release a tag](https://github.com/junchaw/kubekraken/actions/workflows/release.yml/badge.svg)](https://github.com/junchaw/kubekraken/actions/workflows/release.yml)

Kubekraken is a powerful CLI tool that unleashes multiple kubectl commands in parallel—tame your clusters with ease.

```shell
kubekraken --kubeconfig-files ~/kubeconfigs --kubeconfig-filter "us-west-2.*" k -- rollout restart -n argocd deploy/argocd-server
```

```shell
---
TASK START: /Users/junchawu/kubeconfigs/us-west-2-prd-2.yaml@default (2/3)

STDERR:
Warning: Use tokens from the TokenRequest API or manually created secret-based tokens instead of auto-generated secret-based tokens.

STDOUT:
deployment.apps/argocd-server restarted

TASK END: /Users/junchawu/kubeconfigs/us-west-2-prd-2.yaml@default (2/3)
---

---
TASK START: /Users/junchawu/kubeconfigs/us-west-2-prd-1.yaml@default (3/3)

ERROR: Connection refused.

TASK END: /Users/junchawu/kubeconfigs/us-west-2-prd-1.yaml@default (3/3)
---

---
TASK START: /Users/junchawu/kubeconfigs/us-west-2-prd-3.yaml@default (1/3)

STDOUT:
deployment.apps/argocd-server restarted

TASK END: /Users/junchawu/kubeconfigs/us-west-2-prd-3.yaml@default (1/3)
---

---
SUMMARY:
- /Users/junchawu/kubeconfigs/us-west-2-prd-1.yaml@default: error: Connection refused.
- /Users/junchawu/kubeconfigs/us-west-2-prd-2.yaml@default: stderr:
Warning: Use tokens from the TokenRequest API or manually created secret-based tokens instead of auto-generated secret-based tokens.

2 successful (1 with warnings), 1 error, 3 total
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
      --context-filter string      Regex filter for context names (e.g. prd-.*), see --use-current-context if you want to use the default context
      --kubeconfig-files strings   Kubeconfig files, item could be directory or file, in case of directory, all files in the directory will be used, see --kubeconfig-filter (default [/Users/junchawu/.kube/config])
      --kubeconfig-filter string   Regex filter for kubeconfig files, used with kubeconfig from directory, will not filter items specified in --kubeconfig-files (e.g. prd-.*\.yaml)
      --no-stdout                  Do not print kubectl stdout (stderr will still be printed), will be ignored if --output-file or --output-dir is set
      --output-dir string          Output directory for the results, will be ignored if --output-file is set
      --output-file string         Output file for the results, has higher priority than --output-dir
      --use-current-context        Only use the current context from the kubeconfig file, if set, --kubeconfig-filter will be ignored
      --workers int                Number of workers to run concurrently (default 99)
```

Run with docker:

```shell
docker run junchaw/kubekraken -h
```

Or run from source:

```shell
go run .
```
