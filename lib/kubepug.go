package lib

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/rikatz/kubepug/pkg/kubepug"
	fileinput "github.com/rikatz/kubepug/pkg/kubepug/input/file"
	k8sinput "github.com/rikatz/kubepug/pkg/kubepug/input/k8s"
	"github.com/rikatz/kubepug/pkg/store"
	"github.com/rikatz/kubepug/pkg/store/generatedstore"

	"github.com/rikatz/kubepug/pkg/results"
)

// Config configuration object for Kubepug
// configurations for kubernetes and for kubepug functionality
type Config struct {
	// GeneratedStore defines that the new GeneratedStore should be used. This variable should
	// either be a URL (http/s) or a local file location
	GeneratedStore string

	// K8sVersion defines what is the Kubernetes version that the validation should target.
	// Should be on the Kubernetes semver format: v1.24.5
	K8sVersion string

	Input       string
	ConfigFlags *genericclioptions.ConfigFlags
}

// Kubepug defines a kubepug instance to be used
type Kubepug struct {
	Config *Config
}

// NewKubepug returns a new kubepug library
func NewKubepug(config *Config) (*Kubepug, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be null")
	}
	return &Kubepug{Config: config}, nil
}

// GetDeprecated returns the list of deprecated APIs
func (k *Kubepug) GetDeprecated() (result *results.Result, err error) {
	var storer store.DefinitionStorer

	if k.Config.GeneratedStore == "" {
		return nil, fmt.Errorf("a database path should be provided")
	}

	storer, err = generatedstore.NewGeneratedStore(generatedstore.StoreConfig{
		Path:       k.Config.GeneratedStore,
		MinVersion: k.Config.K8sVersion,
	})

	if err != nil {
		return nil, err
	}

	result = k.getResults(storer)

	return result, nil
}

func (k *Kubepug) getResults(storer store.DefinitionStorer) *results.Result {
	var inputMode kubepug.Deprecator
	var err error
	if k.Config.Input != "" {
		inputMode, err = fileinput.NewFileInput(k.Config.Input, storer)
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
			Store:              storer,
			Client:             client,
			DiscoveryClient:    disco,
			IncludePrefixGroup: []string{".k8s.io"},
			// The groups below are: externaldns (not core), anything on x-k8s.io, internal flowcontrol and the autoscaling group that is actually a CRD (the real autoscaling is just autoscaling/version)
			IgnoreExactGroup: []string{"externaldns.k8s.io", "x-k8s.io", "flowcontrol.apiserver.k8s.io", "autoscaling.k8s.io"},
		}
	}

	output, err := kubepug.GetDeprecations(inputMode)
	if err != nil {
		log.Fatal(err)
	}
	return &output
}
