// Package cmd for parsing command line arguments
package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	command "github.com/threefoldtech/grid-agent/grid-cli/internal/cmd"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/filters"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

// deployGatewayNameCmd represents the deploy gateway name command
var deployGatewayNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Deploy a gateway name proxy",
	Long: `Deploy a gateway name proxy for HTTP/HTTPS traffic.

Creates a subdomain on a gateway node that proxies to your backend.

To avoid name collisions, you can follow this convention to creates unique subdomain: solution prefix + twin ID + a chosen name. (e.g., wp41myapp)

IMPORTANT: Gateway must be able to reach your backend. Choose ONE option:

- SAME NETWORK: Deploy gateway and VM on same WireGuard network (private IPs)
   tfcmd deploy vm --name webapp --project-name myapp --ssh ~/.ssh/id_rsa.pub
   tfcmd deploy gateway name --name api --project-name myapp --node 11 --backends http://10.20.2.2:8080 --network myappnetwork
   Note: Even if on same physical node, they still need to be on same WireGuard network

- PUBLIC IPs: Use VM's public IPv4, planetary, or mycelium IP as backend
   tfcmd deploy vm --name webapp --ipv4 --ssh ~/.ssh/id_rsa.pub
   tfcmd deploy gateway name --name api --project-name mygateway --node 11 --backends http://<VM-PUBLIC-IP:PORT>

Examples:
  # Option 1: Same network (recommended for multi-VM)
  tfcmd deploy gateway name --name myapp --project-name myapp --node 11 --backends http://10.20.2.2:8080 --network myappnetwork`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, tls, zosBackends, node, network, projectName, err := parseCommonGatewayFlags(cmd)
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

		gateway := workloads.GatewayNameProxy{
			Name:           name,
			Backends:       zosBackends,
			TLSPassthrough: tls,
			SolutionType:   projectName, // Use project name
			Network:        network,
		}
		farm, err := cmd.Flags().GetUint64("farm")
		if err != nil {
			return err
		}
		noColor, err := cmd.Flags().GetBool("no-color")
		if err != nil {
			return err
		}
		disableSentry, err := cmd.Flags().GetBool("disable-sentry")
		if err != nil {
			return err
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
		if node == 0 {
			nodes, err := deployer.FilterNodes(
				cmd.Context(),
				t,
				filters.BuildGatewayFilter(farm),
				nil,
				nil,
				nil,
			)
			if err != nil {
				log.Fatal().Err(err).Send()
			}

			node = uint32(nodes[0].NodeID)
		}
		gateway.NodeID = node
		resGateway, err := command.DeployGatewayName(cmd.Context(), t, gateway)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		log.Info().Msgf("fqdn: %s", resGateway.FQDN)
		return nil
	},
}

func init() {
	deployGatewayCmd.AddCommand(deployGatewayNameCmd)

	deployGatewayNameCmd.Flags().Uint32("node", 0, "node id gateway should be deployed on")
	deployGatewayNameCmd.Flags().Uint64("farm", 0, "farm ID for deployment (0 = any farm, or specify farm ID)")
	deployGatewayNameCmd.MarkFlagsMutuallyExclusive("node", "farm")
}
