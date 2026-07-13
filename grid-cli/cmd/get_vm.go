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

// getVMCmd represents the get vm command
var getVMCmd = &cobra.Command{
	Use:   "vm <vm-name>",
	Short: "Get deployed vm",
	Long: `Get details of a deployed virtual machine by name.

Requires the project name to locate the VM. The project name is typically
'vm/{vmname}' for VMs deployed with default settings, or a custom name if
specified during deployment.

Examples:
  # Get VM with default project name
  tfcmd get vm myvm --project-name vm/myvm
  
  # Get VM with custom project name
  tfcmd get vm webserver --project-name production`,
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

		vm, err := command.GetVM(cmd.Context(), t, projectName, args[0])
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		s, err := json.MarshalIndent(vm, "", "\t")
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		log.Info().Msg("vm:\n" + string(s))
	},
}

func init() {
	getCmd.AddCommand(getVMCmd)

	getVMCmd.Flags().StringP("project-name", "p", "", "project name of the VM")
	err := getVMCmd.MarkFlagRequired("project-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
