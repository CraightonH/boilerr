# SteamCMD Operator Design Document

## Overview

A Kubernetes operator that simplifies deploying and managing Steam dedicated game servers. Instead of writing bespoke Dockerfiles and K8s manifests for each game, users define a `SteamServer` custom resource and the operator handles the rest.

### Goals

- **Generic by default**: Support any game SteamCMD can install with minimal config
- **Extensible**: Allow game-specific CRDs for first-class support of popular titles
- **Low overhead**: Reduce boilerplate for spinning up new game servers
- **Kubernetes-native**: Leverage existing K8s primitives (StatefulSets, PVCs, Services)

### Non-Goals (v1)

- Automatic backups and restore
- Player-aware update strategies (drain before restart)
- Multi-server clustering (e.g., ARK clusters)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│                                                              │
│  ┌──────────────┐       ┌─────────────────────────────────┐ │
│  │  SteamServer │       │         Operator Pod            │ │
│  │      CR      │──────▶│  - Watches SteamServer CRs      │ │
│  └──────────────┘       │  - Reconciles to K8s resources  │ │
│                         └─────────────────────────────────┘ │
│                                      │                       │
│                                      ▼                       │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Per-Game Resources                     ││
│  │  ┌─────────┐  ┌──────────────┐  ┌─────────────────────┐ ││
│  │  │   PVC   │  │ StatefulSet  │  │      Service        │ ││
│  │  │ (saves) │  │ - init: SCMD │  │ (LoadBalancer/NP)   │ ││
│  │  │         │  │ - main: game │  │                     │ ││
│  │  └─────────┘  └──────────────┘  └─────────────────────┘ ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Reconciliation Loop

1. Watch for `SteamServer` (or game-specific) CR events
2. Generate desired state: PVC, StatefulSet, Service, ConfigMaps
3. Compare with actual state in cluster
4. Apply diffs (create/update/delete resources)
5. Update CR status with server state, IP, ports

---

## Custom Resource Definitions

### Option A: Single Generic CRD

One CRD that works for any game. Users must provide all configuration.

```yaml
apiVersion: boiler.dev/v1alpha1
kind: SteamServer
metadata:
  name: valheim-prod
  namespace: game-servers
spec:
  # Required: Steam App ID for the dedicated server
  appId: 896660
  
  # Optional: Beta branch
  # beta: "public-test"
  
  # Optional: Run steamcmd validate on startup (default: true)
  validate: true
  
  # Optional: Anonymous login. If false, provide credentials secret
  anonymous: true
  # steamCredentialsSecret: steam-login  # if anonymous: false
  
  # Container image for steamcmd (operator may provide default)
  image: steamcmd/steamcmd:latest
  
  # Ports to expose
  ports:
    - name: game
      containerPort: 2456
      servicePort: 2456
      protocol: UDP
    - name: query
      containerPort: 2457
      servicePort: 2457
      protocol: UDP
  
  # Startup command and arguments
  command: ["/bin/bash", "-c"]
  args:
    - |
      cd /serverfiles
      ./valheim_server.x86_64 \
        -name "$(SERVER_NAME)" \
        -world "$(WORLD_NAME)" \
        -password "$(SERVER_PASSWORD)" \
        -port 2456 \
        -public 0
  
  # Environment variables
  env:
    - name: SERVER_NAME
      value: "My Valheim Server"
    - name: WORLD_NAME
      value: "Midgard"
    - name: SERVER_PASSWORD
      valueFrom:
        secretKeyRef:
          name: valheim-secrets
          key: password
  
  # Config files to mount
  configFiles:
    - path: /serverfiles/admins.txt
      content: |
        76561198000000001
        76561198000000002
  
  # Persistent storage
  storage:
    size: 20Gi
    storageClassName: fast-ssd  # optional
  
  # Resource requests/limits
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
  
  # Service type for game ports
  serviceType: LoadBalancer  # or NodePort, ClusterIP

status:
  state: Running  # Pending, Installing, Running, Error
  address: 192.168.1.100
  ports:
    - name: game
      port: 2456
  lastUpdated: "2026-01-18T05:30:00Z"
  appBuildId: "12345678"  # Steam build ID, useful for tracking updates
```

### Option B: Game-Specific CRDs (Layered)

For popular games, provide first-class CRDs with sane defaults and typed config.

#### ValheimServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: ValheimServer
metadata:
  name: valheim-prod
spec:
  # Game-specific, typed configuration
  serverName: "Vikings Only"
  worldName: "Midgard"
  password:
    secretKeyRef:
      name: valheim-secrets
      key: password
  
  # Valheim-specific options
  public: false
  crossplay: true
  
  # Admins by Steam ID
  admins:
    - "76561198000000001"
    - "76561198000000002"
  
  # Common fields still available
  storage:
    size: 20Gi
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
  serviceType: LoadBalancer

status:
  state: Running
  address: 192.168.1.100
  gamePort: 2456
  version: "0.217.22"
```

#### SatisfactoryServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: SatisfactoryServer
metadata:
  name: satisfactory-prod
spec:
  serverName: "Factory Must Grow"
  maxPlayers: 8
  
  # Satisfactory-specific
  autoPause: true
  autoSaveOnDisconnect: true
  networkQuality: 3  # 0-3
  
  storage:
    size: 50Gi  # Satisfactory saves get big
  resources:
    requests:
      cpu: "4"
      memory: "12Gi"
```

#### PalworldServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: PalworldServer
metadata:
  name: palworld-prod
spec:
  serverName: "Pal Paradise"
  adminPassword:
    secretKeyRef:
      name: palworld-secrets
      key: admin-password
  serverPassword:
    secretKeyRef:
      name: palworld-secrets
      key: server-password
  
  # Palworld-specific settings
  maxPlayers: 32
  difficulty: Normal  # Casual, Normal, Hard
  deathPenalty: ItemAndEquipment  # None, Item, ItemAndEquipment, All
  
  # Rate multipliers
  expRate: 1.5
  captureRate: 1.0
  
  storage:
    size: 30Gi
  resources:
    requests:
      cpu: "4"
      memory: "16Gi"
```

#### FactorioServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: FactorioServer
metadata:
  name: factorio-prod
spec:
  # App ID: 427520
  serverName: "The Factory Must Grow"
  description: "A friendly Factorio server"
  
  # Game settings
  maxPlayers: 16
  visibility:
    public: false
    lan: true
  
  # Credentials for public listing (optional)
  factorioCredentials:
    secretKeyRef:
      name: factorio-secrets
      key: credentials  # username:token format
  
  # Gameplay settings
  autosaveInterval: 10  # minutes
  autosaveSlots: 5
  afkKickMinutes: 30
  
  # Allow commands (admins-only, true, false)
  allowCommands: admins-only
  
  # Admins by Factorio username
  admins:
    - "player1"
    - "player2"
  
  # Mods (optional - list of mod names, fetched from mod portal)
  # mods:
  #   - "space-exploration"
  #   - "Krastorio2"
  
  storage:
    size: 10Gi
  resources:
    requests:
      cpu: "2"
      memory: "2Gi"
```

#### SevenDaysServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: SevenDaysServer
metadata:
  name: 7dtd-prod
spec:
  # App ID: 294420
  serverName: "Survival Server"
  serverPassword:
    secretKeyRef:
      name: 7dtd-secrets
      key: password
  
  # Server settings
  maxPlayers: 8
  serverPort: 26900
  
  # World settings
  worldGenSeed: "apocalypse"
  worldGenSize: 6144  # 4096, 6144, or 8192
  gameWorld: RWG  # or Navezgane
  gameName: "MyWorld"
  
  # Gameplay
  gameMode: GameModeSurvival
  difficulty: 3  # 0-5
  
  # Day/night cycle
  dayNightLength: 60  # real minutes per in-game day
  dayLightLength: 18  # in-game hours of daylight
  
  # Zombies
  zombieMove: 0  # 0=walk, 1=jog, 2=run, 3=sprint, 4=nightmare
  zombieMoveNight: 3
  bloodMoonFrequency: 7
  bloodMoonEnemyCount: 8
  
  # Admin
  adminPassword:
    secretKeyRef:
      name: 7dtd-secrets
      key: admin-password
  telnetEnabled: false
  
  storage:
    size: 30Gi
  resources:
    requests:
      cpu: "4"
      memory: "8Gi"
```

#### VRisingServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: VRisingServer
metadata:
  name: vrising-prod
spec:
  # App ID: 1829350
  serverName: "Vampire Domain"
  serverDescription: "PvE Chill Server"
  password:
    secretKeyRef:
      name: vrising-secrets
      key: password
  
  # Server settings
  maxPlayers: 40
  gamePort: 9876
  queryPort: 9877
  
  # Game mode
  gameMode: PvE  # PvE, PvP, FullLoot
  difficultyPreset: Normal  # Easy, Normal, Hard, Brutal
  
  # Clan settings
  clanSize: 4
  
  # Game rules (selected settings, full list is extensive)
  sunDamageModifier: 1.0
  castleDamageMode: Never  # Never, Always, TimeRestricted
  
  # Auto-save
  autoSaveCount: 25
  autoSaveInterval: 120  # seconds
  
  # RCON (optional)
  rcon:
    enabled: true
    password:
      secretKeyRef:
        name: vrising-secrets
        key: rcon-password
    port: 25575
  
  storage:
    size: 20Gi
  resources:
    requests:
      cpu: "2"
      memory: "6Gi"
```

#### EnshroudedServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: EnshroudedServer
metadata:
  name: enshrouded-prod
spec:
  # App ID: 2278520
  serverName: "Embervale Explorers"
  password:
    secretKeyRef:
      name: enshrouded-secrets
      key: password
  
  # Server settings
  maxPlayers: 16
  gamePort: 15636
  queryPort: 15637
  
  # Performance
  saveDirectory: "./savegame"
  logDirectory: "./logs"
  
  storage:
    size: 15Gi
  resources:
    requests:
      cpu: "2"
      memory: "8Gi"  # Enshrouded is memory-hungry
```

#### TerrariaServer

```yaml
apiVersion: boiler.dev/v1alpha1
kind: TerrariaServer
metadata:
  name: terraria-prod
spec:
  # App ID: 105600 (vanilla) - or use tShock
  variant: tshock  # vanilla or tshock
  
  serverName: "Terraria World"
  password:
    secretKeyRef:
      name: terraria-secrets
      key: password
  
  # World settings
  worldName: "MyWorld"
  worldSize: medium  # small, medium, large
  difficulty: normal  # classic, expert, master, journey
  
  # Server settings
  maxPlayers: 8
  port: 7777
  
  # World gen (for new worlds)
  seed: ""  # empty = random
  worldEvil: random  # corruption, crimson, random
  
  # Gameplay
  spawnProtection: true
  announcePlayerJoinLeave: true
  
  # tShock-specific (only if variant: tshock)
  tshock:
    restApiEnabled: true
    restApiPort: 7878
    
  storage:
    size: 5Gi
  resources:
    requests:
      cpu: "1"
      memory: "1Gi"
```

### Game CRD Reference Table

| Game | CRD Kind | App ID | Default Ports | Min Resources |
|------|----------|--------|---------------|---------------|
| Valheim | `ValheimServer` | 896660 | 2456-2457/UDP | 2 CPU, 4Gi |
| Satisfactory | `SatisfactoryServer` | 1690800 | 7777/UDP, 15000/UDP | 4 CPU, 12Gi |
| Palworld | `PalworldServer` | 2394010 | 8211/UDP | 4 CPU, 16Gi |
| Factorio | `FactorioServer` | 427520 | 34197/UDP | 2 CPU, 2Gi |
| 7 Days to Die | `SevenDaysServer` | 294420 | 26900-26902/TCP+UDP | 4 CPU, 8Gi |
| V Rising | `VRisingServer` | 1829350 | 9876-9877/UDP | 2 CPU, 6Gi |
| Enshrouded | `EnshroudedServer` | 2278520 | 15636-15637/UDP | 2 CPU, 8Gi |
| Terraria | `TerrariaServer` | 105600 | 7777/TCP | 1 CPU, 1Gi |

### Planned / Upcoming Games

| Game | Status | Notes |
|------|--------|-------|
| RuneScape: Dragonwilds | ⏳ Waiting | Jagex survival crafting game. Dedicated servers on their 2026 roadmap. Add CRD once server binaries are available. |

---

## Implementation Details

### Directory Structure

```
boiler/
├── api/
│   └── v1alpha1/
│       ├── steamserver_types.go      # Generic CRD
│       ├── valheimserver_types.go    # Game-specific (optional)
│       └── groupversion_info.go
├── controllers/
│   ├── steamserver_controller.go
│   └── valheimserver_controller.go
├── internal/
│   ├── resources/
│   │   ├── statefulset.go            # StatefulSet builder
│   │   ├── service.go                # Service builder
│   │   └── pvc.go                    # PVC builder
│   └── steamcmd/
│       └── scripts.go                # SteamCMD init scripts
├── config/
│   ├── crd/                          # Generated CRD YAML
│   ├── rbac/                         # RBAC manifests
│   └── manager/                      # Operator deployment
├── web/                              # Web UI
│   ├── src/
│   │   ├── components/
│   │   │   ├── GameForm.tsx          # Dynamic form from CRD schema
│   │   │   ├── YamlPreview.tsx       # Live YAML preview
│   │   │   └── ServerList.tsx        # Dashboard of existing servers
│   │   ├── lib/
│   │   │   ├── k8s.ts                # K8s API client
│   │   │   ├── github.ts             # GitHub PR integration
│   │   │   └── schema.ts             # CRD schema → form mapping
│   │   └── pages/
│   │       ├── index.tsx             # Dashboard
│   │       ├── new.tsx               # Create server form
│   │       └── edit/[name].tsx       # Edit existing server
│   ├── package.json
│   └── Dockerfile
├── Dockerfile
├── Makefile
└── go.mod
```

### Key Decisions

1. **StatefulSet vs Deployment**: StatefulSet for stable network identity and ordered pod management. Game servers are inherently stateful.

2. **Init Container for SteamCMD**: Runs steamcmd to download/validate game files before starting the game server. Allows separation of concerns.

3. **Persistent Volume Strategy**: Single PVC mounted at a consistent path. Game files and saves co-located (simplest approach for v1).

4. **Update Strategy**: For v1, require manual restart (delete pod). v2 could add smarter strategies.

5. **Service Type**: Default to LoadBalancer for cloud, but support NodePort for bare-metal.

### SteamCMD Init Script (Generated)

```bash
#!/bin/bash
set -e

steamcmd \
  +force_install_dir /serverfiles \
  +login anonymous \
  +app_update ${APP_ID} ${VALIDATE_FLAG} \
  +quit

echo "SteamCMD complete, starting server..."
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
- [ ] Helm chart for easy operator installation

### Web UI Enhancements
- [ ] Real-time server logs streaming
- [ ] Player count / server query integration
- [ ] Backup management UI (trigger, restore, schedule)
- [ ] Multi-cluster support
- [ ] RCON console in browser
- [ ] Server templates / presets library

---

## Open Questions

1. **Naming**: `SteamServer`? `GameServer`? `DedicatedServer`?
2. **Group/Domain**: `boiler.dev`? `gameserver.io`?
3. **Image Strategy**: Operator-provided base image or user brings their own?
4. **Game Profiles**: Ship a library of known-good configs for popular games?
5. **Web UI Deployment**: Bundled with operator? Separate deployment? Static site?
6. **Web UI Auth**: Rely on K8s RBAC? Separate user management? OAuth?

---

## References

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Operator SDK](https://sdk.operatorframework.io/)
- [SteamCMD Documentation](https://developer.valvesoftware.com/wiki/SteamCMD)
- [minecraft-operator](https://github.com/itzg/minecraft-operator) - Similar concept, good reference
