# Get Commands

Retrieve information about deployed resources on the ThreeFold Grid.

## Get VM

Retrieve detailed information about a deployed virtual machine.

```bash
tfcmd get vm <vm-name> --project-name <project-name>
```

**Required Flags:**
- `--project-name`: Project name of the VM deployment

**Arguments:**
- `vm-name`: Name of the VM used during deployment

**Example:**
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
                                "SSH_KEY": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDcGrS1RT36rHAGLK3/4FMazGXjIYgWVnZ4bCvxxg8KosEEbs/DeUKT2T2LYV91jUq3yibTWwK0nc6O+K5kdShV4qsQlPmIbdur6x2zWHPeaGXqejbbACEJcQMCj8szSbG8aKwH8Nbi8BNytgzJ20Ysaaj2QpjObCZ4Ncp+89pFahzDEIJx2HjXe6njbp6eCduoA+IE2H9vgwbIDVMQz6y/TzjdQjgbMOJRTlP+CzfbDBb6Ux+ed8F184bMPwkFrpHs9MSfQVbqfIz8wuq/wjewcnb3wK9dmIot6CxV2f2xuOZHgNQmVGratK8TyBnOd5x4oZKLIh3qM9Bi7r81xCkXyxAZbWYu3gGdvo3h85zeCPGK8OEPdYWMmIAIiANE42xPmY9HslPz8PAYq6v0WwdkBlDWrG3DD3GX6qTt9lbSHEgpUP2UOnqGL4O1+g5Rm9x16HWefZWMjJsP6OV70PnMjo9MPnH+yrBkXISw4CGEEXryTvupfaO5sL01mn+UOyE= user@host\n"
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

## Get Gateway

Retrieve information about deployed gateway proxies.

### Get Gateway Name

```bash
tfcmd get gateway name <gateway-name> --project-name <project-name>
```

**Example:**
```console
$ tfcmd get gateway name mygateway --project-name myapp
12:15PM INF starting peer session=tf-1522840 twin=192
12:15PM INF gateway name:
{
        "Name": "mygateway",
        "NodeID": 14,
        "SolutionType": "gateway.name/myapp",
        "NetworkName": "myappnetwork",
        "Gateways": [
                {
                        "name": "mygateway",
                        "node_id": 14,
                        "tls_passthrough": false,
                        "backends": ["http://10.20.2.2:8080"],
                        "fqdn": "mygateway.gent01.dev.grid.tf"
                }
        ]
}
```

### Get Gateway FQDN

```bash
tfcmd get gateway fqdn <gateway-name> --project-name <project-name>
```

**Example:**
```console
$ tfcmd get gateway fqdn myapp --project-name myapp
12:20PM INF starting peer session=tf-1522841 twin=192
12:20PM INF gateway fqdn:
{
        "Name": "myapp",
        "NodeID": 14,
        "SolutionType": "gateway.fqdn/myapp",
        "Gateways": [
                {
                        "name": "myapp",
                        "node_id": 14,
                        "fqdn": "myapp.com",
                        "backends": ["http://10.20.2.2:8080"]
                }
        ]
}
```

## Get Kubernetes

Retrieve information about deployed Kubernetes clusters.

```bash
tfcmd get kubernetes <cluster-name> --project-name <project-name>
```

**Example:**
```console
$ tfcmd get kubernetes mycluster --project-name k8s/mycluster
1:30PM INF starting peer session=tf-1522842 twin=192
1:30PM INF kubernetes cluster:
{
        "Name": "mycluster",
        "SolutionType": "kubernetes/mycluster",
        "Masters": [...],
        "Workers": [...],
        "NetworkName": "myclusternetwork"
}
```

## Get ZDB

Retrieve information about deployed ZDB (Zero-DB) instances.

```bash
tfcmd get zdb <zdb-name> --project-name <project-name>
```

**Example:**
```console
$ tfcmd get zdb mydb --project-name db/mydb
2:45PM INF starting peer session=tf-1522843 twin=192
2:45PM INF zdb:
{
        "Name": "mydb",
        "SolutionType": "zdb/mydb",
        "NodeID": 11,
        "Zdbs": [
                {
                        "name": "mydb",
                        "size": 50,
                        "mode": "user",
                        "password": "your-password",
                        "public": false
                }
        ]
}
```

## Get Contract

Retrieve information about a specific contract by ID.

```bash
tfcmd get contract <contract-id>
```

**Arguments:**
- `contract-id`: Numeric contract ID

**Example:**
```console
$ tfcmd get contract 100063
3:00PM INF starting peer session=tf-1522844 twin=192
3:00PM INF contract:
{
        "ContractID": 100063,
        "TwinID": 192,
        "State": "Created",
        "SolutionType": "vm.examplevm",
        "CreatedAt": "2023-12-17T10:30:00Z"
}
```

## Get Contracts

List all contracts for your twin (user account).

```bash
tfcmd get contracts
```

**Example:**
```console
$ tfcmd get contracts
3:15PM INF starting peer session=tf-1522845 twin=192
3:15PM INF contracts:
Contract ID    Type                State    Created At
100063         vm.examplevm        Created  2023-12-17T10:30:00Z
100064         gateway.name.myapp  Created  2023-12-17T11:15:00Z
100065         kubernetes.mycluster Created  2023-12-17T13:00:00Z
```

## Finding Project Names

Use `get contracts` to discover your existing project names:

```bash
$ tfcmd get contracts
# Look at the SolutionType field to find project names
# vm/examplevm → project name: vm/examplevm
# gateway.name.myapp → project name: myapp
```

## Output Formats

### JSON Output (Default)
All get commands return detailed JSON output with complete deployment information including:
- Node IDs and network configuration
- Resource specifications (CPU, memory, disk)
- IP addresses (public, planetary, mycelium)
- Environment variables and mount points
- Contract IDs and deployment status

### Key Fields
- **Name**: Resource name
- **NodeID**: Node where resource is deployed
- **SolutionType**: Project and resource type
- **NetworkName**: Associated network name
- **ContractID**: Unique contract identifier

## Use Cases

### Monitoring Deployments
```bash
# Check VM status and IP addresses
tfcmd get vm webserver --project-name production/webapp

# Verify gateway configuration
tfcmd get gateway name api-gateway --project-name production/api
```

### Troubleshooting
```bash
# Get detailed deployment information
tfcmd get vm problematic-vm --project-name debug/test

# Check contract status
tfcmd get contract 123456
```

### Inventory Management
```bash
# List all deployed resources
tfcmd get contracts

# Get cluster information
tfcmd get kubernetes production-cluster --project-name k8s/prod
```

## Breaking Changes

> **Warning**: Get commands now require the `--project-name` flag:

- Old: `tfcmd get vm <name>`
- New: `tfcmd get vm <name> --project-name <project>`

Use `tfcmd get contracts` to find your existing project names.
