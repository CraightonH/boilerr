# Boilerr Design Document

## Overview

A Kubernetes operator that simplifies deploying and managing Steam dedicated game servers. Users define a `SteamServer` custom resource referencing a `GameDefinition`, and the operator handles the rest—no custom Dockerfiles or complex K8s manifests required.

### Goals

- **User-friendly**: Non-K8s experts can deploy game servers with simple config
- **Extensible**: Anyone can add new games via GameDefinition YAML (no code changes)
- **Low overhead**: Reduce boilerplate for spinning up new game servers
- **Kubernetes-native**: Leverage existing K8s primitives (StatefulSets, PVCs, Services)
- **No custom images**: Use `steamcmd/steamcmd` container directly, configured at runtime

### Non-Goals (v1)

- Automatic backups and restore
- Player-aware update strategies (drain before restart)
- Multi-server clustering (e.g., ARK clusters)

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                      Kubernetes Cluster                           │
│                                                                   │
│  ┌────────────────┐  ┌──────────────┐                            │
│  │ GameDefinition │  │  SteamServer │                            │
│  │  (per game)    │  │   (user CR)  │                            │
│  │                │  │              │                            │
│  │ - appId        │◀─│ - game: ref  │                            │
│  │ - command      │  │ - config     │                            │
│  │ - ports        │  │ - storage    │                            │
│  │ - configSchema │  │ - resources  │                            │
│  └────────────────┘  └──────┬───────┘                            │
│                             │                                     │
│                             ▼                                     │
│                ┌─────────────────────────────┐                   │
│                │       Operator Pod          │                   │
│                │ - Watches GameDefinitions   │                   │
│                │ - Watches SteamServers      │                   │
│                │ - Reconciles to K8s resources│                   │
│                └──────────────┬──────────────┘                   │
│                               │                                   │
│                               ▼                                   │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │                   Per-Server Resources                     │   │
│  │  ┌─────────┐  ┌───────────────────────┐  ┌─────────────┐  │   │
│  │  │   PVC   │  │     StatefulSet       │  │   Service   │  │   │
│  │  │ (saves) │  │ init: steamcmd/steamcmd│  │ (LB/NP/CIP)│  │   │
│  │  │         │  │ main: steamcmd/steamcmd│  │             │  │   │
│  │  └─────────┘  └───────────────────────┘  └─────────────┘  │   │
│  └───────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

### Reconciliation Loop

1. Watch for `SteamServer` CR events
2. Fetch referenced `GameDefinition`
3. Merge GameDefinition defaults with SteamServer config
4. Generate desired state: PVC, StatefulSet, Service, ConfigMaps
5. Compare with actual state in cluster
6. Apply diffs (create/update/delete resources)
7. Update CR status with server state, IP, ports

---

## Custom Resource Definitions

### Two-CRD Architecture

Boilerr uses two CRDs that work together:

1. **GameDefinition** - Defines how to install and run a specific game (created by operator/community)
2. **SteamServer** - User-facing CR that references a GameDefinition to deploy a server instance

This separation allows:
- Users to deploy servers without knowing game internals
- Community to contribute new games without operator code changes
- Helm chart to bundle popular GameDefinitions out of the box

### GameDefinition CRD

Defines everything needed to install and run a Steam game server. Typically created by the operator maintainers or community contributors.

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: GameDefinition
metadata:
  name: valheim
spec:
  # Steam App ID for the dedicated server
  appId: 896660

  # Container image (default: steamcmd/steamcmd:ubuntu-22)
  image: steamcmd/steamcmd:ubuntu-22

  # Where SteamCMD installs the game
  installDir: /data/server

  # Game server startup command
  command: /data/server/valheim_server.x86_64

  # Default startup arguments (can reference config via {{.Config.key}})
  args:
    - "-name"
    - "{{.Config.serverName}}"
    - "-world"
    - "{{.Config.worldName}}"
    - "-password"
    - "{{.Config.password}}"
    - "-port"
    - "2456"
    - "-public"
    - "{{.Config.public}}"

  # Default ports
  ports:
    - name: game
      port: 2456
      protocol: UDP
    - name: query
      port: 2457
      protocol: UDP

  # Default environment variables
  env:
    - name: SteamAppId
      value: "892970"  # Valheim client app ID for Steam networking

  # Config schema - defines user-facing configuration options
  # Maps user config keys to how they're used (args, env, or configFile)
  configSchema:
    serverName:
      description: "Server name shown in browser"
      default: "My Valheim Server"
      required: true
    worldName:
      description: "World/save name"
      default: "Dedicated"
      required: true
    password:
      description: "Server password (min 5 chars)"
      secret: true
      required: true
    public:
      description: "List on public server browser"
      default: "0"
      enum: ["0", "1"]
    crossplay:
      description: "Enable crossplay"
      default: "false"
      mapTo:
        type: arg
        value: "-crossplay"
        condition: "true"  # only add arg if value is "true"
    admins:
      description: "List of Steam IDs for admin access"
      array: true
      mapTo:
        type: configFile
        path: /data/server/adminlist.txt
        template: "{{range .}}{{.}}\n{{end}}"

  # Config file templates (static files the game needs)
  configFiles:
    - path: /data/server/permittedlist.txt
      content: ""  # empty by default

  # Default resource recommendations
  defaultResources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "4"
      memory: "8Gi"

  # Default storage size
  defaultStorage: 20Gi

  # Health check configuration
  healthCheck:
    tcpSocket:
      port: 2456
    initialDelaySeconds: 120
    periodSeconds: 30

status:
  ready: true
  message: "GameDefinition validated successfully"
```

### SteamServer CRD

User-facing CR for deploying a game server instance. References a GameDefinition by name.

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: valheim-prod
  namespace: game-servers
spec:
  # Reference to GameDefinition (required)
  game: valheim

  # Game-specific configuration (keys defined by GameDefinition.configSchema)
  # Values can be literals or secret references
  config:
    serverName: "Vikings Only"
    worldName: "Midgard"
    password:
      secretKeyRef:
        name: valheim-secrets
        key: password
    public: "0"
    crossplay: "true"
    admins:
      - "76561198000000001"
      - "76561198000000002"

  # Override default storage (optional)
  storage:
    size: 30Gi
    storageClassName: fast-ssd

  # Override default resources (optional)
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "4"
      memory: "8Gi"

  # Service type (optional, default: LoadBalancer)
  serviceType: LoadBalancer

  # SteamCMD options (optional)
  beta: ""  # beta branch name, empty = stable
  validate: true  # run steamcmd validate on startup
  anonymous: true  # use anonymous login
  # steamCredentialsSecret: steam-creds  # if anonymous: false

status:
  state: Running  # Pending, Installing, Starting, Running, Error
  address: "192.168.1.100"
  ports:
    - name: game
      port: 2456
    - name: query
      port: 2457
  lastUpdated: "2026-01-20T10:30:00Z"
  appBuildId: "12345678"
  message: "Server running successfully"
```

### Example: Deploying Different Games

All games use the same `SteamServer` kind, just with different `game` references and `config` values.

#### Satisfactory Server

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: satisfactory-prod
spec:
  game: satisfactory
  config:
    maxPlayers: "8"
    autosaveInterval: "300"
  storage:
    size: 50Gi  # Satisfactory saves get big
  resources:
    requests:
      cpu: "4"
      memory: "12Gi"
```

#### Palworld Server

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: palworld-prod
spec:
  game: palworld
  config:
    serverName: "Pal Paradise"
    password:
      secretKeyRef:
        name: palworld-secrets
        key: password
    maxPlayers: "32"
    difficulty: "Normal"
  storage:
    size: 30Gi
  resources:
    requests:
      cpu: "4"
      memory: "16Gi"
```

#### 7 Days to Die Server

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: 7dtd-prod
spec:
  game: 7daystodie
  config:
    serverName: "Survival Server"
    password:
      secretKeyRef:
        name: 7dtd-secrets
        key: password
    worldGenSeed: "apocalypse"
    worldGenSize: "6144"
    difficulty: "3"
  storage:
    size: 30Gi
  resources:
    requests:
      cpu: "4"
      memory: "8Gi"
```

### Bundled GameDefinitions Reference

These GameDefinitions ship with the Helm chart. Users can deploy servers for these games immediately.

| Game | GameDefinition Name | App ID | Default Ports | Min Resources |
|------|---------------------|--------|---------------|---------------|
| Valheim | `valheim` | 896660 | 2456-2458/UDP | 2 CPU, 4Gi |
| Satisfactory | `satisfactory` | 1690800 | 7777/UDP+TCP, 15000/UDP, 15777/UDP | 4 CPU, 12Gi |
| Palworld | `palworld` | 2394010 | 8211/UDP | 4 CPU, 16Gi |
| 7 Days to Die | `7daystodie` | 294420 | 26900/TCP+UDP, 26901-26902/UDP | 4 CPU, 8Gi |
| V Rising | `vrising` | 1829350 | 9876-9877/UDP | 2 CPU, 6Gi |
| Enshrouded | `enshrouded` | 2278520 | 15636-15637/UDP | 2 CPU, 8Gi |
| Project Zomboid | `projectzomboid` | 380870 | 16261/UDP, 16262-16272/TCP | 4 CPU, 8Gi |
| Terraria | `terraria` | 105600 | 7777/TCP | 1 CPU, 1Gi |

### Adding Custom Games

Users can create their own GameDefinition for games not bundled with the operator:

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: GameDefinition
metadata:
  name: my-custom-game
spec:
  appId: 123456
  installDir: /data/server
  command: /data/server/start.sh
  args:
    - "-port"
    - "27015"
  ports:
    - name: game
      port: 27015
      protocol: UDP
  configSchema:
    serverName:
      description: "Server name"
      default: "My Server"
  defaultResources:
    requests:
      cpu: "2"
      memory: "4Gi"
  defaultStorage: 20Gi
```

Then deploy with:

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: my-server
spec:
  game: my-custom-game
  config:
    serverName: "Custom Server"
```

---

## Implementation Details

### Directory Structure

```
boilerr/
├── api/
│   └── v1alpha1/
│       ├── gamedefinition_types.go   # GameDefinition CRD
│       ├── steamserver_types.go      # SteamServer CRD
│       ├── common_types.go           # Shared types (ports, resources, etc.)
│       └── groupversion_info.go
├── internal/
│   ├── controller/
│   │   ├── gamedefinition_controller.go  # Validates GameDefinitions
│   │   └── steamserver_controller.go     # Main reconciliation logic
│   ├── resources/
│   │   ├── statefulset.go            # StatefulSet builder
│   │   ├── service.go                # Service builder
│   │   ├── pvc.go                    # PVC builder
│   │   └── configmap.go              # ConfigMap builder for game configs
│   └── steamcmd/
│       └── command.go                # SteamCMD args builder
├── config/
│   ├── crd/                          # Generated CRD YAML
│   ├── rbac/                         # RBAC manifests
│   └── manager/                      # Operator deployment
├── charts/
│   └── boilerr/
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── templates/
│       │   ├── deployment.yaml       # Operator deployment
│       │   ├── crds/                 # CRD manifests
│       │   └── gamedefinitions/      # Bundled GameDefinitions
│       │       ├── valheim.yaml
│       │       ├── satisfactory.yaml
│       │       ├── palworld.yaml
│       │       └── ...
│       └── README.md
├── web/                              # Web UI (future)
│   └── ...
├── Dockerfile
├── Makefile
└── go.mod
```

### Key Decisions

1. **Two-CRD Architecture**: GameDefinition + SteamServer separation enables extensibility without code changes. Users add games via YAML, not Go code.

2. **StatefulSet vs Deployment**: StatefulSet for stable network identity and ordered pod management. Game servers are inherently stateful.

3. **No Custom Images**: Uses `steamcmd/steamcmd:ubuntu-22` directly for both init and main containers. Game-specific behavior configured via command/args/env at runtime.

4. **Init Container for SteamCMD**: Runs steamcmd to download/validate game files before starting the game server. Same image, different entrypoint.

5. **Persistent Volume Strategy**: Single PVC mounted at `/data`. Game files and saves co-located (simplest approach for v1).

6. **Update Strategy**: For v1, require manual restart (delete pod). v2 could add smarter strategies.

7. **Service Type**: Default to LoadBalancer for cloud, but support NodePort for bare-metal.

### SteamCMD Init Container

No scripts generated. The operator builds args directly from GameDefinition + SteamServer specs:

```yaml
# Generated StatefulSet init container
initContainers:
- name: steamcmd
  image: steamcmd/steamcmd:ubuntu-22
  args:
  - "+login"
  - "anonymous"
  - "+force_install_dir"
  - "/data/server"
  - "+app_update"
  - "896660"        # from GameDefinition.spec.appId
  - "validate"      # if SteamServer.spec.validate: true
  - "+quit"
  volumeMounts:
  - name: game-data
    mountPath: /data
```

### Main Container

Command and args derived from GameDefinition, with config values interpolated:

```yaml
# Generated StatefulSet main container
containers:
- name: game-server
  image: steamcmd/steamcmd:ubuntu-22
  command:
  - "/data/server/valheim_server.x86_64"  # from GameDefinition.spec.command
  args:
  - "-name"
  - "Vikings Only"      # from SteamServer.spec.config.serverName
  - "-world"
  - "Midgard"           # from SteamServer.spec.config.worldName
  - "-password"
  - "$(SERVER_PASSWORD)"
  - "-port"
  - "2456"
  env:
  - name: SERVER_PASSWORD
    valueFrom:
      secretKeyRef:
        name: valheim-secrets
        key: password
  - name: SteamAppId
    value: "892970"     # from GameDefinition.spec.env
  volumeMounts:
  - name: game-data
    mountPath: /data
```

---

## Web UI / Configuration Frontend

A web-based frontend for configuring game servers without writing YAML by hand. Supports multiple deployment workflows.

### Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Web Frontend                                 │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │  - Game selector (dropdown of supported games)                  ││
│  │  - Dynamic form based on CRD schema                             ││
│  │  - Real-time YAML preview                                       ││
│  │  - Validation against CRD OpenAPI schema                        ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                │                                     │
│                                ▼                                     │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                     Output Options                               ││
│  │                                                                  ││
│  │   ┌─────────────┐   ┌─────────────┐   ┌─────────────────────┐  ││
│  │   │   Direct    │   │   Export    │   │   GitOps PR         │  ││
│  │   │   Apply     │   │   YAML      │   │   (GitHub/GitLab)   │  ││
│  │   └──────┬──────┘   └──────┬──────┘   └──────────┬──────────┘  ││
│  │          │                 │                      │             ││
│  │          ▼                 ▼                      ▼             ││
│  │   kubectl apply     Download/Copy        Create PR to           ││
│  │   to cluster        to clipboard         k8s manifests repo     ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Deployment Options

#### 1. Direct Apply to Cluster

- Frontend connects to K8s API (via backend proxy or service account)
- One-click deploy: form → CR → cluster
- Good for: quick iteration, dev environments, users with cluster access

```
[Web UI] → [Backend API] → [K8s API Server] → CR Created
```

#### 2. Export YAML

- Generate valid CR YAML from form inputs
- Copy to clipboard or download as file
- User applies manually: `kubectl apply -f server.yaml`
- Good for: review before apply, learning, manual workflows

#### 3. GitOps PR Integration

- Connect to GitHub/GitLab repository
- Frontend creates a branch and PR with the new/updated CR YAML
- Integrates with existing GitOps tooling (ArgoCD, Flux)
- Good for: production workflows, audit trail, team review

```
[Web UI] → [GitHub API] → PR Created → [ArgoCD] → Cluster Synced
```

### UI Features

#### Game Selection & Form Generation

```
┌──────────────────────────────────────────────────────────────┐
│  New Game Server                                              │
├──────────────────────────────────────────────────────────────┤
│  Game Type: [▼ Valheim                              ]        │
│                                                              │
│  ─── Basic Settings ───────────────────────────────────────  │
│  Server Name:    [ Vikings Only                    ]         │
│  World Name:     [ Midgard                         ]         │
│  Password:       [ ••••••••  ] (stored as Secret)           │
│                                                              │
│  ─── Game Options ─────────────────────────────────────────  │
│  [✓] Public Server                                           │
│  [✓] Enable Crossplay                                        │
│  Admins (Steam IDs):                                         │
│    [ 76561198000000001 ] [×]                                │
│    [ 76561198000000002 ] [×]                                │
│    [+ Add Admin]                                             │
│                                                              │
│  ─── Resources ────────────────────────────────────────────  │
│  CPU Request:    [ 2    ] cores                              │
│  Memory Request: [ 4    ] Gi                                 │
│  Storage Size:   [ 20   ] Gi                                 │
│                                                              │
│  ─── Deployment ───────────────────────────────────────────  │
│  Namespace:      [ game-servers                    ]         │
│  Service Type:   [▼ LoadBalancer                   ]         │
│                                                              │
│  ┌─────────┐  ┌─────────┐  ┌──────────────────────┐        │
│  │  Apply  │  │  YAML   │  │  Create PR (GitOps)  │        │
│  └─────────┘  └─────────┘  └──────────────────────┘        │
└──────────────────────────────────────────────────────────────┘
```

#### Live YAML Preview

Side panel or toggle showing real-time generated YAML as user fills form.

#### Server Management Dashboard

- List existing SteamServer CRs across namespaces
- Status overview (Running, Error, Updating)
- Quick actions: Restart, Edit, Delete
- Connection info (IP, ports, connect string)

### GitOps Integration Details

#### Configuration

```yaml
# UI config for GitOps mode
gitops:
  provider: github  # or gitlab, bitbucket
  repository: myorg/k8s-manifests
  baseBranch: main
  manifestPath: clusters/prod/game-servers/
  
  # PR settings
  branchPrefix: gameserver/
  autoMerge: false
  reviewers:
    - platform-team
  labels:
    - game-server
    - auto-generated
```

#### Generated PR Example

```
Title: [SteamServer] Add valheim-prod server

Description:
  Adds new Valheim dedicated server.
  
  Generated by SteamCMD Operator UI
  
  Server Details:
  - Game: Valheim
  - Name: Vikings Only
  - Resources: 2 CPU, 4Gi RAM
  
Files Changed:
  clusters/prod/game-servers/valheim-prod.yaml
```

### Tech Stack Options

| Component | Options |
|-----------|---------|
| Frontend | React, Vue, Svelte, or plain TypeScript |
| Backend | Go (shared with operator), Node.js, or static + client-side K8s |
| K8s Auth | Service Account, OIDC, kubeconfig upload |
| Git Integration | GitHub API, GitLab API, or generic git |

### Security Considerations

1. **Authentication**: Integrate with existing auth (OIDC, GitHub OAuth)
2. **Authorization**: Map users to K8s RBAC or namespace restrictions
3. **Secrets Handling**: Never expose secrets in UI; create K8s Secrets separately or via sealed-secrets
4. **GitOps Credentials**: Store GitHub/GitLab tokens securely; consider GitHub App for finer permissions

---

## Status & Observability

### CR Status Fields

| Field | Description |
|-------|-------------|
| `state` | Current state: Pending, Installing, Starting, Running, Error |
| `address` | External IP/hostname for the game server |
| `ports` | Map of port names to exposed ports |
| `lastUpdated` | Timestamp of last successful reconciliation |
| `appBuildId` | Current Steam build ID (detect when updates available) |
| `message` | Human-readable status message or error |

### Metrics (Future)

- `steamserver_status{name, namespace, state}` - Gauge of server states
- `steamserver_reconcile_duration_seconds` - Histogram of reconcile times
- `steamserver_players_current` - Current player count (if game supports query)

---

## Security Considerations

1. **Steam Credentials**: If non-anonymous login required, use K8s Secrets with proper RBAC
2. **Game Admin Passwords**: Always via secretKeyRef, never inline
3. **Network Policy**: Consider generating NetworkPolicies to limit traffic
4. **Pod Security**: Run as non-root where possible (some games require root unfortunately)

---

## Future Work (v2+)

### Operator Features
- [ ] Automatic backup CronJobs
- [ ] RCON integration for remote management
- [ ] Player-aware update strategy (wait for empty server)
- [ ] Multi-instance clustering (ARK clusters, etc.)
- [ ] Steam Workshop mod management

### Web UI Enhancements
- [ ] Real-time server logs streaming
- [ ] Player count / server query integration
- [ ] Backup management UI (trigger, restore, schedule)
- [ ] Multi-cluster support
- [ ] RCON console in browser
- [ ] Server templates / presets library

---

## Resolved Decisions

1. **Naming**: `SteamServer` for user-facing CR, `GameDefinition` for game configs ✓
2. **Group/Domain**: `boilerr.dev` ✓
3. **Image Strategy**: Use `steamcmd/steamcmd:ubuntu-22` directly, no custom images ✓
4. **Game Profiles**: `GameDefinition` CRD - bundled via Helm, users can add custom ✓
5. **GameDefinition Scope**: Cluster-scoped - available to all namespaces ✓

## Open Questions

1. **Web UI Deployment**: Bundled with operator? Separate deployment? Static site?
2. **Web UI Auth**: Rely on K8s RBAC? Separate user management? OAuth?
3. **Config Templating**: Go templates for GameDefinition args? Or simpler interpolation?

---

## References

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Operator SDK](https://sdk.operatorframework.io/)
- [SteamCMD Documentation](https://developer.valvesoftware.com/wiki/SteamCMD)
- [steamcmd/steamcmd Docker](https://github.com/steamcmd/docker) - Base container image used for game servers
- [minecraft-operator](https://github.com/itzg/minecraft-operator) - Similar concept, good reference
