// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// listGatewaysCmd represents the list gateways command
var listGatewaysCmd = &cobra.Command{
	Use:   "gateways",
	Short: "List available gateway nodes",
	Long: `List available gateway nodes on the ThreeFold Grid.

Gateway nodes provide domains or public IPs for exposing your workloads.

Available Commands:
  name        List gateways with domains (for name proxy)
  fqdn        List gateways with public IPs (for FQDN proxy)`,
}

func init() {
	listCmd.AddCommand(listGatewaysCmd)
}
