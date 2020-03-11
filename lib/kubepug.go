package kubepug

// Config configuration object for Kubepug
// configurations for kubernetes and for kubepug functionality
type Config struct {
	k8sVersion      string
	forceDownload   bool
	apiWalk         bool
	swaggerDir      string
	showDescription bool
	format          string
	filename        string
}

// NewKubepug returns a new kubepug library
func NewKubepug() {

}
