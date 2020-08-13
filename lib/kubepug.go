package lib

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rikatz/kubepug/pkg/kubepug"
	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
	"github.com/rikatz/kubepug/pkg/utils"
	log "github.com/sirupsen/logrus"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Config configuration object for Kubepug
// configurations for kubernetes and for kubepug functionality
type Config struct {
	K8sVersion       string
	ForceDownload    bool
	APIWalk          bool
	SwaggerDir       string
	ShowDescription  bool
	Input            string
	Monitor          bool
	DeprecatedMetric *prometheus.CounterVec
	ScrapeInterval   time.Duration
	ConfigFlags      *genericclioptions.ConfigFlags
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
func (k *Kubepug) GetDeprecated() (result *results.Result, err error) {
	log.Debugf("Populating the KubernetesAPI map from swagger.json")

	var KubernetesAPIs parser.KubernetesAPIs = make(parser.KubernetesAPIs)

	log.Infof("Downloading the swagger.json file")
	swaggerfile, err := utils.DownloadSwaggerFile(k.Config.K8sVersion, k.Config.SwaggerDir, k.Config.ForceDownload)

	if err != nil {
		return &results.Result{}, err
	}

	log.Infof("Populating the Deprecated Kubernetes APIs Map")
	err = KubernetesAPIs.PopulateKubeAPIMap(swaggerfile)

	if err != nil {
		return &results.Result{}, err
	}

	log.Debugf("Kubernetes APIs Populated: %#v", KubernetesAPIs)

	result = k.getResults(KubernetesAPIs)

	return result, nil
}

// MeasureResults increments prometheus counter for deleted APIs
func (k *Kubepug) MeasureResults(result *results.Result, c *prometheus.CounterVec) {
	for _, d := range result.DeprecatedAPIs {
		for _, item := range d.Items {
			c.With(prometheus.Labels{
				"group":       d.Group,
				"version":     d.Version,
				"kind":        d.Kind,
				"name":        d.Name,
				"scope":       item.Scope,
				"object_name": item.ObjectName,
				"namespace":   item.Namespace,
				"deprocated":  strconv.FormatBool(d.Deprecated),
				"deleted":     "",
			}).Inc()
		}
	}

	for _, d := range result.DeletedAPIs {
		for _, item := range d.Items {
			c.With(prometheus.Labels{
				"group":       d.Group,
				"version":     d.Version,
				"kind":        d.Kind,
				"name":        d.Name,
				"scope":       item.Scope,
				"object_name": item.ObjectName,
				"namespace":   item.Namespace,
				"deprocated":  "",
				"deleted":     strconv.FormatBool(d.Deleted),
			}).Inc()
		}
	}
}

func (k *Kubepug) getResults(kubeapis parser.KubernetesAPIs) (result *results.Result) {
	var inputMode kubepug.Deprecator
	if k.Config.Input != "" {
		inputMode = kubepug.NewFileInput(k.Config.Input, kubeapis)

	} else {
		inputMode = kubepug.K8sInput{
			K8sconfig: k.Config.ConfigFlags,
			K8sapi:    kubeapis,
			Apiwalk:   k.Config.APIWalk,
			Monitor:   k.Config.Monitor,
		}
	}
	results := kubepug.GetDeprecations(inputMode)
	return &results
}
