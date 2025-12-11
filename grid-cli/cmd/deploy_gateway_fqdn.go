// Package cmd for parsing command line arguments
package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	command "github.com/threefoldtech/grid-agent/grid-cli/internal/cmd"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

// deployGatewayFQDNCmd represents the deploy gateway fqdn command
var deployGatewayFQDNCmd = &cobra.Command{
	Use:   "fqdn",
	Short: "Deploy a gateway FQDN proxy",
	Long: `Deploy a gateway FQDN proxy for custom domains.

Use your own domain name. You must configure DNS to point to the gateway node.

Examples:
  # Deploy gateway with custom domain
  tfcmd deploy gateway fqdn --name myapp --fqdn myapp.example.com --node 14 --backends http://10.20.2.2:8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, tls, zosBackends, node, network, projectName, err := parseCommonGatewayFlags(cmd)
		if err != nil {
			return err
		}
		fqdn, err := cmd.Flags().GetString("fqdn")
		if err != nil {
			return err
		}

		// Validate: --network requires --project-name
		if network != "" && projectName == "" {
			return fmt.Errorf("--project-name is required when using --network")
		}

		// Default project name to gateway name if not provided
		if projectName == "" {
			projectName = name
		}

		noColor, err := cmd.Flags().GetBool("no-color")
		if err != nil {
			return err
		}
		disableSentry, err := cmd.Flags().GetBool("disable-sentry")
		if err != nil {
			return err
		}

		gateway := workloads.GatewayFQDNProxy{
			Name:           name,
			Backends:       zosBackends,
			TLSPassthrough: tls,
			SolutionType:   projectName, // Use project name
			FQDN:           fqdn,
			NodeID:         node,
			Network:        network,
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
		err = command.DeployGatewayFQDN(cmd.Context(), t, gateway)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		log.Info().Msg("gateway fqdn deployed")
		return nil
	},
}

func init() {
	deployGatewayCmd.AddCommand(deployGatewayFQDNCmd)

	deployGatewayFQDNCmd.Flags().String("fqdn", "", "fqdn pointing to the specified node")
	err := deployGatewayFQDNCmd.MarkFlagRequired("fqdn")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	deployGatewayFQDNCmd.Flags().Uint32("node", 0, "node id gateway should be deployed on")
	err = deployGatewayFQDNCmd.MarkFlagRequired("node")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
