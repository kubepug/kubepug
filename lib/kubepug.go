package lib

import (
	"fmt"

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
}

// Kubepug struct to be used
type Kubepug struct {
	Config Config
}

// NewKubepug returns a new kubepug library
func NewKubepug(config Config) *Kubepug {
	return &Kubepug{Config: config}
}

// GetDeprecated returns the list of
func (k *Kubepug) GetDeprecated() error {
	kubernetesConfigFlags = genericclioptions.NewConfigFlags(true)

	var KubernetesAPIs kubepug.KubernetesAPIs = make(kubepug.KubernetesAPIs)

	swaggerfile, err := kubepug.DownloadSwaggerFile(k.Config.K8sVersion, k.Config.SwaggerDir, k.Config.ForceDownload)

	if err != nil {
		return err
	}

	config, err := kubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	fmt.Println("6")
	err = KubernetesAPIs.PopulateKubeAPIMap(config, swaggerfile)

	if err != nil {
		return err
	}
	fmt.Println("7")
	// First lets List all the deprecated APIs
	KubernetesAPIs.ListDeprecated(config, k.Config.ShowDescription)
	fmt.Println("8")
	if k.Config.APIWalk {
		KubernetesAPIs.WalkObjects(config)
	}
	fmt.Println("9")
	return nil
}
