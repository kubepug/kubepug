package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/rikatz/kubepug/pkg/version"
)

func Version() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the kubepug version",
		Long:  "Prints the kubepug version",

		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.VersionInfo()
			res := v.String()
			if outputJSON {
				j, err := v.JSONString()
				if err != nil {
					return errors.Wrap(err, "unable to generate JSON from version info")
				}
				res = j
			}
			fmt.Println(res)
			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false,
		"print JSON instead of text")

	return cmd
}
