// Package cmd for parsing command line arguments
package cmd

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/zos_sdk_go/grid-client/deployer"
	"github.com/threefoldtech/zos_sdk_go/grid-client/workloads"
)

// cancelCmd represents the cancel command
var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel resources on Threefold grid",
	Long: `Cancel resources by project name. 
	
Examples:
  # Cancel entire project
  tfcmd cancel --project-name myproject
  
  # Cancel specific VM from multi-node project
  tfcmd cancel --project-name myproject --vm vm02`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetUserConfig()
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		disableSentry, err := cmd.Flags().GetBool("disable-sentry")
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		projectName, err := cmd.Flags().GetString("project-name")
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		vmName, err := cmd.Flags().GetString("vm")
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

		if vmName != "" {
			// Cancel specific VM from project
			log.Info().Msgf("canceling VM '%s' from project '%s'", vmName, projectName)
			err = cancelVMFromProject(cmd.Context(), t, projectName, vmName)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
			log.Info().Msgf("successfully canceled VM '%s'", vmName)
		} else {
			// Cancel entire project
			log.Info().Msgf("canceling project '%s'", projectName)
			err = t.CancelByProjectName(projectName)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
			log.Info().Msgf("successfully canceled project '%s'", projectName)
		}
	},
}

func init() {
	rootCmd.AddCommand(cancelCmd)

	cancelCmd.Flags().StringP("project-name", "p", "", "project name to cancel")
	err := cancelCmd.MarkFlagRequired("project-name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	cancelCmd.Flags().String("vm", "", "specific VM name to cancel (optional, cancels entire project if not specified)")
}

// cancelVMFromProject removes a specific VM from a multi-node project
func cancelVMFromProject(ctx context.Context, t deployer.TFPluginClient, projectName, vmName string) error {
	vms, err := loadProjectVMs(ctx, t, projectName)
	if err != nil {
		return err
	}

	vmToCancel, remainingVMs := separateVM(vms, vmName)
	if vmToCancel == nil {
		return fmt.Errorf("VM '%s' not found in project '%s'", vmName, projectName)
	}

	if err := t.DeploymentDeployer.Cancel(ctx, vmToCancel); err != nil {
		return fmt.Errorf("failed to cancel VM '%s': %w", vmName, err)
	}

	if len(remainingVMs) > 0 && vmToCancel.NetworkName != "" {
		return updateNetworkAfterVMRemoval(ctx, t, projectName, vmToCancel.NetworkName, remainingVMs)
	}

	return nil
}

// loadProjectVMs loads all VM deployments in a project
func loadProjectVMs(ctx context.Context, t deployer.TFPluginClient, projectName string) ([]*workloads.Deployment, error) {
	contracts, err := t.ContractsGetter.ListContractsOfProjectName(projectName, true)
	if err != nil {
		return nil, err
	}

	var vms []*workloads.Deployment
	for _, contract := range contracts.NodeContracts {
		deploymentData, err := workloads.ParseDeploymentData(contract.DeploymentData)
		if err != nil {
			continue
		}

		// Only load VM deployments - skip networks, gateways, etc.
		if deploymentData.Type != workloads.VMType {
			continue
		}

		// Populate state with contract ID before loading deployment
		contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
		if err != nil {
			continue
		}

		// Add to state
		checkIfExistAndAppend(t, contract.NodeID, contractID)

		// Load the deployment using the VM name from deploymentData
		vm, err := t.State.LoadDeploymentFromGrid(ctx, contract.NodeID, deploymentData.Name)
		if err != nil {
			continue
		}
		vms = append(vms, &vm)
	}

	return vms, nil
}

// separateVM splits VMs into target and remaining
func separateVM(vms []*workloads.Deployment, targetName string) (*workloads.Deployment, []*workloads.Deployment) {
	var target *workloads.Deployment
	var remaining []*workloads.Deployment

	for _, vm := range vms {
		if vm.Name == targetName {
			target = vm
		} else {
			remaining = append(remaining, vm)
		}
	}

	return target, remaining
}

// updateNetworkAfterVMRemoval updates network by removing unused nodes
func updateNetworkAfterVMRemoval(ctx context.Context, t deployer.TFPluginClient, projectName, networkName string, remainingVMs []*workloads.Deployment) error {
	// Import command package for GetNetwork
	network, err := getNetworkForCancel(ctx, t, projectName, networkName)
	if err != nil {
		return fmt.Errorf("failed to load network '%s': %w", networkName, err)
	}

	usedNodes := collectUsedNodes(remainingVMs)
	removeUnusedNodes(&network, usedNodes)

	return t.NetworkDeployer.Deploy(ctx, &network)
}

// getNetworkForCancel loads network for cancellation (avoids circular import)
func getNetworkForCancel(ctx context.Context, t deployer.TFPluginClient, projectName, networkName string) (workloads.ZNet, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.NetworkType, networkName)
	if err != nil {
		return workloads.ZNet{}, err
	}

	if len(nodeContractIDs) == 0 {
		return workloads.ZNet{}, fmt.Errorf("network '%s' not found in project '%s'", networkName, projectName)
	}

	for node, contractID := range nodeContractIDs {
		t.State.CurrentNodeDeployments[node] = append(t.State.CurrentNodeDeployments[node], contractID)
	}

	return t.State.LoadNetworkFromGrid(ctx, networkName)
}

// collectUsedNodes extracts unique node IDs from VMs
func collectUsedNodes(vms []*workloads.Deployment) []uint32 {
	usedNodes := make([]uint32, 0, len(vms))
	for _, vm := range vms {
		if !slices.Contains(usedNodes, vm.NodeID) {
			usedNodes = append(usedNodes, vm.NodeID)
		}
	}
	return usedNodes
}

// removeUnusedNodes cleans up network nodes and keys
func removeUnusedNodes(network *workloads.ZNet, usedNodes []uint32) {
	for _, node := range network.Nodes {
		if !slices.Contains(usedNodes, node) {
			delete(network.MyceliumKeys, node)
		}
	}
	network.Nodes = usedNodes
}

// checkIfExistAndAppend adds contract ID to state if not already present
func checkIfExistAndAppend(t deployer.TFPluginClient, node uint32, contractID uint64) {
	for _, id := range t.State.CurrentNodeDeployments[node] {
		if id == contractID {
			return
		}
	}
	t.State.CurrentNodeDeployments[node] = append(t.State.CurrentNodeDeployments[node], contractID)
}
