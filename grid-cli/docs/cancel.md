# Cancel Commands

Cancel resources on the ThreeFold Grid by project name or contract ID.

## Cancel Project

Cancel entire projects or specific VMs within multi-VM projects.

### Cancel Entire Project

```bash
tfcmd cancel --project-name <project-name>
```

**Required Flags:**
- `--project-name`: Project name to cancel

This cancels all resources in the project (all VMs, gateways, and the network).

**Example:**
```console
$ tfcmd cancel --project-name vm/examplevm
3:37PM INF canceling project 'vm/examplevm'
3:37PM INF successfully canceled project 'vm/examplevm'
```

### Cancel Specific VM (Granular Cancellation)

```bash
tfcmd cancel --project-name <project-name> --vm <vm-name>
```

**Required Flags:**
- `--project-name`: Project name containing the VM

**Optional Flags:**
- `--vm`: Specific VM name to cancel (if not specified, cancels entire project)

This removes a specific VM from a multi-VM project while keeping other VMs running. The network is automatically updated to remove the canceled VM's node.

**Example:**
```console
# Cancel vm02 from multi-VM project, keep vm01 and vm03 running
$ tfcmd cancel --project-name myapp --vm vm02
3:40PM INF canceling VM 'vm02' from project 'myapp'
3:40PM INF successfully canceled VM 'vm02'
```

## Cancel Contracts

Cancel twin contracts directly by contract ID.

```bash
tfcmd cancel contracts [flags]
```

**Required Flags:**
- `--contract-id`: Contract ID to cancel

**Example:**
```console
$ tfcmd cancel contracts --contract-id 123456
3:45PM INF canceling contract 123456
3:45PM INF successfully canceled contract 123456
```

## Finding Project Names

To find your existing project names, use:
```bash
tfcmd get contracts
```

For VMs deployed with default project names, use `vm/{vmname}`:
```bash
# Old command
tfcmd get vm myvm

# New command  
tfcmd get vm myvm --project-name vm/myvm
```

## Network Behavior

When canceling VMs from multi-VM projects:
- ✅ The network is automatically updated
- ✅ The canceled VM's node is removed from the network
- ✅ Other VMs in the project continue running
- ✅ Network routes are recalculated

## Use Cases

### Development Cleanup
```bash
# Cancel entire development environment
tfcmd cancel --project-name dev/myapp
```

### Production Maintenance
```bash
# Remove specific faulty VM while keeping others running
tfcmd cancel --project-name production/webapp --vm web03
```

### Contract Management
```bash
# Cancel specific contract by ID
tfcmd cancel contracts --contract-id 987654
```

## Error Handling

**Common errors and solutions:**

- **Project not found**: Check project name with `tfcmd get contracts`
- **VM not in project**: Verify VM name exists in the project
- **Contract already canceled**: Contract may already be canceled
- **Permission denied**: Ensure you own the contract/project

## Breaking Changes

> **Warning**: The cancel command now requires the `--project-name` flag:

- Old: `tfcmd cancel <name>`  
- New: `tfcmd cancel --project-name <project>`

See the migration guide in the VM documentation for details.
