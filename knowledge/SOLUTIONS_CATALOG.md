# ThreeFold Grid Solutions Catalogs

This document provides a comprehensive list of all solutions supported in the ThreeFold Grid Playground weblets, including their configurations, requirements, and specifications.

---

## 1. Algorand

**Flist URL:** `https://hub.grid.tf/tf-official-apps/algorand-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `NETWORK` | Optional | `mainnet` | Network type (mainnet/testnet/betanet/devnet) |
| `NODE_TYPE` | Optional | `default` | Node type (default/relay/indexer) |

### Proposed Specs

- **CPU:** Variable based on node type (2-4 cores)
- **Memory:** Variable based on node type (4-8 GB)
- **Storage:** Variable based on node type (100-1500 GB)

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | 50 GB | `/var/lib/docker` | Yes (indexer only) |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 2. Bitcoin Node

**Flist URL:** `https://hub.grid.tf/tf-official-apps/threefoldtech-tf_btcnode_30.0.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Small:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Medium:** CPU: 4, Memory: 8 GB, Disk: 100 GB
- **Large:** CPU: 8, Memory: 16 GB, Disk: 200 GB

### Disk/Volumes

No additional disks required (uses root filesystem).

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 3. Caprover

**Flist URL:** `https://hub.grid.tf/tf-official-apps/tf-caprover-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

#### Leader Node

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SWM_NODE_MODE` | Required | `leader` | Node mode |
| `CAPROVER_ROOT_DOMAIN` | Required | - | Root domain for Caprover |
| `CAPTAIN_IMAGE_VERSION` | Required | `latest` | Caprover version |
| `PUBLIC_KEY` | Required | - | SSH public key |
| `DEFAULT_PASSWORD` | Required | - | Admin password |
| `CAPTAIN_IS_DEBUG` | Optional | `true` | Debug mode |

#### Worker Nodes

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SWM_NODE_MODE` | Required | `worker` | Node mode |
| `PUBLIC_KEY` | Required | - | SSH public key |

### Proposed Specs

Variable based on leader and worker configurations.

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4:** Required for leader
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Multi-Node Architecture:**

- Caprover supports a **leader + workers** cluster architecture
- All nodes (leader and workers) should deployed on the **same private network**

**Network Linking:**

- Leader and workers communicate via the private network (WireGuard)
- Workers discover the leader automatically through the shared network

**Deployment Flow:**

1. Create a single network for the entire cluster
2. Deploy leader node with `SWM_NODE_MODE=leader`
3. Deploy worker nodes with `SWM_NODE_MODE=worker` on the same network
4. All nodes share the same deployment name and network configuration

**DNS Requirements:**

- Wildcard DNS record required: `*.yourdomain.com` → Leader's public IPv4
- Example: If domain is `example.com`, set DNS: `*.example.com` → `185.x.x.x`
- This allows Caprover to create subdomains for each deployed app

**Special Handling:**

- Leader requires public IPv4 (mandatory)
- Workers can use private IPs only
- Each node gets its own disk mounted at `/var/lib/docker`
- SSH keys are shared across all nodes

---

## 4. Casperlabs

**Flist URL:** `https://hub.grid.tf/tf-official-apps/casperlabs-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `CASPERLABS_HOSTNAME` | Required | - | Hostname/domain |
| `KNOWN_VALIDATOR_IP` | Optional | `3.14.161.135` | Known validator IP |

### Proposed Specs

- **Small:** CPU: 2, Memory: 4 GB, Disk: 100 GB
- **Medium:** CPU: 4, Memory: 16 GB, Disk: 500 GB
- **Large:** CPU: 8, Memory: 32 GB, Disk: 1000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 5. Discourse

**Flist URL:** `https://hub.grid.tf/tf-official-apps/forum.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `DISCOURSE_HOSTNAME` | Required | - | Hostname/domain |
| `DISCOURSE_DEVELOPER_EMAILS` | Required | - | Admin email |
| `DISCOURSE_SMTP_ADDRESS` | Required | - | SMTP server address |
| `DISCOURSE_SMTP_PORT` | Required | - | SMTP server port |
| `DISCOURSE_SMTP_ENABLE_START_TLS` | Required | - | Enable TLS (true/false) |
| `DISCOURSE_SMTP_USER_NAME` | Required | - | SMTP username |
| `DISCOURSE_SMTP_PASSWORD` | Required | - | SMTP password |
| `THREEBOT_PRIVATE_KEY` | Required | - | ED25519 base64 private key |
| `FLASK_SECRET_KEY` | Required | - | Flask secret |

### Proposed Specs

- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **SMTP:** Required
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 6. Freeflow

**Flist URL:** `https://hub.grid.tf/lennertapp2.3bot/threefoldjimber-freeflow-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `USER_ID` | Required | - | 3Bot name |
| `DIGITALTWIN_APPID` | Required | - | Domain |
| `NODE_ENV` | Required | `staging` | Environment |

### Proposed Specs

- **Small:** CPU: 1, Memory: 4 GB, Disk: 100 GB
- **Medium:** CPU: 2, Memory: 16 GB, Disk: 500 GB
- **Large:** CPU: 4, Memory: 32 GB, Disk: 1000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/disk` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 7. Ubuntu OS Images

**Flist URLs:**

- Ubuntu 24.04: `https://hub.grid.tf/tf-official-vms/ubuntu-24.04-full.flist`
- Ubuntu 22.04: `https://hub.grid.tf/tf-official-vms/ubuntu-22.04.flist`
- Ubuntu 20.04: `https://hub.grid.tf/tf-official-vms/ubuntu-20.04-lts.flist`
- Ubuntu 18.04: `https://hub.grid.tf/tf-official-vms/ubuntu-18.04-lts.flist`

**Entry Point:** `` (empty)

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Micro:** CPU: 1, Memory: 0.5 GB, Disk: 20 GB
- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/` or custom | Optional (additional disks) |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** No
- **GPU:** Optional
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Use Cases:**

- Development environments
- Production servers
- GPU computing (ML/AI workloads)
- Custom application hosting

---

## 8. Funkwhale

**Flist URL:** `https://hub.grid.tf/tf-official-apps/funkwhale-1.4.0.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `FUNKWHALE_HOSTNAME` | Required | - | Hostname/domain |
| `FUNKWHALE_SUPERUSER_EMAIL` | Required | - | Admin email |
| `FUNKWHALE_SUPERUSER_USERNAME` | Required | - | Admin username |
| `FUNKWHALE_SUPERUSER_PASSWORD` | Required | - | Admin password |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 50 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 9. Gitea

**Flist URL:** `https://hub.grid.tf/tf-official-apps/gitea-mycelium.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `GITEA__HOSTNAME` | Required | - | Hostname/domain |
| `GITEA__mailer__PROTOCOL` | Optional | `smtp` | Mail protocol |
| `GITEA__mailer__ENABLED` | Optional | `false` | Enable mailer |
| `GITEA__mailer__HOST` | Optional | `smtp.gmail.com` | SMTP host |
| `GITEA__mailer__FROM` | Optional | - | From email |
| `GITEA__mailer__PORT` | Optional | `587` | SMTP port |
| `GITEA__mailer__USER` | Optional | - | SMTP username |
| `GITEA__mailer__PASSWD` | Optional | - | SMTP password |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

No additional disks required (uses root filesystem).

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** Yes (required)
- **SMTP:** Optional
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 10. Jenkins

**Flist URL:** `https://hub.grid.tf/tf-official-apps/jenkins-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `JENKINS_HOSTNAME` | Required | - | Hostname/domain |
| `JENKINS_ADMIN_USERNAME` | Required | - | Admin username |
| `JENKINS_ADMIN_PASSWORD` | Required | - | Admin password |

### Proposed Specs

- **Small:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Medium:** CPU: 4, Memory: 8 GB, Disk: 500 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 1000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 11. Jitsi

**Flist URL:** `https://hub.grid.tf/tf-official-apps/jitsi-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `JITSI_HOSTNAME` | Required | - | Hostname/domain |

### Proposed Specs

- **Small:** CPU: 2, Memory: 4 GB, Disk: 100 GB
- **Medium:** CPU: 4, Memory: 16 GB, Disk: 500 GB
- **Large:** CPU: 8, Memory: 32 GB, Disk: 1000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 12. Kubernetes

**Flist URL:** Uses same flist as micro VMs (varies by OS)

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

Variable based on master and worker configurations.

### Disk/Volumes

Variable based on configuration.

### Additional Requirements

- **IPv4:** Required for master
- **Gateway/Domain:** No
- **Cluster Token:** Required (6-15 characters)
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Cluster Architecture:**

- Kubernetes supports a **master + workers** cluster architecture
- All nodes are deployed on the **same private network**
- Master and workers communicate via the shared network

**Special Requirements:**

- Master node requires public IPv4 for external access
- Cluster token acts as authentication between nodes
- SSH key is mandatory (not optional) for cluster management
- Each node can have different resource specifications

**Use Cases:**

- Container orchestration
- Microservices deployment
- Scalable application hosting

---

## 13. Mattermost

**Flist URL:** `https://hub.grid.tf/tf-official-apps/mattermost-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `DB_PASSWORD` | Required | - | Database password |
| `SITE_URL` | Required | - | Site URL (<https://domain>) |
| `MATTERMOST_DOMAIN` | Required | - | Domain |
| `SMTPUsername` | Optional | - | SMTP username |
| `SMTPPassword` | Optional | - | SMTP password |
| `SMTPServer` | Optional | - | SMTP server |
| `SMTPPort` | Optional | - | SMTP port |

### Proposed Specs

- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional (IPv4 required if SMTP enabled)
- **Gateway/Domain:** Yes (required)
- **SMTP:** Optional
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 14. Micro Virtual Machine

**Flist URLs and Entry Points:**

- Ubuntu 24.04: `https://hub.grid.tf/tf-official-vms/ubuntu-24.04-latest.flist` - Entry Point: `/sbin/zinit init`
- Ubuntu 23.10: `https://hub.grid.tf/tf-official-vms/ubuntu-23.10-mycelium.flist` - Entry Point: `/sbin/zinit init`
- Ubuntu 22.04: `https://hub.grid.tf/tf-official-apps/threefoldtech-ubuntu-22.04.flist` - Entry Point: `/sbin/zinit init`
- Debian 12: `https://hub.grid.tf/tf-official-apps/debian12.flist` - Entry Point: `/sbin/zinit init`
- CentOS 9: `https://hub.grid.tf/tf-official-apps/centos-stream9.flist` - Entry Point: `/entrypoint.sh`
- Arch: `https://hub.grid.tf/petep.3bot/archlinux_20240101.0.204074.flist` - Entry Point: `/sbin/zinit init`
- Alpine 3: `https://hub.grid.tf/tf-official-apps/alpine3.flist` - Entry Point: `/entrypoint.sh`

### Environment Variables

Custom environment variables can be defined.

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Micro:** CPU: 1, Memory: 0.5 GB, Disk: 20 GB
- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | Custom | Optional (additional disks) |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Custom Environment Variables:**

- Unlike other solutions, Micro VMs allow you to define **any custom environment variables**
- Useful for passing configuration to your applications
- Variables are injected into the VM at boot time

---

## 15. Nextcloud

**Flist URL:** `https://hub.grid.tf/tf-official-apps/nextcloud.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `NEXTCLOUD_DOMAIN` | Required | - | Domain |
| `NEXTCLOUD_AIO_LINK` | Required | - | AIO link (domain/aio) |
| `GATEWAY` | Required | - | Has gateway (true/false) |
| `IPV4` | Required | - | Has IPv4 (true/false) |

### Proposed Specs

- **Small:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Medium:** CPU: 4, Memory: 8 GB, Disk: 500 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 1000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/mnt/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Nextcloud AIO (All-in-One):**

- This deployment uses Nextcloud AIO, which includes:
  - Nextcloud server
  - Database (PostgreSQL)
  - Redis cache
  - Full-text search
  - Backup solution
  - All in one container

**Access Setup:**

- Access Nextcloud via: `https://yourdomain.com`
- Access AIO admin interface via: `https://yourdomain.com/aio`
- AIO interface is used for initial setup and configuration

**Environment Variables:**

- `NEXTCLOUD_DOMAIN`: Your full domain
- `NEXTCLOUD_AIO_LINK`: Domain with `/aio` path for admin
- `GATEWAY`: Set to `true` if using gateway
- `IPV4`: Set to `true` if node has public IPv4

**Storage:**

- Large disk recommended (500GB - 1TB)
- All data stored in `/mnt/data`
- Includes files, database, and backups

**First-Time Setup:**

1. Deploy the instance
2. Access the AIO interface at `yourdomain.com/aio`
3. Complete the initial configuration wizard
4. Set admin password and configure apps

---

## 16. Node Pilot

**Flist URL:** `https://hub.grid.tf/tf-official-vms/node-pilot-zdbfs.flist`

**Entry Point:** `/`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `NODE_PILOT_HOSTNAME` | Required | - | Hostname/domain |

### Proposed Specs

- **Small:** CPU: 4, Memory: 8 GB, Disk: 500 GB
- **Medium:** CPU: 8, Memory: 16 GB, Disk: 1000 GB
- **Large:** CPU: 8, Memory: 32 GB, Disk: 2000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/` | Yes |

### Additional Requirements

- **IPv4:** Required
- **IPv6:** Required
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 17. Nostr

**Flist URL:** `https://hub.grid.tf/tf-official-apps/nostr_relay-mycelium.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `NOSTR_HOSTNAME` | Required | - | Hostname/domain |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

No additional disks required (uses root filesystem).

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 18. Open WebUI

**Flist URL:** `https://hub.grid.tf/tf-official-apps/threefoldtech-ubuntu-24.04_fullvm_oi.flist`

**Entry Point:** `` (empty)

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `OPENWEBUI_DOMAIN` | Required | - | Domain |

### Proposed Specs

- **Small:** CPU: 4, Memory: 16 GB, Disk: 125 GB
- **Medium:** CPU: 8, Memory: 32 GB, Disk: 250 GB
- **Large:** CPU: 16, Memory: 64 GB, Disk: 500 GB

### Disk/Volumes

No additional disks required (uses root filesystem).

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** Yes (required)
- **GPU:** Optional
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**GPU Support for AI:**

- Open WebUI supports GPU passthrough for AI model inference
- GPU significantly improves performance for large language models
- Requires dedicated node with GPU when using GPU features
- GPU is optional - CPU-only deployment is also supported

**AI Model Hosting:**

- Compatible with Ollama for running local AI models
- Can run models like Llama, Mistral, CodeLlama, etc.
- GPU recommended for models larger than 7B parameters
- CPU-only works for smaller models (3B-7B)

**Resource Requirements:**

- Minimum: 4 CPU, 16 GB RAM (for small models)
- Recommended: 8+ CPU, 32+ GB RAM, GPU (for large models)
- Large disk space for model storage (125GB+)

**Use Cases:**

- Private AI assistant deployment
- Local LLM experimentation
- AI-powered applications
- Privacy-focused AI hosting

---

## 19. Owncloud

**Flist URL:** `https://hub.grid.tf/tf-official-apps/owncloud-10.9.1.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `OWNCLOUD_ADMIN_USERNAME` | Required | - | Admin username |
| `OWNCLOUD_ADMIN_PASSWORD` | Required | - | Admin password |
| `OWNCLOUD_DOMAIN` | Required | - | Domain |
| `OWNCLOUD_MAIL_SMTP_SECURE` | Optional | - | SMTP security (tls/ssl/none) |
| `OWNCLOUD_MAIL_FROM_ADDRESS` | Optional | - | From email address |
| `OWNCLOUD_MAIL_DOMAIN` | Optional | - | Mail domain |
| `OWNCLOUD_MAIL_SMTP_HOST` | Optional | - | SMTP host |
| `OWNCLOUD_MAIL_SMTP_PORT` | Optional | - | SMTP port |
| `OWNCLOUD_MAIL_SMTP_NAME` | Optional | - | SMTP username |
| `OWNCLOUD_MAIL_SMTP_PASSWORD` | Optional | - | SMTP password |

### Proposed Specs

- **Small:** CPU: 2, Memory: 8 GB, Disk: 250 GB
- **Medium:** CPU: 4, Memory: 16 GB, Disk: 500 GB
- **Large:** CPU: 8, Memory: 32 GB, Disk: 1000 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **SMTP:** Optional
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 20. Peertube

**Flist URL:** `https://hub.grid.tf/tf-official-apps/peertube-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `PEERTUBE_ADMIN_EMAIL` | Required | - | Admin email |
| `PT_INITIAL_ROOT_PASSWORD` | Required | - | Admin password |
| `PEERTUBE_WEBSERVER_HOSTNAME` | Required | - | Hostname/domain |

### Proposed Specs

Uses standard solution flavors (Small/Medium/Large).

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 21. Presearch

**Flist URL:** `https://hub.grid.tf/tf-official-apps/presearch.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `PRESEARCH_REGISTRATION_CODE` | Required | - | Registration code (32 chars) |
| `PRESEARCH_BACKUP_PRI_KEY` | Optional | - | Private restore key |
| `PRESEARCH_BACKUP_PUB_KEY` | Optional | - | Public restore key |

### Proposed Specs

- **CPU:** 1
- **Memory:** 512 MB
- **Root Filesystem:** Calculated
- **Docker Disk:** 10 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | 10 GB | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Farm Limitations:**

- **Important:** Only **one Presearch node per farm** is allowed without a dedicated public IP
- If deploying multiple nodes, each must have its own public IPv4
- This is a Presearch network requirement, not a Grid limitation

**Restore Keys:**

- Private and public restore keys are optional but recommended
- These keys allow you to restore your node if redeployed
- Without restore keys, you'll need to re-register with Presearch

**Registration:**

- Registration code (32 characters) is mandatory
- Obtain from Presearch dashboard before deployment
- Each node requires a unique registration code

**Resource Requirements:**

- Minimal resources: 1 CPU, 512 MB RAM
- Small disk requirement: 10 GB for Docker
- Suitable for low-resource nodes

---

## 22. Static Website

**Flist URL:** `https://hub.grid.tf/tf-official-apps/staticwebsite-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `GITHUB_URL` | Required | - | Git repository URL (https) |
| `GITHUB_BRANCH` | Optional | - | Git branch name |
| `HTML_DIR` | Optional | `website` | HTML directory path |
| `USER_DOMAIN` | Optional | - | Custom domain |
| `STATICWEBSITE_DOMAIN` | Required | - | Domain |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 50 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 23. Subsquid

**Flist URL:** `https://hub.grid.tf/tf-official-apps/subsquid.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `CHAIN_ENDPOINT` | Required | - | Websocket endpoint (wss://) |
| `SUBSQUID_WEBSERVER_HOSTNAME` | Required | - | Hostname/domain |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 50 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 24. Taiga

**Flist URL:** `https://hub.grid.tf/tf-official-apps/grid3_taiga_docker-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `DOMAIN_NAME` | Required | - | Domain |
| `ADMIN_USERNAME` | Required | - | Admin username |
| `ADMIN_PASSWORD` | Required | - | Admin password |
| `ADMIN_EMAIL` | Required | - | Admin email |
| `EMAIL_HOST` | Optional | - | SMTP host |
| `EMAIL_PORT` | Optional | - | SMTP port |
| `EMAIL_HOST_USER` | Optional | - | SMTP username |
| `EMAIL_HOST_PASSWORD` | Optional | - | SMTP password |
| `EMAIL_USE_TLS` | Optional | - | Use TLS (True/False) |
| `EMAIL_USE_SSL` | Optional | - | Use SSL (True/False) |

### Proposed Specs

- **Small:** CPU: 2, Memory: 4 GB, Disk: 100 GB
- **Medium:** CPU: 4, Memory: 8 GB, Disk: 150 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional (IPv4 required if SMTP enabled)
- **Gateway/Domain:** Yes (required)
- **SMTP:** Optional
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 25. TFRobot

**Flist URL:** `https://hub.grid.tf/tf-official-apps/tfrobot.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

Custom environment variables can be defined.

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | Custom | Optional (additional disks) |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

### Notes

**Flexible Deployment:**

- TFRobot is a **generic VM template** for custom deployments
- Allows you to define custom environment variables
- Supports multiple disks with custom mount points
- Most flexible solution for advanced users

**Custom Configuration:**

- Define any environment variables needed for your application
- Add as many disks as needed
- Specify custom mount points for each disk
- Full control over VM configuration

**Use Cases:**

- Custom application deployment
- Testing new services
- Advanced configurations not covered by other solutions
- Prototype development

---

## 26. Umbrel

**Flist URL:** `https://hub.grid.tf/tf-official-apps/umbrel-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `USERNAME` | Required | - | Admin username |
| `PASSWORD` | Required | - | Admin password |
| `UMBREL_DISK` | Required | `/umbrelDisk` | Umbrel disk path |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 10 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | 10 GB | `/var/lib/docker` | Yes |
| SSD | Variable | `/umbrelDisk` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (at least one)
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 27. Wordpress

**Flist URL:** `https://hub.grid.tf/tf-official-apps/tf-wordpress-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `MYSQL_USER` | Required | - | MySQL username |
| `MYSQL_PASSWORD` | Required | - | MySQL password |
| `ADMIN_EMAIL` | Required | - | Admin email |
| `WP_URL` | Required | - | WordPress URL/domain |

### Proposed Specs

- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 16 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/www/html` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

---

## 28. Aydo

**Flist URL:** `https://hub.grid.tf/tf-official-apps/aydo-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (for Aydo to work with ONLYOFFICE)
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- Aydo application deployment
- ONLYOFFICE document collaboration
- Web-based office applications

---

## 29. Dagu

**Flist URL:** `https://hub.grid.tf/tf-official-apps/dagu-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `DAGU_USERNAME` | Optional | `admin` | Dagu UI username |
| `DAGU_PASSWORD` | Optional | `password` | Dagu UI password |
| `CODE_SERVER_PASSWORD` | Optional | `password` | Code server access password |
| `HOMEIP` | Optional | - | IP address for hero callback |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- Workflow automation and orchestration
- Cron job management
- DAG-based task scheduling
- Code server for remote development

---

## 30. Hero

**Flist URL:** `https://hub.grid.tf/tf-official-apps/hero-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- Hero framework development
- Go language development environment
- ThreeFold ecosystem development
- Build and deployment automation

---

## 31. IPFS Cluster

**Flist URL:** `https://hub.grid.tf/tf-official-apps/ipfs-cluster-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `CLUSTER_SECRET` | Optional | - | Cluster secret (same for all peers) |
| `BOOTSTRAP` | Optional | - | Bootstrap peer address for joining cluster |
| `IPFS_PROFILE` | Optional | `server` | IPFS profile optimization |
| `CLUSTER_PINSVCAPI_HTTPLISTENMULTIADDRESS` | Optional | `/ip4/0.0.0.0/tcp/9097` | Pin service REST endpoint |
| `CLUSTER_RESTAPI_HTTPLISTENMULTIADDRESS` | Optional | `/ip4/0.0.0.0/tcp/9094` | Cluster REST API endpoint |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 50 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 100 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 200 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (for cluster communication)
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- Distributed file storage
- IPFS cluster deployment
- Content pinning services
- Peer-to-peer file sharing

---

## 32. Mastodon

**Flist URL:** `https://hub.grid.tf/tf-official-apps/mastodon-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `LOCAL_DOMAIN` | Required | - | Domain name for Mastodon instance |
| `SUPERUSER_EMAIL` | Required | - | Admin email for Mastodon |
| `SUPERUSER_USERNAME` | Required | - | Admin username |
| `SUPERUSER_PASSWORD` | Required | - | Admin password |
| `IS_TF_CONNECT` | Optional | `false` | Enable ThreeFold Connect authenticator |
| `RELAYS_LINKS` | Optional | - | List of relay links for Fediverse |
| `SMTP_SERVER` | Optional | - | SMTP server address |
| `SMTP_PORT` | Optional | `587` | SMTP server port |
| `SMTP_LOGIN` | Optional | - | SMTP username |
| `SMTP_PASSWORD` | Optional | - | SMTP password |
| `SMTP_FROM_ADDRESS` | Optional | - | SMTP from address |

### Proposed Specs

- **Medium:** CPU: 2, Memory: 4 GB, Disk: 100 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 200 GB
- **Extra Large:** CPU: 8, Memory: 16 GB, Disk: 500 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/var/lib/docker` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Required (for Fediverse communication)
- **Gateway/Domain:** Yes (required)
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- Social media platform deployment
- Fediverse instance hosting
- Community communication platform
- Decentralized social networking

---

## 33. Monitor

**Flist URL:** `https://hub.grid.tf/tf-official-apps/monitor-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `PROM_TARGETS` | Optional | - | Comma-separated hosts for Prometheus to scrape |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- System monitoring and alerting
- Metrics collection and visualization
- Grafana dashboard deployment
- Prometheus monitoring setup

---

## 34. Nomad

**Flist URL:** `https://hub.grid.tf/tf-official-apps/nomad-latest.flist`

**Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |
| `FIRST_SERVER_IP` | Optional | - | Private IP of first server (for cluster joining) |
| `NOMAD_SERVERS` | Optional | `1` | Number of expected Nomad servers |

### Proposed Specs

- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB
- **Large:** CPU: 4, Memory: 8 GB, Disk: 100 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/data` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional

**Use Cases:**

- Container orchestration
- Workload scheduling
- Service deployment automation
- Distributed system management

---

## 35. Base OS Images (Additional)

**Flist URLs:**

- Alpine Linux: `https://hub.grid.tf/tf-official-apps/threefoldtech-alpine-3.flist` - **Entry Point:** `/entrypoint.sh`
- Arch Linux (Mycelium): `https://hub.grid.tf/tf-official-apps/arch-mycelium-latest.flist` - **Entry Point:** `/sbin/zinit init`
- CentOS 9: `https://hub.grid.tf/tf-official-apps/centos-9-latest.flist` - **Entry Point:** `/sbin/zinit init`
- Debian Linux: `https://hub.grid.tf/tf-official-apps/threefoldtech-debian-12.flist` - **Entry Point:** `/sbin/zinit init`
- Docker with SSH: `https://hub.grid.tf/tf-official-apps/docker-ssh-latest.flist` - **Entry Point:** `/sbin/zinit init`
- NixOS: `https://hub.grid.tf/tf-official-apps/nixos-latest.flist` - **Entry Point:** `/sbin/zinit init`

### Environment Variables

| Variable | Type | Proposed Default Value | Description |
|----------|------|---------|-------------|
| `SSH_KEY` | Required | - | SSH public key for access |

### Proposed Specs

- **Micro:** CPU: 1, Memory: 1 GB, Disk: 10 GB
- **Small:** CPU: 1, Memory: 2 GB, Disk: 25 GB
- **Medium:** CPU: 2, Memory: 4 GB, Disk: 50 GB

### Disk/Volumes

| Type | Size | Mount Point | Required |
|------|------|-------------|----------|
| SSD | Variable | `/` | Yes |

### Additional Requirements

- **IPv4/IPv6:** Optional
- **Gateway/Domain:** No
- **Planetary Network:** Optional
- **Mycelium:** Optional (except Arch Mycelium variant)

**Use Cases:**

- Custom application deployment
- Development environments
- Lightweight virtual machines
- Container hosting platforms

---

## Summary

This catalog contains **35 different solutions** available on the ThreeFold Grid, ranging from blockchain nodes to full-featured applications, base operating systems, and virtual machines. Each solution has specific requirements for:

- **Compute Resources:** CPU, Memory, and Storage
- **Network Configuration:** IPv4, IPv6, Planetary Network, Mycelium
- **Gateway/Domain:** Some solutions require domain configuration
- **Additional Services:** SMTP, GPU support, etc.

All solutions support SSH key authentication and can be deployed with various resource configurations to match different use cases.
