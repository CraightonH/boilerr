# CLAUDE.md - Agent Guidelines for Boilerr

## Project Overview

Boilerr is a Kubernetes operator for managing Steam dedicated game servers. It uses Kubebuilder to scaffold CRDs and controllers, allowing users to deploy game servers via custom resources.

**Status:** Phase 2 in progress - refactoring to GameDefinition architecture

## Architecture

- **Language:** Go
- **Framework:** Kubebuilder (controller-runtime)
- **CRDs:** `GameDefinition` (defines games) + `SteamServer` (user deploys servers)
- **Container:** Uses `steamcmd/steamcmd:ubuntu-22` directly - no custom images per game
- **Resources Generated:** StatefulSet, PVC, Service, ConfigMap per game server

### Two-CRD Pattern

1. **GameDefinition** (cluster-scoped) - Defines how to install/run a game. Maintained by operator/community. Bundled via Helm.
2. **SteamServer** (namespaced) - User creates to deploy a server. References a GameDefinition by name.

This enables extensibility: adding new games = YAML file, not Go code.

## Key Documents

- `DESIGN.md` - Architecture decisions, CRD schemas, reconciliation patterns
- `ROADMAP.md` - Phased task breakdown, current progress
- `ROADMAP_COMPLETED.md` - Completed roadmap items archive
- `docs/REFACTOR_PLAN.md` - **ACTIVE** - Detailed refactor plan for GameDefinition migration

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
api/v1alpha1/               # CRD type definitions
  gamedefinition_types.go   # GameDefinition CRD
  steamserver_types.go      # SteamServer CRD
  common_types.go           # Shared field types
internal/
  controller/               # Reconciliation logic
  resources/                # K8s resource builders (StatefulSet, Service, PVC)
  steamcmd/                 # SteamCMD command builder (NOT script generator)
  config/                   # Config interpolation utilities
config/
  crd/                      # Generated CRD YAML
  rbac/                     # RBAC manifests
  manager/                  # Operator deployment
charts/boilerr/             # Helm chart
  templates/gamedefinitions/ # Bundled GameDefinitions
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

### Adding a New Game (GameDefinition)

No Go code needed! Create a YAML file:

1. Create `charts/boilerr/templates/gamedefinitions/<game>.yaml`
2. Define: `appId`, `command`, `args`, `ports`, `configSchema`
3. Test with `kubectl apply -f` and create a SteamServer referencing it
4. Add to Helm chart values for selective enablement

See `docs/REFACTOR_PLAN.md` for GameDefinition schema details.

### Reconciler Pattern

```go
func (r *SteamServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch the SteamServer CR
    // 2. Fetch the referenced GameDefinition (by spec.game)
    // 3. Validate config against GameDefinition.configSchema
    // 4. Handle deletion (finalizers)
    // 5. Build desired resources (pass both CRs to builders)
    // 6. Create or update each resource
    // 7. Update CR status
    // 8. Return (requeue if needed)
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
