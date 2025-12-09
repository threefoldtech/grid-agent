// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	command "github.com/threefoldtech/grid-agent/grid-cli/internal/cmd"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// getGatewayNameCmd represents the get gateway name command
var getGatewayNameCmd = &cobra.Command{
	Use:   "name <gateway-name>",
	Short: "Get deployed gateway name",
	Long: `Get details of a deployed gateway name proxy by name.

Requires the project name to locate the gateway.

Examples:
  # Get gateway name proxy
  tfcmd get gateway name mygateway --project-name myproject`,
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

		gateway, err := command.GetGatewayName(cmd.Context(), t, projectName, args[0])
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		s, err := json.MarshalIndent(gateway, "", "\t")
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		log.Info().Msg("gateway name:\n" + string(s))
	},
}

func init() {
	getGatewayCmd.AddCommand(getGatewayNameCmd)

	getGatewayNameCmd.Flags().StringP("project-name", "p", "", "project name of the gateway")
	err := getGatewayNameCmd.MarkFlagRequired("project-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
