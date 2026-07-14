// Package cmd for handling commands
package cmd

import (
	"strings"

	"github.com/threefoldtech/zos_sdk_go/grid-client/deployer"
	"github.com/threefoldtech/zos_sdk_go/grid-client/workloads"
)

// DeploymentExists checks if a deployment with the given name already exists in the project
func DeploymentExists(t deployer.TFPluginClient, projectName, deploymentName string) (bool, error) {
	nodeContractIDs, err := t.ContractsGetter.GetNodeContractsByTypeAndName(projectName, workloads.VMType, deploymentName)
	if err != nil {
		// "could not find any contracts" means no deployment exists - this is not an error for our purpose
		if strings.Contains(err.Error(), "could not find any contracts") {
			return false, nil
		}
		return false, err
	}
	return len(nodeContractIDs) > 0, nil
}

// NetworkExistsForUser checks if a network with the given name already exists for this user (any project)
// Returns: exists, projectName (if found), error
func NetworkExistsForUser(t deployer.TFPluginClient, networkName string) (bool, string, error) {
	contracts, err := t.ContractsGetter.ListContractsByTwinID([]string{"Created"})
	if err != nil {
		// No contracts at all means network doesn't exist
		if strings.Contains(err.Error(), "could not find any contracts") {
			return false, "", nil
		}
		return false, "", err
	}

	for _, contract := range contracts.NodeContracts {
		deploymentData, err := workloads.ParseDeploymentData(contract.DeploymentData)
		if err != nil {
			continue
		}

		// Check if this is a network with the same name
		if deploymentData.Type == workloads.NetworkType && deploymentData.Name == networkName {
			return true, deploymentData.ProjectName, nil
		}
	}

	return false, "", nil
}
