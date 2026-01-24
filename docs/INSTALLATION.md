# Installation Guide

This guide covers different methods for installing Boilerr on your Kubernetes cluster.

## Prerequisites

- Kubernetes cluster v1.24+ (tested on v1.24-1.31)
- `kubectl` configured to access your cluster
- Cluster-admin permissions (for CRD and RBAC installation)

### Verify Prerequisites

```bash
# Check Kubernetes version
kubectl version --short

# Verify cluster access
kubectl cluster-info

# Check permissions
kubectl auth can-i create customresourcedefinitions
```

---

## Installation Methods

### Method 1: Quick Install (Recommended)

Install the latest release with a single command:

```bash
kubectl apply -f https://github.com/CraightonH/boilerr/releases/latest/download/install.yaml
```

This installs:
- Custom Resource Definitions (GameDefinition, SteamServer)
- Namespace (`boilerr-system`)
- RBAC (ServiceAccount, ClusterRole, ClusterRoleBinding)
- Operator Deployment

**Verify installation:**
```bash
# Check operator is running
kubectl get deployment -n boilerr-system
kubectl get pods -n boilerr-system

# Check CRDs are installed
kubectl get crd | grep boilerr
```

**Expected output:**
```
NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
boilerr-controller-manager            1/1     1            1           30s

NAME                                        CREATED AT
gamedefinitions.boilerr.dev                 2024-01-23T10:00:00Z
steamservers.boilerr.dev                    2024-01-23T10:00:00Z
```

---

### Method 2: Specific Version

Pin to a specific version for production stability:

```bash
VERSION=v0.3.0
kubectl apply -f https://github.com/CraightonH/boilerr/releases/download/${VERSION}/install.yaml
```

---

### Method 3: Kustomize Overlays

For custom configurations, use kustomize overlays:

#### Production Deployment (HA)
```bash
# Clone repository
git clone https://github.com/CraightonH/boilerr.git
cd boilerr

# Deploy production overlay
kubectl apply -k config/overlays/production
```

Features:
- 2 replicas with pod anti-affinity
- Higher resource limits
- Production-optimized settings

#### Development Deployment
```bash
kubectl apply -k config/overlays/development
```

Features:
- Debug logging
- Lower resources
- Uses development image builds

#### Minimal Deployment
```bash
kubectl apply -k config/overlays/minimal
```

Features:
- Single replica
- Minimal resource footprint
- Suitable for small clusters

See [config/overlays/README.md](../config/overlays/README.md) for more details.

---

### Method 4: Helm Chart

**Recommended for production** - Easiest way to install with customizable values.

```bash
# Add Helm repository
helm repo add boilerr https://craightonh.github.io/boilerr
helm repo update

# Install with default values
helm install boilerr boilerr/boilerr \
  --namespace boilerr-system \
  --create-namespace

# Or install with custom values
helm install boilerr boilerr/boilerr \
  --namespace boilerr-system \
  --create-namespace \
  --set replicaCount=2 \
  --set resources.limits.memory=512Mi \
  --set gameDefinitions.include={valheim}
```

**Verify installation:**
```bash
helm status boilerr -n boilerr-system
helm get values boilerr -n boilerr-system
```

See [charts/boilerr/README.md](../charts/boilerr/README.md) for all configuration options.

---

## Post-Installation

### 1. Verify Operator Logs

```bash
kubectl logs -n boilerr-system -l app.kubernetes.io/name=boilerr -f
```

Expected:
```
INFO    setup    starting manager
INFO    controller-runtime.metrics    Metrics server is starting to listen
INFO    Starting server
```

### 2. Install GameDefinitions

GameDefinitions are bundled separately. Install included definitions:

```bash
# Install Valheim GameDefinition
kubectl apply -f https://raw.githubusercontent.com/CraightonH/boilerr/main/gamedefinitions/valheim.yaml

# Verify
kubectl get gamedefinitions
```

**Or install from local clone:**
```bash
kubectl apply -f gamedefinitions/valheim.yaml
kubectl apply -f gamedefinitions/TEMPLATE.yaml  # Template for custom games
```

### 3. Deploy a Test Server

Create a SteamServer to verify everything works:

```bash
# Create secret for server password
kubectl create secret generic valheim-secrets \
  --from-literal=server-password='MySecurePass123' \
  -n default

# Create SteamServer
cat <<EOF | kubectl apply -f -
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: test-valheim
  namespace: default
spec:
  gameDefinition: valheim
  config:
    serverName: "Test Server"
    worldName: "TestWorld"
    password:
      secretKeyRef:
        name: valheim-secrets
        key: server-password
    public: "0"
  storage:
    size: 10Gi
  serviceType: NodePort
EOF
```

**Check status:**
```bash
# Watch server creation
kubectl get steamserver test-valheim -w

# Check resources
kubectl get statefulset,service,pvc -l app=test-valheim

# Check server logs
kubectl logs -l app=test-valheim -f
```

---

## Configuration

### Resource Limits

Default operator resources:
- **Requests:** 10m CPU, 64Mi memory
- **Limits:** 500m CPU, 128Mi memory

Adjust via kustomize patch or Helm values (when available).

### Image Configuration

By default, the operator pulls from `ghcr.io/craightonh/boilerr:latest`.

**Use Docker Hub instead:**
```bash
# Edit install.yaml before applying
sed -i 's|ghcr.io/craightonh/boilerr|craighton/boilerr|' install.yaml
kubectl apply -f install.yaml
```

**Or with kustomize:**
```yaml
# config/overlays/my-custom/kustomization.yaml
images:
  - name: controller
    newName: craighton/boilerr
    newTag: v0.3.0
```

### Node Affinity

Operator supports multi-arch by default (amd64, arm64). To restrict:

```yaml
# deployment_patch.yaml
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                - key: kubernetes.io/arch
                  operator: In
                  values:
                    - amd64  # Only amd64 nodes
```

---

## Upgrading

### Upgrade Operator

```bash
# Update to latest
kubectl apply -f https://github.com/CraightonH/boilerr/releases/latest/download/install.yaml

# Or specific version
kubectl apply -f https://github.com/CraightonH/boilerr/releases/download/v0.4.0/install.yaml
```

**Check rollout:**
```bash
kubectl rollout status deployment/boilerr-controller-manager -n boilerr-system
```

### Upgrade CRDs

CRDs are updated automatically when applying install.yaml. To upgrade manually:

```bash
kubectl apply -f config/crd/bases/boilerr.dev_gamedefinitions.yaml
kubectl apply -f config/crd/bases/boilerr.dev_steamservers.yaml
```

### Upgrade GameDefinitions

```bash
kubectl apply -f gamedefinitions/valheim.yaml
```

Existing SteamServers will reconcile with updated GameDefinition specs.

---

## Uninstallation

### Delete All SteamServers First

**Important:** Delete SteamServers before uninstalling the operator to ensure cleanup.

```bash
# List all servers
kubectl get steamservers -A

# Delete specific server
kubectl delete steamserver test-valheim -n default

# Delete all servers (careful!)
kubectl delete steamservers --all -A
```

### Uninstall Operator

```bash
# Delete operator and RBAC
kubectl delete -f https://github.com/CraightonH/boilerr/releases/latest/download/install.yaml

# Or with kustomize
kubectl delete -k config/overlays/production
```

### Delete CRDs (Optional)

**Warning:** This deletes all GameDefinition and SteamServer resources cluster-wide.

```bash
kubectl delete crd gamedefinitions.boilerr.dev
kubectl delete crd steamservers.boilerr.dev
```

### Delete Namespace

```bash
kubectl delete namespace boilerr-system
```

---

## Troubleshooting

### Operator Crashes on Startup

**Check logs:**
```bash
kubectl logs -n boilerr-system -l app.kubernetes.io/name=boilerr
```

**Common causes:**
- Missing CRDs → `kubectl get crd | grep boilerr`
- RBAC issues → `kubectl get clusterrole manager-role`
- Image pull failures → `kubectl describe pod -n boilerr-system`

**Fix:**
```bash
# Reinstall
kubectl delete -f install.yaml
kubectl apply -f install.yaml
```

### SteamServer Stuck in Pending

**Check GameDefinition:**
```bash
kubectl get gamedefinition valheim
```

If not found, install it:
```bash
kubectl apply -f gamedefinitions/valheim.yaml
```

**Check operator logs:**
```bash
kubectl logs -n boilerr-system -l app.kubernetes.io/name=boilerr -f
```

### RBAC Permission Denied

Ensure you have cluster-admin:
```bash
kubectl auth can-i create customresourcedefinitions
kubectl auth can-i create clusterroles
```

If denied, contact your cluster administrator.

### Image Pull Failures

**Verify image exists:**
```bash
docker pull ghcr.io/craightonh/boilerr:latest
```

**Check node connectivity:**
```bash
kubectl run test --image=ghcr.io/craightonh/boilerr:latest --rm -it -- /manager --help
```

**Use Docker Hub mirror:**
```bash
# Edit deployment
kubectl set image deployment/boilerr-controller-manager \
  manager=craighton/boilerr:latest \
  -n boilerr-system
```

---

## Security Considerations

### Pod Security Standards

Operator is configured for **restricted** Pod Security Standards:
- Runs as non-root (UID 65532)
- Read-only root filesystem
- No privilege escalation
- All capabilities dropped

Namespace enforces restricted policy:
```yaml
labels:
  pod-security.kubernetes.io/enforce: restricted
```

### RBAC Principle of Least Privilege

Operator only has permissions for:
- GameDefinition and SteamServer CRDs (cluster-scoped)
- StatefulSets, Services, PVCs, ConfigMaps (namespaced)
- Pods (read-only, for status updates)

### Image Verification

Verify image signatures with cosign:
```bash
cosign verify ghcr.io/craightonh/boilerr:latest \
  --certificate-identity-regexp=https://github.com/CraightonH/boilerr \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

---

## Next Steps

- [Deploy your first game server](../README.md#quick-start)
- [Create custom GameDefinitions](../gamedefinitions/TEMPLATE.yaml)
- [Configure storage and resources](../DESIGN.md#resource-management)
- [Set up monitoring (Phase 8)](../ROADMAP.md#phase-8-observability--operations)

## Support

- [GitHub Issues](https://github.com/CraightonH/boilerr/issues)
- [Documentation](../README.md)
- [DESIGN.md](../DESIGN.md)
