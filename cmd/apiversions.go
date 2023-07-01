package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rikatz/kubepug/lib"
	"github.com/rikatz/kubepug/pkg/formatter"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "api-versions",
	Short: "Print the supported API versions for the requested k8s-version, in the form of \"group/version\"",
	Long:  "Print the supported API versions for the requested k8s-version, in the form of \"group/version\"",
	RunE:  runAPIVersions,
}

func runAPIVersions(cmd *cobra.Command, args []string) error {
	config := lib.Config{
		K8sVersion:    k8sVersion,
		ForceDownload: forceDownload,
		SwaggerDir:    swaggerDir,
		ConfigFlags:   kubernetesConfigFlags,
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	if lvl == logrus.DebugLevel {
		logrus.SetReportCaller(true)
	}

	logrus.Debugf("Starting Kubepug with configs: %+v", config)
	kubepug := lib.NewKubepug(config)

	result, err := kubepug.GetAPIVersions()
	if err != nil {
		return err
	}

	logrus.Debug("Starting API versions printing")
	format := formatter.NewFormatter(getOutputFormat(format))
	bytes, err := format.Output(*result)
	if err != nil {
		return err
	}

	if filename != "" {
		err = os.WriteFile(filename, bytes, 0o644) // nolint: gosec
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%s", string(bytes))
	}

	return nil
}

// Allow structured formats
func getOutputFormat(t string) string {
	switch t {
	case "json":
		return t
	case "yaml":
		return t
	default:
		return "apiversions"
	}
}
