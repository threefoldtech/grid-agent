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
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// listGatewaysFQDNCmd represents the list gateways fqdn command
var listGatewaysFQDNCmd = &cobra.Command{
	Use:   "fqdn",
	Short: "List gateways with public IPs",
	Long: `List gateway nodes with public IP addresses for FQDN Gateway deployments.

Shows node IDs and public IPs you need for deploying FQDN Gateways.
You'll point your custom domain's DNS to these IPs.

Examples:
  # List all FQDN gateways
  tfcmd list gateways fqdn
  
  # Filter by country
  tfcmd list gateways fqdn --country Germany
  
  # Filter by farm
  tfcmd list gateways fqdn --farm 2`,
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

		err = listGatewaysFQDN(ctx, t, farmID, country, cmd.OutOrStdout())
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	listGatewaysCmd.AddCommand(listGatewaysFQDNCmd)

	listGatewaysFQDNCmd.Flags().Uint64("farm", 0, "filter by farm ID")
	listGatewaysFQDNCmd.Flags().String("country", "", "filter by country")
}

func listGatewaysFQDN(ctx context.Context, t deployer.TFPluginClient, farmID uint64, country string, writer io.Writer) error {
	// Build filter for nodes with public IPs (no domain requirement)
	farmIDs := []uint64{}
	if farmID > 0 {
		farmIDs = []uint64{farmID}
	}

	rented := false
	freeIPs := uint64(1) // Require at least 1 public IP

	filter := types.NodeFilter{
		Status:  []string{"up"},
		FarmIDs: farmIDs,
		FreeIPs: &freeIPs,
		Rented:  &rented,
	}

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

	// Filter nodes with public IPs
	var gatewaysWithPublicIPs []types.Node
	for _, node := range nodes {
		if node.PublicConfig.Ipv4 != "" || node.PublicConfig.Ipv6 != "" {
			gatewaysWithPublicIPs = append(gatewaysWithPublicIPs, node)
		}
	}

	if len(gatewaysWithPublicIPs) == 0 {
		log.Info().Msg("No gateway nodes with public IPs found")
		return nil
	}

	printFQDNGatewayTable(gatewaysWithPublicIPs, writer)
	return nil
}

func printFQDNGatewayTable(nodes []types.Node, writer io.Writer) {
	table := tabwriter.NewWriter(writer, 0, 0, 4, ' ', 0)
	fmt.Fprintln(table, "Node ID\tPublic IPv4\tPublic IPv6\tFarm ID\tCountry\tCity")

	for _, node := range nodes {
		ipv4 := node.PublicConfig.Ipv4
		if ipv4 == "" {
			ipv4 = "-"
		}

		ipv6 := node.PublicConfig.Ipv6
		if ipv6 == "" {
			ipv6 = "-"
		}

		fmt.Fprintf(table, "%d\t%s\t%s\t%d\t%s\t%s\n",
			node.NodeID,
			ipv4,
			ipv6,
			node.FarmID,
			node.Location.Country,
			node.Location.City,
		)
	}

	table.Flush()
}
