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
- [x] Define `GameDefinition` API types (`api/v1alpha1/gamedefinition_types.go`)
- [x] Implement spec fields:
  - [x] `appId` - Steam application ID
  - [x] `image` - Container image (default: `steamcmd/steamcmd:ubuntu-22`)
  - [x] `command` - Game server startup command
  - [x] `args` - Default startup arguments
  - [x] `installDir` - SteamCMD install directory path
  - [x] `ports` - Default ports with protocol (TCP/UDP)
  - [x] `env` - Default environment variables
  - [x] `configFiles` - Config file templates (path + content template)
  - [x] `healthCheck` - Health check configuration
- [x] Implement config mapping for user-facing settings:
  - [x] `configSchema` - Map of user config keys to arg/env/file mappings
- [x] Generate CRD manifests with `make manifests`

#### SteamServer CRD
- [x] Define `SteamServer` API types (`api/v1alpha1/steamserver_types.go`)
- [x] Implement spec fields:
  - [x] `game` - Reference to GameDefinition by name
  - [x] `config` - User-provided game configuration (map[string]string)
  - [x] `storage` - PVC size and storage class
  - [x] `resources` - CPU/memory requests and limits
  - [x] `serviceType` - LoadBalancer, NodePort, ClusterIP
- [x] Add optional fields: `beta`, `validate`, `anonymous`, `steamCredentialsSecret`
- [x] Define status fields: `state`, `address`, `ports`, `lastUpdated`, `appBuildId`, `message`
- [x] Add OpenAPI validation schema with proper defaults

### 2.2 Resource Builders
- [x] Create StatefulSet builder (`internal/resources/statefulset.go`)
  - [x] Init container using `steamcmd/steamcmd` image
  - [x] Build SteamCMD args from GameDefinition (login, force_install_dir, app_update, quit)
  - [x] Main container command/args from GameDefinition + SteamServer config
  - [x] Volume mounts for persistent storage
  - [x] Environment variable injection from GameDefinition + SteamServer
  - [x] Config file generation from templates
- [x] Create Service builder (`internal/resources/service.go`)
  - [x] Support LoadBalancer, NodePort, ClusterIP
  - [x] Handle multiple ports (game, query, RCON)
  - [x] Proper protocol handling (TCP/UDP)
- [x] Create PVC builder (`internal/resources/pvc.go`)
  - [x] Configurable size and storage class
  - [x] Proper access modes

### 2.3 Controller Implementation
- [x] Scaffold controllers with Kubebuilder
- [x] Implement GameDefinition controller
  - [x] Watch `GameDefinition` resources
  - [x] Validate GameDefinition spec
  - [x] Update status (ready/error)
- [x] Implement SteamServer controller
  - [x] Watch `SteamServer` resources
  - [x] Fetch referenced GameDefinition
  - [x] Validate SteamServer config against GameDefinition schema
  - [x] Generate desired state (PVC, StatefulSet, Service)
  - [x] Compare with actual cluster state
  - [x] Apply diffs with proper ownership references
- [x] Implement status updates
  - [x] Track server state transitions
  - [x] Update external address from Service
  - [x] Error handling and message propagation
- [x] Add finalizers for cleanup on deletion

### 2.4 SteamCMD Command Builder
- [x] Create command builder (`internal/steamcmd/command.go`)
- [x] Build args slice from GameDefinition + SteamServer spec
- [x] Handle anonymous vs authenticated login
- [x] Implement beta branch selection (`+app_update <id> -beta <branch>`)
- [x] Add validation flag support
- [x] Document Steam Guard limitations

### 2.5 Testing - Core CRDs
- [x] Unit tests for resource builders
- [x] Unit tests for controller reconciliation logic
- [x] Unit tests for SteamCMD command builder
- [x] Integration tests with envtest
- [x] E2E tests with kind cluster
- [x] Test with sample GameDefinition + SteamServer
