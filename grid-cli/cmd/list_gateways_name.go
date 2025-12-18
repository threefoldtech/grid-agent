package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/filters"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// listGatewaysNameCmd represents the list gateways name command
var listGatewaysNameCmd = &cobra.Command{
	Use:   "name",
	Short: "List gateways with domains",
	Long: `List gateway nodes with configured domains for Name Gateway deployments.

Shows available domains you'll get when deploying a Name Gateway.

Examples:
  # List all name gateways
  tfcmd list gateways name
  
  # Filter by country
  tfcmd list gateways name --country Belgium
  
  # Filter by farm
  tfcmd list gateways name --farm 1`,
	Run: func(cmd *cobra.Command, args []string) {
		farmID, err := cmd.Flags().GetUint64("farm")
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		country, err := cmd.Flags().GetString("country")
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

		if disableSentry {
			opts = append(opts, deployer.WithDisableSentry())
		}

		t, err := deployer.NewTFPluginClient(cfg.Mnemonics, opts...)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = listGatewaysName(ctx, t, farmID, country, cmd.OutOrStdout())
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	listGatewaysCmd.AddCommand(listGatewaysNameCmd)

	listGatewaysNameCmd.Flags().Uint64("farm", 0, "filter by farm ID")
	listGatewaysNameCmd.Flags().String("country", "", "filter by country")
}

func listGatewaysName(ctx context.Context, t deployer.TFPluginClient, farmID uint64, country string, writer io.Writer) error {
	// Build filter for gateways with domains
	filter := filters.BuildGatewayFilter(farmID)

	// Query nodes
	nodes, err := deployer.FilterNodes(ctx, t, filter, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to query gateway nodes: %w", err)
	}

	// Filter by country if specified
	if country != "" {
		var filtered []types.Node
		for _, node := range nodes {
			if strings.EqualFold(node.Location.Country, country) {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Filter nodes with domains
	var gatewaysWithDomains []types.Node
	for _, node := range nodes {
		if node.PublicConfig.Domain != "" {
			gatewaysWithDomains = append(gatewaysWithDomains, node)
		}
	}

	if len(gatewaysWithDomains) == 0 {
		log.Info().Msg("No gateway nodes with domains found")
		return nil
	}

	printNameGatewayTable(gatewaysWithDomains, writer)
	return nil
}

func printNameGatewayTable(nodes []types.Node, writer io.Writer) {
	table := tabwriter.NewWriter(writer, 0, 0, 4, ' ', 0)
	fmt.Fprintln(table, "Node ID\tDomain\tFarm ID\tCountry\tCity")

	for _, node := range nodes {
		fmt.Fprintf(table, "%d\t%s\t%d\t%s\t%s\n",
			node.NodeID,
			node.PublicConfig.Domain,
			node.FarmID,
			node.Location.Country,
			node.Location.City,
		)
	}

	table.Flush()
}
