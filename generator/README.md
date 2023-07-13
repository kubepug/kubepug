# Deprecated Generator

This generator fetches the last stable version of Kubernetes API package, and 
generates a deprecations database based on this API markers.

This program is heavily based on (almost a copy) of https://github.com/kubernetes/code-generator/tree/master/cmd/prerelease-lifecycle-gen

It outputs today a json containing an array of deprecated items (this structure should be documented soon).

The idea is that this json can be consumed either by a status page, or by Kubepug in a much smaller and faster way than
the whole swagger.json file

## Running
On a working directory (and after building the generator):

```
GOPATH=$(pwd) go get k8s.io/api
GOPATH=$(pwd) generator -i k8s.io/api/./... > results.json
```
