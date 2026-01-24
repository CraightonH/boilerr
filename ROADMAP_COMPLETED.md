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

---

## Phase 3: Bundled GameDefinitions (Partial)

**Note:** Only template and Valheim completed. Remaining games (3.3+) deferred until CRD schema stabilizes post-Phase 4.

### 3.1 GameDefinition Template
- [x] Create example/template GameDefinition with documentation (`gamedefinitions/TEMPLATE.yaml`)
- [x] Document configSchema mapping patterns (arg, env, configFile)
- [x] Create contribution guide for adding new games
- [x] Include comprehensive examples of all field types

### 3.2 Valheim
- [x] Create `gamedefinitions/valheim.yaml`
- [x] App ID: 896660, ports: 2456-2458/UDP
- [x] Config mappings: serverName, worldName, password, public, crossplay
- [x] Config file templates: adminlist.txt, permittedlist.txt, bannedlist.txt
- [x] Document preset and modifier options (difficulty, combat, resources, raids, portals)
- [x] Include example SteamServer manifest

---

## Phase 4: Deployment & Distribution

### 4.1 Container Images
- [x] Optimize multi-stage Dockerfile (golang:1.24 → distroless/static:nonroot)
- [x] Add OCI labels for metadata (title, description, source, licenses, etc.)
- [x] Add build args for version/commit/date injection
- [x] Document `steamcmd/steamcmd:ubuntu-22` as game server base (`docs/CONTAINER_IMAGES.md`)
- [x] Set up automated image builds on release (GHCR + Docker Hub)
- [x] Multi-platform support (linux/amd64, linux/arm64)
- [x] Implement cosign keyless signing in release workflow
- [x] Update CI workflows to skip on docs-only changes

### 4.2 Kubernetes Manifests
- [x] Enable CRDs in default kustomization (`config/default/kustomization.yaml`)
- [x] Improve Namespace manifest with proper labels
  - [x] Pod Security Standards labels (enforce/audit/warn: restricted)
  - [x] Standard app.kubernetes.io labels (name, component, part-of)
- [x] Enhance Deployment manifest
  - [x] Enable multi-arch node affinity (amd64, arm64)
  - [x] Production-ready labels
  - [x] Security context with restricted PSS compliance
- [x] Create kustomize overlays
  - [x] `config/overlays/production/` - HA deployment (2 replicas, pod anti-affinity, higher resources)
  - [x] `config/overlays/development/` - Dev deployment (debug logging, dev image tag)
  - [x] `config/overlays/minimal/` - Minimal footprint (low resources, single replica)
  - [x] `config/overlays/README.md` - Usage documentation with examples
- [x] Verify `make build-installer` generates complete install.yaml (~1400 lines)
- [x] Create comprehensive installation documentation (`docs/INSTALLATION.md`)
  - [x] Multiple installation methods (kubectl apply, kustomize, future Helm)
  - [x] Post-installation verification steps
  - [x] Configuration options
  - [x] Upgrade procedures
  - [x] Uninstallation guide
  - [x] Troubleshooting section
  - [x] Security considerations

### 4.3 Helm Chart
- [x] Create Helm chart structure (`charts/boilerr/`)
  - [x] Chart.yaml with metadata, keywords, maintainers
  - [x] values.yaml with comprehensive configuration options
  - [x] templates/ directory with all Kubernetes resources
  - [x] crds/ directory with CRD definitions
  - [x] .helmignore for package exclusions
- [x] Parameterize operator configuration
  - [x] Image (repository, tag, pullPolicy, pullSecrets)
  - [x] Replica count
  - [x] Resource limits and requests
  - [x] Node affinity (multi-arch support)
  - [x] Tolerations, pod anti-affinity
  - [x] Security contexts (pod and container)
  - [x] Priority class, termination grace period
- [x] Controller manager configuration
  - [x] Leader election settings
  - [x] Health probe configuration
  - [x] Metrics server settings (enabled, bind address, service config)
  - [x] Logging level and development mode
- [x] RBAC configuration
  - [x] Service account creation and annotations
  - [x] ClusterRole and ClusterRoleBinding
  - [x] Configurable RBAC enablement
- [x] GameDefinitions bundling
  - [x] Valheim GameDefinition template in `templates/gamedefinitions/`
  - [x] Conditional rendering based on `gameDefinitions.enabled`
  - [x] Selective game installation (`include`/`exclude` lists)
  - [x] Helper function for enabled game calculation
- [x] CRD management
  - [x] CRDs in `crds/` directory (auto-installed by Helm 3)
  - [x] `crds.install` and `crds.keep` values
- [x] Templates
  - [x] _helpers.tpl with common template functions
  - [x] NOTES.txt with post-install instructions
  - [x] Namespace template (conditional)
  - [x] ServiceAccount, ClusterRole, ClusterRoleBinding
  - [x] Deployment with full parameterization
  - [x] Service for metrics
- [x] Documentation
  - [x] charts/boilerr/README.md - comprehensive usage guide
    - [x] Installation instructions
    - [x] Configuration parameter table
    - [x] Example configurations (HA, minimal, custom registry, debug)
    - [x] Upgrade procedures
    - [x] Troubleshooting
  - [x] docs/HELM_CHART.md - development and publishing guide
    - [x] Local development and testing
    - [x] Packaging and versioning
    - [x] GitHub Pages publishing (automated and manual)
    - [x] Adding new GameDefinitions
    - [x] Testing checklist
- [x] GitHub Actions workflow
  - [x] `.github/workflows/helm-publish.yml`
  - [x] Automatic chart publishing on tag push
  - [x] GitHub Pages deployment
  - [x] Release asset attachment
- [x] Updated docs/INSTALLATION.md with Helm installation method

### 4.4 Installation Documentation
- [x] Created comprehensive installation guide (`docs/INSTALLATION.md`)
  - [x] Prerequisites and verification steps
  - [x] Method 1: Quick install (kubectl apply from GitHub release)
  - [x] Method 2: Specific version installation
  - [x] Method 3: Kustomize overlays (production, development, minimal)
  - [x] Method 4: Helm chart installation (with examples)
  - [x] Post-installation verification
  - [x] GameDefinition installation
  - [x] Test server deployment example
  - [x] Configuration options (resources, images, node affinity)
  - [x] Upgrade procedures (operator and CRDs)
  - [x] Uninstallation guide with safety warnings
  - [x] Troubleshooting section (operator crashes, RBAC, image pulls)
  - [x] Security considerations (Pod Security Standards, RBAC, image verification)
  - [x] Next steps and support links
- [x] Updated with Helm repository installation method
- [x] Documented prerequisites and compatibility (Kubernetes 1.24+, kubectl, cluster-admin)
- [ ] OLM (Operator Lifecycle Manager) bundle - deferred (optional feature)
