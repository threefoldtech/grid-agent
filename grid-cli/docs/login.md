# Login

Authenticate with the ThreeFold Grid using your mnemonic phrase.

## Login Command

```bash
tfcmd login
```

The login command provides an interactive authentication flow that prompts you for:
- **Mnemonic phrase**: Your 12 or 24-word seed phrase
- **Network selection**: Choose between mainnet, testnet, devnet, or QA network

## Interactive Login Process

### Step 1: Enter Mnemonic
```console
$ tfcmd login
Enter your mnemonic phrase (12 or 24 words):
```

Paste your mnemonic phrase when prompted. The phrase will be hidden for security.

### Step 2: Select Network
```console
Select network:
1) mainnet
2) testnet  
3) devnet
4) qa
Enter choice (1-4):
```

Choose the appropriate network for your deployment:
- **mainnet**: Production environment with real costs
- **testnet**: Testing environment with free resources
- **devnet**: Development environment
- **qa**: Quality assurance environment

### Step 3: Confirmation
```console
Login successful! Configuration saved to ~/.config/tfgrid/config.yaml
Network: testnet
Twin ID: 192
```

## Configuration Storage

Your credentials and network selection are saved locally in:
```
~/.config/tfgrid/config.yaml
```

This file contains:
- Encrypted mnemonic phrase
- Selected network
- Twin ID (automatically retrieved)
- Account information

## Network Options

### Mainnet
- **Purpose**: Production deployments
- **Costs**: Real TFT tokens required
- **Resources**: Full production capacity
- **Use Case**: Production applications and services

### Testnet
- **Purpose**: Testing and development
- **Costs**: Free testnet tokens
- **Resources**: Limited but sufficient for testing
- **Use Case**: Application testing and development

### Devnet
- **Purpose**: Development and experimentation
- **Costs**: Free
- **Resources**: Development environment
- **Use Case**: Feature development and debugging

### QA
- **Purpose**: Quality assurance and testing
- **Costs**: Free
- **Resources**: QA environment
- **Use Case**: Pre-production testing

## Security Considerations

### Mnemonic Security
- ✅ **Never share your mnemonic** with anyone
- ✅ **Store it securely** offline
- ✅ **Use different mnemonics** for different networks
- ❌ **Don't commit to version control**
- ❌ **Don't share in screenshots or logs**

### Configuration File
- ✅ **Locally stored** on your machine only
- ✅ **Encrypted at rest**
- ✅ **Permissions restricted** to user only
- ❌ **Don't share the config file**

## Troubleshooting

### Common Issues

#### Invalid Mnemonic
```console
Error: invalid mnemonic phrase
```
**Solution**: Verify your mnemonic phrase is correct (12 or 24 words)

#### Network Connection
```console
Error: failed to connect to grid network
```
**Solution**: Check internet connection and network availability

#### Twin Not Found
```console
Error: twin not found for this mnemonic
```
**Solution**: Ensure you've created a twin on the selected network

### Reset Configuration
If you need to reset your configuration:
```bash
# Remove configuration file
rm ~/.config/tfgrid/config.yaml

# Login again
tfcmd login
```

### Switch Networks
To switch to a different network:
```bash
# Remove current configuration
rm ~/.config/tfgrid/config.yaml

# Login and select new network
tfcmd login
```

## Advanced Usage

### Environment Variables
You can set these environment variables to customize behavior:

```bash
# Override config location
export TFGRID_CONFIG_DIR=/custom/path

# Disable sentry error reporting
export TFGRID_DISABLE_SENTRY=true
```

### Multiple Accounts
For multiple accounts, use different configuration directories:
```bash
# Account 1
TFGRID_CONFIG_DIR=~/.tfgrid/account1 tfcmd login

# Account 2  
TFGRID_CONFIG_DIR=~/.tfgrid/account2 tfcmd login
```

## Verification

After login, verify your configuration:

### Check Twin ID
```bash
# Your twin ID is shown in login confirmation
# Or check with any command
tfcmd get contracts
```

### Test Connection
```bash
# Test basic connectivity
tfcmd list gateways name
```

### Verify Network
```bash
# Check you're on the right network
# (Different networks have different available resources)
tfcmd list gateways name --farm 1
```

## Wallet Connector Integration

If you don't have mnemonics yet, you can:
1. Use the [ThreeFold Wallet Connector](https://manual.grid.tf/documentation/dashboard/wallet_connector.html)
2. Create a new wallet
3. Export your mnemonic phrase
4. Use it with `tfcmd login`

## Best Practices

### Development Workflow
```bash
# 1. Use testnet for development
tfcmd login
# Select: testnet

# 2. Test deployments
tfcmd deploy vm --name test --ssh ~/.ssh/id_rsa.pub

# 3. Cleanup
tfcmd cancel --project-name vm/test
```

### Production Workflow
```bash
# 1. Use mainnet for production
tfcmd login  
# Select: mainnet

# 2. Verify TFT balance
# (Check your wallet has sufficient TFT)

# 3. Deploy production resources
tfcmd deploy vm --name prod --ssh ~/.ssh/id_rsa.pub --cpu 4 --memory 8
```

### Security Checklist
- [ ] Mnemonic stored securely offline
- [ ] Different mnemonics for mainnet/testnet
- [ ] Config file permissions set correctly
- [ ] Regular backup of configuration
- [ ] Never share credentials in logs/screenshots

## Getting Help

If you encounter issues:
1. Check network connectivity
2. Verify mnemonic phrase
3. Ensure twin exists on selected network
4. Try resetting configuration
5. Contact ThreeFold support

For more information about mnemonics and wallet setup, see the [ThreeFold documentation](https://manual.grid.tf/documentation/dashboard/wallet_connector.html).
