# Factorio Server

Deploy a Factorio dedicated server on Kubernetes using Boilerr.

## Quick Start

Minimal configuration to get a Factorio server running:

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: my-factorio-server
  namespace: games
spec:
  gameDefinition: factorio
  
  config:
    serverName: "My Factorio Server"
    
    password:
      secretKeyRef:
        name: factorio-secrets
        key: server-password
    
    rconPassword:
      secretKeyRef:
        name: factorio-secrets
        key: rcon-password
  
  storage:
    size: 20Gi
  
  serviceType: LoadBalancer
```

**Required Secrets:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: factorio-secrets
  namespace: games
type: Opaque
stringData:
  server-password: "your-server-password-here"
  rcon-password: "your-rcon-password-here"
```

### Minimum Required Configuration

| Setting | Description | Example |
|---------|-------------|---------|
| `serverName` | Server name shown in browser | `"My Factorio Server"` |
| `password` | Server password (Secret ref) | `secretKeyRef: {name: factorio-secrets, key: server-password}` |
| `rconPassword` | RCON admin password (Secret ref) | `secretKeyRef: {name: factorio-secrets, key: rcon-password}` |

## Advanced Configuration

### All Configuration Options

| Setting | Type | Default | Description | Example |
|---------|------|---------|-------------|---------|
| **Server Identity** |
| `serverName` | string | `"Factorio Server"` | Server name displayed in browser | `"Factory Must Grow"` |
| `description` | string | `"A Factorio server managed by Boilerr"` | Server description | `"Automation paradise"` |
| **Network** |
| `port` | string | `"34197"` | Game port (UDP) | `"34197"` |
| `maxPlayers` | string | `"0"` | Max players (0 = unlimited) | `"10"` |
| `visibility` | enum | `"public"` | Server visibility | `"public"`, `"lan"`, `"steam"` |
| `requireUserVerification` | enum | `"true"` | Require factorio.com login | `"true"`, `"false"` |
| **Authentication** |
| `password` | secret | - | Server password (optional) | Secret reference |
| `rconPassword` | secret | - | RCON password (required) | Secret reference |
| **Game Settings** |
| `gameName` | string | `"default"` | Save file name | `"factory1"` |
| `autosaveInterval` | string | `"5"` | Autosave interval (minutes) | `"10"` |
| `autosaveSlots` | string | `"3"` | Number of autosave slots | `"5"` |
| `afkAutokickInterval` | string | `"0"` | AFK kick timer (0 = disabled) | `"30"` |
| `allowCommands` | enum | `"admins-only"` | Console command permissions | `"true"`, `"false"`, `"admins-only"` |
| **Mods** |
| `updateMods` | enum | `"false"` | Auto-update mods on start | `"true"`, `"false"` |
| **Administration** |
| `rconPort` | string | `"27015"` | RCON port (TCP) | `"27015"` |
| `admins` | array | - | Admin usernames (comma-separated) | `"admin1,admin2"` |
| `whitelist` | array | - | Whitelisted users (empty = all allowed) | `"user1,user2"` |
| `banned` | array | - | Banned usernames | `"griefer1,griefer2"` |

### Complete Example

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: advanced-factorio
  namespace: games
spec:
  gameDefinition: factorio
  
  config:
    # Server identity
    serverName: "The Factory Must Grow"
    description: "A dedicated server for automation enthusiasts"
    
    # Network settings
    port: "34197"
    maxPlayers: "20"
    visibility: "public"
    requireUserVerification: "true"
    
    # Authentication
    password:
      secretKeyRef:
        name: factorio-secrets
        key: server-password
    
    rconPassword:
      secretKeyRef:
        name: factorio-secrets
        key: rcon-password
    
    # Game settings
    gameName: "megabase"
    autosaveInterval: "10"
    autosaveSlots: "10"
    afkAutokickInterval: "30"
    allowCommands: "admins-only"
    
    # Mods
    updateMods: "true"
    
    # Administration
    rconPort: "27015"
    admins: "admin1,admin2,admin3"
    whitelist: ""  # Empty = allow all
    banned: "griefer123"
  
  # Resource overrides (optional)
  resources:
    requests:
      memory: "4Gi"
      cpu: "2000m"
    limits:
      memory: "8Gi"
      cpu: "4000m"
  
  # Storage configuration
  storage:
    size: 30Gi
    storageClassName: fast-ssd
  
  # Service type
  serviceType: LoadBalancer
```

## Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 34197 | UDP | Game traffic |
| 27015 | TCP | RCON administration |

## Resource Requirements

**Default resources:**
- CPU: 1-2 cores (can be CPU-intensive with large factories)
- Memory: 2-4 GiB (scales with factory complexity and player count)
- Storage: 10 GiB default (save files, mods)

**Recommended for large factories:**
- CPU: 2-4 cores
- Memory: 4-8 GiB
- Storage: 20-30 GiB

## Notes

- **Factorio account**: Public servers require a factorio.com account for listing
- **RCON**: Remote administration requires `rconPassword` to be set
- **Usernames**: Admin, whitelist, and ban lists use factorio.com usernames (not Steam IDs)
- **Mods**: Place mods in the persistent storage volume at `/factorio/mods/`
- **Saves**: Server saves are stored in `/factorio/saves/`
- **Image**: Uses `factoriotools/factorio:stable` (Factorio doesn't support SteamCMD)

## Troubleshooting

**Server not appearing in browser:**
- Check `visibility` is set to `"public"` or `"steam"`
- Verify `requireUserVerification` is configured correctly
- Ensure ports are accessible (LoadBalancer/NodePort configured)

**Can't connect to RCON:**
- Verify `rconPassword` secret is set
- Check `rconPort` is accessible (usually TCP 27015)
- Ensure firewall allows TCP traffic on RCON port

**Server crashes or restarts:**
- Check resource limits (factories can be memory-intensive)
- Review logs for mod conflicts
- Verify save file isn't corrupted

## Admin Tasks

**Connect to RCON:**
```bash
# Using kubectl port-forward
kubectl port-forward -n games svc/my-factorio-server 27015:27015

# Then connect with an RCON client
rcon -H localhost -P 27015 -p your-rcon-password
```

**Managing saves:**
```bash
# List saves
kubectl exec -n games my-factorio-server-0 -- ls -lh /factorio/saves/

# Backup a save
kubectl cp games/my-factorio-server-0:/factorio/saves/megabase.zip ./megabase-backup.zip

# Restore a save
kubectl cp ./megabase-backup.zip games/my-factorio-server-0:/factorio/saves/megabase.zip
```

**Installing mods:**
```bash
# Copy mod zip to server
kubectl cp mod-name_1.0.0.zip games/my-factorio-server-0:/factorio/mods/

# Restart server to load mods
kubectl rollout restart -n games statefulset/my-factorio-server
```
