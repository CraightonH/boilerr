# Adding Games to Boilerr

Guide for contributing new GameDefinitions to the Boilerr operator.

## Overview

Boilerr uses a **two-CRD pattern** to separate game configuration from server deployment:

1. **GameDefinition** (cluster-scoped) - Defines how to install and run a specific game. Maintained by operator developers/community.
2. **SteamServer** (namespaced) - User creates this to deploy a server instance, referencing a GameDefinition.

**Adding a new game = writing a YAML file, no Go code required.**

## Quick Start

1. Copy `gamedefinitions/TEMPLATE.yaml` to `gamedefinitions/<game-name>.yaml`
2. Find the game's Steam App ID (search SteamDB for "<game> dedicated server")
3. Fill in required fields: `appId`, `command`, `ports`
4. Add `configSchema` for user-configurable options
5. Test locally with `kubectl apply`
6. Submit a PR with your GameDefinition

## How GameDefinition Works

### Core Concept

A GameDefinition tells the operator:
- **Installation**: Which Steam app to download (`appId`)
- **Execution**: How to start the server (`command`, `args`)
- **Networking**: Which ports to expose (`ports`)
- **Configuration**: What users can customize (`configSchema`)

When a user creates a SteamServer referencing your GameDefinition, the operator:
1. Validates user config against your `configSchema`
2. Interpolates config values into `args`, `env`, and `configFiles`
3. Creates a StatefulSet, PVC, Service, and ConfigMap
4. Runs SteamCMD to install the game
5. Starts the server with the configured settings

### Required Fields

```yaml
apiVersion: boilerr.io/v1alpha1
kind: GameDefinition
metadata:
  name: valheim  # lowercase, hyphens
spec:
  appId: 896660  # Steam App ID
  command: ./valheim_server.x86_64  # Server executable
  ports:  # At least one port
    - name: game
      containerPort: 2456
      protocol: UDP
```

### Finding the App ID

1. Search SteamDB: `<game name> dedicated server`
2. Look for the **dedicated server** app ID, not the game client
3. Example: "Valheim" (game) is 892970, "Valheim Dedicated Server" is 896660

### Command and Args

The `command` field is the server executable, typically:
- Relative to `installDir` (default: `/data/server`)
- Or an absolute path

```yaml
command: ./valheim_server.x86_64  # Linux binary
# OR
command: /data/server/ShooterGame/Binaries/Linux/ShooterGameServer  # Absolute
```

The `args` field provides startup arguments. Use Go template syntax to reference user config:

```yaml
args:
  - -batchmode
  - -nographics
  - -port
  - "{{.Config.port}}"  # User provides config.port
  - -name
  - "{{.Config.serverName}}"  # User provides config.serverName
```

## ConfigSchema - The Powerful Part

`configSchema` defines what users can configure **declaratively**. Each entry becomes a field in `SteamServer.spec.config`.

### Basic String Config

```yaml
configSchema:
  serverName:
    description: "Server name shown in browser"
    default: "My Server"
```

User sets it:
```yaml
spec:
  config:
    serverName:
      value: "Vikings Only"
```

### Required Config

```yaml
worldName:
  description: "World/save file name"
  required: true  # User must provide
```

### Secret Config

For passwords, tokens, etc:

```yaml
password:
  description: "Server password"
  secret: true  # Indicates sensitive value
```

User references a Secret:
```yaml
spec:
  config:
    password:
      secretKeyRef:
        name: server-secrets
        key: password
```

### Enum (Limited Choices)

```yaml
difficulty:
  description: "Difficulty level"
  default: "normal"
  enum:
    - easy
    - normal
    - hard
```

### Array Config

```yaml
admins:
  description: "Admin Steam IDs"
  array: true
```

User provides comma-separated or multiple values:
```yaml
config:
  admins:
    value: "76561198012345678,76561198087654321"
```

## ConfigSchema Mapping Patterns

By default, config values are available in `args` templates as `{{.Config.key}}`. Use `mapTo` for advanced patterns.

### Pattern 1: Command Line Flag (Conditional)

Add a flag only if config is set to a specific value:

```yaml
crossplay:
  description: "Enable crossplay"
  default: "false"
  enum: ["true", "false"]
  mapTo:
    type: arg
    value: "-crossplay"
    condition: "true"  # Only add flag if user sets "true"
```

Result:
- User sets `crossplay.value = "true"` → args include `-crossplay`
- User sets `crossplay.value = "false"` → flag omitted

### Pattern 2: Environment Variable

Map config to an environment variable:

```yaml
maxPlayers:
  description: "Max players"
  default: "10"
  mapTo:
    type: env
    value: MAX_PLAYERS  # Env var name
```

Result: Container gets `MAX_PLAYERS=10` environment variable.

### Pattern 3: Config File Generation

Generate a config file from user values:

```yaml
serverSettings:
  description: "Server settings JSON"
  mapTo:
    type: configFile
    path: /data/server/settings.json
    template: |
      {
        "serverName": "{{.Config.serverName}}",
        "maxPlayers": {{.Config.maxPlayers}},
        "difficulty": "{{.Config.difficulty}}"
      }
```

Result: Operator creates a ConfigMap with this file, mounts it at the path.

### Pattern 4: INI File Section

For games using INI files:

```yaml
pvpEnabled:
  description: "Enable PvP"
  default: "false"
  enum: ["true", "false"]
  mapTo:
    type: configFile
    path: /data/server/game.ini
    template: |
      [Server]
      PvP={{.Config.pvpEnabled}}
      MaxPlayers={{.Config.maxPlayers}}
```

### Combining Patterns

Config can appear in both args and env:

```yaml
configSchema:
  port:
    description: "Game port"
    default: "2456"
    # No mapTo - available in args as {{.Config.port}}

  maxPlayers:
    description: "Max players"
    default: "10"
    mapTo:
      type: env
      value: MAX_PLAYERS

# Then in args:
args:
  - -port
  - "{{.Config.port}}"
  # maxPlayers becomes env var, not arg
```

## Port Configuration

Define all ports the game uses. Operator creates a Service with these ports.

```yaml
ports:
  - name: game        # Unique identifier
    containerPort: 2456
    protocol: UDP
  - name: query       # Server browser queries
    containerPort: 2457
    protocol: UDP
  - name: rcon        # Remote console
    containerPort: 2458
    protocol: TCP
```

Notes:
- `servicePort` defaults to `containerPort` if omitted
- `protocol` defaults to `UDP`
- At least one port required

## Static Config Files

For files that don't need user customization, use `configFiles`:

```yaml
configFiles:
  - path: /data/server/engine.ini
    content: |
      [Engine]
      MaxFPS=60
      LogLevel=Info
```

For dynamic files based on user config, use `configSchema` with `mapTo.type: configFile` instead.

## Resource Recommendations

Provide sensible defaults for CPU, memory, and storage:

```yaml
defaultResources:
  requests:
    memory: "4Gi"
    cpu: "2000m"
  limits:
    memory: "8Gi"
    cpu: "4000m"

defaultStorage: "30Gi"
```

Users can override these in their SteamServer.

## Health Checks

Define how to check server health:

```yaml
healthCheck:
  tcpSocket:
    port: game  # Port name from ports list
  initialDelaySeconds: 120  # Wait before first check
  periodSeconds: 30
```

Most games accept TCP connections on their game port even if using UDP.

## Step-by-Step Guide

### 1. Research the Game

Before writing YAML, gather:
- **App ID**: SteamDB entry for dedicated server
- **Installation**: Does SteamCMD install work? (`+app_update <id> validate`)
- **Executable**: Path to server binary after install
- **Arguments**: What flags/args does server accept? Check game docs/wikis
- **Ports**: What ports does server listen on?
- **Config**: How does game read config? (CLI args, env vars, INI files, JSON?)

### 2. Create GameDefinition YAML

```bash
cd gamedefinitions/
cp TEMPLATE.yaml valheim.yaml
```

Fill in required fields:

```yaml
apiVersion: boilerr.io/v1alpha1
kind: GameDefinition
metadata:
  name: valheim
spec:
  appId: 896660
  command: ./valheim_server.x86_64

  args:
    - -nographics
    - -batchmode
    - -port
    - "{{.Config.port}}"
    - -name
    - "{{.Config.serverName}}"
    - -world
    - "{{.Config.worldName}}"
    - -password
    - "{{.Config.password}}"

  ports:
    - name: game
      containerPort: 2456
      protocol: UDP
    - name: game-2
      containerPort: 2457
      protocol: UDP

  configSchema:
    serverName:
      description: "Server name in browser"
      default: "Valheim Server"

    worldName:
      description: "World save name"
      required: true

    port:
      description: "Game port"
      default: "2456"

    password:
      description: "Server password (5+ chars)"
      secret: true
      required: true

  defaultResources:
    requests:
      memory: "4Gi"
      cpu: "2000m"

  defaultStorage: "20Gi"
```

### 3. Test Locally

Apply the GameDefinition:

```bash
kubectl apply -f gamedefinitions/valheim.yaml
kubectl get gamedefinitions
```

Create a test SteamServer:

```yaml
# test-server.yaml
apiVersion: boilerr.io/v1alpha1
kind: SteamServer
metadata:
  name: test-valheim
  namespace: default
spec:
  game: valheim
  config:
    serverName:
      value: "Test Server"
    worldName:
      value: "TestWorld"
    port:
      value: "2456"
    password:
      value: "test123"  # In prod, use secretKeyRef
  serviceType: NodePort
```

```bash
kubectl apply -f test-server.yaml
kubectl get steamservers
kubectl describe steamserver test-valheim
```

Watch logs:

```bash
kubectl logs -f <pod-name>
```

Look for:
- SteamCMD install success
- Server startup
- Port binding
- No errors

### 4. Refine Configuration

Test different config values:
- Change port, server name, etc.
- Verify args are interpolated correctly
- Test enum values work
- Test secret references

### 5. Document in GameDefinition

Add helpful comments in your YAML:

```yaml
configSchema:
  serverName:
    description: "Server name shown in browser (max 50 chars)"
    default: "Valheim Server"

  crossplay:
    description: "Enable crossplay (requires public server)"
    default: "false"
    enum: ["true", "false"]
    mapTo:
      type: arg
      value: "-crossplay"
      condition: "true"
```

### 6. Submit PR

1. Commit your GameDefinition:
   ```bash
   git add gamedefinitions/valheim.yaml
   git commit -m "feat: add Valheim GameDefinition"
   ```

2. Create PR with description:
   - Game name and App ID
   - What you tested
   - Any quirks or limitations

3. In your PR, include example SteamServer YAML

## Testing Checklist

Before submitting:

- [ ] GameDefinition applies without errors
- [ ] SteamServer references it successfully
- [ ] SteamCMD installs game files
- [ ] Server starts and binds to ports
- [ ] Config values are interpolated correctly
- [ ] Service exposes ports (check `kubectl get svc`)
- [ ] Server is reachable from game client
- [ ] Required config validation works (try omitting required fields)
- [ ] Enum validation works (try invalid enum value)
- [ ] Secret references work (if using secrets)

## Common Pitfalls

### Port Mismatches

Make sure `containerPort` matches the port your server actually binds to:

```yaml
ports:
  - name: game
    containerPort: 2456  # Must match what server listens on
```

### Template Syntax Errors

Quote template values in args:

```yaml
# CORRECT:
args:
  - -port
  - "{{.Config.port}}"  # Quoted

# WRONG:
args:
  - -port
  - {{.Config.port}}  # Unquoted may cause YAML parsing issues
```

### Default vs Required

Don't make a field both `required: true` and have a `default`:

```yaml
# WRONG:
serverName:
  default: "My Server"
  required: true  # Contradictory

# CORRECT (pick one):
serverName:
  default: "My Server"  # Optional with default
# OR
worldName:
  required: true  # Required, no default
```

### Config Not Available in Args

If using `mapTo`, the config is **not** available in args templates:

```yaml
maxPlayers:
  mapTo:
    type: env
    value: MAX_PLAYERS

# DON'T use in args:
args:
  - -players
  - "{{.Config.maxPlayers}}"  # Won't work, it's mapped to env
```

Only configs **without** `mapTo` are available in args.

### Command Path

Use paths relative to `installDir` or absolute:

```yaml
# CORRECT:
command: ./valheim_server.x86_64  # Relative to /data/server

# WRONG:
command: valheim_server.x86_64  # Missing ./
```

## Advanced Patterns

### Multiple Config Files

Generate multiple files from one config:

```yaml
configSchema:
  serverConfig:
    mapTo:
      type: configFile
      path: /data/server/server.json
      template: |
        {"name": "{{.Config.serverName}}"}

  adminConfig:
    mapTo:
      type: configFile
      path: /data/server/admins.txt
      template: |
        {{.Config.admins}}
```

### Conditional Args

Add flags only when enabled:

```yaml
configSchema:
  enablePvP:
    default: "false"
    enum: ["true", "false"]

args:
  - "{{if eq .Config.enablePvP \"true\"}}-pvp{{end}}"
```

### Complex Template Logic

Use Go template features:

```yaml
template: |
  [Server]
  {{range split .Config.admins ","}}
  Admin={{.}}
  {{end}}
```

## Reference: CRD Field Documentation

See `api/v1alpha1/gamedefinition_types.go` for definitive field documentation.

Key types:
- `GameDefinitionSpec` - Top-level spec fields
- `ConfigSchemaEntry` - Config option definition
- `ConfigMapping` - How config maps to args/env/files
- `ServerPort` - Port definition
- `HealthCheckSpec` - Health check config

## Getting Help

- Check existing GameDefinitions in `gamedefinitions/` for examples
- Read `DESIGN.md` for architecture details
- Open an issue with "Question: Adding <game>" label
- Join community discussions

## Contributing Guidelines

- One game per file: `gamedefinitions/<game-name>.yaml`
- Use lowercase-with-hyphens for names
- Include comments explaining non-obvious config
- Test with a real server deployment
- Update this doc if you discover new patterns
