package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/rikatz/kubepug/lib"
	"github.com/rikatz/kubepug/pkg/formatter"

	// Import the Kubernetes Authentication plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/release-utils/version"
)

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags

	generatedStore    string
	k8sVersion        string
	forceDownload     bool
	errorOnDeprecated bool
	errorOnDeleted    bool
	swaggerDir        string
	format            string
	filename          string
	inputFile         string
	logLevel          string

	outputFormatter formatter.Formatter

	rootCmd = &cobra.Command{
		Use:          filepath.Base(os.Args[0]),
		SilenceUsage: true,
		Short:        "Shows all the deprecated objects in a Kubernetes cluster allowing the operator to verify them before upgrading the cluster.\nIt uses the Kubernetes API source code markers to define deprecated and deleted versions.",
		Example:      filepath.Base(os.Args[0]),
		Args:         cobra.MinimumNArgs(0),
		PreRunE:      Complete,
		RunE:         runPug,
	}
)

func Complete(_ *cobra.Command, _ []string) error {
	var errComplete error
	var err error

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	if k8sVersion != "master" && k8sVersion != "main" && !semver.IsValid(k8sVersion) {
		errComplete = errors.Join(errComplete, fmt.Errorf("invalid Kubernetes version, should be 'master' or a valid semantic version"))
	}

	outputFormatter, err = formatter.NewFormatterWithError(format)
	if err != nil {
		errComplete = errors.Join(errComplete, err)
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		errComplete = errors.Join(errComplete, err)
	}

	if errComplete != nil {
		return errComplete
	}

	logrus.SetLevel(lvl)

	if lvl == logrus.DebugLevel {
		logrus.SetReportCaller(true)
	}
	return nil
}

func runPug(_ *cobra.Command, _ []string) error {
	config := lib.Config{
		GeneratedStore: generatedStore,
		K8sVersion:     k8sVersion,
		ConfigFlags:    kubernetesConfigFlags,
		Input:          inputFile,
	}

	logrus.Debugf("Starting Kubepug with configs: %+v", config)
	kubepug, err := lib.NewKubepug(&config)
	if err != nil {
		return err
	}

	result, err := kubepug.GetDeprecated()
	if err != nil {
		return err
	}

	logrus.Debug("Starting deprecated objects printing")
	bytes, err := outputFormatter.Output(*result)
	if err != nil {
		return err
	}

	if filename != "" {
		err = os.WriteFile(filename, bytes, 0o644) //nolint: gosec
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%s", string(bytes))
	}

	if (errorOnDeleted && len(result.DeletedAPIs) > 0) || (errorOnDeprecated && len(result.DeprecatedAPIs) > 0) {
		return fmt.Errorf("found %d Deleted APIs and %d Deprecated APIs", len(result.DeletedAPIs), len(result.DeprecatedAPIs))
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

	rootCmd.Flags().MarkHidden("as")                       //nolint: errcheck
	rootCmd.Flags().MarkHidden("as-group")                 //nolint: errcheck
	rootCmd.Flags().MarkHidden("cache-dir")                //nolint: errcheck
	rootCmd.Flags().MarkHidden("certificate-authority")    //nolint: errcheck
	rootCmd.Flags().MarkHidden("client-certificate")       //nolint: errcheck
	rootCmd.Flags().MarkHidden("client-key")               //nolint: errcheck
	rootCmd.Flags().MarkHidden("insecure-skip-tls-verify") //nolint: errcheck
	rootCmd.Flags().MarkHidden("namespace")                //nolint: errcheck
	rootCmd.Flags().MarkHidden("request-timeout")          //nolint: errcheck
	rootCmd.Flags().MarkHidden("server")                   //nolint: errcheck
	rootCmd.Flags().MarkHidden("token")                    //nolint: errcheck
	rootCmd.Flags().MarkHidden("user")                     //nolint: errcheck

	rootCmd.PersistentFlags().BoolVar(&errorOnDeprecated, "error-on-deprecated", false, "If a deprecated object is found, the program will exit with return code 1 instead of 0. Defaults to false")
	rootCmd.PersistentFlags().BoolVar(&errorOnDeleted, "error-on-deleted", false, "If a deleted object is found, the program will exit with return code 1 instead of 0. Defaults to false")
	rootCmd.PersistentFlags().StringVar(&k8sVersion, "k8s-version", "master", "Which Kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master")
	rootCmd.PersistentFlags().StringVar(&swaggerDir, "swagger-dir", "", "Where to keep swagger.json downloaded file. If not provided will use the system temporary directory")
	rootCmd.PersistentFlags().BoolVar(&forceDownload, "force-download", false, "Whether to force the download of a new swagger.json file even if one exists. Defaults to false")
	rootCmd.PersistentFlags().StringVar(&format, "format", "stdout", "Format in which the list will be displayed [stdout, plain, json, yaml]")
	rootCmd.PersistentFlags().StringVar(&filename, "filename", "", "Name of the file the results will be saved to, if empty it will display to stdout")
	rootCmd.PersistentFlags().StringVar(&inputFile, "input-file", "", "Location of a file or directory containing k8s manifests to be analysed. Use \"-\" to read from STDIN")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "verbosity", "v", logrus.WarnLevel.String(), "Log level: debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringVar(&generatedStore, "database", "https://kubepug.xyz/data/data.json", "Sets the generated database location. Can be remote file or local")
	rootCmd.AddCommand(version.WithFont("starwars"))

	rootCmd.PersistentFlags().MarkDeprecated("swagger-dir", "flag is deprecated and will be removed on next version. database flag should be used instead") //nolint: errcheck
	rootCmd.PersistentFlags().MarkDeprecated("force-download", "flag is deprecated and will be removed on next version. This flag is no-op.")               //nolint: errcheck
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("An error has occurred: %v", err)
		os.Exit(1)
	}
}
