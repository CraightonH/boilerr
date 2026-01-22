# Project Status - 2026-01-21

## Summary

Boilerr is a Kubernetes operator for managing Steam dedicated game servers. The operator successfully reconciles GameDefinition and SteamServer CRDs into running game servers.

## ✅ Working Features

### Core Operator Functionality
- **CRD Generation**: GameDefinition and SteamServer CRDs generated via kubebuilder
- **Controller Reconciliation**: Both GameDefinition and SteamServer controllers functional
- **Resource Generation**: Automatically creates StatefulSet, Service, PVC from SteamServer specs
- **Config Interpolation**: Template values from GameDefinition args interpolated with user config
- **Secret References**: Config values support both direct strings and secretKeyRef
- **Status Updates**: SteamServer status reflects pod state (Pending, Installing, Starting, Running, Error)

### Clean Config Syntax
Custom `UnmarshalJSON` on ConfigValue allows clean YAML:
```yaml
config:
  serverName: "My Server"    # Direct string
  password:                  # Secret ref
    secretKeyRef:
      name: secrets
      key: password
```

Backward compatible with structured syntax:
```yaml
config:
  serverName:
    value: "My Server"
```

### Resource Architecture
- **Init Container**: steamcmd downloads game files once on startup
- **Main Container**: Runs game server process
- **PVC**: Persistent storage for game files and saves
- **Service**: LoadBalancer or NodePort for external access
- **Labels**: Standard Kubernetes labels + `boilerr.dev/game` for filtering

## ⚠️ Known Requirements

### Storage Class Required
PVCs need `storageClassName` specified to provision storage:
```yaml
spec:
  storage:
    size: 30Gi
    storageClassName: nfs-client  # REQUIRED - use your cluster's storage class
```

Without this, PVC remains Pending and pod won't schedule.

### Storage Class Field
The `StorageSpec` type already includes `storageClassName` field:
```go
type StorageSpec struct {
    Size             resource.Quantity `json:"size"`
    StorageClassName *string          `json:"storageClassName,omitempty"`
}
```

Users must populate this field based on their cluster's available storage classes.

## 📝 Implementation Details

### Field Refactoring
- Renamed `game` → `gameDefinition` throughout codebase for clarity
- Updated all controllers, builders, tests, and docs

### Test Coverage
- `api/v1alpha1`: 5.0% (ConfigValue unmarshaling)
- `internal/config`: 93.4% (config interpolation & validation)
- `internal/controller`: 64.7% (reconciliation logic)
- `internal/resources`: 82.5% (resource builders)
- `internal/steamcmd`: 100.0% (command generation)

All tests passing (`make test`).

### Files Modified Today
- `api/v1alpha1/common_types.go` - Added UnmarshalJSON to ConfigValue
- `api/v1alpha1/common_types_test.go` - Created tests for unmarshaling
- `api/v1alpha1/steamserver_types.go` - Renamed game → gameDefinition
- `internal/controller/steamserver_controller.go` - Updated field references
- `internal/resources/*.go` - Updated builders to use gameDefinition
- `internal/resources/*_test.go` - Updated test fixtures
- `config/crd/kustomization.yaml` - Created (was missing)
- `config/crd/bases/*.yaml` - Regenerated with updated schema
- `gamedefinitions/valheim.yaml` - Updated examples with storageClassName
- `config/samples/*.yaml` - Updated with storageClassName
- `docs/local-testing.md` - Complete rewrite with current state
- `docs/STATUS.md` - This file

## 🚀 Tested End-to-End

Successfully deployed to local cluster (`admin@home`):
1. ✅ CRDs installed
2. ✅ GameDefinition applied and validated
3. ✅ Secret created for password
4. ✅ SteamServer created with clean config syntax
5. ✅ Controller reconciled and created:
   - StatefulSet (with init + main containers)
   - Service (LoadBalancer, external IP assigned)
   - PVC (pending until storageClassName added)
6. ✅ Status updated: State=Pending, Address=192.168.1.152

### Current Test Status
- Controller running locally via `make run`
- Deployed controller scaled to 0 replicas
- Resources created in `games` namespace
- Waiting on PVC provisioning (storageClassName needed)

## 📋 Next Steps

### For Users
1. Choose storage class: `kubectl get storageclass`
2. Add to SteamServer spec: `storage.storageClassName: <name>`
3. Wait for steamcmd download (~1-2GB, 5-15 minutes for Valheim)
4. Connect to external IP on port 2456

### For Development
1. Consider default storage class annotation support
2. Add storage class validation/suggestions in status
3. Improve documentation for different cluster types (kind, k3s, cloud)
4. Add more game definitions (Satisfactory, 7 Days to Die, etc.)

## 📂 Project Structure

```
api/v1alpha1/              # CRD type definitions
├── gamedefinition_types.go
├── steamserver_types.go
└── common_types.go        # ConfigValue with UnmarshalJSON

internal/controller/       # Reconciliation logic
├── gamedefinition_controller.go
└── steamserver_controller.go

internal/resources/        # K8s resource builders
├── statefulset.go        # Init + main containers
├── service.go
├── pvc.go
└── *_test.go

internal/config/          # Config interpolation
└── interpolate.go        # {{.Config.foo}} → values

internal/steamcmd/        # SteamCMD command builder
└── command.go            # +login +app_update args

gamedefinitions/          # Bundled game configs
└── valheim.yaml

config/
├── crd/bases/           # Generated CRD YAML
├── samples/             # Example SteamServers
└── manager/             # Operator deployment

docs/
├── local-testing.md     # Deployment guide
├── STATUS.md            # This file
├── REFACTOR_PLAN.md     # Migration plan (completed)
└── adding-games.md      # Guide for new games
```

## 🔧 Development Commands

```bash
# Run tests
make test

# Run locally against cluster
make run

# Generate CRDs and code
make manifests generate

# Install CRDs
make install

# Build operator binary
make build
```

## 📖 Documentation

- **[local-testing.md](local-testing.md)** - How to deploy and test
- **[DESIGN.md](../DESIGN.md)** - Architecture decisions
- **[REFACTOR_PLAN.md](REFACTOR_PLAN.md)** - GameDefinition migration (completed)
- **[adding-games.md](adding-games.md)** - How to add new games

## 🎮 Supported Games

Currently bundled:
- **Valheim** (`gamedefinitions/valheim.yaml`)
  - AppId: 896660
  - Ports: 2456-2458 UDP
  - Config: serverName, worldName, password, public, crossplay, modifiers

More games can be added by creating GameDefinition YAML files - no Go code required.
