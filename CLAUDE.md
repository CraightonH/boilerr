# CLAUDE.md - Agent Guidelines for Boilerr

## Project Overview

Boilerr is a Kubernetes operator for managing Steam dedicated game servers. It uses Kubebuilder to scaffold CRDs and controllers, allowing users to deploy game servers via custom resources.

**Status:** Design phase - implementation starting from scratch following ROADMAP.md

## Architecture

- **Language:** Go
- **Framework:** Kubebuilder (controller-runtime)
- **CRDs:** Generic `SteamServer` + game-specific CRDs (ValheimServer, SatisfactoryServer, etc.)
- **Resources Generated:** StatefulSet, PVC, Service per game server

## Key Documents

- `DESIGN.md` - Architecture decisions, CRD schemas, reconciliation patterns
- `ROADMAP.md` - Phased task breakdown, current progress
- `ROADMAP_COMPLETED.md` - Completed roadmap items archive

## Development Workflow

**Before submitting PRs:**
- Move completed roadmap items from `ROADMAP.md` to `ROADMAP_COMPLETED.md`
- Keep ROADMAP.md focused on remaining work

**When adding new documentation files:**
- Update `paths-ignore` in GitHub Actions workflows (build.yml, test.yml, lint.yml, test-e2e.yml)
- Add new docs to ignore list to skip CI on docs-only changes

## Development Commands

```bash
# Generate CRD manifests and Go code
make manifests
make generate

# Run tests
make test

# Build operator binary
make build

# Build container image
make docker-build IMG=<tag>

# Deploy to cluster
make deploy IMG=<tag>

# Run locally against cluster
make run
```

## Code Structure (Kubebuilder conventions)

```
api/v1alpha1/           # CRD type definitions
  steamserver_types.go  # Generic SteamServer CRD
  valheimserver_types.go # Game-specific CRDs
  common_types.go       # Shared field types
internal/
  controller/           # Reconciliation logic
  resources/            # K8s resource builders (StatefulSet, Service, PVC)
  steamcmd/             # SteamCMD script generation
config/
  crd/                  # Generated CRD YAML
  rbac/                 # RBAC manifests
  manager/              # Operator deployment
```

## Coding Guidelines

### Go Style
- Follow standard Go conventions and `golangci-lint` rules
- Use controller-runtime patterns for reconciliation
- Return `ctrl.Result{}` with appropriate requeue behavior
- Always set owner references on child resources

### Kubernetes Patterns
- Use `StatefulSet` (not Deployment) for game servers - stable network identity matters
- PVCs should be created separately, not via volumeClaimTemplates, for easier management
- Support all service types: LoadBalancer, NodePort, ClusterIP
- Handle both TCP and UDP ports correctly

### CRD Design
- Embed common fields (storage, resources, serviceType) in shared types
- Use `secretKeyRef` pattern for sensitive values (passwords, Steam credentials)
- Status should reflect: state, address, ports, lastUpdated, error messages
- Add printer columns for `kubectl get` output

### Testing
- Unit test resource builders with table-driven tests
- Use `envtest` for controller integration tests
- E2E tests with `kind` cluster
- Mock external dependencies (Steam API if added later)

## Common Tasks

### Adding a New Game CRD

1. Define types in `api/v1alpha1/<game>server_types.go`
2. Add game-specific fields (world name, mods, etc.)
3. Embed `SteamServerSpec` for common fields
4. Run `make manifests generate`
5. Add controller or extend existing reconciler
6. Add game config defaults (app ID, default ports, startup command template)

### Reconciler Pattern

```go
func (r *SteamServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch the CR
    // 2. Handle deletion (finalizers)
    // 3. Build desired resources
    // 4. Create or update each resource
    // 5. Update CR status
    // 6. Return (requeue if needed)
}
```

## Important Constraints

- **No hardcoded game configs in controller** - use CRD fields or ConfigMaps
- **Graceful shutdown** - game servers need time to save; handle SIGTERM properly
- **Resource cleanup** - use owner references so child resources are garbage collected
- **Idempotent reconciliation** - same input should produce same output, safe to re-run

## Security Notes

- Steam credentials (if non-anonymous) must come from Secrets, never inline
- Server passwords must use secretKeyRef
- RCON passwords (future) same treatment
- Operator needs minimal RBAC - only what's necessary for managed resources

## Out of Scope (for now)

Per DESIGN.md non-goals:
- Automatic backups (Phase 9)
- Player-aware updates (Phase 9)
- Multi-server clustering (e.g., ARK clusters)
