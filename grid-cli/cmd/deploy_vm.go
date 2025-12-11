// Package cmd for parsing command line arguments
package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	command "github.com/threefoldtech/grid-agent/grid-cli/internal/cmd"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/config"
	"github.com/threefoldtech/grid-agent/grid-cli/internal/filters"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	client "github.com/threefoldtech/tfgrid-sdk-go/grid-client/node"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/subi"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
)

var (
	ubuntuFlist           = "https://hub.grid.tf/tf-official-apps/threefoldtech-ubuntu-22.04.flist"
	ubuntuFlistEntrypoint = "/sbin/zinit init"
)

// DiskSpec represents a disk or volume specification with size and mount point
type DiskSpec struct {
	SizeGB     uint64
	MountPoint string
}

func convertGPUsToZosGPUs(gpus []string) (zosGPUs []zos.GPU) {
	for _, g := range gpus {
		zosGPUs = append(zosGPUs, zos.GPU(g))
	}
	return
}

// parseDiskSpecs parses disk/volume specifications in format "size:mountpoint" or just "size"
// For backward compatibility, if only size is provided, defaultMountPoint is used
func parseDiskSpecs(specs []string, defaultMountPoint string) ([]DiskSpec, error) {
	if len(specs) == 0 {
		return []DiskSpec{}, nil
	}

	result := make([]DiskSpec, 0, len(specs))
	mountCounter := 0

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		parts := strings.SplitN(spec, ":", 2)

		// Parse size
		sizeGB, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid disk size '%s': %w", parts[0], err)
		}

		if sizeGB == 0 {
			continue // Skip zero-sized disks
		}

		// Determine mount point
		mountPoint := defaultMountPoint
		if len(parts) == 2 {
			mountPoint = strings.TrimSpace(parts[1])
			if mountPoint == "" {
				return nil, fmt.Errorf("mount point cannot be empty in spec '%s'", spec)
			}
			if !strings.HasPrefix(mountPoint, "/") {
				return nil, fmt.Errorf("mount point must be an absolute path (start with /): '%s'", mountPoint)
			}
		} else if mountCounter > 0 {
			// For multiple disks without explicit mount points, append counter
			mountPoint = fmt.Sprintf("%s%d", defaultMountPoint, mountCounter)
		}

		result = append(result, DiskSpec{
			SizeGB:     sizeGB,
			MountPoint: mountPoint,
		})
		mountCounter++
	}

	return result, nil
}

// deployVMCmd represents the deploy vm command
var deployVMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy a vm",
	Long: `Deploy a virtual machine to the ThreeFold Grid.

Supports single VM deployment or multi-VM deployments with shared networks.
Use --project-name to organize VMs into projects and --network to deploy
multiple VMs on the same network.

**Disks vs Volumes**:
- Disks: Local SSD storage on the node (faster, not shared between VMs)
- Volumes: Distributed QSFS storage (slower, can be shared across VMs)

Use disks for: databases, caches, temporary files
Use volumes for: shared data, backups, large files

**Network Limitations**:
IMPORTANT: Networks can only span nodes within the same farm. Nodes in different
farms cannot share a network. For cross-farm deployments, use separate networks
or planetary network IPs for inter-VM communication.

Examples:
  # Deploy single VM with default settings
  tfcmd deploy vm --name myvm --ssh ~/.ssh/id_rsa.pub
  
  # Deploy VM with custom resources and project name
  tfcmd deploy vm --name webserver --ssh ~/.ssh/id_rsa.pub --cpu 4 --memory 8 --project-name production
  
  # Deploy multiple VMs on shared network (same farm)
  tfcmd deploy vm --name vm01 --ssh ~/.ssh/id_rsa.pub --project-name myapp
  tfcmd deploy vm --name vm02 --ssh ~/.ssh/id_rsa.pub --network myappnetwork --project-name myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		env, err := cmd.Flags().GetStringToString("env")
		if err != nil {
			return err
		}
		sshFile, err := cmd.Flags().GetString("ssh")
		if err != nil {
			return err
		}
		sshKey, err := os.ReadFile(sshFile)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		env["SSH_KEY"] = string(sshKey)
		node, err := cmd.Flags().GetUint32("node")
		if err != nil {
			return err
		}
		farm, err := cmd.Flags().GetUint64("farm")
		if err != nil {
			return err
		}
		cpu, err := cmd.Flags().GetUint8("cpu")
		if err != nil {
			return err
		}
		memory, err := cmd.Flags().GetUint64("memory")
		if err != nil {
			return err
		}
		rootfs, err := cmd.Flags().GetUint64("rootfs")
		if err != nil {
			return err
		}
		diskSpecs, err := cmd.Flags().GetStringSlice("disk")
		if err != nil {
			return err
		}
		volumeSpecs, err := cmd.Flags().GetStringSlice("volume")
		if err != nil {
			return err
		}

		// Parse disk and volume specifications
		disks, err := parseDiskSpecs(diskSpecs, "/data")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse disk specifications")
		}
		volumes, err := parseDiskSpecs(volumeSpecs, "/volume")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse volume specifications")
		}

		flist, err := cmd.Flags().GetString("flist")
		if err != nil {
			return err
		}
		entrypoint, err := cmd.Flags().GetString("entrypoint")
		if err != nil {
			return err
		}
		gpus, err := cmd.Flags().GetStringSlice("gpus")
		if err != nil {
			return err
		}
		if len(gpus) > 0 && node == 0 {
			log.Fatal().Msg("must specify node ID when using GPUs")
		}

		ipv4, err := cmd.Flags().GetBool("ipv4")
		if err != nil {
			return err
		}
		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}
		ygg, err := cmd.Flags().GetBool("ygg")
		if err != nil {
			return err
		}
		mycelium, err := cmd.Flags().GetBool("mycelium")
		if err != nil {
			return err
		}
		noColor, err := cmd.Parent().Flags().GetBool("no-color")
		if err != nil {
			return err
		}
		disableSentry, err := cmd.Parent().Flags().GetBool("disable-sentry")
		if err != nil {
			return err
		}

		var seed []byte
		if mycelium {
			seed, err = workloads.RandomMyceliumIPSeed()
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		}

		networkName, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		projectName, err := cmd.Flags().GetString("project-name")
		if err != nil {
			return err
		}

		// Validate: --node and --farm are mutually exclusive
		if node != 0 && farm != 0 {
			return fmt.Errorf("--node and --farm flags are mutually exclusive")
		}

		// Validate: --gpus requires --node
		if len(gpus) > 0 && node == 0 {
			return fmt.Errorf("--gpus requires --node flag to specify which node has the GPU")
		}

		// Validate: --network requires --project-name
		if networkName != "" && projectName == "" {
			return fmt.Errorf("--project-name is required when using --network")
		}

		// Default project name if not provided
		if projectName == "" {
			projectName = fmt.Sprintf("vm/%s", name)
		}

		cfg, err := config.GetUserConfig()
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		opts := []deployer.PluginOpt{
			deployer.WithNetwork(cfg.Network),
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

		// if no public ips or yggdrasil then we should go for the light deployment
		isLight := !ipv4 && !ipv6 && !ygg

		if node != 0 {
			isLight, err = isZos4Node(cmd.Context(), t.NcPool, t.SubstrateConn, node)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		}

		if isLight {
			vm := workloads.VMLight{
				Name:           name,
				EnvVars:        env,
				CPU:            cpu,
				MemoryMB:       memory * 1024,
				GPUs:           convertGPUsToZosGPUs(gpus),
				RootfsSizeMB:   rootfs * 1024,
				Flist:          flist,
				Entrypoint:     entrypoint,
				MyceliumIPSeed: seed,
			}
			err = executeVMLight(cmd.Context(), t, vm, node, farm, disks, volumes, projectName, networkName)
			if err == nil {
				return nil
			}

			if !errors.Is(err, deployer.ErrNoNodesMatchesResources) {
				log.Fatal().Err(err).Send()
			}
		}

		vm := workloads.VM{
			Name:           name,
			EnvVars:        env,
			CPU:            cpu,
			MemoryMB:       memory * 1024,
			GPUs:           convertGPUsToZosGPUs(gpus),
			RootfsSizeMB:   rootfs * 1024,
			Flist:          flist,
			Entrypoint:     entrypoint,
			PublicIP:       ipv4,
			PublicIP6:      ipv6,
			MyceliumIPSeed: seed,
			Planetary:      ygg,
		}
		err = executeVM(cmd.Context(), t, vm, node, farm, disks, volumes, projectName, networkName)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		return nil
	},
}

func init() {
	deployCmd.AddCommand(deployVMCmd)

	deployVMCmd.Flags().StringP("name", "n", "", "name of the virtual machine")
	err := deployVMCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	deployVMCmd.Flags().StringP("ssh", "s", "", "path to public SSH key file (e.g., ~/.ssh/id_rsa.pub)")
	// should it be required?
	err = deployVMCmd.MarkFlagRequired("ssh")
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	deployVMCmd.Flags().Uint32("node", 0, "node id vm should be deployed on")
	deployVMCmd.Flags().Uint64("farm", 0, "farm ID for deployment (0 = any farm, or specify farm ID)")
	deployVMCmd.MarkFlagsMutuallyExclusive("node", "farm")
	deployVMCmd.Flags().String("network", "", "name of existing network to deploy VM on. If not specified, a new network will be created default to '{vmname}network'")
	deployVMCmd.Flags().String("project-name", "", "project name for the VM deployment. Defaults to 'vm/{vmname}' if not specified. Required when using --network")

	deployVMCmd.Flags().Uint8("cpu", 1, "number of cpu units")
	deployVMCmd.Flags().Uint64("memory", 1, "memory size in GB (e.g., --memory 2 for 2GB RAM)")
	deployVMCmd.Flags().Uint64("rootfs", 2, "root filesystem size in gb")
	deployVMCmd.Flags().StringSlice("disk", []string{}, "disk specification in format 'size:mountpoint' (e.g., '10:/data'). Can be specified multiple times. For backward compatibility, just 'size' defaults to '/data'")
	deployVMCmd.Flags().String("flist", ubuntuFlist, "flist for vm")
	deployVMCmd.Flags().StringSlice("gpus", []string{}, "gpus for vm")
	deployVMCmd.Flags().StringSlice("volume", []string{}, "volume specification in format 'size:mountpoint' (e.g., '50:/shared'). Can be specified multiple times. For backward compatibility, just 'size' defaults to '/volume'")

	deployVMCmd.Flags().String("entrypoint", ubuntuFlistEntrypoint, "entrypoint for vm")
	// to ensure entrypoint is provided for custom flist
	deployVMCmd.MarkFlagsRequiredTogether("flist", "entrypoint")

	deployVMCmd.Flags().Bool("ipv4", false, "assign public ipv4 for vm")
	deployVMCmd.Flags().Bool("ipv6", false, "assign public ipv6 for vm")
	deployVMCmd.Flags().Bool("ygg", true, "assign planetary network IP (Yggdrasil) for VM")
	deployVMCmd.Flags().Bool("mycelium", true, "assign mycelium ip for vm")
	deployVMCmd.Flags().StringToStringP("env", "e", make(map[string]string), "environment variables for the vm")
}

func executeVM(
	ctx context.Context, t deployer.TFPluginClient,
	vm workloads.VM,
	node uint32,
	farm uint64, diskSpecs, volumeSpecs []DiskSpec, projectName, networkName string,
) error {
	// Build disk mounts from specifications
	diskMounts := make([]workloads.Disk, 0, len(diskSpecs))
	for i, spec := range diskSpecs {
		diskName := fmt.Sprintf("%sdisk%d", vm.Name, i)
		diskMounts = append(diskMounts, workloads.Disk{
			Name:   diskName,
			SizeGB: spec.SizeGB,
		})
		vm.Mounts = append(vm.Mounts, workloads.Mount{
			Name:       diskName,
			MountPoint: spec.MountPoint,
		})
	}

	// Build volume mounts from specifications
	volumeMounts := make([]workloads.Volume, 0, len(volumeSpecs))
	for i, spec := range volumeSpecs {
		volumeName := fmt.Sprintf("%svolume%d", vm.Name, i)
		volumeMounts = append(volumeMounts, workloads.Volume{
			Name:   volumeName,
			SizeGB: spec.SizeGB,
		})
		vm.Mounts = append(vm.Mounts, workloads.Mount{
			Name:       volumeName,
			MountPoint: spec.MountPoint,
		})
	}

	if node == 0 {
		// Calculate total disk and volume sizes for node filtering
		var totalDiskSize, totalVolumeSize uint64
		for _, spec := range diskSpecs {
			totalDiskSize += spec.SizeGB
		}
		for _, spec := range volumeSpecs {
			totalVolumeSize += spec.SizeGB
		}

		// Use a single disk/volume for filtering (backward compatible)
		diskForFilter := workloads.Disk{SizeGB: totalDiskSize}
		volumeForFilter := workloads.Volume{SizeGB: totalVolumeSize}

		filter, ssd, rootfss := filters.BuildVMFilter(diskForFilter, volumeForFilter, farm, vm.MemoryMB, vm.RootfsSizeMB, vm.PublicIP, false)
		nodes, err := deployer.FilterNodes(
			ctx,
			t,
			filter,
			ssd,
			nil,
			rootfss,
		)
		if err != nil {
			return err
		}

		node = uint32(nodes[0].NodeID)
	}

	vm.NodeID = node
	resVM, err := command.DeployVM(ctx, t, vm, diskMounts, volumeMounts, projectName, networkName)
	if err != nil {
		return err
	}

	if vm.PublicIP {
		log.Info().Msgf("vm ipv4: %s", resVM.ComputedIP)
	}
	if vm.PublicIP6 {
		log.Info().Msgf("vm ipv6: %s", resVM.ComputedIP6)
	}
	if vm.Planetary {
		log.Info().Msgf("vm planetary ip: %s", resVM.PlanetaryIP)
	}
	if len(resVM.MyceliumIP) != 0 {
		log.Info().Msgf("vm mycelium ip: %s", resVM.MyceliumIP)
	}

	return nil
}

func executeVMLight(
	ctx context.Context, t deployer.TFPluginClient,
	vm workloads.VMLight,
	node uint32,
	farm uint64, diskSpecs, volumeSpecs []DiskSpec, projectName, networkName string,
) error {
	// Build disk mounts from specifications
	diskMounts := make([]workloads.Disk, 0, len(diskSpecs))
	for i, spec := range diskSpecs {
		diskName := fmt.Sprintf("%sdisk%d", vm.Name, i)
		diskMounts = append(diskMounts, workloads.Disk{
			Name:   diskName,
			SizeGB: spec.SizeGB,
		})
		vm.Mounts = append(vm.Mounts, workloads.Mount{
			Name:       diskName,
			MountPoint: spec.MountPoint,
		})
	}

	// Build volume mounts from specifications
	volumeMounts := make([]workloads.Volume, 0, len(volumeSpecs))
	for i, spec := range volumeSpecs {
		volumeName := fmt.Sprintf("%svolume%d", vm.Name, i)
		volumeMounts = append(volumeMounts, workloads.Volume{
			Name:   volumeName,
			SizeGB: spec.SizeGB,
		})
		vm.Mounts = append(vm.Mounts, workloads.Mount{
			Name:       volumeName,
			MountPoint: spec.MountPoint,
		})
	}

	if node == 0 {
		// Calculate total disk and volume sizes for node filtering
		var totalDiskSize, totalVolumeSize uint64
		for _, spec := range diskSpecs {
			totalDiskSize += spec.SizeGB
		}
		for _, spec := range volumeSpecs {
			totalVolumeSize += spec.SizeGB
		}

		// Use a single disk/volume for filtering (backward compatible)
		diskForFilter := workloads.Disk{SizeGB: totalDiskSize}
		volumeForFilter := workloads.Volume{SizeGB: totalVolumeSize}

		filter, ssd, rootfss := filters.BuildVMFilter(diskForFilter, volumeForFilter, farm, vm.MemoryMB, vm.RootfsSizeMB, false, true)
		nodes, err := deployer.FilterNodes(
			ctx,
			t,
			filter,
			ssd,
			nil,
			rootfss,
		)
		if err != nil {
			return err
		}

		node = uint32(nodes[0].NodeID)
	}

	vm.NodeID = node
	resVM, err := command.DeployVMLight(ctx, t, vm, diskMounts, volumeMounts, projectName, networkName)
	if err != nil {
		return err
	}

	if len(resVM.MyceliumIP) != 0 {
		log.Info().Msgf("vm mycelium ip: %s", resVM.MyceliumIP)
	}

	return nil
}

func isZos4Node(ctx context.Context, client client.NodeClientGetter, sub subi.SubstrateExt, node uint32) (isLight bool, err error) {
	cli, err := client.GetNodeClient(sub, node)
	if err != nil {
		return
	}
	feat, err := cli.SystemGetNodeFeatures(ctx)
	if err != nil {
		return
	}

	return slices.Contains(feat, zos.NetworkLightType), nil
}
