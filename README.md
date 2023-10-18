# Deprecations  AKA KubePug - Pre UpGrade (Checker)
[![Build Status](https://github.com/kubepug/kubepug/actions/workflows/build.yml/badge.svg)](https://github.com/kubepug/kubepug/actions/workflows/build.yml)
[![codecov](https://codecov.io/github/rikatz/kubepug/graph/badge.svg?token=BIAQ7JIYD1)](https://codecov.io/github/rikatz/kubepug)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubepug/kubepug)](https://goreportcard.com/report/github.com/kubepug/kubepug)
[![kubepug](https://snapcraft.io/kubepug/badge.svg)](https://snapcraft.io/kubepug)


![Kubepug](assets/kubepug.png)

KubePug/Deprecations is intended to be a kubectl plugin, which:

* Downloads a data.json generated containing Kubernetes APIs deprecation information
* Verifies the current Kubernetes cluster or input files checking whether exists objects in this deprecated API Versions, allowing the user to check before migrating

## Features
* Can run against a Kubernetes cluster, using kubeconfig or the current cluster
* Can run against a different set of manifest/files
* Allows specifying the target Kubernetes version to be validated
* Provides the replacement API that should be used
* Informs the version that the API was deprecated or deleted, based on the target cluster version

## How to use it as a krew plugin

Just run `kubectl krew install deprecations`

## How to use it with Helm

If you want to verify the generated manifests by Helm, you can run the program as following:

```console
helm template -f values.yaml .0 | kubepug --k8s-version v1.22.0 --input-file=-
```

Change the arguments in kubepug program (and Helm template!) as desired!

## How to Use it as a standalone program

Download the correct version from [Releases](https://github.com/kubepug/kubepug/releases/latest) page.

After that, the command can be used just as kubectl, but with the following flags:

```console
$ kubepug --help
[...]
Flags:
      --cluster string           The name of the kubeconfig cluster to use
      --context string           The name of the kubeconfig context to use
      --database string          Sets the generated database location. Can be remote file or local (default "https://kubepug.xyz/data/data.json")
      --error-on-deleted         If a deleted object is found, the program will exit with return code 1 instead of 0. Defaults to false
      --error-on-deprecated      If a deprecated object is found, the program will exit with return code 1 instead of 0. Defaults to false
      --filename string          Name of the file the results will be saved to, if empty it will display to stdout
      --format string            Format in which the list will be displayed [stdout, plain, json, yaml] (default "stdout")
  -h, --help                     help for kubepug
      --input-file string        Location of a file or directory containing k8s manifests to be analized
      --k8s-version string       Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master (default "master")
      --kubeconfig string        Path to the kubeconfig file to use for CLI requests.
      --tls-server-name string   Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
  -v, --verbosity string         Log level: debug, info, warn, error, fatal, panic (default "warning")
```

#### Alternatively you can install using go install
```
$ go install github.com/kubepug/kubepug@latest
```

### Checking a Kubernetes Cluster

You can check the status of a running cluster with the following command.

```console
$ kubepug --k8s-version=v1.22 # Will verify the current context against v1.22 version
[...]
RESULTS:
Deprecated APIs:
PodSecurityPolicy found in policy/v1beta1
	 ├─ Deprecated at: 1.21
	 ├─ PodSecurityPolicy governs the ability to make requests that affect the Security Contextthat will be applied to a pod and container.Deprecated in 1.21.
		-> OBJECT: restrictive namespace: default

Deleted APIs:
	 APIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!
Ingress found in extensions/v1beta1
	 ├─ Deleted at: 1.22
	 ├─ Replacement: networking.k8s.io/v1/Ingress
	 ├─ Ingress is a collection of rules that allow inbound connections to reach theendpoints defined by a backend. An Ingress can be configured to give servicesexternally-reachable urls, load balance traffic, terminate SSL, offer namebased virtual hosting etc.DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.
		-> OBJECT: bla namespace: blabla
```

### Putting Kubepug in your CI / Checking input files

You can verify files with the following:

```console
$ kubepug --input-file=./deployment/ --error-on-deleted --error-on-deprecated
```

With the command above
* The data.json from `https://kubepug.xyz/data/data.json` will be used
* All YAML files (excluding subdirectories) will be verified
* The program will exit with an error if deprecated or deleted objects are found.

### Air-gapped environment

This happens when you have a secure environment that does not have an internet connectivity.

The data.json file is generated every hour, based on the latest stable version of Kubernetes API. 
You can download it from `https://kubepug.xyz/data/data.json` and move it to a safe location.

Then run kubepug pointing to the location of this file:

```console
kubepug --k8s-version=v1.22 --database=location/of/your/data.json
```

### Building your own data.json file

Steps to follow:

1. Clone/Download this repository, and build the container on `generator/` directory

```console
git clone https://github.com/kubepug/kubepug
docker build -t generator -f generator/Dockerfile generator
```

2. Generate the data.json
```console
docker run generator > data.json
```

Generator uses the latest stable Kubernetes API version, if you want the latest dev version you should run as:
```
docker run -e VERSION=master generator > data.json
```

3. Securely move the json file to your Air-Gapped environment, to the folder of your choosing. This folder will be used by `kubepug`.

4. Execute `kubepug` with the option `database`, like this

```console
kubepug --k8s-version=v1.22 --database=location/of/your/data.json
```

### Example of Usage in CI with Github Actions

```yaml
name: Sample CI Workflow
# This workflow is triggered on pushes to the repository.
on: [push]
env:
  HELM_VERSION: "v3.9.0"
  K8S_TARGET_VERSION: "v1.22.0"

jobs:
 api-deprecations-test:
    runs-on: ubuntu-latest
    steps:
      - name: Check-out repo
        uses: actions/checkout@v2

      - uses: azure/setup-helm@v1
        with:
          version: $HELM_VERSION
        id: install

      - uses: cpanato/kubepug-installer@v1.0.0

      - name: Run Kubepug with your Helm Charts Repository
        run: |
          find charts -mindepth 1 -maxdepth 1 -type d | xargs -t -n1 -I% /bin/bash -c 'helm template % --api-versions ${K8S_TARGET_VERSION} | kubepug --error-on-deprecated --error-on-deleted --k8s-version ${K8S_TARGET_VERSION} --input-file /dev/stdin'
```

## Screenshot

![Kubepug](assets/screenshot.png)

## References

As I've used this project to learn Go and also some Kubernetes [client-go](https://github.com/kubernetes/client-go/) some parts of this plugin are based in Caio Begotti's [Pod-Tree](https://github.com/caiobegotti/Pod-Dive), Ahmet Balkan [kubectl-tree](https://github.com/ahmetb/kubectl-tree) and Bitnami [Kubecfg](https://github.com/bitnami/kubecfg)

Logo based in <a href="https://br.freepik.com/fotos-vetores-gratis/mao">Mão vetor criado por freepik - br.freepik.com</a>
