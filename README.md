# Boiler 🔥

A Kubernetes operator for managing Steam dedicated game servers.

🚧 **Status: Design Phase** 🚧

See [DESIGN.md](./DESIGN.md) for the design document.

## Overview

Boiler simplifies deploying and managing Steam dedicated game servers on Kubernetes. Define a custom resource, and Boiler handles the rest - SteamCMD downloads, persistent storage, networking, and lifecycle management.

```yaml
apiVersion: boiler.dev/v1alpha1
kind: ValheimServer
metadata:
  name: vikings-only
spec:
  serverName: "Vikings Only"
  worldName: "Midgard"
  password:
    secretKeyRef:
      name: valheim-secrets
      key: password
  storage:
    size: 20Gi
```

## Supported Games

- Valheim
- Satisfactory
- Palworld
- Factorio
- 7 Days to Die
- V Rising
- Enshrouded
- Terraria

Plus a generic `SteamServer` CRD for any SteamCMD-compatible game.

## Quick Start

*Coming soon*

## License

TBD
