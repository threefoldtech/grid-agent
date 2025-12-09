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

// getK8sCmd represents the get kubernetes command
var getK8sCmd = &cobra.Command{
	Use:   "kubernetes <cluster-name>",
	Short: "Get deployed kubernetes cluster",
	Long: `Get details of a deployed Kubernetes cluster by name.

Requires the project name to locate the cluster. The project name is typically
'kubernetes/{clustername}' for clusters deployed with tfcmd.

Examples:
  # Get Kubernetes cluster
  tfcmd get kubernetes mycluster --project-name kubernetes/mycluster`,
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

		cluster, err := command.GetK8sCluster(cmd.Context(), t, projectName, args[0])
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		s, err := json.MarshalIndent(cluster, "", "\t")
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		log.Info().Msg("k8s cluster:\n" + string(s))
	},
}

func init() {
	getCmd.AddCommand(getK8sCmd)

	getK8sCmd.Flags().StringP("project-name", "p", "", "project name of the cluster")
	err := getK8sCmd.MarkFlagRequired("project-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
