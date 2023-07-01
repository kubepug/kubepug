package lib

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/rikatz/kubepug/pkg/kubepug"
	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
	"github.com/rikatz/kubepug/pkg/utils"
)

// Config configuration object for Kubepug
// configurations for kubernetes and for kubepug functionality
type Config struct {
	// K8sVersion defines what is the Kubernetes version that the validation should target.
	// Should be on the Kubernetes semver format: v1.24.5
	K8sVersion string
	// ForceDownload defines if the download should happen even if the swagger file already exists
	ForceDownload bool
	// APIWalk defines if the expensive operation of checking every object should happen
	APIWalk bool
	// SwaggerDir defines where the swagger file should be saved. If empty, a temporary directory will be created and used.
	SwaggerDir string
	// ShowDescription defines if the description of the API should be copied to the output result
	ShowDescription bool
	Input           string
	ConfigFlags     *genericclioptions.ConfigFlags
}

// Kubepug defines a kubepug instance to be used
type Kubepug struct {
	Config Config
}

// NewKubepug returns a new kubepug library
func NewKubepug(config Config) *Kubepug {
	return &Kubepug{Config: config}
}

// GetDeprecated returns the list of deprecated APIs
func (k *Kubepug) GetDeprecated() (result *results.Result, err error) {

	log.Infof("Downloading the swagger.json file")
	swaggerfile, err := utils.DownloadSwaggerFile(k.Config.K8sVersion, k.Config.SwaggerDir, k.Config.ForceDownload)
	if err != nil {
		return &results.Result{}, err
	}

	kubernetesAPIs := make(parser.KubernetesAPIs)
	log.Infof("Populating the Deprecated Kubernetes APIs Map")
	err = kubernetesAPIs.PopulateKubeAPIMap(swaggerfile)

	if err != nil {
		return &results.Result{}, err
	}

	log.Debugf("Kubernetes APIs Populated: %#v", kubernetesAPIs)

	result = k.getResults(kubernetesAPIs)

	return result, nil
}

func (k *Kubepug) getResults(kubeapis parser.KubernetesAPIs) *results.Result {
	var inputMode kubepug.Deprecator
	if k.Config.Input != "" {
		inputMode = kubepug.NewFileInput(k.Config.Input, kubeapis)
	} else {
		inputMode = kubepug.K8sInput{
			K8sconfig: k.Config.ConfigFlags,
			K8sapi:    kubeapis,
			APIWalk:   k.Config.APIWalk,
		}
	}

	output := kubepug.GetDeprecations(inputMode)
	return &output
}
