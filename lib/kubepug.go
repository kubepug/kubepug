package lib

import (
	"github.com/rikatz/kubepug/pkg/kubepug"
	log "github.com/sirupsen/logrus"

	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	log.Debugf("Getting the deprecated APIs")
	var KubernetesAPIs kubepug.KubernetesAPIs = make(kubepug.KubernetesAPIs)

	log.Infof("Downloading the swagger.json file")
	swaggerfile, err := kubepug.DownloadSwaggerFile(k.Config.K8sVersion, k.Config.SwaggerDir, k.Config.ForceDownload)

	if err != nil {
		return nil, err
	}

	k8sConfig, err := k.Config.ConfigFlags.ToRESTConfig()
	log.Debugf("Will connect to cluster running in: %s", k8sConfig.Host)
	if err != nil {
		return nil, err
	}

	log.Infof("Populating the Deprecated Kubernetes APIs Map")
	err = KubernetesAPIs.PopulateKubeAPIMap(k8sConfig, swaggerfile)

	if err != nil {
		return nil, err
	}

	result := kubepug.Result{}
	// First lets List all the deprecated APIs
	log.Info("Getting existing objects that are marked as deprecated in swagger.json")
	result.DeprecatedAPIs = KubernetesAPIs.ListDeprecated(k8sConfig, k.Config.ShowDescription)

	if k.Config.APIWalk {
		log.Info("Getting existing objects that does not exist in swagger.json anymore")
		result.DeletedAPIs = KubernetesAPIs.WalkObjects(k8sConfig)
	}

	return &result, nil
}
