// Package cmd for handling commands
package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zos_sdk_go/grid-client/deployer"
	"github.com/threefoldtech/zos_sdk_go/grid-client/workloads"
	"github.com/threefoldtech/zos_sdk_go/grid-client/zos"
)

// generateClusterToken generates a random 16-character alphanumeric token for K8s cluster authentication
func generateClusterToken() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 15)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// DeployVM deploys a VM with mounts
func DeployVM(ctx context.Context, t deployer.TFPluginClient, vm workloads.VM, diskMounts []workloads.Disk, volumeMounts []workloads.Volume, projectName, existingNetworkName string) (workloads.VM, error) {
	var networkName string
	var network workloads.ZNet
	var err error

	// Validate: deployment name must be unique within the project
	exists, err := DeploymentExists(t, projectName, vm.Name)
	if err != nil {
		return workloads.VM{}, errors.Wrapf(err, "failed to check if deployment '%s' exists", vm.Name)
	}
	if exists {
		return workloads.VM{}, fmt.Errorf("deployment '%s' already exists in project '%s'", vm.Name, projectName)
	}

	if existingNetworkName != "" {
		// Use existing network
		networkName = existingNetworkName
		log.Info().Msgf("loading existing network '%s'", networkName)

		// Load network using GetNetwork helper
		network, err = GetNetwork(ctx, t, projectName, networkName)
		if err != nil {
			return workloads.VM{}, errors.Wrapf(err, "failed to load existing network '%s'", networkName)
		}
		if !slices.Contains(network.Nodes, vm.NodeID) {
			log.Info().Msgf("adding node %d to network '%s'", vm.NodeID, networkName)
			network.Nodes = append(network.Nodes, vm.NodeID)

			// Add mycelium key if VM uses mycelium
			if len(vm.MyceliumIPSeed) != 0 {
				key, err := workloads.RandomMyceliumKey()
				if err != nil {
					return workloads.VM{}, err
				}
				if network.MyceliumKeys == nil {
					network.MyceliumKeys = make(map[uint32][]byte)
				}
				network.MyceliumKeys[vm.NodeID] = key
			}

			// Redeploy network with new node
			log.Info().Msg("updating network")
			err = t.NetworkDeployer.Deploy(ctx, &network)
			if err != nil {
				return workloads.VM{}, errors.Wrapf(err, "failed to update network with node %d", vm.NodeID)
			}
		}
	} else {
		// Create new network - derive name from project name
		// Strip path separators from project name (e.g., "vm/myapp" -> "myapp")
		baseName := projectName
		if idx := strings.LastIndex(projectName, "/"); idx != -1 {
			baseName = projectName[idx+1:]
		}
		networkName = fmt.Sprintf("%snetwork", baseName)

		// Validate: network name must be unique for this user
		exists, existingProject, err := NetworkExistsForUser(t, networkName)
		if err != nil {
			return workloads.VM{}, errors.Wrapf(err, "failed to check if network '%s' exists", networkName)
		}
		if exists {
			return workloads.VM{}, fmt.Errorf("network '%s' already exists in project '%s'. Use --network flag to join existing network or choose a different project name", networkName, existingProject)
		}

		network, err = buildNetwork(networkName, projectName, []uint32{vm.NodeID}, len(vm.MyceliumIPSeed) != 0)
		if err != nil {
			return workloads.VM{}, err
		}

		log.Info().Msg("deploying network")
		err = t.NetworkDeployer.Deploy(ctx, &network)
		if err != nil {
			return workloads.VM{}, errors.Wrapf(err, "failed to deploy network on node %d", vm.NodeID)
		}
	}

	vm.NetworkName = networkName
	dl := workloads.NewDeployment(vm.Name, vm.NodeID, projectName, nil, networkName, diskMounts, nil, []workloads.VM{vm}, nil, nil, volumeMounts)

	log.Info().Msg("deploying vm")
	err = t.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		// Only remove network if we created it
		if existingNetworkName == "" {
			log.Warn().Msg("error happened while deploying. removing network")
			revertErr := t.NetworkDeployer.Cancel(ctx, &network)
			if revertErr != nil {
				log.Error().Err(revertErr).Msg("failed to remove network")
			}
		}
		return workloads.VM{}, errors.Wrapf(err, "failed to deploy vm on node %d", vm.NodeID)
	}
	resVM, err := t.State.LoadVMFromGrid(ctx, vm.NodeID, vm.Name, dl.Name)
	if err != nil {
		return workloads.VM{}, errors.Wrapf(err, "failed to load vm from node %d", vm.NodeID)
	}
	return resVM, nil
}

// DeployVMLight deploys a VM-light with mounts
func DeployVMLight(ctx context.Context, t deployer.TFPluginClient, vm workloads.VMLight, diskMounts []workloads.Disk, volumeMounts []workloads.Volume, projectName, existingNetworkName string) (workloads.VMLight, error) {
	var networkName string
	var network workloads.ZNetLight
	var err error

	// Validate: deployment name must be unique within the project
	exists, err := DeploymentExists(t, projectName, vm.Name)
	if err != nil {
		return workloads.VMLight{}, errors.Wrapf(err, "failed to check if deployment '%s' exists", vm.Name)
	}
	if exists {
		return workloads.VMLight{}, fmt.Errorf("deployment '%s' already exists in project '%s'", vm.Name, projectName)
	}

	if existingNetworkName != "" {
		// Use existing network
		networkName = existingNetworkName
		log.Info().Msgf("loading existing network '%s'", networkName)

		// Load network using GetNetworkLight helper
		network, err = GetNetworkLight(ctx, t, projectName, networkName)
		if err != nil {
			return workloads.VMLight{}, errors.Wrapf(err, "failed to load existing network '%s'", networkName)
		}

		// Add VM node to network if not already present
		if !slices.Contains(network.Nodes, vm.NodeID) {
			log.Info().Msgf("adding node %d to network '%s'", vm.NodeID, networkName)
			network.Nodes = append(network.Nodes, vm.NodeID)

			// Add mycelium key (light VMs always use mycelium)
			key, err := workloads.RandomMyceliumKey()
			if err != nil {
				return workloads.VMLight{}, err
			}
			if network.MyceliumKeys == nil {
				network.MyceliumKeys = make(map[uint32][]byte)
			}
			network.MyceliumKeys[vm.NodeID] = key

			// Redeploy network with new node
			log.Info().Msg("updating network")
			err = t.NetworkDeployer.Deploy(ctx, &network)
			if err != nil {
				return workloads.VMLight{}, errors.Wrapf(err, "failed to update network with node %d", vm.NodeID)
			}
		}
	} else {
		// Create new network - derive name from project name
		// Strip path separators from project name (e.g., "vm/myapp" -> "myapp")
		baseName := projectName
		if idx := strings.LastIndex(projectName, "/"); idx != -1 {
			baseName = projectName[idx+1:]
		}
		networkName = fmt.Sprintf("%snetwork", baseName)

		// Validate: network name must be unique for this user
		exists, existingProject, err := NetworkExistsForUser(t, networkName)
		if err != nil {
			return workloads.VMLight{}, errors.Wrapf(err, "failed to check if network '%s' exists", networkName)
		}
		if exists {
			return workloads.VMLight{}, fmt.Errorf("network '%s' already exists in project '%s'. Use --network flag to join existing network or choose a different project name", networkName, existingProject)
		}

		network, err = buildNetworkLight(networkName, projectName, []uint32{vm.NodeID})
		if err != nil {
			return workloads.VMLight{}, err
		}

		log.Info().Msg("deploying network")
		err = t.NetworkDeployer.Deploy(ctx, &network)
		if err != nil {
			return workloads.VMLight{}, errors.Wrapf(err, "failed to deploy network on node %d", vm.NodeID)
		}
	}

	vm.NetworkName = networkName
	dl := workloads.NewDeployment(vm.Name, vm.NodeID, projectName, nil, networkName, diskMounts, nil, nil, []workloads.VMLight{vm}, nil, volumeMounts)

	log.Info().Msg("deploying vm")
	err = t.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		// Only remove network if we created it
		if existingNetworkName == "" {
			log.Warn().Msg("error happened while deploying. removing network")
			revertErr := t.NetworkDeployer.Cancel(ctx, &network)
			if revertErr != nil {
				log.Error().Err(revertErr).Msg("failed to remove network")
			}
		}
		return workloads.VMLight{}, errors.Wrapf(err, "failed to deploy vm on node %d", vm.NodeID)
	}

	resVM, err := t.State.LoadVMLightFromGrid(ctx, vm.NodeID, vm.Name, dl.Name)
	if err != nil {
		return workloads.VMLight{}, errors.Wrapf(err, "failed to load vm from node %d", vm.NodeID)
	}

	return resVM, nil
}

// DeployKubernetesCluster deploys a kubernetes cluster
func DeployKubernetesCluster(ctx context.Context, t deployer.TFPluginClient, master workloads.K8sNode, workers []workloads.K8sNode, sshKey, k8sFlist string) (workloads.K8sCluster, error) {
	networkName := fmt.Sprintf("%snetwork", master.Name)
	projectName := fmt.Sprintf("kubernetes/%s", master.Name)
	networkNodes := []uint32{master.NodeID}
	for _, worker := range workers {
		if !slices.Contains(networkNodes, worker.NodeID) {
			networkNodes = append(networkNodes, worker.NodeID)
		}
	}

	network, err := buildNetwork(networkName, projectName, networkNodes, len(master.MyceliumIPSeed) != 0)
	if err != nil {
		return workloads.K8sCluster{}, err
	}

	master.NetworkName = networkName
	for i := range workers {
		workers[i].NetworkName = networkName
	}

	cluster := workloads.K8sCluster{
		Master:       &master,
		Workers:      workers,
		Token:        generateClusterToken(),
		SolutionType: projectName,
		SSHKey:       sshKey,
		Flist:        k8sFlist,
		NetworkName:  networkName,
	}
	log.Info().Msg("deploying network")
	err = t.NetworkDeployer.Deploy(ctx, &network)
	if err != nil {
		return workloads.K8sCluster{}, errors.Wrapf(err, "failed to deploy network on nodes %v", network.Nodes)
	}

	log.Info().Msg("deploying cluster")
	err = t.K8sDeployer.Deploy(ctx, &cluster)
	if err != nil {
		log.Warn().Msg("error happened while deploying. removing network")
		revertErr := t.NetworkDeployer.Cancel(ctx, &network)
		if revertErr != nil {
			log.Error().Err(revertErr).Msg("failed to remove network")
		}
		return workloads.K8sCluster{}, errors.Wrap(err, "failed to deploy kubernetes cluster")
	}
	nodeIDs := []uint32{master.NodeID}
	for _, worker := range workers {
		nodeIDs = append(nodeIDs, worker.NodeID)
	}
	return t.State.LoadK8sFromGrid(
		ctx,
		nodeIDs,
		master.Name,
	)
}

// DeployGatewayName deploys a gateway name
func DeployGatewayName(ctx context.Context, t deployer.TFPluginClient, gateway workloads.GatewayNameProxy) (workloads.GatewayNameProxy, error) {
	// If network is specified, ensure gateway node is part of the network
	if gateway.Network != "" {
		log.Info().Msgf("checking if gateway node %d is part of network '%s'", gateway.NodeID, gateway.Network)

		// Load the network - use SolutionType (project name) to find it
		network, err := GetNetwork(ctx, t, gateway.SolutionType, gateway.Network)
		if err != nil {
			return workloads.GatewayNameProxy{}, errors.Wrapf(err, "failed to load network '%s'", gateway.Network)
		}

		// Check if gateway node is already in the network
		if !slices.Contains(network.Nodes, gateway.NodeID) {
			log.Info().Msgf("extending network '%s' to include gateway node %d", gateway.Network, gateway.NodeID)
			network.Nodes = append(network.Nodes, gateway.NodeID)

			// Add mycelium key for the gateway node
			key, err := workloads.RandomMyceliumKey()
			if err != nil {
				return workloads.GatewayNameProxy{}, err
			}
			if network.MyceliumKeys == nil {
				network.MyceliumKeys = make(map[uint32][]byte)
			}
			network.MyceliumKeys[gateway.NodeID] = key

			// Redeploy network with gateway node
			log.Info().Msg("updating network")
			err = t.NetworkDeployer.Deploy(ctx, &network)
			if err != nil {
				return workloads.GatewayNameProxy{}, errors.Wrapf(err, "failed to extend network to gateway node %d", gateway.NodeID)
			}
		}
	}

	log.Info().Msg("deploying gateway name")
	err := t.GatewayNameDeployer.Deploy(ctx, &gateway)
	if err != nil {
		return workloads.GatewayNameProxy{}, errors.Wrapf(err, "failed to deploy gateway on node %d", gateway.NodeID)
	}

	return t.State.LoadGatewayNameFromGrid(ctx, gateway.NodeID, gateway.Name, gateway.Name)
}

// DeployGatewayFQDN deploys a gateway fqdn
func DeployGatewayFQDN(ctx context.Context, t deployer.TFPluginClient, gateway workloads.GatewayFQDNProxy) error {
	// If network is specified, ensure gateway node is part of the network
	if gateway.Network != "" {
		log.Info().Msgf("checking if gateway node %d is part of network '%s'", gateway.NodeID, gateway.Network)

		// Load the network - use SolutionType (project name) to find it
		network, err := GetNetwork(ctx, t, gateway.SolutionType, gateway.Network)
		if err != nil {
			return errors.Wrapf(err, "failed to load network '%s'", gateway.Network)
		}

		// Check if gateway node is already in the network
		if !slices.Contains(network.Nodes, gateway.NodeID) {
			log.Info().Msgf("extending network '%s' to include gateway node %d", gateway.Network, gateway.NodeID)
			network.Nodes = append(network.Nodes, gateway.NodeID)

			// Add mycelium key for the gateway node
			key, err := workloads.RandomMyceliumKey()
			if err != nil {
				return err
			}
			if network.MyceliumKeys == nil {
				network.MyceliumKeys = make(map[uint32][]byte)
			}
			network.MyceliumKeys[gateway.NodeID] = key

			// Redeploy network with gateway node
			log.Info().Msg("updating network")
			err = t.NetworkDeployer.Deploy(ctx, &network)
			if err != nil {
				return errors.Wrapf(err, "failed to extend network to gateway node %d", gateway.NodeID)
			}
		}
	}

	log.Info().Msg("deploying gateway fqdn")
	err := t.GatewayFQDNDeployer.Deploy(ctx, &gateway)
	if err != nil {
		return errors.Wrapf(err, "failed to deploy gateway on node %d", gateway.NodeID)
	}
	return nil
}

// DeployZDBs deploys multiple zdbs
func DeployZDBs(ctx context.Context, t deployer.TFPluginClient, projectName string, zdbs []workloads.ZDB, n int, node uint32) ([]workloads.ZDB, error) {
	dl := workloads.NewDeployment(projectName, node, projectName, nil, "", nil, zdbs, nil, nil, nil, nil)
	log.Info().Msgf("deploying zdbs")
	err := t.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		return []workloads.ZDB{}, errors.Wrapf(err, "failed to deploy zdbs on node %d", node)
	}

	var resZDBs []workloads.ZDB
	for _, zdb := range zdbs {
		resZDB, err := t.State.LoadZdbFromGrid(ctx, node, zdb.Name, dl.Name)
		if err != nil {
			return []workloads.ZDB{}, errors.Wrapf(err, "failed to load zdb '%s' from node %d", zdb.Name, node)
		}

		resZDBs = append(resZDBs, resZDB)
	}

	return resZDBs, nil
}

func buildNetwork(name, projectName string, nodes []uint32, addMycelium bool) (workloads.ZNet, error) {
	keys := make(map[uint32][]byte)
	if addMycelium {
		for _, node := range nodes {
			key, err := workloads.RandomMyceliumKey()
			if err != nil {
				return workloads.ZNet{}, err
			}
			keys[node] = key
		}
	}
	return workloads.ZNet{
		Name:  name,
		Nodes: nodes,
		IPRange: zos.IPNet{IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}},
		MyceliumKeys: keys,
		SolutionType: projectName,
	}, nil
}

func buildNetworkLight(name, projectName string, nodes []uint32) (workloads.ZNetLight, error) {
	keys := make(map[uint32][]byte)
	for _, node := range nodes {
		key, err := workloads.RandomMyceliumKey()
		if err != nil {
			return workloads.ZNetLight{}, err
		}
		keys[node] = key
	}

	return workloads.ZNetLight{
		Name:  name,
		Nodes: nodes,
		IPRange: zos.IPNet{IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}},
		MyceliumKeys: keys,
		SolutionType: projectName,
	}, nil
}
