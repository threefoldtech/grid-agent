# Gateway Name

This document explains Gateway Name related commands using tfcmd.

## Deploy

```bash
tfcmd deploy gateway name [flags]
```

### Required Flags

- name: name for the gateway deployment also used for canceling the deployment. must be unique.
- backends: list of backends the gateway will forward requests to.

### Optional Flags

- node: node id gateway should be deployed on.
- farm: farm id gateway should be deployed on, if set choose available node from farm that fits vm specs (default 1). note: node and farm flags cannot be set both.
- tls: add TLS passthrough option (default false).
- network: network name (optional, for reference). Required when backend uses private IP.
- project-name: project name for grouping deployments (required when using --network).

### Network Flag Usage

The `--network` flag is optional but important for gateway-backend connectivity:

**IMPORTANT: Gateway must be able to reach your backend. Choose ONE option:**

#### 1️⃣ SAME NETWORK (Recommended - uses private WireGuard IPs)
Deploy gateway and VM on the same WireGuard network using private IPs:

```bash
# Deploy VM first
tfcmd deploy vm --name webapp --project-name myapp --ssh ~/.ssh/id_rsa.pub

# Deploy gateway on same network (can be same or different node within same farm)
tfcmd deploy gateway name --name api \
  --project-name myapp \
  --node 11 \
  --backends http://10.20.2.2:8080 \
  --network myappnetwork
```

> **Note**: Even if gateway and VM are on the same physical node, they still need to be on the same WireGuard network to communicate via private IPs (10.x.x.x).

#### 2️⃣ PUBLIC IPv4 (Cross-farm deployment)
Use VM's public IPv4 address as backend:

```bash
# Deploy VM with public IPv4
tfcmd deploy vm --name webapp --ipv4 --ssh ~/.ssh/id_rsa.pub
# Note the public IPv4 from deployment output

# Deploy gateway using VM's public IPv4
tfcmd deploy gateway name --name api \
  --project-name mygateway \
  --node 11 \
  --backends http://203.0.113.1:8080
```

#### 3️⃣ PLANETARY/MYCELIUM IP (Cross-network, no IPv4 needed)
Use planetary (Yggdrasil) or mycelium IP as backend:

```bash
# Deploy VM (planetary and mycelium IPs are assigned by default)
tfcmd deploy vm --name webapp --ssh ~/.ssh/id_rsa.pub
tfcmd get vm webapp --project-name vm/webapp  # Note planetary or mycelium IP

# Deploy gateway using planetary IP
tfcmd deploy gateway name --name api \
  --project-name mygateway \
  --node 11 \
  --backends http://[302:9e63:7d43:b742:c2e0:ab69:e101:8032]:8080

# Or using mycelium IP
tfcmd deploy gateway name --name api \
  --project-name mygateway \
  --node 11 \
  --backends http://[544:b74f:ceef:cc7e:ff0f:6b18:921f:8031]:8080
```

**Network Flag Requirements:**
- **Required** when backend uses private WireGuard IP (10.x.x.x from same network)
- **Not needed** when backend uses public IPv4, planetary, or mycelium IP
- **Must match** the VM's network name when using `--network` (e.g., if VM is on `myappnetwork`, gateway must use `--network myappnetwork`)

Example:

```console
$ tfcmd deploy gateway name -n gatewaytest --node 14 --backends http://93.184.216.34:80
3:34PM INF deploying gateway name
3:34PM INF fqdn: gatewaytest.gent01.dev.grid.tf
```

## Get

```bash
tfcmd get gateway name <gateway>
```

gateway is the name used when deploying gateway-name using tfcmd.

Example:

```console
$ tfcmd get gateway name gatewaytest
1:56PM INF gateway name:
{
        "NodeID": 14,
        "Name": "gatewaytest",
        "Backends": [
                "http://93.184.216.34:80"
        ],
        "TLSPassthrough": false,
        "Description": "",
        "SolutionType": "gatewaytest",
        "NodeDeploymentID": {
                "14": 19644
        },
        "FQDN": "gatewaytest.gent01.dev.grid.tf",
        "NameContractID": 19643,
        "ContractID": 19644
}
```

## Cancel

```bash
tfcmd cancel <deployment-name>
```

deployment-name is the name of the deployment specified in while deploying using tfcmd.

Example:

```console
$ tfcmd cancel gatewaytest
3:37PM INF canceling contracts for project gatewaytest
3:37PM INF gatewaytest canceled
```
