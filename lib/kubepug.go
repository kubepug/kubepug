package lib

import (
	"github.com/rikatz/kubepug/pkg/kubepug"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	// needed for auth
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags
)

// Config configuration object for Kubepug
// configurations for kubernetes and for kubepug functionality
type Config struct {
	K8sVersion      string
	ForceDownload   bool
	APIWalk         bool
	SwaggerDir      string
	ShowDescription bool
	ConfigFlags     *genericclioptions.ConfigFlags
}

// Kubepug struct to be used
type Kubepug struct {
	Config Config
}

// NewKubepug returns a new kubepug library
func NewKubepug(config Config) *Kubepug {
	return &Kubepug{Config: config}
}

// GetDeprecated returns the list of deprecated APIs
func (k *Kubepug) GetDeprecated() (*kubepug.Result, error) {
	var KubernetesAPIs kubepug.KubernetesAPIs = make(kubepug.KubernetesAPIs)

	swaggerfile, err := kubepug.DownloadSwaggerFile(k.Config.K8sVersion, k.Config.SwaggerDir, k.Config.ForceDownload)

	if err != nil {
		return nil, err
	}

	config, err := k.Config.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	err = KubernetesAPIs.PopulateKubeAPIMap(config, swaggerfile)

	if err != nil {
		return nil, err
	}

	result := kubepug.Result{}
	// First lets List all the deprecated APIs
	result.DeprecatedAPIs = KubernetesAPIs.ListDeprecated(config, k.Config.ShowDescription)

	if k.Config.APIWalk {
		result.DeletedAPIs = KubernetesAPIs.WalkObjects(config)
	}

	return &result, nil
}
