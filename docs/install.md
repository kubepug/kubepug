---
title: Installing Kubepug
---

# Installing Kubepug

Kubepug can be installed on some different ways:

!!! info "Additional installation methods"

    If you want an installation method that is not supported yet, feel free to
    open an issue on the [github project](https://github.com/rikatz/kubepug) asking for the new method.
    No promises! But we will do our best to support it!

## Krew plugin
Krew is a Kubernetes / kubectl plugin manager. Kubepug can be installed 
as a krew plugin with the following command:

```
kubectl krew install deprecations
```

## Snap
Kubepug can be installed with snap using the following command:

```
sudo snap install kubepug
```

If you want to install the development version, just use:

```
sudo snap install kubepug --edge
```

## Getting the binary
Kubepug is compiled as a binary for various Operational Systems and architectures. 

You can get the latest release from [the Release page](https://github.com/rikatz/kubepug/releases/latest)

## Github Action
Kubepug can be used as a Github Action with the following definition:

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
