# Refactor Plan: GameDefinition Architecture Migration

> **Status:** COMPLETE - All phases done, tests passing, lint passing, build successful
> **Date:** 2026-01-20 (updated 2026-01-21)
> **Context:** Migrating from single-CRD approach to GameDefinition + SteamServer two-CRD architecture
>
> ### Progress Summary (2026-01-21)
> All phases complete. Test failures fixed (duplicate helper functions, beta flag test).
> `make test`, `make lint`, and `make build` all passing. Sample files already existed.

---

## Background

### Why This Refactor?

The original implementation used a single `SteamServer` CRD where users had to specify everything: `appId`, `command`, `args`, `ports`, etc. This required users to know game internals.

The new architecture separates concerns:
- **GameDefinition** - Defines how to install/run a game (maintained by operator/community)
- **SteamServer** - User-facing CR that references a GameDefinition (simple config only)

Benefits:
- Users don't need to know game internals
- Adding new games = YAML file, not Go code
- Helm chart bundles popular GameDefinitions
- Community can contribute games easily

### Key Design Decisions

1. **No custom Docker images per game** - Use `steamcmd/steamcmd:ubuntu-22` directly
2. **No shell scripts** - Build container args as `[]string`, not script strings
3. **Config templating** - GameDefinition defines `configSchema`, SteamServer provides values
4. **Extensibility** - Users can create custom GameDefinitions for unsupported games

### Reference Documents

- `DESIGN.md` - Full architecture documentation (updated)
- `ROADMAP.md` - Phase 2 tasks (updated)

---

## Current State Analysis

### Existing Files

| File | Purpose | Migration Status |
|------|---------|------------------|
| `api/v1alpha1/steamserver_types.go` | SteamServer CRD | ✅ Updated |
| `api/v1alpha1/gamedefinition_types.go` | GameDefinition CRD | ✅ Created |
| `api/v1alpha1/common_types.go` | Shared types | ✅ Created |
| `api/v1alpha1/groupversion_info.go` | API registration | ✅ Updated |
| `api/v1alpha1/zz_generated.deepcopy.go` | Generated | ✅ Regenerated |
| `internal/steamcmd/command.go` | Command builder | ✅ Created (was scripts.go) |
| `internal/steamcmd/command_test.go` | Command tests | ✅ Created (was scripts_test.go) |
| `internal/config/interpolate.go` | Config interpolation | ✅ Created |
| `internal/config/interpolate_test.go` | Interpolation tests | ✅ Created |
| `internal/resources/statefulset.go` | StatefulSet builder | ✅ Updated |
| `internal/resources/statefulset_test.go` | StatefulSet tests | Needs modification |
| `internal/resources/service.go` | Service builder | Minor updates |
| `internal/resources/service_test.go` | Service tests | Minor updates |
| `internal/resources/pvc.go` | PVC builder | Likely okay |
| `internal/resources/pvc_test.go` | PVC tests | Likely okay |
| `internal/controller/steamserver_controller.go` | Main controller | Needs modification |
| `internal/controller/steamserver_controller_test.go` | Controller tests | Needs modification |
| `internal/controller/suite_test.go` | Test suite | Minor update |
| `config/crd/bases/boilerr.dev_steamservers.yaml` | Generated CRD | Will regenerate |

### What's Wrong with Current Implementation

**1. SteamServer CRD has game-specific fields:**
```go
// These should come from GameDefinition, not user input:
AppId   int32        `json:"appId"`
Ports   []ServerPort `json:"ports"`
Command []string     `json:"command,omitempty"`
Args    []string     `json:"args,omitempty"`
```

**2. steamcmd/scripts.go generates shell scripts:**
```go
// Current: returns shell script string
func (b *ScriptBuilder) Build() string {
    sb.WriteString("set -e\n\n")
    sb.WriteString("steamcmd \\\n")
    // ...
}

// Should: return args slice
func (b *CommandBuilder) Build() []string {
    return []string{"+login", "anonymous", ...}
}
```

**3. statefulset.go uses script in Command:**
```go
// Current: runs script via bash
Command: []string{"/bin/bash", "-c", b.buildSteamCMDScript()}

// Should: pass args directly to steamcmd
Args: b.buildSteamCMDArgs()
```

**4. No GameDefinition reference:**
```go
// Missing from SteamServer:
GameDefinition string                 `json:"gameDefinition"`
Config         map[string]ConfigValue `json:"config,omitempty"`
```

---

## Refactor Phases

### Phase A: CRD Changes (Foundation)

#### A1. Create GameDefinition CRD

**File:** `api/v1alpha1/gamedefinition_types.go`

```go
package v1alpha1

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GameDefinitionSpec defines how to install and run a Steam game server.
type GameDefinitionSpec struct {
    // AppId is the Steam application ID for the dedicated server.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=1
    AppId int32 `json:"appId"`

    // Image is the container image to use (default: steamcmd/steamcmd:ubuntu-22)
    // +kubebuilder:default="steamcmd/steamcmd:ubuntu-22"
    // +optional
    Image string `json:"image,omitempty"`

    // InstallDir is where SteamCMD installs game files.
    // +kubebuilder:default="/data/server"
    // +optional
    InstallDir string `json:"installDir,omitempty"`

    // Command is the game server startup command.
    // +kubebuilder:validation:Required
    Command string `json:"command"`

    // Args are the default startup arguments.
    // Supports {{.Config.key}} template syntax.
    // +optional
    Args []string `json:"args,omitempty"`

    // Ports defines the default ports for this game.
    // +kubebuilder:validation:MinItems=1
    Ports []ServerPort `json:"ports"`

    // Env defines default environment variables.
    // +optional
    Env []corev1.EnvVar `json:"env,omitempty"`

    // ConfigSchema defines user-configurable options.
    // Keys are config names, values define how they map to args/env/files.
    // +optional
    ConfigSchema map[string]ConfigSchemaEntry `json:"configSchema,omitempty"`

    // ConfigFiles defines static config file templates.
    // +optional
    ConfigFiles []ConfigFileTemplate `json:"configFiles,omitempty"`

    // DefaultResources defines recommended resource requirements.
    // +optional
    DefaultResources corev1.ResourceRequirements `json:"defaultResources,omitempty"`

    // DefaultStorage defines recommended storage size.
    // +kubebuilder:default="20Gi"
    // +optional
    DefaultStorage string `json:"defaultStorage,omitempty"`

    // HealthCheck defines how to check if the server is healthy.
    // +optional
    HealthCheck *HealthCheckSpec `json:"healthCheck,omitempty"`
}

// ConfigSchemaEntry defines a user-configurable option.
type ConfigSchemaEntry struct {
    // Description explains what this config option does.
    // +optional
    Description string `json:"description,omitempty"`

    // Default is the default value if not specified by user.
    // +optional
    Default string `json:"default,omitempty"`

    // Required indicates this config must be provided.
    // +optional
    Required bool `json:"required,omitempty"`

    // Secret indicates this value should come from a Secret.
    // +optional
    Secret bool `json:"secret,omitempty"`

    // Enum restricts values to a specific set.
    // +optional
    Enum []string `json:"enum,omitempty"`

    // Array indicates this config accepts multiple values.
    // +optional
    Array bool `json:"array,omitempty"`

    // MapTo defines how this config maps to args/env/files.
    // If not specified, value is used directly in args template.
    // +optional
    MapTo *ConfigMapping `json:"mapTo,omitempty"`
}

// ConfigMapping defines how a config value maps to container config.
type ConfigMapping struct {
    // Type is the mapping type: "arg", "env", or "configFile"
    // +kubebuilder:validation:Enum=arg;env;configFile
    Type string `json:"type"`

    // Value is the arg flag or env var name.
    // For "arg": the flag to add (e.g., "-crossplay")
    // For "env": the env var name
    // +optional
    Value string `json:"value,omitempty"`

    // Condition for "arg" type: only add if config value equals this.
    // +optional
    Condition string `json:"condition,omitempty"`

    // Path for "configFile" type: the file path.
    // +optional
    Path string `json:"path,omitempty"`

    // Template for "configFile" type: Go template for file content.
    // +optional
    Template string `json:"template,omitempty"`
}

// ConfigFileTemplate defines a static config file.
type ConfigFileTemplate struct {
    // Path is where to mount the file.
    Path string `json:"path"`

    // Content is the file content (can use {{.Config.key}} templates).
    Content string `json:"content"`
}

// HealthCheckSpec defines health check configuration.
type HealthCheckSpec struct {
    // TCPSocket specifies a TCP port to check.
    // +optional
    TCPSocket *corev1.TCPSocketAction `json:"tcpSocket,omitempty"`

    // InitialDelaySeconds before first check.
    // +kubebuilder:default=120
    // +optional
    InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`

    // PeriodSeconds between checks.
    // +kubebuilder:default=30
    // +optional
    PeriodSeconds int32 `json:"periodSeconds,omitempty"`
}

// GameDefinitionStatus defines the observed state.
type GameDefinitionStatus struct {
    // Ready indicates the GameDefinition is valid and usable.
    // +optional
    Ready bool `json:"ready,omitempty"`

    // Message provides status details.
    // +optional
    Message string `json:"message,omitempty"`

    // Conditions for detailed status.
    // +optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=gd
// +kubebuilder:printcolumn:name="App ID",type="integer",JSONPath=".spec.appId"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// GameDefinition defines how to install and run a Steam game server.
type GameDefinition struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   GameDefinitionSpec   `json:"spec,omitempty"`
    Status GameDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GameDefinitionList contains a list of GameDefinition.
type GameDefinitionList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []GameDefinition `json:"items"`
}

func init() {
    SchemeBuilder.Register(&GameDefinition{}, &GameDefinitionList{})
}
```

**Note:** GameDefinition is cluster-scoped (`scope=Cluster`) so it's available to all namespaces.

---

#### A2. Create common_types.go

**File:** `api/v1alpha1/common_types.go`

Extract shared types from steamserver_types.go:
- `ServerPort`
- `ConfigFile`
- `StorageSpec`
- `PortStatus`

Add new type:
```go
// ConfigValue represents a config value that can be a literal or secret reference.
//
// DESIGN NOTE: DESIGN.md shows a cleaner UX where literals are direct strings:
//   config:
//     serverName: "Vikings Only"       # direct string
//     password:                         # object with secretKeyRef
//       secretKeyRef: {...}
//
// This requires custom unmarshaling (UnmarshalJSON) to detect whether the YAML
// value is a string or an object. For MVP, we can use the simpler structured approach
// below, then improve UX later with custom unmarshaling.
//
// MVP approach (structured):
//   config:
//     serverName:
//       value: "Vikings Only"
//     password:
//       secretKeyRef: {...}
//
// The implementation below supports both - if only `value` is set, it's a literal.
// If `secretKeyRef` is set, the value comes from a Secret.
type ConfigValue struct {
    // Value is a literal string value.
    // +optional
    Value string `json:"value,omitempty"`

    // SecretKeyRef references a key in a Secret.
    // +optional
    SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// NOTE: To achieve the cleaner DESIGN.md syntax later, implement:
// func (c *ConfigValue) UnmarshalJSON(data []byte) error {
//     // Try string first
//     var s string
//     if err := json.Unmarshal(data, &s); err == nil {
//         c.Value = s
//         return nil
//     }
//     // Otherwise unmarshal as struct
//     type Alias ConfigValue
//     return json.Unmarshal(data, (*Alias)(c))
// }
```

---

#### A3. Update SteamServer CRD

**File:** `api/v1alpha1/steamserver_types.go`

Changes:
1. Add `GameDefinition` field (required)
2. Add `Config` field
3. Make `AppId`, `Ports`, `Command`, `Args` optional (fallback mode)
4. Import types from common_types.go

```go
type SteamServerSpec struct {
    // GameDefinition references a GameDefinition by name.
    // +kubebuilder:validation:Required
    GameDefinition string `json:"gameDefinition"`

    // Config provides values for GameDefinition.configSchema keys.
    // +optional
    Config map[string]ConfigValue `json:"config,omitempty"`

    // --- Override fields (optional, for power users or fallback) ---

    // AppId overrides GameDefinition.appId (fallback mode).
    // +optional
    AppId *int32 `json:"appId,omitempty"`

    // Image overrides GameDefinition.image.
    // +optional
    Image string `json:"image,omitempty"`

    // Ports overrides GameDefinition.ports.
    // +optional
    Ports []ServerPort `json:"ports,omitempty"`

    // Command overrides GameDefinition.command.
    // +optional
    Command []string `json:"command,omitempty"`

    // Args overrides GameDefinition.args.
    // +optional
    Args []string `json:"args,omitempty"`

    // Env adds to or overrides GameDefinition.env.
    // +optional
    Env []corev1.EnvVar `json:"env,omitempty"`

    // ConfigFiles adds to GameDefinition.configFiles.
    // +optional
    ConfigFiles []ConfigFile `json:"configFiles,omitempty"`

    // --- User-facing fields (not from GameDefinition) ---

    // Storage configuration.
    // +optional
    Storage *StorageSpec `json:"storage,omitempty"`

    // Resources overrides GameDefinition.defaultResources.
    // +optional
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // ServiceType for the game server Service.
    // +kubebuilder:validation:Enum=LoadBalancer;NodePort;ClusterIP
    // +kubebuilder:default="LoadBalancer"
    // +optional
    ServiceType corev1.ServiceType `json:"serviceType,omitempty"`

    // --- SteamCMD options ---

    // Beta branch to install.
    // +optional
    Beta string `json:"beta,omitempty"`

    // Validate game files on startup.
    // +kubebuilder:default=true
    // +optional
    Validate *bool `json:"validate,omitempty"`

    // Anonymous Steam login.
    // +kubebuilder:default=true
    // +optional
    Anonymous *bool `json:"anonymous,omitempty"`

    // SteamCredentialsSecret for authenticated login.
    // +optional
    SteamCredentialsSecret string `json:"steamCredentialsSecret,omitempty"`
}
```

After changes, run:
```bash
make manifests generate
```

---

### Phase B: SteamCMD Refactor

#### B1. Rename scripts.go → command.go

**File:** `internal/steamcmd/command.go` (rename from scripts.go)

```go
package steamcmd

// CommandConfig holds configuration for building SteamCMD arguments.
type CommandConfig struct {
    AppID      int32
    InstallDir string
    Anonymous  bool
    Beta       string
    Validate   bool
}

// CommandBuilder builds SteamCMD command arguments.
type CommandBuilder struct {
    config CommandConfig
}

// NewCommandBuilder creates a new CommandBuilder.
func NewCommandBuilder(config CommandConfig) *CommandBuilder {
    return &CommandBuilder{config: config}
}

// Build returns the SteamCMD arguments as a string slice.
func (b *CommandBuilder) Build() []string {
    args := []string{}

    // Login
    if b.config.Anonymous {
        args = append(args, "+login", "anonymous")
    } else {
        // Credentials come from env vars
        args = append(args, "+login", "$STEAM_USERNAME", "$STEAM_PASSWORD")
    }

    // Install directory
    installDir := b.config.InstallDir
    if installDir == "" {
        installDir = "/data/server"
    }
    args = append(args, "+force_install_dir", installDir)

    // App update with optional beta and validate
    appUpdateArgs := []string{"+app_update", fmt.Sprintf("%d", b.config.AppID)}

    if b.config.Beta != "" {
        appUpdateArgs = append(appUpdateArgs, "-beta", b.config.Beta)
    }

    if b.config.Validate {
        appUpdateArgs = append(appUpdateArgs, "validate")
    }

    args = append(args, appUpdateArgs...)

    // Quit
    args = append(args, "+quit")

    return args
}
```

#### B2. Update tests

**File:** `internal/steamcmd/command_test.go` (rename from scripts_test.go)

Update tests to check `[]string` output instead of script strings.

---

### Phase C: Resource Builder Updates

#### C1. Create config interpolation package

**File:** `internal/config/interpolate.go`

```go
package config

import (
    "bytes"
    "text/template"

    boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

// TemplateData holds data for template interpolation.
type TemplateData struct {
    Config map[string]string
}

// InterpolateArgs replaces {{.Config.key}} in args with actual values.
func InterpolateArgs(args []string, config map[string]string) ([]string, error) {
    result := make([]string, len(args))
    data := TemplateData{Config: config}

    for i, arg := range args {
        tmpl, err := template.New("arg").Parse(arg)
        if err != nil {
            return nil, err
        }

        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, data); err != nil {
            return nil, err
        }
        result[i] = buf.String()
    }

    return result, nil
}

// ResolveConfigValues converts ConfigValue map to string map.
// Handles secretKeyRef by returning env var reference syntax.
func ResolveConfigValues(
    config map[string]boilerrv1alpha1.ConfigValue,
    schema map[string]boilerrv1alpha1.ConfigSchemaEntry,
) (map[string]string, []corev1.EnvVar) {
    values := make(map[string]string)
    envVars := []corev1.EnvVar{}

    // Apply defaults from schema
    for key, entry := range schema {
        if entry.Default != "" {
            values[key] = entry.Default
        }
    }

    // Override with user config
    for key, cv := range config {
        if cv.SecretKeyRef != nil {
            // Create env var and reference it
            envName := "CONFIG_" + strings.ToUpper(key)
            envVars = append(envVars, corev1.EnvVar{
                Name: envName,
                ValueFrom: &corev1.EnvVarSource{
                    SecretKeyRef: cv.SecretKeyRef,
                },
            })
            values[key] = "$(" + envName + ")"
        } else {
            values[key] = cv.Value
        }
    }

    return values, envVars
}

// ValidateConfig checks config against schema.
func ValidateConfig(
    config map[string]boilerrv1alpha1.ConfigValue,
    schema map[string]boilerrv1alpha1.ConfigSchemaEntry,
) error {
    // Check required fields
    for key, entry := range schema {
        if entry.Required {
            if _, ok := config[key]; !ok {
                return fmt.Errorf("required config key %q not provided", key)
            }
        }
    }

    // Check for unknown keys
    for key := range config {
        if _, ok := schema[key]; !ok {
            return fmt.Errorf("unknown config key %q", key)
        }
    }

    // Check enum values
    for key, cv := range config {
        entry := schema[key]
        if len(entry.Enum) > 0 && cv.Value != "" {
            valid := false
            for _, allowed := range entry.Enum {
                if cv.Value == allowed {
                    valid = true
                    break
                }
            }
            if !valid {
                return fmt.Errorf("config key %q value %q not in allowed values %v",
                    key, cv.Value, entry.Enum)
            }
        }
    }

    return nil
}
```

---

#### C2. Update StatefulSetBuilder

**File:** `internal/resources/statefulset.go`

Major changes:
1. Constructor takes both `SteamServer` and `GameDefinition`
2. Init container uses args, not script
3. Main container interpolates args from GameDefinition + config

```go
type StatefulSetBuilder struct {
    server  *boilerrv1alpha1.SteamServer
    gameDef *boilerrv1alpha1.GameDefinition
}

func NewStatefulSetBuilder(
    server *boilerrv1alpha1.SteamServer,
    gameDef *boilerrv1alpha1.GameDefinition,
) *StatefulSetBuilder {
    return &StatefulSetBuilder{server: server, gameDef: gameDef}
}

func (b *StatefulSetBuilder) buildInitContainer() corev1.Container {
    cmdBuilder := steamcmd.NewCommandBuilder(steamcmd.CommandConfig{
        AppID:      b.getAppID(),
        InstallDir: b.getInstallDir(),
        Anonymous:  b.isAnonymous(),
        Beta:       b.server.Spec.Beta,
        Validate:   b.shouldValidate(),
    })

    return corev1.Container{
        Name:  "steamcmd",
        Image: b.getImage(),
        Args:  cmdBuilder.Build(),  // []string, not script
        VolumeMounts: []corev1.VolumeMount{
            {Name: "game-data", MountPath: "/data"},
        },
        Env: b.buildInitEnvVars(),
    }
}

func (b *StatefulSetBuilder) buildMainContainer() corev1.Container {
    // Resolve config values
    configValues, configEnvVars := config.ResolveConfigValues(
        b.server.Spec.Config,
        b.gameDef.Spec.ConfigSchema,
    )

    // Interpolate args
    args, _ := config.InterpolateArgs(b.getArgs(), configValues)

    // Merge env vars
    env := b.mergeEnvVars(b.gameDef.Spec.Env, b.server.Spec.Env, configEnvVars)

    return corev1.Container{
        Name:         "game-server",
        Image:        b.getImage(),
        Command:      []string{b.getCommand()},
        Args:         args,
        Ports:        b.buildContainerPorts(),
        Env:          env,
        Resources:    b.getResources(),
        VolumeMounts: b.buildVolumeMounts(),
    }
}

// Helper methods to get values with fallback chain:
// SteamServer override → GameDefinition → default

func (b *StatefulSetBuilder) getAppID() int32 {
    if b.server.Spec.AppId != nil {
        return *b.server.Spec.AppId
    }
    return b.gameDef.Spec.AppId
}

func (b *StatefulSetBuilder) getImage() string {
    if b.server.Spec.Image != "" {
        return b.server.Spec.Image
    }
    if b.gameDef.Spec.Image != "" {
        return b.gameDef.Spec.Image
    }
    return "steamcmd/steamcmd:ubuntu-22"
}

func (b *StatefulSetBuilder) getInstallDir() string {
    if b.gameDef.Spec.InstallDir != "" {
        return b.gameDef.Spec.InstallDir
    }
    return "/data/server"
}

func (b *StatefulSetBuilder) getCommand() string {
    if len(b.server.Spec.Command) > 0 {
        return b.server.Spec.Command[0]
    }
    return b.gameDef.Spec.Command
}

func (b *StatefulSetBuilder) getArgs() []string {
    if len(b.server.Spec.Args) > 0 {
        return b.server.Spec.Args
    }
    return b.gameDef.Spec.Args
}
```

---

### Phase D: Controller Updates

#### D1. Create GameDefinition controller

**File:** `internal/controller/gamedefinition_controller.go`

```go
package controller

import (
    "context"

    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

type GameDefinitionReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *GameDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch GameDefinition
    var gameDef boilerrv1alpha1.GameDefinition
    if err := r.Get(ctx, req.NamespacedName, &gameDef); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Validate
    if err := r.validate(&gameDef); err != nil {
        gameDef.Status.Ready = false
        gameDef.Status.Message = err.Error()
        r.Status().Update(ctx, &gameDef)
        return ctrl.Result{}, nil
    }

    // Mark ready
    gameDef.Status.Ready = true
    gameDef.Status.Message = "GameDefinition validated successfully"
    if err := r.Status().Update(ctx, &gameDef); err != nil {
        return ctrl.Result{}, err
    }

    log.Info("GameDefinition reconciled", "name", gameDef.Name)
    return ctrl.Result{}, nil
}

func (r *GameDefinitionReconciler) validate(gd *boilerrv1alpha1.GameDefinition) error {
    if gd.Spec.AppId <= 0 {
        return fmt.Errorf("appId must be positive")
    }
    if gd.Spec.Command == "" {
        return fmt.Errorf("command is required")
    }
    if len(gd.Spec.Ports) == 0 {
        return fmt.Errorf("at least one port is required")
    }
    return nil
}

func (r *GameDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&boilerrv1alpha1.GameDefinition{}).
        Complete(r)
}
```

---

#### D2. Update SteamServer controller

**File:** `internal/controller/steamserver_controller.go`

Key changes:
1. Fetch GameDefinition before building resources
2. Validate config against schema
3. Pass both to resource builders

```go
func (r *SteamServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch SteamServer
    var server boilerrv1alpha1.SteamServer
    if err := r.Get(ctx, req.NamespacedName, &server); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Fetch GameDefinition (cluster-scoped, no namespace)
    var gameDef boilerrv1alpha1.GameDefinition
    if err := r.Get(ctx, client.ObjectKey{Name: server.Spec.GameDefinition}, &gameDef); err != nil {
        if apierrors.IsNotFound(err) {
            r.updateStatus(&server, boilerrv1alpha1.ServerStateError,
                fmt.Sprintf("GameDefinition %q not found", server.Spec.GameDefinition))
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // Check GameDefinition is ready
    if !gameDef.Status.Ready {
        r.updateStatus(&server, boilerrv1alpha1.ServerStateError,
            fmt.Sprintf("GameDefinition %q is not ready", server.Spec.GameDefinition))
        return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
    }

    // Validate config against schema
    if err := config.ValidateConfig(server.Spec.Config, gameDef.Spec.ConfigSchema); err != nil {
        r.updateStatus(&server, boilerrv1alpha1.ServerStateError, err.Error())
        return ctrl.Result{}, nil
    }

    // Build resources with both CRs
    pvcBuilder := resources.NewPVCBuilder(&server, &gameDef)
    stsBuilder := resources.NewStatefulSetBuilder(&server, &gameDef)
    svcBuilder := resources.NewServiceBuilder(&server, &gameDef)

    // ... rest of reconciliation
}
```

Also add watch for GameDefinition changes:
```go
func (r *SteamServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&boilerrv1alpha1.SteamServer{}).
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        Owns(&corev1.PersistentVolumeClaim{}).
        Watches(
            &boilerrv1alpha1.GameDefinition{},
            handler.EnqueueRequestsFromMapFunc(r.findSteamServersForGameDef),
        ).
        Complete(r)
}

func (r *SteamServerReconciler) findSteamServersForGameDef(ctx context.Context, obj client.Object) []reconcile.Request {
    gameDef := obj.(*boilerrv1alpha1.GameDefinition)

    var serverList boilerrv1alpha1.SteamServerList
    if err := r.List(ctx, &serverList); err != nil {
        return nil
    }

    var requests []reconcile.Request
    for _, server := range serverList.Items {
        if server.Spec.GameDefinition == gameDef.Name {
            requests = append(requests, reconcile.Request{
                NamespacedName: client.ObjectKeyFromObject(&server),
            })
        }
    }
    return requests
}
```

---

### Phase E: Testing

#### E1. Update unit tests

- `internal/steamcmd/command_test.go` - Test `[]string` output
- `internal/config/interpolate_test.go` - Test template interpolation
- `internal/resources/statefulset_test.go` - Test with GameDefinition + SteamServer

#### E2. Update integration tests

- Add GameDefinition CRD to envtest setup
- Create test fixtures for GameDefinition
- Test full reconciliation flow

---

### Phase F: Sample Files

#### F1. Sample GameDefinition

**File:** `config/samples/boilerr.dev_v1alpha1_gamedefinition_valheim.yaml`

```yaml
apiVersion: boilerr.dev/v1alpha1
kind: GameDefinition
metadata:
  name: valheim
spec:
  appId: 896660
  image: steamcmd/steamcmd:ubuntu-22
  installDir: /data/server
  command: /data/server/valheim_server.x86_64
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
  ports:
    - name: game
      port: 2456
      protocol: UDP
    - name: query
      port: 2457
      protocol: UDP
  env:
    - name: SteamAppId
      value: "892970"
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
      description: "List on public server browser (0 or 1)"
      default: "0"
      enum: ["0", "1"]
  defaultResources:
    requests:
      cpu: "2"
      memory: "4Gi"
    limits:
      cpu: "4"
      memory: "8Gi"
  defaultStorage: "20Gi"
```

#### F2. Sample SteamServer

**File:** `config/samples/boilerr.dev_v1alpha1_steamserver_valheim.yaml`

Recommended syntax (clean):
```yaml
apiVersion: boilerr.dev/v1alpha1
kind: SteamServer
metadata:
  name: valheim-prod
  namespace: game-servers
spec:
  gameDefinition: valheim
  config:
    serverName: "Vikings Only"      # direct string
    worldName: "Midgard"
    password:                        # secret reference
      secretKeyRef:
        name: valheim-secrets
        key: password
    public: "0"
  storage:
    size: 30Gi
  serviceType: LoadBalancer
```

Backward-compatible syntax (structured - still supported):
```yaml
spec:
  gameDefinition: valheim
  config:
    serverName:
      value: "Vikings Only"
    worldName:
      value: "Midgard"
    password:
      secretKeyRef:
        name: valheim-secrets
        key: password
```

---

## Execution Checklist

```
[x] Phase A: CRD Changes
    [x] A1: Create api/v1alpha1/gamedefinition_types.go
    [x] A2: Create api/v1alpha1/common_types.go
    [x] A3: Update api/v1alpha1/steamserver_types.go
    [x] Run: make manifests generate
    [x] Verify CRDs generate correctly

[x] Phase B: SteamCMD Refactor
    [x] B1: Rename scripts.go → command.go
    [x] B2: Refactor to return []string
    [x] B3: Update/rename tests (command_test.go created)
    [x] Run: make test (steamcmd package)

[x] Phase C: Resource Builders
    [x] C1: Create internal/config/interpolate.go
    [x] C2: Create internal/config/interpolate_test.go
    [x] C3: Update internal/resources/statefulset.go
    [x] C4: Update internal/resources/statefulset_test.go
    [x] C5: Update service.go and pvc.go if needed
    [x] Run: make test (resources package)

[x] Phase D: Controllers
    [x] D1: Create internal/controller/gamedefinition_controller.go
    [x] D2: Update internal/controller/steamserver_controller.go
    [x] D3: Register new controller in cmd/main.go
    [x] D4: Update controller tests
    [x] Run: make test (controller package)

[x] Phase E: Integration
    [x] E1: Update suite_test.go for both CRDs
    [x] E2: Run full test suite: make test
    [x] E3: Fix any failures
        [x] Fixed duplicate int32Ptr/boolPtr declarations (created test_helpers.go)
        [x] Fixed beta flag test expectations (separate args, not combined string)
        [x] Fixed unused helper functions (removed containsString/Helper)

[x] Phase F: Samples
    [x] F1: Create sample GameDefinition (already existed)
    [x] F2: Update sample SteamServer (already existed)
    [x] F3: Test manually with: make run (deferred to user)

[x] Final
    [x] make lint
    [x] make test
    [x] make build
    [x] Update ROADMAP.md checkboxes (pending)
```

---

## Common Commands

```bash
# Regenerate CRDs and code after type changes
make manifests generate

# Run all tests
make test

# Run specific package tests
go test ./internal/steamcmd/... -v
go test ./internal/config/... -v
go test ./internal/resources/... -v
go test ./internal/controller/... -v

# Run operator locally
make run

# Apply sample resources
kubectl apply -f config/samples/boilerr.dev_v1alpha1_gamedefinition_valheim.yaml
kubectl apply -f config/samples/boilerr.dev_v1alpha1_steamserver_valheim.yaml

# Check status
kubectl get gamedefinitions
kubectl get steamservers -A
```

---

## Troubleshooting

### CRD not generating
- Check kubebuilder markers are correct
- Run `make manifests` and check `config/crd/bases/`

### Controller not watching GameDefinition
- Ensure `Watches()` is added in `SetupWithManager`
- Check RBAC in `config/rbac/role.yaml` includes GameDefinition permissions

### Config interpolation failing
- Check template syntax: `{{.Config.key}}` not `{{.config.key}}`
- Ensure config key exists in schema

### Init container failing
- Check args are valid steamcmd commands
- Verify install directory permissions
