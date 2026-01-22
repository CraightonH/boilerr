---

## Phase 1: Foundation & Project Setup

Establish the project structure, tooling, and basic operator scaffolding.

### 1.1 Project Scaffolding
- [x] Initialize Go module (`go mod init github.com/CraightonH/boilerr`)
- [x] Set up Kubebuilder project structure
- [x] Create initial Makefile with common targets (build, test, generate, deploy)
- [x] Add `.gitignore` for Go/Kubernetes projects
- [x] Configure linting (golangci-lint)
- [x] Set up pre-commit hooks

### 1.2 CI/CD Pipeline
- [x] GitHub Actions workflow for PR checks (lint, test, build)
- [x] Container image build and push workflow
- [x] Release workflow with semantic versioning
- [x] CRD schema validation in CI

### 1.3 Documentation Foundation
- [x] Set up contributing guidelines (CONTRIBUTING.md)
- [x] Create issue and PR templates
- [x] Add code of conduct
- [x] License selection and LICENSE file

### 2.1 CRD Definitions

#### GameDefinition CRD
- [ ] Define `GameDefinition` API types (`api/v1alpha1/gamedefinition_types.go`)
- [ ] Implement spec fields:
  - [ ] `appId` - Steam application ID
  - [ ] `image` - Container image (default: `steamcmd/steamcmd:ubuntu-22`)
  - [ ] `command` - Game server startup command
  - [ ] `args` - Default startup arguments
  - [ ] `installDir` - SteamCMD install directory path
  - [ ] `ports` - Default ports with protocol (TCP/UDP)
  - [ ] `env` - Default environment variables
  - [ ] `configFiles` - Config file templates (path + content template)
  - [ ] `healthCheck` - Health check configuration
- [ ] Implement config mapping for user-facing settings:
  - [ ] `configSchema` - Map of user config keys to arg/env/file mappings
- [ ] Generate CRD manifests with `make manifests`

#### SteamServer CRD
- [ ] Define `SteamServer` API types (`api/v1alpha1/steamserver_types.go`)
- [ ] Implement spec fields:
  - [ ] `game` - Reference to GameDefinition by name
  - [ ] `config` - User-provided game configuration (map[string]string)
  - [ ] `storage` - PVC size and storage class
  - [ ] `resources` - CPU/memory requests and limits
  - [ ] `serviceType` - LoadBalancer, NodePort, ClusterIP
- [ ] Add optional fields: `beta`, `validate`, `anonymous`, `steamCredentialsSecret`
- [ ] Define status fields: `state`, `address`, `ports`, `lastUpdated`, `appBuildId`, `message`
- [ ] Add OpenAPI validation schema with proper defaults

### 2.2 Resource Builders
- [ ] Create StatefulSet builder (`internal/resources/statefulset.go`)
  - [ ] Init container using `steamcmd/steamcmd` image
  - [ ] Build SteamCMD args from GameDefinition (login, force_install_dir, app_update, quit)
  - [ ] Main container command/args from GameDefinition + SteamServer config
  - [ ] Volume mounts for persistent storage
  - [ ] Environment variable injection from GameDefinition + SteamServer
  - [ ] Config file generation from templates
- [ ] Create Service builder (`internal/resources/service.go`)
  - [ ] Support LoadBalancer, NodePort, ClusterIP
  - [ ] Handle multiple ports (game, query, RCON)
  - [ ] Proper protocol handling (TCP/UDP)
- [ ] Create PVC builder (`internal/resources/pvc.go`)
  - [ ] Configurable size and storage class
  - [ ] Proper access modes

### 2.3 Controller Implementation
- [ ] Scaffold controllers with Kubebuilder
- [ ] Implement GameDefinition controller
  - [ ] Watch `GameDefinition` resources
  - [ ] Validate GameDefinition spec
  - [ ] Update status (ready/error)
- [ ] Implement SteamServer controller
  - [ ] Watch `SteamServer` resources
  - [ ] Fetch referenced GameDefinition
  - [ ] Validate SteamServer config against GameDefinition schema
  - [ ] Generate desired state (PVC, StatefulSet, Service)
  - [ ] Compare with actual cluster state
  - [ ] Apply diffs with proper ownership references
- [ ] Implement status updates
  - [ ] Track server state transitions
  - [ ] Update external address from Service
  - [ ] Error handling and message propagation
- [ ] Add finalizers for cleanup on deletion

### 2.4 SteamCMD Command Builder
- [ ] Create command builder (`internal/steamcmd/command.go`)
- [ ] Build args slice from GameDefinition + SteamServer spec
- [ ] Handle anonymous vs authenticated login
- [ ] Implement beta branch selection (`+app_update <id> -beta <branch>`)
- [ ] Add validation flag support
- [ ] Document Steam Guard limitations

### 2.5 Testing - Core CRDs
- [ ] Unit tests for resource builders
- [ ] Unit tests for controller reconciliation logic
- [ ] Unit tests for SteamCMD command builder
- [ ] Integration tests with envtest
- [ ] E2E tests with kind cluster
- [ ] Test with sample GameDefinition + SteamServer
