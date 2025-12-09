// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	command "github.com/threefoldtech/grid-agent/grid-cli/internal/cmd"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with mnemonics to a grid network",
	Long: `Login to the ThreeFold Grid with your mnemonics.

This command will prompt you to enter your mnemonic phrase and select
a network (mainnet, testnet, devnet, or qa). Your credentials will be
saved locally for future commands.

Examples:
  # Interactive login
  tfcmd login`,
	Run: func(cmd *cobra.Command, args []string) {
		err := command.Login()
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
