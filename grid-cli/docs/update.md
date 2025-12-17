# Update Commands

Update and manage deployed resources on the ThreeFold Grid.

## Update Kubernetes

Manage Kubernetes cluster workers (add or remove nodes).

### Add Worker Node

Add a new worker node to an existing Kubernetes cluster.

```bash
tfcmd update kubernetes add-worker [flags]
```

**Required Flags:**
- `--cluster-name`: Name of the Kubernetes cluster
- `--project-name`: Project name of the cluster
- `--node`: Node ID where to deploy the worker
- `--ssh`: Path to SSH public key

**Optional Flags:**
- `--disk`: Worker disk size (default: 50GB)
- `--memory`: Worker memory size (default: 4GB)
- `--farm`: Farm ID for node selection (if not using --node)

**Example:**
```console
$ tfcmd update kubernetes add-worker \
  --cluster-name mycluster \
  --project-name k8s/mycluster \
  --node 15 \
  --ssh ~/.ssh/id_rsa.pub \
  --disk 100 \
  --memory 8
2:30PM INF starting peer session=tf-1522850 twin=192
2:30PM INF adding worker to kubernetes cluster 'mycluster'
2:31PM INF successfully added worker to cluster
```

### Delete Worker Node

Remove a worker node from a Kubernetes cluster.

```bash
tfcmd update kubernetes delete-worker [flags]
```

**Required Flags:**
- `--cluster-name`: Name of the Kubernetes cluster
- `--project-name`: Project name of the cluster
- `--worker-node`: Node ID of worker to remove

**Example:**
```console
$ tfcmd update kubernetes delete-worker \
  --cluster-name mycluster \
  --project-name k8s/mycluster \
  --worker-node 15
2:45PM INF starting peer session=tf-1522851 twin=192
2:45PM INF removing worker node 15 from kubernetes cluster 'mycluster'
2:46PM INF successfully removed worker from cluster
```

## Cluster Management

### Scaling Workers

Add multiple workers to scale your cluster:

```bash
# Add first worker
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 11 \
  --ssh ~/.ssh/id_rsa.pub

# Add second worker  
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 12 \
  --ssh ~/.ssh/id_rsa.pub

# Add third worker
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 13 \
  --ssh ~/.ssh/id_rsa.pub
```

### Worker Configuration

Customize worker specifications:

```bash
# High-memory worker for databases
tfcmd update kubernetes add-worker \
  --cluster-name db-cluster \
  --project-name k8s/db-cluster \
  --node 14 \
  --ssh ~/.ssh/id_rsa.pub \
  --memory 16 \
  --disk 200

# Compute worker for CPU-intensive tasks
tfcmd update kubernetes add-worker \
  --cluster-name compute-cluster \
  --project-name k8s/compute-cluster \
  --node 15 \
  --ssh ~/.ssh/id_rsa.pub \
  --memory 4 \
  --disk 50
```

## Finding Clusters and Nodes

### List Your Clusters
```bash
# Get all contracts to find your clusters
tfcmd get contracts

# Look for kubernetes.* solution types
# k8s/mycluster → project name: k8s/mycluster
```

### Get Cluster Details
```bash
# Get current cluster information
tfcmd get kubernetes mycluster --project-name k8s/mycluster

# Check current workers and their nodes
```

### Find Available Nodes
```bash
# List nodes in a specific farm
tfcmd list gateways name --farm 1

# Or check node status directly
# (Node IDs should be known from your infrastructure planning)
```

## Worker Node Requirements

### Minimum Specifications
- **CPU**: 2 vCPUs
- **Memory**: 4GB RAM
- **Disk**: 50GB storage
- **Network**: Same farm as master nodes

### Recommended Specifications
- **Production workers**: 4+ vCPUs, 8GB+ RAM, 100GB+ disk
- **Development workers**: 2 vCPUs, 4GB RAM, 50GB disk
- **Database workers**: 4+ vCPUs, 16GB+ RAM, 200GB+ disk

### Network Requirements
- **Same farm**: Workers must be in same farm as cluster
- **Network access**: Must be able to join cluster network
- **Connectivity**: Must reach master nodes on private network

## Use Cases

### Horizontal Scaling
```bash
# Scale out for increased capacity
tfcmd update kubernetes add-worker \
  --cluster-name webapp \
  --project-name k8s/webapp \
  --node 16 \
  --ssh ~/.ssh/id_rsa.pub

# Scale down to reduce costs
tfcmd update kubernetes delete-worker \
  --cluster-name webapp \
  --project-name k8s/webapp \
  --worker-node 13
```

### Node Maintenance
```bash
# Replace failing node
# 1. Add new worker
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 17 \
  --ssh ~/.ssh/id_rsa.pub

# 2. Drain old node (kubectl drain nodeXX)
# 3. Delete old worker
tfcmd update kubernetes delete-worker \
  --cluster-name production \
  --project-name k8s/production \
  --worker-node 12
```

### Workload Separation
```bash
# Add dedicated database workers
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 18 \
  --ssh ~/.ssh/id_rsa.pub \
  --memory 16 \
  --disk 200

# Add dedicated compute workers
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 19 \
  --ssh ~/.ssh/id_rsa.pub \
  --memory 8 \
  --disk 100
```

## Monitoring and Verification

### Check Cluster Status
```bash
# Get updated cluster information
tfcmd get kubernetes mycluster --project-name k8s/mycluster

# Verify new workers are listed
```

### Kubernetes Verification
```bash
# After adding workers, verify in kubectl
kubectl get nodes
# Should show new worker nodes

# Check node status
kubectl describe node <new-worker-node>
```

### Network Verification
```bash
# Test connectivity between nodes
# Use kubectl to test pod networking
kubectl run test-pod --image=busybox --rm -it -- /bin/sh
# ping other node IPs
```

## Troubleshooting

### Common Issues

#### Worker Fails to Join Cluster
```bash
Error: worker failed to join kubernetes cluster
```
**Solutions:**
- Check network connectivity between nodes
- Verify worker is in same farm as master
- Check SSH key configuration
- Verify cluster network configuration

#### Node Not Available
```bash
Error: node 15 is not available
```
**Solutions:**
- Check if node exists and is online
- Verify node has sufficient resources
- Try a different node in the same farm

#### Project Not Found
```bash
Error: project k8s/mycluster not found
```
**Solutions:**
- Check project name with `tfcmd get contracts`
- Verify cluster was deployed successfully
- Use correct project name format

### Recovery Operations

#### Failed Worker Addition
```bash
# If worker addition fails, check cluster status
tfcmd get kubernetes mycluster --project-name k8s/mycluster

# Remove any partially created workers
tfcmd update kubernetes delete-worker \
  --cluster-name mycluster \
  --project-name k8s/mycluster \
  --worker-node <failed-node-id>
```

#### Cluster Corruption
```bash
# If cluster is corrupted, you may need to:
# 1. Cancel entire deployment
tfcmd cancel --project-name k8s/mycluster

# 2. Deploy new cluster
tfcmd deploy kubernetes --name mycluster --project-name k8s/mycluster ...

# 3. Add workers back
tfcmd update kubernetes add-worker ...
```

## Best Practices

### Production Deployment
```bash
# Use consistent worker specifications
tfcmd update kubernetes add-worker \
  --cluster-name production \
  --project-name k8s/production \
  --node 20 \
  --ssh ~/.ssh/id_rsa.pub \
  --memory 8 \
  --disk 100

# Document worker configurations
# Monitor worker health
# Plan node maintenance windows
```

### Cost Optimization
```bash
# Scale down during off-peak hours
tfcmd update kubernetes delete-worker \
  --cluster-name dev-cluster \
  --project-name k8s/dev-cluster \
  --worker-node 21

# Scale up when needed
tfcmd update kubernetes add-worker \
  --cluster-name dev-cluster \
  --project-name k8s/dev-cluster \
  --node 22 \
  --ssh ~/.ssh/id_rsa.pub
```

### Security Considerations
- Use consistent SSH keys across workers
- Monitor worker node access
- Regular security updates on workers
- Network segmentation between clusters

## Integration with kubectl

After updating workers, use kubectl to manage the Kubernetes cluster:

```bash
# Get cluster credentials (if applicable)
# (This depends on your Kubernetes setup)

# Verify new workers
kubectl get nodes

# Schedule workloads on new workers
kubectl label nodes <worker-node> workload-type=database

# Remove old workers gracefully
kubectl drain <old-worker-node>
```

## Future Updates

The update command may expand to support:
- VM resource updates
- Gateway configuration updates  
- ZDB capacity modifications
- Network configuration changes

Check `tfcmd update --help` for the latest available update operations.
