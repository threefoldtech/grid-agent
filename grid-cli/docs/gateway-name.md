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

#### 1️⃣ SAME NETWORK (Recommended for multi-VM)
Deploy gateway and VM on the same network using private IPs:

```bash
# Deploy VM first
tfcmd deploy vm --name webapp --project-name myapp --ssh ~/.ssh/id_rsa.pub

# Deploy gateway on same network
tfcmd deploy gateway name --name api --node 11 \
  --backends http://10.20.2.2:8080 \
  --network myappnetwork \
  --project-name myapp
```

#### 2️⃣ SAME NODE (Simplest option)
Deploy gateway and VM on the same node:

```bash
# Deploy VM on specific node
tfcmd deploy vm --name webapp --node 11 --ssh ~/.ssh/id_rsa.pub

# Deploy gateway on same node
tfcmd deploy gateway name --name api --node 11 \
  --backends http://10.20.2.2:8080
```

#### 3️⃣ PUBLIC IP (Cross-farm deployment)
Use VM's public IP as backend:

```bash
# Deploy VM with public IP
tfcmd deploy vm --name webapp --ipv4 --ssh ~/.ssh/id_rsa.pub

# Deploy gateway using VM's public IP
tfcmd deploy gateway name --name api --node 11 \
  --backends http://203.0.113.1:8080
```

#### 4️⃣ PLANETARY/MYCELIUM IP (Cross-network)
Use planetary or mycelium IP as backend:

```bash
# Deploy VM (get planetary IP from get vm output)
tfcmd deploy vm --name webapp --ssh ~/.ssh/id_rsa.pub
tfcmd get vm webapp --project-name vm/webapp  # Note planetary IP

# Deploy gateway using planetary IP
tfcmd deploy gateway name --name api --node 11 \
  --backends http://302:9e63:7d43:b742:c2e0:ab69:e101:8032:8080
```

**Network Flag Requirements:**
- **Required** when backend uses private IP (10.x.x.x from same network)
- **Optional** when backend uses public, planetary, or mycelium IP
- **Must match** the VM's project name when using `--network`

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
