# tfcmd

<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-53%25-brightgreen.svg?longCache=true&style=flat)</a>

Threefold CLI to manage deployments on Threefold Grid.

## Usage

First [download](#download) tfcmd binaries.

Login using your mnemonics and specify which grid network (mainnet/testnet) to deploy on by running:

```bash
tfcmd login
```

Check [Wallet Connector](https://manual.grid.tf/documentation/dashboard/wallet_connector.html) for more details if you do not have mnemonics yet.

For examples and description of tfcmd commands check out:

- [vm](docs/vm.md)
- [gateway-fqdn](docs/gateway-fqdn.md)
- [gateway-name](docs/gateway-name.md)
- [kubernetes](docs/kubernetes.md)
- [ZDB](docs/zdb.md)
- [cancel](docs/cancel.md)
- [get](docs/get.md)
- [list](docs/list.md)
- [login](docs/login.md)
- [update](docs/update.md)
- [version](docs/version.md)
- [contracts](docs/contracts.md)

## Quick Examples

```bash
# Login to ThreeFold Grid
tfcmd login

# Deploy a VM
tfcmd deploy vm --name myvm --ssh ~/.ssh/id_rsa.pub --cpu 2 --memory 4

# Get VM information
tfcmd get vm myvm --project-name vm/myvm

# List available gateways
tfcmd list gateways name

# Deploy a gateway name
tfcmd deploy gateway name --name myapp --project-name vm/myvm --network myvmnetwork --node 11 --backends http://10.20.2.2:8080

# Cancel deployment
tfcmd cancel --project-name vm/myvm
```

## Command Reference

```bash
tfcmd
├── deploy
│   ├── vm              # Deploy virtual machines
│   ├── gateway
│   │   ├── name        # Deploy name gateway (subdomain)
│   │   └── fqdn        # Deploy FQDN gateway
│   ├── kubernetes      # Deploy Kubernetes clusters
│   └── zdb             # Deploy Zero-DB instances
├── get
│   ├── vm              # Get VM information
│   ├── gateway
│   │   ├── name        # Get name gateway info
│   │   └── fqdn        # Get FQDN gateway info
│   ├── kubernetes      # Get Kubernetes cluster info
│   ├── zdb             # Get ZDB instance info
│   ├── contract        # Get specific contract info
│   └── contracts       # List all contracts
├── list
│   └── gateways
│       ├── name        # List name gateways
│       └── fqdn        # List FQDN gateways
├── cancel
│   └── contracts       # Cancel contracts
├── update
│   └── kubernetes      # Update Kubernetes workers
├── login               # Authenticate with mnemonics
├── version             # Show version information
└── completion          # Generate shell completions
```

## Download

- Download the binaries from [releases](https://github.com/threefoldtech/tfgrid-sdk-go/releases)
- Extract the downloaded files
- Move the binary to any of `$PATH` directories, for example:

```bash
mv tfcmd /usr/local/bin
```

## Configuration

tf-grid saves user configuration in `.tfgridconfig` under default configuration directory for your system see: [UserConfigDir()](https://pkg.go.dev/os#UserConfigDir)

## Build

Clone the repo and run the following command inside the repo directory:

```bash
make build
```
