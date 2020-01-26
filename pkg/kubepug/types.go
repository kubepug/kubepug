package kubepug

import "github.com/fatih/color"

var gvColor = color.New(color.FgHiBlack).Add(color.Bold).SprintFunc()
var resourceColor = color.New(color.FgRed).Add(color.Bold).SprintFunc()
var globalColor = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
var namespaceColor = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
var errorColor = color.New(color.FgWhite).Add(color.BgRed).Add(color.Bold).SprintFunc()

// KubeAPI represents a Kubernetes API defined in swagger.json
type KubeAPI struct {
	description string
	group       string
	kind        string
	version     string
	name        string
	deprecated  bool
}

// KubernetesAPIs is a map of KubeAPI objects
type KubernetesAPIs map[string]KubeAPI
