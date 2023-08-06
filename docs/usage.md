---
title: Usage
---

## Quick start
Simply calling Kubepug without flags will result on the `current Kubernetes context` to be checked against `the latest stable version`.

As an example, assuming you have a running Kubernetes cluster on version v1.19, and you have PodSecurityPolicies (an API deprecated on v1.21 and deleted on v1.24):

```
$ kubepug 
RESULTS:
Deleted APIs:
	 APIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!
PodSecurityPolicy found in policy/v1beta1
	 ├─ Deleted at: 1.25
	 ├─ PodSecurityPolicy governs the ability to make requests that affect the Security Context that will be applied to a pod and container.Deprecated in 1.21.
		-> OBJECT: restrictive namespace: default
```

## Specifying a target version
If you are not ready to migrate to the latest stable version, but still want to 
verify against a specific version, the flag `--k8s-version` can be used to check 
about the deprecation and deletions.

Assuming the same scenario above, on a Kubernetes cluster v1.19 with a PodSecurityPolicy, running Kubepug will give you the following report:

```
$ kubepug --k8s-version=v1.22 
[...]
RESULTS:
Deprecated APIs:
PodSecurityPolicy found in policy/v1beta1
	 ├─ Deprecated at: 1.21
	 ├─ PodSecurityPolicy governs the ability to make requests that affect the Security Context that will be applied to a pod and container.Deprecated in 1.21.
		-> OBJECT: restrictive namespace: default
```

## Checking manifests / local files
Kubepug can check local files instead of a Kubernetes cluster, using the flag `--input-file`.

!!! note "Checking files in a directory"
    Besides of the bad flag name, `--input-file` can also be used to check all files in a directory as well!

On this example, we have two types of resources: An `extensions/v1beta1/Ingress` that was deleted at Kubernetes v1.22 and `policy/v1beta1/PodSecurityPolicy` that was deprecated on v1.21.

```
kubepug --k8s-version=v1.22 --input-file=./manifests/
RESULTS:
Deprecated APIs:
PodSecurityPolicy found in policy/v1beta1
	 ├─ Deprecated at: 1.21
	 ├─ PodSecurityPolicy governs the ability to make requests that affect the Security Context that will be applied to a pod and container.Deprecated in 1.21.
		-> OBJECT: restrictive namespace: default location: ./manifests/psp1.yaml

Deleted APIs:
	 APIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!
Ingress found in extensions/v1beta1
	 ├─ Deleted at: 1.22
	 ├─ Replacement: networking.k8s.io/v1/Ingress
	 ├─ Ingress is a collection of rules that allow inbound connections to reach theendpoints defined by a backend. An Ingress can be configured to give servicesexternally-reachable urls, load balance traffic, terminate SSL, offer namebased virtual hosting etc. DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.
		-> OBJECT: bla namespace: blabla location: ./manifests/ingress.yaml
```

## Reporting on other formats
The following formats can be passed to the `--format` flag:
* `stdout` (default) - Prints the output to stdout formatted and with colors 
* `plain` - Prints the output unformatted to stdout
* `json` - Prints the output in a JSON format
* `yaml` - Prints the output in YAML format

!!! note "Additional formats"
    We have on a roadmap to support additional formats! Feel free to open an issue on the Github project if you miss any format that you need!

## Using your own data file
In case you don't want to always download the [data.json](https://kubepug.xyz/data/data.json) file, you can generate yours, or download it once and use it locally with the flag `--database`. 

The flag accepts remote paths, like `--database=https://my.location.tld/data.json` or a local path, like `--database=/home/rkatz/kubepug/data.json`

See the [database](database.md) page for more information on generating your own file.

## Other command flags

The other flags of the command are:

```
      --as-uid string            UID to impersonate for the operation.
      --cluster string           The name of the kubeconfig cluster to use
      --context string           The name of the kubeconfig context to use
      --database string          Sets the generated database location. Can be remote file or local (default "https://kubepug.xyz/data/data.json")
      --disable-compression      If true, opt-out of response compression for all requests to the server
      --error-on-deleted         If a deleted object is found, the program will exit with return code 1 instead of 0. Defaults to false
      --error-on-deprecated      If a deprecated object is found, the program will exit with return code 1 instead of 0. Defaults to false
      --filename string          Name of the file the results will be saved to, if empty it will display to stdout
      --format string            Format in which the list will be displayed [stdout, plain, json, yaml] (default "stdout")
  -h, --help                     help for kubepug
      --input-file string        Location of a file or directory containing k8s manifests to be analysed. Use "-" to read from STDIN
      --k8s-version string       Which Kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master (default "master")
      --kubeconfig string        Path to the kubeconfig file to use for CLI requests.
      --tls-server-name string   Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
  -v, --verbosity string         Log level: debug, info, warn, error, fatal, panic (default "warning")
```
