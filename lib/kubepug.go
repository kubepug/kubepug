package lib

import (
	"fmt"
	"sort"

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
	K8sVersion      string
	ForceDownload   bool
	APIWalk         bool
	SwaggerDir      string
	ShowDescription bool
	Input           string
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

// GetAPIVersions returns the list of available API Groups and Versions
func (k *Kubepug) GetAPIVersions() (result *results.Result, err error) {
	kubernetesAPIs, err := k.getKubernetesAPIs()
	if err != nil {
		return &results.Result{}, err
	}

	// Avoid duplicate group/version strings
	apiVersionsMap := make(map[string]bool)
	var apiVersions []string
	for _, kubeAPI := range kubernetesAPIs {
		var apiVersion string
		if kubeAPI.Group != "" {
			apiVersion = fmt.Sprintf("%s/%s", kubeAPI.Group, kubeAPI.Version)
		} else {
			apiVersion = kubeAPI.Version
		}
		if _, value := apiVersionsMap[apiVersion]; !value {
			apiVersionsMap[apiVersion] = true
			apiVersions = append(apiVersions, apiVersion)
		}
	}

	// Sort the apiversions in the same way as kubectl
	sort.Strings(apiVersions)

	return &results.Result{
		APIVersions: apiVersions,
	}, nil
}

// GetDeprecated returns the list of deprecated APIs
func (k *Kubepug) GetDeprecated() (result *results.Result, err error) {
	kubernetesAPIs, err := k.getKubernetesAPIs()
	if err != nil {
		return &results.Result{}, err
	}

	result = k.getResults(kubernetesAPIs)

	return result, nil
}

func (k *Kubepug) getKubernetesAPIs() (kubernetesAPIs parser.KubernetesAPIs, err error) {
	log.Debugf("Populating the KubernetesAPI map from swagger.json")

	kubernetesAPIs = make(parser.KubernetesAPIs)

	log.Infof("Downloading the swagger.json file")
	swaggerfile, err := utils.DownloadSwaggerFile(k.Config.K8sVersion, k.Config.SwaggerDir, k.Config.ForceDownload)
	if err != nil {
		return nil, err
	}

	log.Infof("Populating the Deprecated Kubernetes APIs Map")
	err = kubernetesAPIs.PopulateKubeAPIMap(swaggerfile)

	if err != nil {
		return nil, err
	}

	log.Debugf("Kubernetes APIs Populated: %#v", kubernetesAPIs)

	return kubernetesAPIs, nil
}

func (k *Kubepug) getResults(kubeapis parser.KubernetesAPIs) *results.Result {
	var inputMode kubepug.Deprecator
	if k.Config.Input != "" {
		inputMode = kubepug.NewFileInput(k.Config.Input, kubeapis)
	} else {
		inputMode = kubepug.K8sInput{
			K8sconfig: k.Config.ConfigFlags,
			K8sapi:    kubeapis,
			Apiwalk:   k.Config.APIWalk,
		}
	}

	output := kubepug.GetDeprecations(inputMode)
	return &output
}
