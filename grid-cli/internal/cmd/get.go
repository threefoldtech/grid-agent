// Package cmd for handling commands
package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/threefoldtech/zos_sdk_go/grid-client/deployer"
	"github.com/threefoldtech/zos_sdk_go/grid-client/workloads"
)

// checkIfExistAndAppend checks if a contract exists in CurrentNodeDeployments and adds it if not
func checkIfExistAndAppend(t deployer.TFPluginClient, nodeID uint32, contractID uint64) {
	if contractIDs, exists := t.State.CurrentNodeDeployments[nodeID]; exists {
		for _, id := range contractIDs {
			if id == contractID {
				return // Contract already exists
			}
		}
		t.State.CurrentNodeDeployments[nodeID] = append(contractIDs, contractID)
	} else {
		t.State.CurrentNodeDeployments[nodeID] = []uint64{contractID}
	}
}

// GetVM gets a VM by project name and VM name
func GetVM(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.Deployment, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.VMType, name)
	if err != nil {
		return workloads.Deployment{}, err
	}

	if len(nodeContractIDs) == 0 {
		return workloads.Deployment{}, fmt.Errorf("no contracts found for VM '%s' in project '%s'", name, projectName)
	}

	var nodeID uint32
	for node, contractID := range nodeContractIDs {
		checkIfExistAndAppend(t, node, contractID)
		nodeID = node
	}

	return t.State.LoadDeploymentFromGrid(ctx, nodeID, name)
}

// GetK8sCluster gets a Kubernetes cluster by project name and cluster name
func GetK8sCluster(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.K8sCluster, error) {
	// Get all contracts for the project
	contracts, err := t.ContractsGetter.ListContractsOfProjectName(projectName, true)
	if err != nil {
		return workloads.K8sCluster{}, err
	}

	if len(contracts.NodeContracts) == 0 {
		return workloads.K8sCluster{}, fmt.Errorf("no contracts found for cluster '%s' in project '%s'", name, projectName)
	}

	// Filter contracts to only include those belonging to this specific cluster
	// Kubernetes clusters have multiple contracts (master + workers) with the same cluster name
	var nodeIDs []uint32
	for _, contract := range contracts.NodeContracts {
		deploymentData, err := workloads.ParseDeploymentData(contract.DeploymentData)
		if err != nil {
			continue
		}

		// Only include contracts that belong to this specific cluster
		// For K8s, the deployment name matches the cluster name
		if deploymentData.Name == name {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				continue
			}

			// Populate state with only the specific contracts we need
			checkIfExistAndAppend(t, contract.NodeID, contractID)
			nodeIDs = append(nodeIDs, contract.NodeID)
		}
	}

	if len(nodeIDs) == 0 {
		return workloads.K8sCluster{}, fmt.Errorf("no nodes found for cluster '%s' in project '%s'", name, projectName)
	}

	// Now LoadK8sFromGrid will work reliably
	return t.State.LoadK8sFromGrid(ctx, nodeIDs, name)
}

// GetNetwork gets a network by project name and network name
func GetNetwork(ctx context.Context, t deployer.TFPluginClient, projectName, networkName string) (workloads.ZNet, error) {
	// Use targeted contract search instead of loading all user contracts
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.NetworkType, networkName)
	if err != nil {
		return workloads.ZNet{}, fmt.Errorf("failed to find network contracts: %w", err)
	}

	if len(nodeContractIDs) == 0 {
		return workloads.ZNet{}, fmt.Errorf("network '%s' not found in project '%s'", networkName, projectName)
	}

	// Populate state with only the specific contracts we found
	for nodeID, contractID := range nodeContractIDs {
		checkIfExistAndAppend(t, nodeID, contractID)
	}

	// Load network from grid
	network, err := t.State.LoadNetworkFromGrid(ctx, networkName)
	if err != nil {
		return workloads.ZNet{}, fmt.Errorf("failed to get network %s: %w", networkName, err)
	}

	// Verify the network belongs to the expected project
	if network.SolutionType != projectName {
		return workloads.ZNet{}, fmt.Errorf("network '%s' found but belongs to project '%s', not '%s'", networkName, network.SolutionType, projectName)
	}

	return network, nil
}

// GetNetworkLight gets a light network by project name and network name
func GetNetworkLight(ctx context.Context, t deployer.TFPluginClient, projectName, networkName string) (workloads.ZNetLight, error) {
	// Use targeted contract search instead of loading all user contracts
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.NetworkType, networkName)
	if err != nil {
		return workloads.ZNetLight{}, fmt.Errorf("failed to find network contracts: %w", err)
	}

	if len(nodeContractIDs) == 0 {
		return workloads.ZNetLight{}, fmt.Errorf("network '%s' not found in project '%s'", networkName, projectName)
	}

	// Populate state with only the specific contracts we found
	for nodeID, contractID := range nodeContractIDs {
		checkIfExistAndAppend(t, nodeID, contractID)
	}

	// Load network from grid (using LoadNetworkFromGrid as it works better)
	fullNetwork, err := t.State.LoadNetworkFromGrid(ctx, networkName)
	if err != nil {
		return workloads.ZNetLight{}, fmt.Errorf("failed to get network %s: %w", networkName, err)
	}

	// Convert ZNet to ZNetLight
	network := workloads.ZNetLight{
		Name:             fullNetwork.Name,
		Description:      fullNetwork.Description,
		Nodes:            fullNetwork.Nodes,
		IPRange:          fullNetwork.IPRange,
		SolutionType:     fullNetwork.SolutionType,
		MyceliumKeys:     fullNetwork.MyceliumKeys,
		PublicNodeID:     fullNetwork.PublicNodeID,
		NodesIPRange:     fullNetwork.NodesIPRange,
		NodeDeploymentID: fullNetwork.NodeDeploymentID,
	}

	// Verify the network belongs to the expected project
	if network.SolutionType != projectName {
		return workloads.ZNetLight{}, fmt.Errorf("network '%s' found but belongs to project '%s', not '%s'", networkName, network.SolutionType, projectName)
	}

	return network, nil
}

// GetGatewayName gets a gateway name by project name and gateway name
func GetGatewayName(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.GatewayNameProxy, error) {
	contracts, err := t.ContractsGetter.ListContractsOfProjectName(projectName, true)
	if err != nil {
		return workloads.GatewayNameProxy{}, err
	}

	// Find the gateway contract by name
	var nodeID uint32
	var found bool
	for _, contract := range contracts.NodeContracts {
		deploymentData, err := workloads.ParseDeploymentData(contract.DeploymentData)
		if err != nil {
			continue
		}

		// Check if this is the gateway we're looking for
		// SDK uses workloads.GatewayNameType constant ("Gateway Name")
		// Also check "gateway" for backward compatibility with old deployments
		if deploymentData.Name == name && (deploymentData.Type == "gateway" || deploymentData.Type == workloads.GatewayNameType) {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				continue
			}

			// Populate state with only the specific contract we need
			checkIfExistAndAppend(t, contract.NodeID, contractID)
			nodeID = contract.NodeID
			found = true
			break
		}
	}

	if !found {
		return workloads.GatewayNameProxy{}, fmt.Errorf("no Gateway Name with name '%s' found in project '%s'", name, projectName)
	}

	// Now LoadGatewayNameFromGrid will work reliably
	return t.State.LoadGatewayNameFromGrid(ctx, nodeID, name, name)
}

// GetGatewayFQDN gets a gateway FQDN by project name and gateway name
func GetGatewayFQDN(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.GatewayFQDNProxy, error) {
	contracts, err := t.ContractsGetter.ListContractsOfProjectName(projectName, true)
	if err != nil {
		return workloads.GatewayFQDNProxy{}, err
	}

	// Find the gateway contract by name
	var nodeID uint32
	var found bool
	for _, contract := range contracts.NodeContracts {
		deploymentData, err := workloads.ParseDeploymentData(contract.DeploymentData)
		if err != nil {
			continue
		}

		// Check if this is the gateway we're looking for
		// SDK uses workloads.GatewayFQDNType constant ("Gateway Fqdn")
		// Also check "gateway" for backward compatibility with old deployments
		if deploymentData.Name == name && (deploymentData.Type == "gateway" || deploymentData.Type == workloads.GatewayFQDNType) {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				continue
			}

			// Populate state with only the specific contract we need
			checkIfExistAndAppend(t, contract.NodeID, contractID)
			nodeID = contract.NodeID
			found = true
			break
		}
	}

	if !found {
		return workloads.GatewayFQDNProxy{}, fmt.Errorf("no Gateway FQDN with name '%s' found in project '%s'", name, projectName)
	}

	// Now LoadGatewayFQDNFromGrid will work reliably
	return t.State.LoadGatewayFQDNFromGrid(ctx, nodeID, name, name)
}

// GetDeployment gets a deployment by project name and deployment name
func GetDeployment(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.Deployment, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.VMType, name)
	if err != nil {
		return workloads.Deployment{}, err
	}

	if len(nodeContractIDs) == 0 {
		return workloads.Deployment{}, fmt.Errorf("no contracts found for deployment '%s' in project '%s'", name, projectName)
	}

	var nodeID uint32
	for node, contractID := range nodeContractIDs {
		// Populate state with only the specific contract we need
		checkIfExistAndAppend(t, node, contractID)
		nodeID = node
	}

	// Now LoadDeploymentFromGrid will work reliably
	return t.State.LoadDeploymentFromGrid(ctx, nodeID, name)
}
