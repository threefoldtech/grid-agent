// Package cmd for handling commands
package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func checkIfExistAndAppend(t deployer.TFPluginClient, node uint32, contractID uint64) {
	for _, n := range t.State.CurrentNodeDeployments[node] {
		if n == contractID {
			return
		}
	}

	t.State.CurrentNodeDeployments[node] = append(t.State.CurrentNodeDeployments[node], contractID)
}

// GetVM gets a VM by project name and VM name
func GetVM(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.Deployment, error) {
	contracts, err := t.ContractsGetter.ListContractsOfProjectName(projectName, true)
	if err != nil {
		return workloads.Deployment{}, err
	}

	if len(contracts.NodeContracts) == 0 {
		return workloads.Deployment{}, fmt.Errorf("no contracts found for VM '%s' in project '%s'", name, projectName)
	}

	var nodeID uint32
	for _, contract := range contracts.NodeContracts {
		contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
		if err != nil {
			return workloads.Deployment{}, err
		}

		nodeID = contract.NodeID
		checkIfExistAndAppend(t, nodeID, contractID)
	}

	return t.State.LoadDeploymentFromGrid(ctx, nodeID, name)
}

// GetK8sCluster gets a Kubernetes cluster by project name and cluster name
func GetK8sCluster(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.K8sCluster, error) {
	contracts, err := t.ContractsGetter.ListContractsOfProjectName(projectName, true)
	if err != nil {
		return workloads.K8sCluster{}, err
	}

	if len(contracts.NodeContracts) == 0 {
		return workloads.K8sCluster{}, fmt.Errorf("no contracts found for cluster '%s' in project '%s'", name, projectName)
	}

	var nodeIDs []uint32
	for _, contract := range contracts.NodeContracts {
		contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
		if err != nil {
			return workloads.K8sCluster{}, err
		}

		checkIfExistAndAppend(t, contract.NodeID, contractID)
		nodeIDs = append(nodeIDs, contract.NodeID)
	}

	return t.State.LoadK8sFromGrid(ctx, nodeIDs, name)
}

// GetNetwork gets a network by project name and network name
func GetNetwork(ctx context.Context, t deployer.TFPluginClient, projectName, networkName string) (workloads.ZNet, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.NetworkType, networkName)
	if err != nil {
		return workloads.ZNet{}, err
	}

	if len(nodeContractIDs) == 0 {
		return workloads.ZNet{}, fmt.Errorf("network '%s' not found in project '%s'", networkName, projectName)
	}

	// Populate state with contract IDs
	for node, contractID := range nodeContractIDs {
		checkIfExistAndAppend(t, node, contractID)
	}

	return t.State.LoadNetworkFromGrid(ctx, networkName)
}

// GetNetworkLight gets a light network by project name and network name
func GetNetworkLight(ctx context.Context, t deployer.TFPluginClient, projectName, networkName string) (workloads.ZNetLight, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.NetworkType, networkName)
	if err != nil {
		return workloads.ZNetLight{}, err
	}

	if len(nodeContractIDs) == 0 {
		return workloads.ZNetLight{}, fmt.Errorf("network '%s' not found in project '%s'", networkName, projectName)
	}

	// Populate state with contract IDs
	for node, contractID := range nodeContractIDs {
		checkIfExistAndAppend(t, node, contractID)
	}

	return t.State.LoadNetworkLightFromGrid(ctx, networkName)
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
		if deploymentData.Name == name && deploymentData.Type == "gateway" {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				continue
			}

			t.State.CurrentNodeDeployments[contract.NodeID] = []uint64{contractID}
			nodeID = contract.NodeID
			found = true
			break
		}
	}

	if !found {
		return workloads.GatewayNameProxy{}, fmt.Errorf("no Gateway Name with name '%s' found in project '%s'", name, projectName)
	}

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
		if deploymentData.Name == name && deploymentData.Type == "gateway" {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				continue
			}

			t.State.CurrentNodeDeployments[contract.NodeID] = []uint64{contractID}
			nodeID = contract.NodeID
			found = true
			break
		}
	}

	if !found {
		return workloads.GatewayFQDNProxy{}, fmt.Errorf("no Gateway FQDN with name '%s' found in project '%s'", name, projectName)
	}

	return t.State.LoadGatewayFQDNFromGrid(ctx, nodeID, name, name)
}

// GetDeployment gets a deployment by project name and deployment name
func GetDeployment(ctx context.Context, t deployer.TFPluginClient, projectName, name string) (workloads.Deployment, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.VMType, name)
	if err != nil {
		return workloads.Deployment{}, err
	}

	var nodeID uint32
	for node, contractID := range nodeContractIDs {
		checkIfExistAndAppend(t, node, contractID)
		nodeID = node
	}

	return t.State.LoadDeploymentFromGrid(ctx, nodeID, name)
}
