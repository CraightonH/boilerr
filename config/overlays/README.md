# Kustomize Overlays

This directory contains kustomize overlays for different deployment scenarios.

## Available Overlays

### Production (`production/`)

**Use case:** Production deployments requiring high availability and reliability.

**Features:**
- 2 replicas with pod anti-affinity
- Higher resource limits (1 CPU, 512Mi memory)
- Uses `latest` stable tag
- Deploys to `boilerr-system` namespace

**Deploy:**
```bash
kubectl apply -k config/overlays/production
```

**Verify:**
```bash
kubectl get deployment -n boilerr-system
kubectl get pods -n boilerr-system
```

---

### Development (`development/`)

**Use case:** Development/testing environments, faster iteration cycles.

**Features:**
- 1 replica
- Moderate resources (500m CPU, 256Mi memory)
- Uses `main` branch image (latest development build)
- Debug logging enabled
- Deploys to `boilerr-dev` namespace

**Deploy:**
```bash
kubectl apply -k config/overlays/development
```

**Verify:**
```bash
kubectl get deployment -n boilerr-dev
kubectl logs -n boilerr-dev -l app.kubernetes.io/name=boilerr -f
```

---

### Minimal (`minimal/`)

**Use case:** Resource-constrained environments, testing, personal clusters.

**Features:**
- 1 replica
- Minimal resources (200m CPU, 128Mi memory)
- Uses `latest` stable tag
- Deploys to `boilerr-system` namespace

**Deploy:**
```bash
kubectl apply -k config/overlays/minimal
```

**Verify:**
```bash
kubectl get deployment -n boilerr-system
kubectl top pod -n boilerr-system
```

---

## Customization

To create a custom overlay:

1. Create a new directory:
   ```bash
   mkdir config/overlays/my-custom
   ```

2. Create `kustomization.yaml`:
   ```yaml
   apiVersion: kustomize.config.k8s.io/v1beta1
   kind: Kustomization

   namespace: my-namespace

   resources:
     - ../../default

   patches:
     - path: deployment_patch.yaml

   images:
     - name: controller
       newName: ghcr.io/craightonh/boilerr
       newTag: v0.3.0  # Pin to specific version
   ```

3. Create `deployment_patch.yaml` with your customizations

4. Apply:
   ```bash
   kubectl apply -k config/overlays/my-custom
   ```

## Image Tags

- `latest` - Latest stable release (production)
- `vX.Y.Z` - Specific version tag (recommended for production)
- `main` - Latest development build (unstable)
- `<commit-sha>` - Specific commit build

**Production recommendation:** Pin to specific version tags:
```yaml
images:
  - name: controller
    newName: ghcr.io/craightonh/boilerr
    newTag: v0.3.0  # Pin version
```

## Common Patches

### Change namespace
```yaml
# kustomization.yaml
namespace: my-custom-namespace
```

### Change replicas
```yaml
# deployment_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
spec:
  replicas: 3
```

### Add resource limits
```yaml
# deployment_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
spec:
  template:
    spec:
      containers:
        - name: manager
          resources:
            limits:
              cpu: 2000m
              memory: 1Gi
            requests:
              cpu: 200m
              memory: 512Mi
```

### Add tolerations for dedicated nodes
```yaml
# deployment_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
spec:
  template:
    spec:
      tolerations:
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
          effect: NoSchedule
```

### Pin to specific nodes
```yaml
# deployment_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
spec:
  template:
    spec:
      nodeSelector:
        node-role.kubernetes.io/operator: "true"
```

## Validation

Test overlay rendering without applying:
```bash
kubectl kustomize config/overlays/production
```

Validate YAML syntax:
```bash
kubectl kustomize config/overlays/production | kubectl apply --dry-run=client -f -
```

## Troubleshooting

**Issue:** Image pull failures
```bash
# Check image exists
docker pull ghcr.io/craightonh/boilerr:latest

# Check cluster can pull
kubectl run test --image=ghcr.io/craightonh/boilerr:latest --rm -it -- /manager --version
```

**Issue:** RBAC errors
```bash
# Check service account
kubectl get sa -n boilerr-system

# Check RBAC bindings
kubectl get clusterrolebinding | grep boilerr
```

**Issue:** CRDs not installed
```bash
# Check CRDs exist
kubectl get crd | grep boilerr

# Reinstall
kubectl apply -k config/overlays/production
```
