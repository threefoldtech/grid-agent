// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	command "github.com/threefoldtech/grid-agent/grid-cli/internal/cmd"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/zos_sdk_go/grid-client/deployer"
)

// getZDBCmd represents the get zdb command
var getZDBCmd = &cobra.Command{
	Use:   "zdb <deployment-name>",
	Short: "Get deployed zdb",
	Long: `Get details of a deployed ZDB (Zero-DB) by deployment name.

Requires the project name to locate the ZDB deployment.

Examples:
  # Get ZDB deployment
  tfcmd get zdb myzdb --project-name myproject`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		noColor, err := cmd.Flags().GetBool("no-color")
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		disableSentry, err := cmd.Flags().GetBool("disable-sentry")
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		cfg, err := config.GetUserConfig()
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		opts := []deployer.PluginOpt{
			deployer.WithNetwork(cfg.Network),
			deployer.WithRMBTimeout(100),
		}

		if noColor {
			opts = append(opts, deployer.WithNoColorLogs())
		}

		if disableSentry {
			opts = append(opts, deployer.WithDisableSentry())
		}
		t, err := deployer.NewTFPluginClient(cfg.Mnemonics, opts...)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		projectName, err := cmd.Flags().GetString("project-name")
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		zdb, err := command.GetDeployment(cmd.Context(), t, projectName, args[0])
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		s, err := json.MarshalIndent(zdb, "", "\t")
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		log.Info().Msg("zdb:\n" + string(s))
	},
}

func init() {
	getCmd.AddCommand(getZDBCmd)

	getZDBCmd.Flags().StringP("project-name", "p", "", "project name of the ZDB deployment")
	err := getZDBCmd.MarkFlagRequired("project-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
