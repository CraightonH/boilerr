# Boilerr Project Roadmap

A phased approach to building the Kubernetes operator for Steam dedicated game servers.

---

## Phase 1: Foundation & Project Setup

Establish the project structure, tooling, and basic operator scaffolding.

### 1.1 Project Scaffolding
- [ ] Initialize Go module (`go mod init github.com/CraightonH/boilerr`)
- [ ] Set up Kubebuilder project structure
- [ ] Create initial Makefile with common targets (build, test, generate, deploy)
- [ ] Add `.gitignore` for Go/Kubernetes projects
- [ ] Configure linting (golangci-lint)
- [ ] Set up pre-commit hooks

### 1.2 CI/CD Pipeline
- [ ] GitHub Actions workflow for PR checks (lint, test, build)
- [ ] Container image build and push workflow
- [ ] Release workflow with semantic versioning
- [ ] CRD schema validation in CI

### 1.3 Documentation Foundation
- [ ] Set up contributing guidelines (CONTRIBUTING.md)
- [ ] Create issue and PR templates
- [ ] Add code of conduct
- [ ] License selection and LICENSE file

---

## Phase 2: Core Operator - Generic SteamServer CRD

Build the generic `SteamServer` CRD that can run any SteamCMD-compatible game.

### 2.1 CRD Definition
- [ ] Define `SteamServer` API types (`api/v1alpha1/steamserver_types.go`)
- [ ] Implement spec fields: `appId`, `ports`, `command`, `args`, `env`
- [ ] Implement spec fields: `storage`, `resources`, `serviceType`
- [ ] Add optional fields: `beta`, `validate`, `anonymous`, `steamCredentialsSecret`
- [ ] Define status fields: `state`, `address`, `ports`, `lastUpdated`, `appBuildId`, `message`
- [ ] Generate CRD manifests with `make manifests`
- [ ] Add OpenAPI validation schema with proper defaults

### 2.2 Resource Builders
- [ ] Create StatefulSet builder (`internal/resources/statefulset.go`)
  - [ ] Init container with SteamCMD download logic
  - [ ] Main container with game server command
  - [ ] Volume mounts for persistent storage
  - [ ] Environment variable injection
- [ ] Create Service builder (`internal/resources/service.go`)
  - [ ] Support LoadBalancer, NodePort, ClusterIP
  - [ ] Handle multiple ports (game, query, RCON)
  - [ ] Proper protocol handling (TCP/UDP)
- [ ] Create PVC builder (`internal/resources/pvc.go`)
  - [ ] Configurable size and storage class
  - [ ] Proper access modes

### 2.3 Controller Implementation
- [ ] Scaffold controller with Kubebuilder
- [ ] Implement reconciliation loop
  - [ ] Watch `SteamServer` resources
  - [ ] Generate desired state (PVC, StatefulSet, Service)
  - [ ] Compare with actual cluster state
  - [ ] Apply diffs with proper ownership references
- [ ] Implement status updates
  - [ ] Track server state transitions
  - [ ] Update external address from Service
  - [ ] Error handling and message propagation
- [ ] Add finalizers for cleanup on deletion

### 2.4 SteamCMD Integration
- [ ] Create init container script generator (`internal/steamcmd/scripts.go`)
- [ ] Handle anonymous vs authenticated login
- [ ] Implement beta branch selection
- [ ] Add validation flag support
- [ ] Handle Steam Guard (document limitations)

### 2.5 Testing - Generic CRD
- [ ] Unit tests for resource builders
- [ ] Unit tests for controller reconciliation logic
- [ ] Integration tests with envtest
- [ ] E2E tests with kind cluster
- [ ] Test manual game server deployment (pick one simple game)

---

## Phase 3: Game-Specific CRDs

Add first-class support for popular games with typed configurations.

### 3.1 Shared Infrastructure
- [ ] Create base types for common fields (`api/v1alpha1/common_types.go`)
- [ ] Implement CRD inheritance/embedding pattern
- [ ] Create game config template system

### 3.2 ValheimServer CRD
- [ ] Define `ValheimServer` types with game-specific fields
  - [ ] `serverName`, `worldName`, `password`
  - [ ] `public`, `crossplay`, `admins`
- [ ] Create Valheim-specific controller (or extend generic)
- [ ] Generate startup command and config files
- [ ] Add config file mounting for admins.txt, etc.
- [ ] Write Valheim-specific tests
- [ ] Document Valheim-specific usage

### 3.3 SatisfactoryServer CRD
- [ ] Define `SatisfactoryServer` types
  - [ ] `serverName`, `maxPlayers`
  - [ ] `autoPause`, `autoSaveOnDisconnect`, `networkQuality`
- [ ] Implement controller logic
- [ ] Handle Satisfactory's specific startup requirements
- [ ] Write tests and documentation

### 3.4 PalworldServer CRD
- [ ] Define `PalworldServer` types
  - [ ] `serverName`, `adminPassword`, `serverPassword`
  - [ ] `maxPlayers`, `difficulty`, `deathPenalty`
  - [ ] Rate multipliers: `expRate`, `captureRate`
- [ ] Implement controller logic
- [ ] Generate PalWorldSettings.ini
- [ ] Write tests and documentation

### 3.5 FactorioServer CRD
- [ ] Define `FactorioServer` types
  - [ ] `serverName`, `description`, `maxPlayers`
  - [ ] `visibility`, `autosaveInterval`, `allowCommands`
  - [ ] `admins`, `factorioCredentials`
- [ ] Implement controller logic
- [ ] Generate server-settings.json
- [ ] Write tests and documentation

### 3.6 SevenDaysServer CRD
- [ ] Define `SevenDaysServer` types
  - [ ] World settings: `worldGenSeed`, `worldGenSize`, `gameWorld`
  - [ ] Gameplay: `gameMode`, `difficulty`
  - [ ] Day/night: `dayNightLength`, `dayLightLength`
  - [ ] Zombies: `zombieMove`, `bloodMoonFrequency`
- [ ] Implement controller logic
- [ ] Generate serverconfig.xml
- [ ] Write tests and documentation

### 3.7 VRisingServer CRD
- [ ] Define `VRisingServer` types
  - [ ] `serverName`, `password`, `maxPlayers`
  - [ ] `gameMode`, `difficultyPreset`, `clanSize`
  - [ ] RCON configuration
- [ ] Implement controller logic
- [ ] Generate ServerHostSettings.json and ServerGameSettings.json
- [ ] Write tests and documentation

### 3.8 EnshroudedServer CRD
- [ ] Define `EnshroudedServer` types
  - [ ] `serverName`, `password`, `maxPlayers`
  - [ ] `saveDirectory`, `logDirectory`
- [ ] Implement controller logic
- [ ] Generate enshrouded_server.json
- [ ] Write tests and documentation

### 3.9 TerrariaServer CRD
- [ ] Define `TerrariaServer` types
  - [ ] `variant` (vanilla/tshock)
  - [ ] World settings: `worldName`, `worldSize`, `difficulty`, `seed`
  - [ ] tShock-specific: `restApiEnabled`, `restApiPort`
- [ ] Implement controller logic
- [ ] Generate serverconfig.txt
- [ ] Write tests and documentation

---

## Phase 4: Deployment & Distribution

Package the operator for easy installation.

### 4.1 Container Images
- [ ] Create optimized multi-stage Dockerfile for operator
- [ ] Create/select base SteamCMD image
- [ ] Set up automated image builds on release
- [ ] Push to container registry (GHCR, Docker Hub)
- [ ] Implement image signing (cosign)

### 4.2 Kubernetes Manifests
- [ ] Generate production-ready RBAC manifests
- [ ] Create operator Deployment manifest
- [ ] Create Namespace manifest with proper labels
- [ ] Bundle CRD manifests
- [ ] Create kustomize overlays for common configurations

### 4.3 Helm Chart
- [ ] Create Helm chart structure
- [ ] Parameterize common options (image, resources, replicas)
- [ ] Add values for RBAC configuration
- [ ] Add values for operator configuration
- [ ] Document Helm installation
- [ ] Publish to Helm repository (or GitHub Pages)

### 4.4 Installation Documentation
- [ ] Quick start guide (kubectl apply)
- [ ] Helm installation guide
- [ ] OLM (Operator Lifecycle Manager) bundle (optional)
- [ ] Document prerequisites and compatibility matrix

---

## Phase 5: Web UI - Core

Build the web-based configuration interface.

### 5.1 Project Setup
- [ ] Initialize frontend project (React + TypeScript recommended)
- [ ] Set up build tooling (Vite or Next.js)
- [ ] Configure ESLint and Prettier
- [ ] Add testing framework (Vitest/Jest + Testing Library)
- [ ] Create component library foundation

### 5.2 CRD Schema Integration
- [ ] Fetch CRD OpenAPI schemas from cluster or static files
- [ ] Create schema parser (`lib/schema.ts`)
- [ ] Map OpenAPI schema to form field types
- [ ] Handle nested objects and arrays
- [ ] Support validation constraints from schema

### 5.3 Dynamic Form Generation
- [ ] Create `GameForm` component
- [ ] Implement field types: text, number, boolean, select, array
- [ ] Add secret reference field type (special handling)
- [ ] Implement collapsible sections for field groups
- [ ] Add form validation with real-time feedback
- [ ] Support default values from schema

### 5.4 YAML Preview & Export
- [ ] Create `YamlPreview` component
- [ ] Real-time YAML generation from form state
- [ ] Syntax highlighting for YAML
- [ ] Copy to clipboard functionality
- [ ] Download as file option
- [ ] Validate generated YAML against schema

### 5.5 Game Selector
- [ ] Create game selection UI
- [ ] Display supported games with icons/logos
- [ ] Show game-specific information (ports, resources)
- [ ] Pre-populate form with game defaults on selection

### 5.6 Basic Styling & UX
- [ ] Choose and integrate UI framework (Tailwind, Shadcn, etc.)
- [ ] Create responsive layout
- [ ] Dark/light mode support
- [ ] Loading states and error handling
- [ ] Accessibility audit (ARIA labels, keyboard nav)

---

## Phase 6: Web UI - Kubernetes Integration

Connect the UI to Kubernetes for direct apply and server management.

### 6.1 Backend API (Optional)
- [ ] Decide: client-side K8s API or backend proxy
- [ ] If backend: Create Go/Node.js API server
- [ ] Implement K8s client wrapper
- [ ] Add authentication middleware
- [ ] Create API endpoints for CRUD operations

### 6.2 Kubernetes Authentication
- [ ] Implement service account authentication (in-cluster)
- [ ] Add kubeconfig upload option (development/testing)
- [ ] OIDC integration for multi-user environments
- [ ] Token refresh handling

### 6.3 Direct Apply Feature
- [ ] Create K8s API client wrapper (`lib/k8s.ts`)
- [ ] Implement CR create operation
- [ ] Implement CR update operation
- [ ] Implement CR delete operation
- [ ] Handle API errors gracefully
- [ ] Show operation progress/status

### 6.4 Server Dashboard
- [ ] Create `ServerList` component
- [ ] List existing SteamServer CRs
- [ ] Display status (state, address, ports)
- [ ] Filter by namespace, game type, state
- [ ] Quick actions: restart (delete pod), edit, delete
- [ ] Auto-refresh or watch for updates

### 6.5 Server Details View
- [ ] Show full server configuration
- [ ] Display connection information prominently
- [ ] Show related K8s resources (Pod, Service, PVC)
- [ ] Event history from CR status
- [ ] Edit mode with form pre-populated

---

## Phase 7: Web UI - GitOps Integration

Add GitHub/GitLab PR workflow for GitOps deployments.

### 7.1 Git Provider Authentication
- [ ] GitHub OAuth App / GitHub App setup
- [ ] GitLab OAuth integration
- [ ] Token storage and refresh
- [ ] Repository access permissions

### 7.2 GitHub Integration (`lib/github.ts`)
- [ ] List user's repositories
- [ ] Get repository contents
- [ ] Create branch
- [ ] Create/update file
- [ ] Create pull request
- [ ] Configure PR reviewers and labels

### 7.3 GitOps Workflow
- [ ] Repository selection UI
- [ ] Manifest path configuration
- [ ] Branch naming strategy
- [ ] PR creation flow
  - [ ] Generate YAML from form
  - [ ] Create feature branch
  - [ ] Commit manifest file
  - [ ] Open PR with description
- [ ] Link to created PR
- [ ] Optional: Monitor PR status

### 7.4 Configuration Persistence
- [ ] Store GitOps configuration per-user
- [ ] Remember repository and path preferences
- [ ] Support multiple target repos/clusters

---

## Phase 8: Observability & Operations

Add monitoring, logging, and operational features.

### 8.1 Operator Metrics
- [ ] Instrument operator with Prometheus metrics
- [ ] `steamserver_status` gauge (by name, namespace, state)
- [ ] `steamserver_reconcile_duration_seconds` histogram
- [ ] `steamserver_reconcile_errors_total` counter
- [ ] Create ServiceMonitor for Prometheus Operator

### 8.2 Grafana Dashboard
- [ ] Create Grafana dashboard JSON
- [ ] Panels: server states, reconcile performance, errors
- [ ] Alert definitions for common issues
- [ ] Include in Helm chart as ConfigMap

### 8.3 Logging
- [ ] Structured logging (JSON) in operator
- [ ] Log levels configurable via flag/env
- [ ] Include relevant context (server name, namespace)
- [ ] Document log aggregation setup (Loki, ELK)

### 8.4 Events
- [ ] Emit K8s Events for state transitions
- [ ] Events for errors with actionable messages
- [ ] Events for successful operations

---

## Phase 9: Advanced Features

Implement features from the "Future Work" section.

### 9.1 Automatic Backups
- [ ] Add backup configuration to CRD spec
  - [ ] `backup.enabled`, `backup.schedule`, `backup.retention`
- [ ] Generate CronJob for backup execution
- [ ] Implement backup script (tar + upload to S3/GCS/PVC)
- [ ] Add backup status to CR
- [ ] Restore functionality (manual trigger)

### 9.2 RCON Integration
- [ ] Add RCON configuration to game CRDs
- [ ] Create RCON client library
- [ ] Expose RCON commands via CR annotation or sub-resource
- [ ] Web UI: RCON console component
- [ ] Common commands: broadcast, save, kick, ban

### 9.3 Server Query Integration
- [ ] Implement Steam server query protocol (A2S)
- [ ] Fetch player count, server info
- [ ] Update CR status with live player data
- [ ] Web UI: Display player count on dashboard
- [ ] Metrics: `steamserver_players_current`

### 9.4 Steam Workshop Mods
- [ ] Add `mods` field to CRDs (list of workshop IDs)
- [ ] Extend SteamCMD init to download workshop items
- [ ] Handle mod load order configuration
- [ ] Document mod support per game

### 9.5 Player-Aware Updates
- [ ] Add update strategy to CRD spec
- [ ] Implement "drain" strategy (wait for 0 players)
- [ ] Configurable wait timeout
- [ ] Force update option
- [ ] Web UI: Show update pending status

---

## Phase 10: Polish & Production Readiness

Final polish, security hardening, and production preparation.

### 10.1 Security Hardening
- [ ] Audit RBAC permissions (principle of least privilege)
- [ ] Pod Security Standards compliance
- [ ] Network Policy generation (optional feature)
- [ ] Secret encryption at rest documentation
- [ ] Security scanning in CI (Trivy, Snyk)

### 10.2 Performance & Scalability
- [ ] Optimize reconciliation (avoid unnecessary updates)
- [ ] Implement rate limiting for API calls
- [ ] Test with many (100+) game servers
- [ ] Document resource requirements for operator

### 10.3 Documentation Site
- [ ] Set up documentation site (Docusaurus, MkDocs, etc.)
- [ ] User guide: installation, configuration, usage
- [ ] Game-specific guides with examples
- [ ] API reference (auto-generated from CRD)
- [ ] Troubleshooting guide
- [ ] FAQ

### 10.4 Community Readiness
- [ ] Finalize open source license
- [ ] Create public roadmap (GitHub Projects)
- [ ] Set up discussion forums (GitHub Discussions)
- [ ] Write announcement blog post
- [ ] Submit to Awesome Kubernetes lists

### 10.5 Release 1.0
- [ ] Version bump to v1
- [ ] Final testing across all supported games
- [ ] Update all documentation for release
- [ ] Create GitHub release with changelog
- [ ] Announce release

---

## Backlog / Future Considerations

Items to consider for future versions:

- [ ] Multi-instance clustering (ARK clusters, etc.)
- [ ] Server templates/presets library
- [ ] Multi-cluster support in Web UI
- [ ] Real-time log streaming in Web UI
- [ ] Cost estimation (cloud resource pricing)
- [ ] Mobile-friendly Web UI
- [ ] Discord/Slack notifications for server events
- [ ] Scheduled server restarts
- [ ] Game-specific health checks
- [ ] Support for non-Steam games (custom game servers)

---

## Version Milestones

| Version | Target | Key Features |
|---------|--------|--------------|
| v0.1.0 | Phase 2 complete | Generic SteamServer CRD working |
| v0.2.0 | Phase 3 complete | All game-specific CRDs |
| v0.3.0 | Phase 4 complete | Helm chart, easy installation |
| v0.4.0 | Phase 5-6 complete | Web UI with direct apply |
| v0.5.0 | Phase 7 complete | GitOps integration |
| v0.6.0 | Phase 8 complete | Observability |
| v1.0.0 | Phase 9-10 complete | Production ready |

---

## Notes

- Prioritize the generic `SteamServer` CRD first - it unblocks all other games
- Game-specific CRDs can be added incrementally; start with Valheim (most popular)
- Web UI can be developed in parallel with operator once CRD schemas are stable
- Consider community contributions for game-specific CRDs after v0.2.0
