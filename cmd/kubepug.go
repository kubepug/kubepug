package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rikatz/kubepug/lib"
	"github.com/rikatz/kubepug/pkg/formatter"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags
)

var (
	k8sVersion      string
	forceDownload   bool
	apiWalk         bool
	swaggerDir      string
	showDescription bool
	format          string
	filename        string

	rootCmd = &cobra.Command{
		Use:          filepath.Base(os.Args[0]),
		SilenceUsage: true,
		Short:        "Shows all the deprecated objects in a Kubernetes cluster allowing the operator to verify them before upgrading the cluster. It uses the swagger.json version available in master branch of Kubernetes repository (github.com/kubernetes/kubernetes) as a reference.",
		Example:      filepath.Base(os.Args[0]),
		Args:         cobra.MinimumNArgs(0),
		RunE:         runPug,
	}
)

func runPug(cmd *cobra.Command, args []string) error {

	config := lib.Config{
		K8sVersion:      k8sVersion,
		ForceDownload:   forceDownload,
		APIWalk:         apiWalk,
		SwaggerDir:      swaggerDir,
		ShowDescription: showDescription,
		ConfigFlags:     kubernetesConfigFlags,
	}

	kubepug := lib.NewKubepug(config)

	result, err := kubepug.GetDeprecated()
	if err != nil {
		return err
	}

	formatter := formatter.NewFormatter(format)
	bytes, err := formatter.Output(*result)
	if err != nil {
		return err
	}

	if filename != "" {
		err = ioutil.WriteFile(filename, bytes, 0644)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%s", string(bytes))
	}

	return nil
}

func init() {
	if strings.Contains(filepath.Base(os.Args[0]), "kubectl-deprecations") {
		cmdValue := "kubectl deprecations"
		rootCmd.Use = cmdValue
		rootCmd.Example = cmdValue
	}

	kubernetesConfigFlags = genericclioptions.NewConfigFlags(true)
	kubernetesConfigFlags.AddFlags(rootCmd.Flags())

	rootCmd.Flags().MarkHidden("as")                       // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("as-group")                 // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("cache-dir")                // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("certificate-authority")    // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("client-certificate")       // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("client-key")               // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("insecure-skip-tls-verify") // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("namespace")                // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("request-timeout")          // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("server")                   // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("token")                    // Ignoring error in deepsource. skipcq: GSC-G104
	rootCmd.Flags().MarkHidden("user")                     // Ignoring error in deepsource. skipcq: GSC-G104

	rootCmd.PersistentFlags().BoolVar(&apiWalk, "api-walk", true, "Wether to walk in the whole API, checking if all objects type still exists in the current swagger.json. May be IO intensive to APIServer. Defaults to true")
	rootCmd.PersistentFlags().BoolVar(&showDescription, "description", true, "Wether to show the description of the deprecated object. The description may contain the solution for the deprecation. Defaults to true")
	rootCmd.PersistentFlags().StringVar(&k8sVersion, "k8s-version", "master", "Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master")
	rootCmd.PersistentFlags().StringVar(&swaggerDir, "swagger-dir", "", "Where to keep swagger.json downloaded file. If not provided will use the system temporary directory")
	rootCmd.PersistentFlags().BoolVar(&forceDownload, "force-download", false, "Wether to force the download of a new swagger.json file even if one exists. Defaults to false")
	rootCmd.PersistentFlags().StringVar(&format, "format", "stdout", "Format in which the list will be displayed [stdout, plain, json, yaml]")
	rootCmd.PersistentFlags().StringVar(&filename, "filename", "", "Name of the file the results will be saved to, if empty it will display to stdout")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
