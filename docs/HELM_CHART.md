# Helm Chart Development & Publishing

This guide covers developing, testing, and publishing the Boilerr Helm chart.

## Chart Structure

```
charts/boilerr/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default configuration values
├── README.md              # Chart documentation
├── .helmignore            # Files to ignore when packaging
├── templates/             # Kubernetes resource templates
│   ├── _helpers.tpl       # Template helpers
│   ├── NOTES.txt          # Post-install notes
│   ├── namespace.yaml     # Namespace (if namespaceOverride)
│   ├── serviceaccount.yaml
│   ├── clusterrole.yaml
│   ├── clusterrolebinding.yaml
│   ├── deployment.yaml
│   ├── service.yaml       # Metrics service
│   └── gamedefinitions/   # GameDefinition resources
│       └── valheim.yaml
└── crds/                  # Custom Resource Definitions
    ├── boilerr.dev_gamedefinitions.yaml
    └── boilerr.dev_steamservers.yaml
```

## Local Development

### Prerequisites

- Helm 3.8+
- Kubernetes cluster (for testing)
- kubectl configured

### Testing the Chart

#### 1. Lint the Chart

```bash
helm lint charts/boilerr
```

Expected output:
```
==> Linting charts/boilerr
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
```

#### 2. Render Templates

```bash
# Render with default values
helm template boilerr charts/boilerr

# Render with custom values
helm template boilerr charts/boilerr \
  --set replicaCount=2 \
  --set gameDefinitions.include={valheim}

# Save rendered output
helm template boilerr charts/boilerr > /tmp/boilerr-rendered.yaml
kubectl apply --dry-run=client -f /tmp/boilerr-rendered.yaml
```

#### 3. Dry-Run Install

```bash
helm install boilerr charts/boilerr \
  --namespace boilerr-system \
  --create-namespace \
  --dry-run --debug
```

#### 4. Install Locally

```bash
# Install from local chart
helm install boilerr charts/boilerr \
  --namespace boilerr-system \
  --create-namespace

# Verify
kubectl get pods -n boilerr-system
helm status boilerr -n boilerr-system
```

#### 5. Upgrade Test

```bash
# Modify values.yaml or use --set
helm upgrade boilerr charts/boilerr \
  -n boilerr-system \
  --set replicaCount=2

# Check diff before upgrade
helm diff upgrade boilerr charts/boilerr -n boilerr-system
```

#### 6. Uninstall

```bash
helm uninstall boilerr -n boilerr-system
```

### Validating Values

Create test values files:

**values-test-minimal.yaml:**
```yaml
replicaCount: 1
resources:
  limits:
    cpu: 100m
    memory: 64Mi
gameDefinitions:
  include: [valheim]
```

Test:
```bash
helm template boilerr charts/boilerr -f charts/boilerr/values-test-minimal.yaml
```

**values-test-ha.yaml:**
```yaml
replicaCount: 3
podAntiAffinity:
  enabled: true
  type: requiredDuringSchedulingIgnoredDuringExecution
```

## Packaging

### Create Chart Package

```bash
# Package the chart
helm package charts/boilerr

# Output: boilerr-0.3.0.tgz
```

### Generate Chart Index

```bash
# Create repository index
helm repo index . --url https://craightonh.github.io/boilerr

# Generates index.yaml
```

## Publishing to GitHub Pages

### Setup

1. **Create `gh-pages` branch:**
   ```bash
   git checkout --orphan gh-pages
   git rm -rf .
   git commit --allow-empty -m "Initialize gh-pages"
   git push origin gh-pages
   git checkout main
   ```

2. **Enable GitHub Pages:**
   - Go to repository Settings → Pages
   - Source: Deploy from branch `gh-pages`
   - Save

### Automated Publishing Workflow

Create `.github/workflows/helm-publish.yml`:

```yaml
name: Publish Helm Chart

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install Helm
        uses: azure/setup-helm@v4
        with:
          version: 'latest'

      - name: Package Chart
        run: |
          helm package charts/boilerr -d .deploy

      - name: Checkout gh-pages
        uses: actions/checkout@v4
        with:
          ref: gh-pages
          path: gh-pages

      - name: Update Helm Repository
        run: |
          cp .deploy/*.tgz gh-pages/
          helm repo index gh-pages --url https://craightonh.github.io/boilerr

      - name: Commit and Push
        working-directory: gh-pages
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add .
          git commit -m "Publish Helm chart ${{ github.ref_name }}"
          git push
```

### Manual Publishing

```bash
# 1. Package chart
helm package charts/boilerr

# 2. Checkout gh-pages
git checkout gh-pages

# 3. Move package
mv boilerr-*.tgz .

# 4. Update index
helm repo index . --url https://craightonh.github.io/boilerr --merge index.yaml

# 5. Commit and push
git add .
git commit -m "Publish chart version X.Y.Z"
git push origin gh-pages

# 6. Return to main
git checkout main
```

## Using the Published Chart

Once published, users can install via:

```bash
# Add repository
helm repo add boilerr https://craightonh.github.io/boilerr
helm repo update

# Install
helm install boilerr boilerr/boilerr \
  --namespace boilerr-system \
  --create-namespace
```

## Versioning

### Chart Version vs App Version

- **Chart Version** (`version` in Chart.yaml): Helm chart version
- **App Version** (`appVersion` in Chart.yaml): Boilerr operator version

Both should be updated for releases:

```yaml
# Chart.yaml
version: 0.3.0      # Chart version
appVersion: "0.3.0" # Operator version
```

### Version Bumping Strategy

- **Patch (0.3.0 → 0.3.1)**: Bug fixes, minor template changes
- **Minor (0.3.0 → 0.4.0)**: New features, new values, backward compatible
- **Major (0.3.0 → 1.0.0)**: Breaking changes, incompatible upgrades

### Updating Versions

```bash
# Update Chart.yaml
sed -i 's/version: .*/version: 0.4.0/' charts/boilerr/Chart.yaml
sed -i 's/appVersion: .*/appVersion: "0.4.0"/' charts/boilerr/Chart.yaml

# Update default image tag in values.yaml (if needed)
sed -i 's/tag: .*/tag: "v0.4.0"/' charts/boilerr/values.yaml
```

## Adding New GameDefinitions

When a new game is added:

1. **Copy GameDefinition template:**
   ```bash
   cp gamedefinitions/newgame.yaml charts/boilerr/templates/gamedefinitions/
   ```

2. **Wrap with conditional:**
   ```yaml
   {{- if and .Values.gameDefinitions.enabled (has "newgame" (include "boilerr.enabledGameDefinitions" . | fromJson)) }}
   ---
   # GameDefinition content
   {{- end }}
   ```

3. **Update `_helpers.tpl`:**
   ```go
   {{- $allGames := list "valheim" "newgame" }}
   ```

4. **Update chart README:**
   Add to GameDefinitions section in `charts/boilerr/README.md`

5. **Bump chart version** (minor version)

## Testing Checklist

Before publishing a new chart version:

- [ ] `helm lint charts/boilerr` passes
- [ ] `helm template charts/boilerr` renders without errors
- [ ] Dry-run install succeeds
- [ ] Install on test cluster succeeds
- [ ] Post-install NOTES.txt displays correctly
- [ ] All GameDefinitions install when enabled
- [ ] Selective game installation works (`include`/`exclude`)
- [ ] CRDs are installed
- [ ] Operator pod starts successfully
- [ ] Metrics service created (if enabled)
- [ ] RBAC permissions are correct
- [ ] Upgrade from previous version works
- [ ] Uninstall cleans up resources (except CRDs if `keep: true`)
- [ ] Chart README is accurate
- [ ] values.yaml documented
- [ ] NOTES.txt reflects actual installed resources

## Troubleshooting

### Chart Lint Warnings

**Missing icon:**
```
[INFO] Chart.yaml: icon is recommended
```

Add icon URL to Chart.yaml:
```yaml
icon: https://raw.githubusercontent.com/CraightonH/boilerr/main/assets/logo.png
```

### Template Rendering Errors

**Undefined variable:**
```
Error: template: boilerr/templates/deployment.yaml:10:14: executing "boilerr/templates/deployment.yaml" at <.Values.foo>: nil pointer evaluating interface {}.foo
```

Add default value in values.yaml or use conditional:
```yaml
{{- if .Values.foo }}
  foo: {{ .Values.foo }}
{{- end }}
```

### CRD Installation Issues

If CRDs aren't installing:
- Check `crds.install: true` in values
- Verify CRD files are in `crds/` directory
- Ensure CRD files don't have Helm templates (CRDs are installed as-is)

### GameDefinition Not Installing

Debug:
```bash
# Check what GameDefinitions would be installed
helm template boilerr charts/boilerr \
  --set gameDefinitions.enabled=true \
  --set gameDefinitions.include={valheim} \
  | grep "kind: GameDefinition" -A 5
```

## Best Practices

1. **Pin versions in production:** Use specific chart versions, not `latest`
2. **Use values files:** Store configuration in values.yaml files
3. **Test upgrades:** Always test upgrade path from previous version
4. **Document breaking changes:** Note in Chart.yaml annotations
5. **Keep CRDs on uninstall:** Set `crds.keep: true` by default
6. **Semantic versioning:** Follow semver for chart versions
7. **Validate schemas:** Use JSON schema for values validation (future)

## References

- [Helm Chart Best Practices](https://helm.sh/docs/chart_best_practices/)
- [Helm Template Guide](https://helm.sh/docs/chart_template_guide/)
- [GitHub Pages Helm Repository](https://helm.sh/docs/topics/chart_repository/#github-pages-example)
