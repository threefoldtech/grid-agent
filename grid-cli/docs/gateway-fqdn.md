# Gateway FQDN

This document explains Gateway FQDN related commands using tfcmd.

## Deploy

```bash
tfcmd deploy gateway fqdn [flags]
```

### Required Flags

- name: name for the gateway deployment also used for canceling the deployment. must be unique.
- node: node id to deploy gateway on.
- backends: list of backends the gateway will forward requests to.
- fqdn: FQDN pointing to the specified node.

### Optional Flags

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
tfcmd deploy gateway fqdn --name api --fqdn api.example.com --node 11 \
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
tfcmd deploy gateway fqdn --name api --fqdn api.example.com --node 11 \
  --backends http://10.20.2.2:8080
```

#### 3️⃣ PUBLIC IP (Cross-farm deployment)
Use VM's public IP as backend:

```bash
# Deploy VM with public IP
tfcmd deploy vm --name webapp --ipv4 --ssh ~/.ssh/id_rsa.pub

# Deploy gateway using VM's public IP
tfcmd deploy gateway fqdn --name api --fqdn api.example.com --node 11 \
  --backends http://203.0.113.1:8080
```

#### 4️⃣ PLANETARY/MYCELIUM IP (Cross-network)
Use planetary or mycelium IP as backend:

```bash
# Deploy VM (get planetary IP from get vm output)
tfcmd deploy vm --name webapp --ssh ~/.ssh/id_rsa.pub
tfcmd get vm webapp --project-name vm/webapp  # Note planetary IP

# Deploy gateway using planetary IP
tfcmd deploy gateway fqdn --name api --fqdn api.example.com --node 11 \
  --backends http://302:9e63:7d43:b742:c2e0:ab69:e101:8032:8080
```

**Network Flag Requirements:**
- **Required** when backend uses private IP (10.x.x.x from same network)
- **Optional** when backend uses public, planetary, or mycelium IP
- **Must match** the VM's project name when using `--network`

**DNS Configuration:**
After deployment, configure your DNS:
```
api.example.com → <gateway-node-public-IP>
```

Example:

```console
$ tfcmd deploy gateway fqdn -n gatewaytest --node 14 --backends http://93.184.216.34:80 --fqdn example.com
3:34PM INF deploying gateway fqdn
3:34PM INF gateway fqdn deployed
```

## Get

```bash
tfcmd get gateway fqdn <gateway>
```

gateway is the name used when deploying gateway-fqdn using tfcmd.

Example:

```console
$ tfcmd get gateway fqdn gatewaytest
2:05PM INF gateway fqdn:
{
        "NodeID": 14,
        "Backends": [
                "http://93.184.216.34:80"
        ],
        "FQDN": "awady.gridtesting.xyz",
        "Name": "gatewaytest",
        "TLSPassthrough": false,
        "Description": "",
        "NodeDeploymentID": {
                "14": 19653
        },
        "SolutionType": "gatewaytest",
        "ContractID": 19653
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
