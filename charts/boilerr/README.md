# Boilerr Helm Chart

A Helm chart for deploying the Boilerr operator - a Kubernetes operator for managing Steam dedicated game servers.

## TL;DR

```bash
helm repo add boilerr https://craightonh.github.io/boilerr
helm install boilerr boilerr/boilerr --namespace boilerr-system --create-namespace
```

## Prerequisites

- Kubernetes 1.24+
- Helm 3.8+
- Cluster-admin permissions (for CRD installation)

## Installing the Chart

```bash
# Add the Boilerr Helm repository
helm repo add boilerr https://craightonh.github.io/boilerr
helm repo update

# Install with default values
helm install boilerr boilerr/boilerr \
  --namespace boilerr-system \
  --create-namespace

# Install with custom values
helm install boilerr boilerr/boilerr \
  --namespace boilerr-system \
  --create-namespace \
  --set replicaCount=2 \
  --set resources.limits.memory=512Mi \
  --set gameDefinitions.include={valheim}
```

## Uninstalling the Chart

```bash
# Delete all SteamServers first (important!)
kubectl delete steamservers --all -A

# Uninstall the chart
helm uninstall boilerr -n boilerr-system

# Optionally delete CRDs (WARNING: deletes all GameDefinitions and SteamServers)
kubectl delete crd gamedefinitions.boilerr.dev steamservers.boilerr.dev
```

## Configuration

The following table lists the configurable parameters of the Boilerr chart and their default values.

### Operator Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Operator image repository | `ghcr.io/craightonh/boilerr` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag (defaults to chart appVersion) | `""` |
| `imagePullSecrets` | Image pull secrets | `[]` |
| `replicaCount` | Number of operator replicas | `1` |
| `namespaceOverride` | Override namespace (creates if set) | `""` |

### Controller Manager

| Parameter | Description | Default |
|-----------|-------------|---------|
| `controllerManager.leaderElection.enabled` | Enable leader election | `true` |
| `controllerManager.health.bindAddress` | Health probe bind address | `:8081` |
| `controllerManager.health.livenessProbe.initialDelaySeconds` | Liveness probe initial delay | `15` |
| `controllerManager.health.livenessProbe.periodSeconds` | Liveness probe period | `20` |
| `controllerManager.health.readinessProbe.initialDelaySeconds` | Readiness probe initial delay | `5` |
| `controllerManager.health.readinessProbe.periodSeconds` | Readiness probe period | `10` |
| `controllerManager.metrics.enabled` | Enable metrics server | `true` |
| `controllerManager.metrics.bindAddress` | Metrics bind address | `:8443` |
| `controllerManager.metrics.service.type` | Metrics service type | `ClusterIP` |
| `controllerManager.metrics.service.port` | Metrics service port | `8443` |
| `controllerManager.logging.level` | Log level (debug, info, warn, error) | `info` |
| `controllerManager.logging.development` | Development mode logging | `false` |

### Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `resources.requests.memory` | Memory request | `64Mi` |

### Pod Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podLabels` | Additional pod labels | `{}` |
| `podAnnotations` | Additional pod annotations | `{}` |
| `podSecurityContext.runAsNonRoot` | Run as non-root | `true` |
| `podSecurityContext.seccompProfile.type` | Seccomp profile | `RuntimeDefault` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `true` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.capabilities.drop` | Capabilities to drop | `["ALL"]` |
| `priorityClassName` | Priority class name | `""` |
| `terminationGracePeriodSeconds` | Termination grace period | `10` |

### Scheduling

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Tolerations | `[]` |
| `affinity` | Node affinity (defaults to multi-arch) | See values.yaml |
| `podAntiAffinity.enabled` | Enable pod anti-affinity | `false` |
| `podAntiAffinity.type` | Type of pod anti-affinity | `preferredDuringSchedulingIgnoredDuringExecution` |

### RBAC & Service Account

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name | `""` |
| `rbac.create` | Create RBAC resources | `true` |

### GameDefinitions

| Parameter | Description | Default |
|-----------|-------------|---------|
| `gameDefinitions.enabled` | Install bundled GameDefinitions | `true` |
| `gameDefinitions.include` | List of games to include (empty = all) | `[]` |
| `gameDefinitions.exclude` | List of games to exclude | `[]` |

### CRDs

| Parameter | Description | Default |
|-----------|-------------|---------|
| `crds.install` | Install CRDs | `true` |
| `crds.keep` | Keep CRDs on uninstall | `true` |

### Advanced

| Parameter | Description | Default |
|-----------|-------------|---------|
| `extraEnv` | Extra environment variables | `[]` |
| `extraVolumes` | Extra volumes | `[]` |
| `extraVolumeMounts` | Extra volume mounts | `[]` |

## Examples

### High Availability Deployment

```yaml
# values-ha.yaml
replicaCount: 2

resources:
  limits:
    cpu: 1000m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi

podAntiAffinity:
  enabled: true
  type: requiredDuringSchedulingIgnoredDuringExecution
```

```bash
helm install boilerr boilerr/boilerr -f values-ha.yaml
```

### Minimal Resource Footprint

```yaml
# values-minimal.yaml
replicaCount: 1

resources:
  limits:
    cpu: 200m
    memory: 128Mi
  requests:
    cpu: 10m
    memory: 64Mi

gameDefinitions:
  include:
    - valheim  # Only install Valheim
```

```bash
helm install boilerr boilerr/boilerr -f values-minimal.yaml
```

### Custom Image Registry

```yaml
# values-custom-registry.yaml
image:
  repository: myregistry.example.com/boilerr
  tag: v0.3.0
  pullPolicy: Always

imagePullSecrets:
  - name: registry-credentials
```

### Debug Mode

```yaml
# values-debug.yaml
controllerManager:
  logging:
    level: debug
    development: true

resources:
  limits:
    cpu: 500m
    memory: 256Mi
```

### Install Only Specific Games

```yaml
# values-specific-games.yaml
gameDefinitions:
  include:
    - valheim
    # Add more games here as they're added to the chart
```

## Upgrading

### Upgrade to Latest Version

```bash
helm repo update
helm upgrade boilerr boilerr/boilerr -n boilerr-system
```

### Upgrade with Value Changes

```bash
helm upgrade boilerr boilerr/boilerr \
  -n boilerr-system \
  --set replicaCount=3 \
  --reuse-values
```

### Upgrade CRDs

Helm does not upgrade CRDs automatically. To upgrade CRDs:

```bash
kubectl apply -f https://github.com/CraightonH/boilerr/releases/latest/download/crds.yaml
```

## Post-Installation

After installing the chart:

1. **Verify operator is running:**
   ```bash
   kubectl get pods -n boilerr-system
   ```

2. **Check available GameDefinitions:**
   ```bash
   kubectl get gamedefinitions
   ```

3. **Deploy a test server:**
   ```bash
   kubectl create secret generic valheim-secrets \
     --from-literal=server-password='TestPass123'

   cat <<EOF | kubectl apply -f -
   apiVersion: boilerr.dev/v1alpha1
   kind: SteamServer
   metadata:
     name: test-server
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
     storage:
       size: 10Gi
     serviceType: NodePort
   EOF
   ```

4. **Watch server creation:**
   ```bash
   kubectl get steamserver test-server -w
   ```

## Troubleshooting

### Operator CrashLoopBackOff

Check logs:
```bash
kubectl logs -n boilerr-system -l app.kubernetes.io/name=boilerr
```

Common causes:
- Missing CRDs
- RBAC issues
- Image pull failures

### CRDs Not Found

Reinstall CRDs:
```bash
helm upgrade boilerr boilerr/boilerr -n boilerr-system --force
```

### GameDefinitions Not Installing

Check values:
```bash
helm get values boilerr -n boilerr-system
```

Verify GameDefinitions are enabled:
```yaml
gameDefinitions:
  enabled: true
```

## Development

### Testing Locally

```bash
# Lint the chart
helm lint charts/boilerr

# Render templates
helm template boilerr charts/boilerr

# Dry run install
helm install boilerr charts/boilerr --dry-run --debug
```

### Package Chart

```bash
helm package charts/boilerr
```

## Support

- [GitHub Issues](https://github.com/CraightonH/boilerr/issues)
- [Documentation](https://github.com/CraightonH/boilerr/blob/main/README.md)
- [Installation Guide](https://github.com/CraightonH/boilerr/blob/main/docs/INSTALLATION.md)

## License

MIT - See [LICENSE](https://github.com/CraightonH/boilerr/blob/main/LICENSE)
