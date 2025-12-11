# List Commands

Discover available resources on the ThreeFold Grid.

## List Gateways

### List Name Gateways

List gateway nodes with configured domains for Name Gateway deployments.

```bash
tfcmd list gateways name [--country <country>] [--farm <id>]
```

**Flags:**
- `--country`: Filter by country name (optional)
- `--farm`: Filter by farm ID (optional)

**Examples:**
```bash
# List all name gateways
tfcmd list gateways name

# Filter by country
tfcmd list gateways name --country Belgium

# Filter by farm
tfcmd list gateways name --farm 1

# Combined filters
tfcmd list gateways name --country Germany --farm 2
```

**Output:**
```
Node ID    Domain                     Farm ID    Country        City
14         gent01.dev.grid.tf         1          Belgium        Ghent
15         gent02.dev.grid.tf         1          Belgium        Ghent
```

---

### List FQDN Gateways

List gateway nodes with public IP addresses for FQDN Gateway deployments.

```bash
tfcmd list gateways fqdn [--country <country>] [--farm <id>]
```

**Flags:**
- `--country`: Filter by country name (optional)
- `--farm`: Filter by farm ID (optional)

**Examples:**
```bash
# List all FQDN gateways
tfcmd list gateways fqdn

# Filter by country
tfcmd list gateways fqdn --country Germany

# Combined filters
tfcmd list gateways fqdn --country Belgium --farm 1
```

**Output:**
```
Node ID    Public IPv4        Public IPv6                           Farm ID    Country        City
14         185.206.122.33     2a10:b600:1::cc4:c5ff:fe88:2e4d      1          Belgium        Ghent
15         185.206.122.35     2a10:b600:1::cc4:c5ff:fe88:2e4e      1          Belgium        Ghent
```

---

## Use Cases

### For Name Gateway Deployments

1. **Discover available domains:**
   ```bash
   tfcmd list gateways name
   ```

2. **Deploy gateway:**
   ```bash
   tfcmd deploy gateway name --name myapp --node 14 --backends http://10.20.2.2:8080
   ```

3. **Your app will be available at:** `myapp.<gateway-domain>`

### For FQDN Gateway Deployments

1. **Find gateway with public IP:**
   ```bash
   tfcmd list gateways fqdn
   ```

2. **Deploy to specific node:**
   ```bash
   tfcmd deploy gateway fqdn --name myapp --fqdn myapp.com --node 14 --backends http://10.20.2.2:8080
   ```

3. **Configure DNS:**
   ```
   myapp.com → 185.206.122.33 (the public IPv4 from the list)
   ```

---

## Gateway Selection Considerations

When selecting a gateway node:

1. **Farm**: Gateway doesn't need to be in same farm as your VM
2. **Network**: If using private IP backend, specify `--network` so gateway can route to VM
3. **Location**: Choose gateway close to your users for better latency
4. **Availability**: Check node uptime and reliability
