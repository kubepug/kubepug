Kubepug is a program that helps you on your journey migrating between Kubernetes
versions.

Kubernetes [deprecates apis](https://kubernetes.io/docs/reference/using-api/deprecation-guide/) 
between releases, and when upgrading the cluster, the admin may face some of this deprecations
and a need to migrate or remove some resources.

Usually this API types are marked as `Deprecated` on a version and then `Deleted`
on a future version, depending on the decided lifecycle.

Kubepug can verify a Kubernetes cluster or manifests file, to check if some API 
has been deprecated or deleted. It allows you to set against which version of 
Kubernetes you want to check your manifests.

## Quick start

Below is a snippet of Kubepug output:
```
$ kubepug
RESULTS:
Deprecated APIs:
PodSecurityPolicy found in policy/v1beta1
	 ├─ Deprecated at: 1.21
	 ├─ PodSecurityPolicy governs the ability to make requests that affect the Security Context that will be applied to a pod and container.Deprecated in 1.21.
		-> OBJECT: restrictive namespace: default

Deleted APIs:
	 APIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!
Ingress found in extensions/v1beta1
	 ├─ Deleted at: 1.22
	 ├─ Replacement: networking.k8s.io/v1/Ingress
	 ├─ Ingress is a collection of rules that allow inbound connections to reach theendpoints defined by a backend. An Ingress can be configured to give servicesexternally-reachable urls, load balance traffic, terminate SSL, offer namebased virtual hosting etc. DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.
		-> OBJECT: bla namespace: blabla
```

## Features

* Verify if there is any deprecated or deleted resource on a running cluster
* Verify if there is any deprecated or deleted resource on a manifest
* Allow specifying different Kubernetes versions, as an API may not have been deprecated
or deleted yet depending on the version you want to migrate
* Provide the replacement API
* Inform what version the API was deprecated or deleted
* Output the results as colored and non colored text, json and yaml

## How it works?

Kubepug generates a `json` file based on [Kubernetes API definitions](https://github.com/kubernetes/api) containing the deprecated and deleted API definitions every 30 minutes. This definition file is called internally as `store`.

The API definition contains comments that represents the API lifecycle and can be 
found on the code as `+k8s:prerelease-lifecycle-gen`.

This file is downloaded on every Kubepug execution, and its size is about ~64kb and grows as new APIs are marked as deprecated or deleted.

!!! note "Air gapped scenarios"
    This file can be generated locally and it can also be downloaded and referred locally. Check on [Database](database.md) for more information

Using the store, Kubepug checks the input to verify if any existent manifest or resource uses a deprecated or deleted API, and reports it.

!!! note "Previous versions"
    On previously versions, Kubepug relied on [Kubernetes swagger](https://github.com/kubernetes/kubernetes/blob/master/api/openapi-spec/swagger.json) directives to define if an API was deprecated or deleted. This method wasn't reliable, as it needed the API to contain a "Deprecated" note on its description

## Acknowledges
As I've used this project to learn Go and also some Kubernetes client-go some parts of this plugin are based in Caio Begotti's Pod-Tree, Ahmet Balkan kubectl-tree and Bitnami Kubecfg

Logo based in Mão vetor criado por freepik - br.freepik.com
