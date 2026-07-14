// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/zos_base/pkg/gridtypes/zos"
)

// deployGatewayCmd represents the deploy gateway command
var deployGatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Deploy a gateway proxy",
}

func init() {
	deployCmd.AddCommand(deployGatewayCmd)

	deployGatewayCmd.PersistentFlags().StringP("name", "n", "", "name of the gateway (alphanumeric only)")
	err := deployGatewayCmd.MarkPersistentFlagRequired("name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	deployGatewayCmd.PersistentFlags().StringSlice("backends", []string{}, "backends for the gateway")
	err = deployGatewayCmd.MarkPersistentFlagRequired("backends")
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	deployGatewayCmd.PersistentFlags().Bool("tls", false, "add tls passthrough")
	deployGatewayCmd.PersistentFlags().String("network", "", "name of existing network to extend to gateway node (required when backend uses private WireGuard IPs)")
	deployGatewayCmd.PersistentFlags().String("project-name", "", "project name for grouping deployments (only alphanumeric and '/' allowed). Required when using --network)")
}

func parseCommonGatewayFlags(cmd *cobra.Command) (
	name string,
	tls bool,
	zosBackends []zos.Backend,
	node uint32,
	network string,
	projectName string,
	err error,
) {
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return
	}
	tls, err = cmd.Flags().GetBool("tls")
	if err != nil {
		return
	}
	backends, err := cmd.Flags().GetStringSlice("backends")
	if err != nil {
		return
	}
	for _, backend := range backends {
		zosBackends = append(zosBackends, zos.Backend(backend))
	}
	node, err = cmd.Flags().GetUint32("node")
	if err != nil {
		return
	}
	network, err = cmd.Flags().GetString("network")
	if err != nil {
		return
	}
	projectName, err = cmd.Flags().GetString("project-name")
	if err != nil {
		return
	}
	return
}
