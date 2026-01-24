# Container Images

Boilerr uses two distinct container images:

## 1. Operator Image

**Purpose:** Runs the Kubernetes operator controller that manages GameDefinitions and SteamServers.

**Built from:** `Dockerfile` in repository root

**Base image:** `gcr.io/distroless/static:nonroot`

**Registries:**
- GitHub Container Registry: `ghcr.io/craightonh/boilerr`
- Docker Hub: `craighton/boilerr`

**Tags:**
- `latest` - Latest stable release
- `vX.Y.Z` - Semantic version tags
- `vX.Y` - Minor version tags
- `vX` - Major version tags
- `<commit-sha>` - Development builds from main branch

**Platforms:** `linux/amd64`, `linux/arm64`

**Build process:**
- Multi-stage build optimized for size
- Includes version, commit, and build date metadata
- Signed with cosign (keyless)
- Automated builds via GitHub Actions

**Verification:**
```bash
# Verify image signature
cosign verify ghcr.io/craightonh/boilerr:latest \
  --certificate-identity-regexp=https://github.com/CraightonH/boilerr \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com

# Pull operator image
docker pull ghcr.io/craightonh/boilerr:latest
# or
docker pull craighton/boilerr:latest
```

## 2. Game Server Base Image

**Purpose:** Base container for running Steam dedicated game servers.

**Image:** `steamcmd/steamcmd:ubuntu-22`

**Source:** [steamcmd/docker](https://github.com/steamcmd/docker)

**Registry:** Docker Hub

**Why this image:**
- Official SteamCMD Docker image
- Ubuntu 22.04 LTS base
- Pre-installed dependencies for Steam servers
- Maintained by the community
- No custom game-specific images needed

**How it's used:**

Each GameDefinition can specify a container image (defaults to `steamcmd/steamcmd:ubuntu-22`):

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: GameDefinition
metadata:
  name: valheim
spec:
  image: steamcmd/steamcmd:ubuntu-22  # Optional, this is the default
  appId: 896660
  command: ./valheim_server.x86_64
  # ...
```

The operator creates a StatefulSet with:
1. **Init container** (using `steamcmd/steamcmd:ubuntu-22`):
   - Runs SteamCMD to download/update game server files
   - Uses persistent volume for game data

2. **Main container** (using same image):
   - Runs the game server executable
   - Mounts same persistent volume

**Why no custom images per game?**

Boilerr takes a **manifest-driven approach** instead of building custom images:

✅ **Advantages:**
- No Docker image builds required for adding games
- Faster updates (just update GameDefinition YAML)
- Community can add games via simple PRs (no registry access needed)
- Single base image to maintain
- Users can override image if needed

❌ **Traditional approach** (build images per game):
- Requires CI/CD for each game
- Registry storage for every game variant
- Harder for community to contribute
- Slower to add new games

**Override example:**

If a game needs special dependencies, users can build a custom image and override:

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: special-server
spec:
  gameDefinition: my-game
  image: myregistry/custom-game-image:latest  # Override default
```

## Image Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Operator Deployment                                 │    │
│  │  Image: ghcr.io/craightonh/boilerr:latest          │    │
│  │  - Watches GameDefinition CRDs                       │    │
│  │  - Watches SteamServer CRDs                          │    │
│  │  - Reconciles StatefulSets/Services/PVCs            │    │
│  └─────────────────────────────────────────────────────┘    │
│                          │                                    │
│                          │ creates/manages                    │
│                          ▼                                    │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  SteamServer StatefulSet (per game instance)        │    │
│  │  Image: steamcmd/steamcmd:ubuntu-22                 │    │
│  │                                                       │    │
│  │  Init Container:                                     │    │
│  │    - Run SteamCMD                                    │    │
│  │    - Download game (appId from GameDefinition)      │    │
│  │    - Install to /data/server                         │    │
│  │                                                       │    │
│  │  Main Container:                                     │    │
│  │    - Run game server (command from GameDefinition)  │    │
│  │    - Serve players                                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Build & Push

**Manual build:**
```bash
# Build locally
make docker-build IMG=ghcr.io/craightonh/boilerr:dev

# Push to registry
make docker-push IMG=ghcr.io/craightonh/boilerr:dev

# Multi-platform build (requires buildx)
make docker-buildx IMG=ghcr.io/craightonh/boilerr:dev
```

**Automated builds:**
- PRs and main branch pushes → GHCR (untagged, SHA-based)
- Release tags (`v*.*.*`) → GHCR + Docker Hub (versioned + signed)

## Security

**Operator image:**
- Distroless base (minimal attack surface)
- Non-root user (UID 65532)
- No shell or package managers
- Signed with cosign
- Scanned for vulnerabilities in CI

**Game server image:**
- Community-maintained steamcmd/steamcmd image
- Ubuntu 22.04 LTS base
- Regular updates via upstream
- Can be overridden if security requirements dictate

**Recommendations:**
- Use specific version tags in production (`v0.3.0`), not `latest`
- Verify cosign signatures before deployment
- Pin `steamcmd/steamcmd` to specific SHA if immutability required
- Use private registry mirrors for air-gapped environments

## Troubleshooting

**Operator image pull failures:**
```bash
# Check registry credentials
kubectl get secret -n boilerr-system

# Pull manually to test
docker pull ghcr.io/craightonh/boilerr:latest
```

**Game server image pull failures:**
```bash
# Verify steamcmd image is accessible
docker pull steamcmd/steamcmd:ubuntu-22

# Check node image cache
kubectl describe pod <pod-name> -n <namespace>
```

**Image override not working:**
```bash
# Check StatefulSet spec
kubectl get sts <server-name> -o yaml | grep image:

# Verify override in SteamServer CR
kubectl get steamserver <name> -o jsonpath='{.spec.image}'
```

## References

- [Dockerfile best practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Distroless images](https://github.com/GoogleContainerTools/distroless)
- [steamcmd/docker](https://github.com/steamcmd/docker)
- [Cosign keyless signing](https://docs.sigstore.dev/cosign/overview/)
