package lib

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/rikatz/kubepug/pkg/kubepug"
	fileinput "github.com/rikatz/kubepug/pkg/kubepug/input/file"
	k8sinput "github.com/rikatz/kubepug/pkg/kubepug/input/k8s"

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

	log.Infof("Populating the Deprecated Kubernetes APIs Map")
	kubernetesAPIs, err := parser.NewAPIGroupsFromSwaggerFile(swaggerfile)
	if err != nil {
		return nil, err
	}

	log.Debugf("Kubernetes APIs Populated: %#v", kubernetesAPIs)

	result = k.getResults(kubernetesAPIs)

	return result, nil
}

func (k *Kubepug) getResults(kubeapis parser.APIGroups) *results.Result {
	var inputMode kubepug.Deprecator
	var err error
	if k.Config.Input != "" {
		inputMode, err = fileinput.NewFileInput(k.Config.Input, kubeapis)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		configRest, err := k.Config.ConfigFlags.ToRESTConfig()
		if err != nil {
			log.Fatalf("Failed to create the K8s config parameters while listing Deprecated objects: %s", err)
		}
		rest.SetDefaultWarningHandler(rest.NoWarnings{})

		client, err := dynamic.NewForConfig(configRest)
		if err != nil {
			log.Fatalf("Failed to create the K8s client while listing Deprecated objects: %s", err)
		}

		// Feed the KubeAPIs with the resourceName as this is used to the K8s Resource lister
		disco, err := discovery.NewDiscoveryClientForConfig(configRest)
		if err != nil {
			log.Fatalf("Failed to create the K8s Discovery client: %s", err)
		}
		// TODO: Use a constructor
		inputMode = &k8sinput.K8sInput{
			K8sconfig:          k.Config.ConfigFlags,
			Database:           kubeapis,
			APIWalk:            k.Config.APIWalk,
			Client:             client,
			DiscoveryClient:    disco,
			IncludePrefixGroup: []string{".k8s.io"},
			IgnoreExactGroup:   []string{"externaldns.k8s.io", "x-k8s.io", "flowcontrol.apiserver.k8s.io"},
		}
	}

	output, err := kubepug.GetDeprecations(inputMode)
	if err != nil {
		log.Fatal(err)
	}
	return &output
}
