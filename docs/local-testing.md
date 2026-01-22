# Local Testing Guide

## Current Status (2026-01-21)

✅ **Working:**
- CRD generation and installation
- Controller reconciliation loop
- ConfigValue clean syntax (direct strings + secret refs)
- Resource generation (StatefulSet, Service, PVC)
- GameDefinition validation

⚠️ **Known Requirements:**
- StorageClass must be specified in SteamServer spec
- PVC will remain Pending without storage class

## Prerequisites

- Kubernetes cluster with storage provisioner
- kubectl configured
- Storage class available (check with `kubectl get storageclass`)

## Quick Start (Local Controller)

**Recommended for testing - no Docker required:**

```bash
# 1. Install CRDs
make install

# 2. Verify CRDs installed
kubectl get crd | grep boilerr
# Should show:
# gamedefinitions.boilerr.dev
# steamservers.boilerr.dev

# 3. Create namespace
kubectl create namespace games

# 4. Create secret for server password
kubectl create secret generic valheim-secrets \
  --from-literal=server-password=test123 \
  -n games

# 5. Check available storage classes
kubectl get storageclass
# Note the NAME of your storage class (e.g., nfs-client, longhorn, etc.)

# 6. Edit sample to add storage class
# Edit config/samples/boilerr.dev_v1alpha1_steamserver_valheim.yaml
# Add under spec.storage:
#   storage:
#     size: 30Gi
#     storageClassName: nfs-client  # Use your storage class name

# 7. Apply GameDefinition and SteamServer
kubectl apply -f gamedefinitions/valheim.yaml

# 8. If controller is deployed, scale it down to run locally
kubectl scale deployment boilerr-controller-manager -n boilerr-system --replicas=0

# 9. Run controller locally (keep this terminal open)
make run

# In another terminal:
# 10. Watch resources get created
kubectl get steamserver,statefulset,service,pvc -n games -w
```

## Deployment Flow

1. **CRDs installed** → GameDefinition and SteamServer types available
2. **GameDefinition applied** → Operator validates and marks ready
3. **SteamServer applied** → Operator reconciles:
   - Creates PVC (requires storageClassName)
   - Creates StatefulSet with init container (downloads game via steamcmd)
   - Creates Service (LoadBalancer or NodePort)
   - Updates SteamServer status

## Resource Details

### StatefulSet Structure

```yaml
initContainers:
- name: steamcmd              # Downloads game files
  image: steamcmd/steamcmd:ubuntu-22
  command: [steamcmd]
  args: [+login, anonymous, +app_update, 896660, validate, +quit]

containers:
- name: gameserver            # Runs the game server
  image: steamcmd/steamcmd:ubuntu-22
  command: [./valheim_server.x86_64]
  args: [-nographics, -batchmode, -port, 2456, ...]
  env:
  - name: CONFIG_PASSWORD     # From secret
    valueFrom:
      secretKeyRef:
        name: valheim-secrets
        key: server-password
```

### Config Value Syntax

**Clean syntax (recommended):**
```yaml
config:
  serverName: "Vikings Only"        # Direct string
  worldName: "Midgard"
  password:                          # Secret reference
    secretKeyRef:
      name: valheim-secrets
      key: server-password
  public: "0"
```

**Structured syntax (also works):**
```yaml
config:
  serverName:
    value: "Vikings Only"
  worldName:
    value: "Midgard"
```

## Storage Class Configuration

**IMPORTANT:** PVC requires a storageClassName to provision storage.

### Check Available Storage Classes
```bash
kubectl get storageclass
```

### Update SteamServer Spec
```yaml
spec:
  storage:
    size: 30Gi
    storageClassName: nfs-client  # Replace with your storage class
```

Common storage classes:
- `nfs-client` - NFS provisioner
- `longhorn` - Longhorn storage
- `local-path` - Local path provisioner (kind/k3s)
- `standard` - Cloud provider default

### Example: Complete SteamServer

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: my-valheim-server
  namespace: games
spec:
  gameDefinition: valheim

  config:
    serverName: "Vikings Valhalla"
    worldName: "Midgard"
    password:
      secretKeyRef:
        name: valheim-secrets
        key: server-password
    public: "0"
    crossplay: "false"

  storage:
    size: 30Gi
    storageClassName: nfs-client  # REQUIRED

  serviceType: LoadBalancer
```

## Monitoring

### Check Status
```bash
# SteamServer status
kubectl get steamserver -n games
# Shows: GAME, STATE, ADDRESS

# Detailed status
kubectl get steamserver my-valheim-server -n games -o yaml | grep -A 10 status:
```

### Server States
- `Pending` - Pod scheduled, waiting for resources
- `Installing` - Init container downloading game files (5-15 min for Valheim)
- `Starting` - Game server process starting
- `Running` - Server ready
- `Error` - Check logs for issues

### View Logs

```bash
# Controller logs (if running locally)
# Output appears in the terminal where `make run` is running

# Init container (steamcmd download)
kubectl logs -n games my-valheim-server-0 -c steamcmd

# Game server
kubectl logs -n games my-valheim-server-0 -c gameserver -f

# All containers
kubectl logs -n games my-valheim-server-0 --all-containers
```

## Accessing the Server

### LoadBalancer Service
```bash
kubectl get svc -n games my-valheim-server
# Note EXTERNAL-IP and PORT
# Connect to: <external-ip>:2456
```

### NodePort Service
```bash
kubectl get svc -n games my-valheim-server
# Note NodePort (e.g., 31923)
# Connect to: <node-ip>:31923
```

## Troubleshooting

### PVC Pending - No Storage Class

**Symptom:**
```
PVC: Pending
Pod: 0/1 - unbound immediate PersistentVolumeClaims
```

**Fix:** Add storageClassName to SteamServer spec:
```yaml
spec:
  storage:
    size: 30Gi
    storageClassName: nfs-client
```

### GameDefinition Not Ready

```bash
kubectl get gamedefinition valheim -o yaml
# Check status.ready and status.message
```

**Common causes:**
- ConfigSchema validation failed
- Required fields missing
- AppId invalid

### Pod Stuck in Init:0/1

```bash
kubectl logs -n games my-valheim-server-0 -c steamcmd
```

**Expected:** steamcmd downloading game files (1-2 GB, takes 5-15 minutes)

### Config Values Not Interpolating

Check controller logs for interpolation errors:
```bash
# If running locally, check terminal output
# If deployed:
kubectl logs -n boilerr-system deployment/boilerr-controller-manager
```

### Clean Syntax Not Working

Ensure CRDs are updated:
```bash
# Regenerate and reinstall CRDs
make manifests
kubectl replace -f config/crd/bases/boilerr.dev_steamservers.yaml
```

The ConfigValue type needs `x-kubernetes-preserve-unknown-fields: true` to accept both strings and objects.

## Cleanup

```bash
# Delete SteamServer (deletes StatefulSet, Service, PVC via owner references)
kubectl delete steamserver my-valheim-server -n games

# Delete namespace
kubectl delete namespace games

# If controller running locally, Ctrl+C to stop

# If deployed, undeploy controller
make undeploy

# Uninstall CRDs
make uninstall
```

## Development Workflow

```bash
# 1. Make code changes
# 2. Ctrl+C to stop local controller
# 3. make run (automatically rebuilds)
# 4. Test changes

# No need to rebuild Docker images or redeploy when running locally
```
