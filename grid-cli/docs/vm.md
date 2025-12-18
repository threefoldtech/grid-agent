# Virtual Machine

This document explains Virtual Machine related commands using tfcmd.

## Deploy

```bash
tfcmd deploy vm [flags]
```

### Required Flags

- name: name for the VM deployment also used for canceling the deployment. must be unique.
- ssh: path to public ssh key to set in the VM.

### Optional Flags

- node: node id vm should be deployed on.
- farm: farm id vm should be deployed on, if set choose available node from farm that fits vm specs (default 1). note: node and farm flags cannot be set both.
- project-name: project name for the VM deployment. Defaults to 'vm/{vmname}' if not specified. Required when using --network flag.
- network: existing network name to deploy VM on. If not specified, a new network will be created. If the VM's node is not part of the network, it will be automatically added. Note: --project-name is required when using this flag.
- cpu: number of cpu units (default 1).
- disk: disk specification in format 'size:mountpoint' (e.g., '10:/data'). Can be specified multiple times for multiple disks. For backward compatibility, just 'size' defaults to '/data'. If multiple disks are specified without mount points, they will be mounted at /data, /data1, /data2, etc.
- volume: volume specification in format 'size:mountpoint' (e.g., '50:/shared'). Can be specified multiple times for multiple volumes. For backward compatibility, just 'size' defaults to '/volume'. If multiple volumes are specified without mount points, they will be mounted at /volume, /volume1, /volume2, etc.
- entrypoint: entrypoint for VM flist (default "/sbin/zinit init"). note: setting this without the flist option will fail.
- flist: flist used in VM (default "<https://hub.grid.tf/tf-official-apps/threefoldtech-ubuntu-22.04.flist>"). note: setting this without the entrypoint option will fail.
- ipv4: assign public ipv4 for VM (default false).
- ipv6: assign public ipv6 for VM (default false).
- memory: memory size in GB (default 1).
- rootfs: root filesystem size in GB (default 2).
- ygg: assign yggdrasil ip for VM (default true).
- mycelium: assign mycelium ip for VM (default true).
- gpus: assign a list of gpus' ids to the VM. note: setting this without the node option will fail.
- env: environment variables for the VM.

### Project Names and Networks

When deploying a VM:

- If `--project-name` is not specified, it defaults to `vm/{vmname}`
- The network name is derived from the project name (e.g., project `vm/myapp` creates network `myappnetwork`)
- Multiple VMs can share the same project and network by using the same `--project-name` and `--network`

### Network Limitations

**IMPORTANT**: Networks can only span nodes within the same farm.

- ✅ Nodes 11 and 14 (both in Farm 1) can share a network
- ❌ Node 327 (Farm 2) cannot join Farm 1's network

For cross-farm deployments:

- Use separate networks per farm
- Use planetary network IPs for inter-VM communication

### Disks vs Volumes

- **Disks (zmount)**: Legacy sparse files on host filesystem (slower, but considered more reliable and stable)
- **Volumes**: Modern btrfs subvolumes with quota (faster, future of storage, more experimental)

**Key Advantages of Volumes:**
- ✅ **Better Performance**: Direct subvolume access vs sparse files
- ✅ **Snapshots**: Backups without interrupting the VM
- ✅ **Live Resize**: Grow/shrink while VM is running
- ✅ **Better Caching**: Optimized storage utilization
- ✅ **Future-Proof**: Will migrate to bcachefs

**Migration Path:**
```
Past:    zmount (sparse files) ✅ Stable & Reliable
Present: btrfs subvolumes ⚠️ Experimental  
Future:  bcachefs ⏳ Coming Soon
```

**Recommendation:**

- Use **disks** for all new deployments (more reliable and stable)
- Use **volumes** only if you explicitly need their advanced features

> **Note**: QSFS is a separate distributed storage technology and is NOT related to volumes. Volumes are local storage.

Example:

- Deploying VM without GPU

```console
$ tfcmd deploy vm --name examplevm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 --disk 10
12:06PM INF starting peer session=tf-1508255 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:7da2:ac99:99db:8821
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:6b18:921f:8031
```

- Deploying VM with GPU

```console
$ tfcmd deploy vm --name examplevm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 --disk 10 --gpus '0000:0e:00.0/1882/543f' --gpus '0000:0e:00.0/1887/593f' --node 12
12:06PM INF starting peer session=tf-1508255 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:7da2:ac99:99db:8821
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:6b18:921f:8031
```

- Deploying VM with multiple disks and custom mount points

```console
$ tfcmd deploy vm --name examplevm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --disk 10:/data \
  --disk 20:/mnt/storage \
  --disk 50:/var/lib/docker
12:06PM INF starting peer session=tf-1508255 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:7da2:ac99:99db:8821
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:6b18:921f:8031
```

- Deploying VM with mixed disks and volumes

```console
$ tfcmd deploy vm --name examplevm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --disk 10:/data \
  --volume 100:/shared
12:06PM INF starting peer session=tf-1508255 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:7da2:ac99:99db:8821
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:6b18:921f:8031
```

- Backward compatible: deploying VM with single disk (old format)

```console
$ tfcmd deploy vm --name examplevm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 --disk 10
12:06PM INF starting peer session=tf-1508255 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:7da2:ac99:99db:8821
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:6b18:921f:8031
```

- Deploy VM on existing network

```console
$ tfcmd deploy vm --name vm2 --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --network vm1network --node 14
12:06PM INF starting peer session=tf-1508256 twin=192
12:06PM INF loading existing network 'vm1network'
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:8da3:bd10:10ec:9932
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:7c29:a32g:9142
```

- Deploy VM with custom project name

```console
$ tfcmd deploy vm --name myvm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --project-name mycoolproject
12:06PM INF starting peer session=tf-1508257 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm
12:07PM INF vm planetary ip: 300:e9c4:9048:57cf:7da2:ac99:99db:8821
12:07PM INF vm mycelium ip: 544:b74f:ceef:cc7e:ff0f:6b18:921f:8031
```

Note: This creates project `mycoolproject` with network `mycoolprojectnetwork`.

- Deploy multiple VMs in the same project (multi-VM deployment)

```console
# Deploy first VM
$ tfcmd deploy vm --name vm01 --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --project-name myapp
12:06PM INF starting peer session=tf-1508258 twin=192
12:06PM INF deploying network
12:06PM INF deploying vm

# Deploy second VM to same project/network
$ tfcmd deploy vm --name vm02 --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --network myappnetwork --project-name myapp
12:08PM INF starting peer session=tf-1508259 twin=192
12:08PM INF loading existing network 'myappnetwork'
12:08PM INF adding node 15 to network 'myappnetwork'
12:08PM INF updating network
12:08PM INF deploying vm

# Deploy third VM to same project/network
$ tfcmd deploy vm --name vm03 --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4 \
  --network myappnetwork --project-name myapp
12:10PM INF starting peer session=tf-1508260 twin=192
12:10PM INF loading existing network 'myappnetwork'
12:10PM INF adding node 16 to network 'myappnetwork'
12:10PM INF updating network
12:10PM INF deploying vm
```

Note: All three VMs share the same project (`myapp`) and network (`myappnetwork`).

## Get

```bash
tfcmd get vm <vm> --project-name <project-name>
```

**Required Flags:**
- project-name: project name of the VM deployment

vm is the name used when deploying vm using tfcmd.

Example:

```console
$ tfcmd get vm examplevm --project-name vm/examplevm
11:56AM INF starting peer session=tf-1522837 twin=192
11:56AM INF vm:
{
        "Name": "examplevm",
        "NodeID": 11,
        "SolutionType": "vm/examplevm",
        "SolutionProvider": null,
        "NetworkName": "examplevmnetwork",
        "Disks": [
                {
                        "name": "examplevmdisk",
                        "size": 10,
                        "description": ""
                }
        ],
        "Zdbs": [],
        "Vms": [
                {
                        "name": "examplevm",
                        "flist": "https://hub.grid.tf/tf-official-apps/threefoldtech-ubuntu-22.04.flist",
                        "flist_checksum": "",
                        "publicip": false,
                        "publicip6": false,
                        "planetary": true,
                        "corex": false,
                        "computedip": "",
                        "computedip6": "",
                        "planetary_ip": "302:9e63:7d43:b742:c2e0:ab69:e101:8032",
                        "mycelium_ip": "45f:6cd6:8c6d:6c73:ff0f:6c97:11e3:4e84",
                        "ip": "10.20.2.2",
                        "mycelium_ip_seed": "bJcR406E",
                        "description": "",
                        "gpus": null,
                        "cpu": 2,
                        "memory": 4096,
                        "rootfs_size": 2048,
                        "entrypoint": "/sbin/zinit init",
                        "mounts": [
                                {
                                        "disk_name": "examplevmdisk",
                                        "mount_point": "/data"
                                }
                        ],
                        "zlogs": null,
                        "env_vars": {
                                "SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDcGrS1RT36rHAGLK3/4FMazGXjIYgWVnZ4bCvxxg8KosEEbs/DeUKT2T2LYV91jUq3yibTWwK0nc6O+K5kdShV4qsQlPmIbdur6x2zWHPeaGXqejbbACEJcQMCj8szSbG8aKwH8Nbi8BNytgzJ20Ysaaj2QpjObCZ4Ncp+89pFahzDEIJx2HjXe6njbp6eCduoA+IE2H9vgwbIDVMQz6y/TzjdQjgbMOJRTlP+CzfbDBb6Ux+ed8F184bMPwkFrpHs9MSfQVbqfIz8wuq/wjewcnb3wK9dmIot6CxV2f2xuOZHgNQmVGratK8TyBnOd5x4oZKLIh3qM9Bi7r81xCkXyxAZbWYu3gGdvo3h85zeCPGK8OEPdYWMmIAIiANE42xPmY9HslPz8PAYq6v0WwdkBlDWrG3DD3GX6qTt9lbSHEgpUP2UOnqGL4O1+g5Rm9x16HWefZWMjJsP6OV70PnMjo9MPnH+yrBkXISw4CGEEXryTvupfaO5sL01mn+UOyE= abdulrahman@AElawady-PC\n"
                        },
                        "network_name": "examplevmnetwork",
                        "console_url": "10.20.2.0:20002"
                }
        ],
        "QSFS": [],
        "NodeDeploymentID": {
                "11": 100063
        },
        "ContractID": 100063,
        "IPrange": "10.20.2.0/24"
}
```

## Cancel

### Cancel Entire Project

```bash
tfcmd cancel --project-name <project-name>
```

**Required Flags:**
- project-name: project name to cancel

This cancels all resources in the project (all VMs and the network).

Example:

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
- project-name: project name containing the VM

**Optional Flags:**
- vm: specific VM name to cancel (if not specified, cancels entire project)

This removes a specific VM from a multi-VM project while keeping other VMs running. The network is automatically updated to remove the canceled VM's node.

Example:

```console
# Cancel vm02 from multi-VM project, keep vm01 and vm03 running
$ tfcmd cancel --project-name myapp --vm vm02
3:40PM INF canceling VM 'vm02' from project 'myapp'
3:40PM INF successfully canceled VM 'vm02'
```

Note: The network `myappnetwork` is automatically updated to remove vm02's node. VMs vm01 and vm03 continue running.

## Breaking Changes

> **Warning**: The following commands now require the `--project-name` flag:

- `tfcmd get vm <name>` → `tfcmd get vm <name> --project-name <project>`
- `tfcmd cancel <name>` → `tfcmd cancel --project-name <project>`

### Migration Guide

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
