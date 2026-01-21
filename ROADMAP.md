# Boilerr Project Roadmap

A phased approach to building the Kubernetes operator for Steam dedicated game servers.

---

## Phase 2: Core Operator - SteamServer + GameDefinition CRDs

Build the `GameDefinition` and `SteamServer` CRDs. GameDefinition defines how to install/run a game; SteamServer references a GameDefinition to deploy an instance.

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

---

## Phase 3: Bundled GameDefinitions

Create GameDefinition manifests for popular games. These ship with the Helm chart.

### 3.1 GameDefinition Template
- [ ] Create example/template GameDefinition with documentation
- [ ] Document configSchema mapping patterns (arg, env, configFile)
- [ ] Create contribution guide for adding new games

### 3.2 Valheim
- [ ] Create `gamedefinitions/valheim.yaml`
- [ ] App ID: 896660, ports: 2456-2458/UDP
- [ ] Config mappings: serverName, worldName, password, public, crossplay
- [ ] Config file template: adminlist.txt
- [ ] Test deployment and document

### 3.3 Satisfactory
- [ ] Create `gamedefinitions/satisfactory.yaml`
- [ ] App ID: 1690800, ports: 7777/UDP+TCP, 15000/UDP, 15777/UDP
- [ ] Config mappings: maxPlayers, autosaveInterval, networkQuality
- [ ] Test deployment and document

### 3.4 Palworld
- [ ] Create `gamedefinitions/palworld.yaml`
- [ ] App ID: 2394010, ports: 8211/UDP
- [ ] Config mappings: serverName, password, maxPlayers, difficulty
- [ ] Config file template: PalWorldSettings.ini
- [ ] Test deployment and document

### 3.5 7 Days to Die
- [ ] Create `gamedefinitions/7daystodie.yaml`
- [ ] App ID: 294420, ports: 26900/TCP+UDP, 26901-26902/UDP
- [ ] Config mappings: serverName, worldGenSeed, worldGenSize, gameMode
- [ ] Config file template: serverconfig.xml
- [ ] Test deployment and document

### 3.6 V Rising
- [ ] Create `gamedefinitions/vrising.yaml`
- [ ] App ID: 1829350, ports: 9876-9877/UDP
- [ ] Config mappings: serverName, password, maxPlayers, gameMode
- [ ] Config file templates: ServerHostSettings.json, ServerGameSettings.json
- [ ] Test deployment and document

### 3.7 Enshrouded
- [ ] Create `gamedefinitions/enshrouded.yaml`
- [ ] App ID: 2278520, ports: 15636-15637/UDP
- [ ] Config mappings: serverName, password, maxPlayers
- [ ] Config file template: enshrouded_server.json
- [ ] Test deployment and document

### 3.8 Project Zomboid
- [ ] Create `gamedefinitions/projectzomboid.yaml`
- [ ] App ID: 380870, ports: 16261/UDP, 16262-16272/TCP
- [ ] Config mappings: serverName, password, maxPlayers, publicServer
- [ ] Config file template: server.ini
- [ ] Test deployment and document

### 3.9 Terraria (tShock)
- [ ] Create `gamedefinitions/terraria.yaml`
- [ ] App ID: 105600 (or tShock container), ports: 7777/TCP
- [ ] Config mappings: worldName, maxPlayers, password, difficulty
- [ ] Config file template: serverconfig.txt
- [ ] Test deployment and document

### 3.10 Additional Games (community-driven)
- [ ] ARK: Survival Evolved
- [ ] Rust
- [ ] Counter-Strike 2
- [ ] Team Fortress 2
- [ ] Left 4 Dead 2
- [ ] Conan Exiles
- [ ] The Forest

---

## Phase 4: Deployment & Distribution

Package the operator for easy installation.

### 4.1 Container Images
- [ ] Create optimized multi-stage Dockerfile for operator
- [ ] Document use of `steamcmd/steamcmd:ubuntu-22` as game server base
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
- [ ] Bundle GameDefinitions from Phase 3
  - [ ] Include all bundled GameDefinitions in `templates/gamedefinitions/`
  - [ ] Add `gameDefinitions.enabled` value (default: true)
  - [ ] Allow selective game enablement via values
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
- [ ] Add RCON configuration to GameDefinition spec
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
- [ ] Add `mods` field to SteamServer spec (list of workshop IDs)
- [ ] Extend SteamCMD init to download workshop items
- [ ] Handle mod load order configuration
- [ ] Document mod support per GameDefinition

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
| v0.1.0 | Phase 2 complete | GameDefinition + SteamServer CRDs working |
| v0.2.0 | Phase 3 complete | Bundled GameDefinitions for popular games |
| v0.3.0 | Phase 4 complete | Helm chart with bundled games, easy installation |
| v0.4.0 | Phase 5-6 complete | Web UI with direct apply |
| v0.5.0 | Phase 7 complete | GitOps integration |
| v0.6.0 | Phase 8 complete | Observability |
| v1.0.0 | Phase 9-10 complete | Production ready |

---

## Notes

- `GameDefinition` CRD enables extensibility - users can add games without operator changes
- `SteamServer` CRD references a GameDefinition - keeps user-facing config simple
- Uses `steamcmd/steamcmd` container directly - no custom images per game
- Adding a new game = PR a GameDefinition YAML file (no Go code required)
- Web UI can be developed in parallel with operator once CRD schemas are stable
- Community contributions welcome for GameDefinitions after v0.1.0
