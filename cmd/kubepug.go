package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rikatz/kubepug/lib"
	"github.com/rikatz/kubepug/pkg/formatter"
	kpug "github.com/rikatz/kubepug/pkg/kubepug"
	"github.com/spf13/cobra"
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

const formatDesc = `choose a format for the list of deprecated APIs.
Options:
- plain: prints all deprecated APIs
- stdout: prints all deprecated APIs to STDOUT beautiffied
- json: outputs all deprecated APIs in JSON format
- yaml: outputs all deprecated APIs in YAML format
`

func runPug(cmd *cobra.Command, args []string) error {

	config := lib.Config{
		K8sVersion:      k8sVersion,
		ForceDownload:   forceDownload,
		APIWalk:         apiWalk,
		SwaggerDir:      swaggerDir,
		ShowDescription: showDescription,
	}

	kubepug := lib.NewKubepug(config)

	results := &kpug.Result{}
	deprecatedAPIs, err := kubepug.GetDeprecated()
	if err != nil {
		return err
	}
	results.DeprecatedAPIs = deprecatedAPIs

	if apiWalk {
		err = kubepug.WalkObjects()
		if err != nil {
			return err
		}
	}

	formatter := formatter.NewFormatter(format)
	bytes, err := formatter.Output(*results)
	if err != nil {
		return err
	}

	if filename != "" {
		err = ioutil.WriteFile(filename, bytes, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Printf("%s", string(bytes))
	return nil
}

func init() {
	if strings.Contains(filepath.Base(os.Args[0]), "kubectl-deprecations") {
		cmdValue := "kubectl deprecations"
		rootCmd.Use = cmdValue
		rootCmd.Example = cmdValue
	}

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
	rootCmd.PersistentFlags().StringVar(&format, "format", "stdout", formatDesc)
	rootCmd.PersistentFlags().StringVar(&filename, "filename", "", formatDesc)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
