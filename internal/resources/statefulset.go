package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
	"github.com/CraightonH/boilerr/internal/config"
	"github.com/CraightonH/boilerr/internal/steamcmd"
)

const (
	// ServerFilesVolumeName is the volume name for game server files.
	ServerFilesVolumeName = "serverfiles"
	// ServerFilesMountPath is the mount path for game server files.
	ServerFilesMountPath = "/serverfiles"
	// InitContainerName is the name of the SteamCMD init container.
	InitContainerName = "steamcmd"
	// GameServerContainerName is the name of the main game server container.
	GameServerContainerName = "gameserver"
	// DefaultImage is the default container image.
	DefaultImage = "steamcmd/steamcmd:ubuntu-22"
)

// StatefulSetBuilder builds a StatefulSet for a SteamServer.
type StatefulSetBuilder struct {
	server  *boilerrv1alpha1.SteamServer
	gameDef *boilerrv1alpha1.GameDefinition
}

// NewStatefulSetBuilder creates a new StatefulSetBuilder.
// gameDef can be nil for backwards compatibility (fallback mode).
func NewStatefulSetBuilder(server *boilerrv1alpha1.SteamServer, gameDef *boilerrv1alpha1.GameDefinition) *StatefulSetBuilder {
	return &StatefulSetBuilder{server: server, gameDef: gameDef}
}

// Build creates the StatefulSet for the SteamServer.
func (b *StatefulSetBuilder) Build() *appsv1.StatefulSet {
	labels := b.labels()
	replicas := int32(1)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.server.Name,
			Namespace: b.server.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: b.server.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						b.buildInitContainer(),
					},
					Containers: []corev1.Container{
						b.buildMainContainer(),
					},
					Volumes: b.buildVolumes(),
				},
			},
		},
	}
}

// labels returns the common labels for the StatefulSet.
func (b *StatefulSetBuilder) labels() map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       "steamserver",
		"app.kubernetes.io/instance":   b.server.Name,
		"app.kubernetes.io/managed-by": "boilerr",
		"boilerr.dev/game":             b.server.Spec.Game,
	}
	appID := b.getAppID()
	if appID > 0 {
		labels["boilerr.dev/app-id"] = fmt.Sprintf("%d", appID)
	}
	return labels
}

// buildInitContainer creates the SteamCMD init container.
func (b *StatefulSetBuilder) buildInitContainer() corev1.Container {
	return corev1.Container{
		Name:    InitContainerName,
		Image:   b.getImage(),
		Command: []string{"steamcmd"},
		Args:    b.buildSteamCMDArgs(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      ServerFilesVolumeName,
				MountPath: ServerFilesMountPath,
			},
		},
		Env: b.buildInitEnvVars(),
	}
}

// buildMainContainer creates the main game server container.
func (b *StatefulSetBuilder) buildMainContainer() corev1.Container {
	// Resolve config values and get env vars for secrets
	configValues, configEnvVars := b.resolveConfigValues()

	// Interpolate args with config values
	args := b.getInterpolatedArgs(configValues)

	// Build env vars: GameDefinition defaults + SteamServer overrides + config secret refs
	env := b.buildMainEnvVars(configEnvVars)

	return corev1.Container{
		Name:         GameServerContainerName,
		Image:        b.getImage(),
		Command:      b.getCommand(),
		Args:         args,
		Ports:        b.buildContainerPorts(),
		Env:          env,
		Resources:    b.getResources(),
		VolumeMounts: b.buildVolumeMounts(),
	}
}

// resolveConfigValues resolves config values from SteamServer.Config against GameDefinition.ConfigSchema.
func (b *StatefulSetBuilder) resolveConfigValues() (map[string]string, []corev1.EnvVar) {
	schema := make(map[string]boilerrv1alpha1.ConfigSchemaEntry)
	if b.gameDef != nil {
		schema = b.gameDef.Spec.ConfigSchema
	}
	return config.ResolveConfigValues(b.server.Spec.Config, schema)
}

// getInterpolatedArgs returns args with config values interpolated.
func (b *StatefulSetBuilder) getInterpolatedArgs(configValues map[string]string) []string {
	args := b.getArgs()
	if len(args) == 0 {
		return nil
	}

	interpolated, err := config.InterpolateArgs(args, configValues)
	if err != nil {
		// Fall back to raw args on error
		return args
	}
	return interpolated
}

// getResources returns the resource requirements.
// Fallback: SteamServer.Resources -> GameDefinition.DefaultResources -> empty
func (b *StatefulSetBuilder) getResources() corev1.ResourceRequirements {
	if b.server.Spec.Resources != nil {
		return *b.server.Spec.Resources
	}
	if b.gameDef != nil {
		return b.gameDef.Spec.DefaultResources
	}
	return corev1.ResourceRequirements{}
}

// getImage returns the container image to use.
// Fallback: SteamServer.Image -> GameDefinition.Image -> DefaultImage
func (b *StatefulSetBuilder) getImage() string {
	if b.server.Spec.Image != "" {
		return b.server.Spec.Image
	}
	if b.gameDef != nil && b.gameDef.Spec.Image != "" {
		return b.gameDef.Spec.Image
	}
	return DefaultImage
}

// getInstallDir returns the install directory for SteamCMD.
// Fallback: GameDefinition.InstallDir -> ServerFilesMountPath
func (b *StatefulSetBuilder) getInstallDir() string {
	if b.gameDef != nil && b.gameDef.Spec.InstallDir != "" {
		return b.gameDef.Spec.InstallDir
	}
	return ServerFilesMountPath
}

// getCommand returns the command for the main container.
// Fallback: SteamServer.Command -> GameDefinition.Command -> nil
func (b *StatefulSetBuilder) getCommand() []string {
	if len(b.server.Spec.Command) > 0 {
		return b.server.Spec.Command
	}
	if b.gameDef != nil && b.gameDef.Spec.Command != "" {
		return []string{b.gameDef.Spec.Command}
	}
	return nil
}

// getArgs returns the args for the main container.
// Fallback: SteamServer.Args -> GameDefinition.Args -> nil
func (b *StatefulSetBuilder) getArgs() []string {
	if len(b.server.Spec.Args) > 0 {
		return b.server.Spec.Args
	}
	if b.gameDef != nil {
		return b.gameDef.Spec.Args
	}
	return nil
}

// buildSteamCMDArgs generates the SteamCMD arguments using the steamcmd package.
func (b *StatefulSetBuilder) buildSteamCMDArgs() []string {
	cmdConfig := steamcmd.CommandConfig{
		AppID:      b.getAppID(),
		InstallDir: b.getInstallDir(),
		Anonymous:  b.isAnonymous(),
		Beta:       b.server.Spec.Beta,
		Validate:   b.shouldValidate(),
	}

	builder := steamcmd.NewCommandBuilder(cmdConfig)
	return builder.Build()
}

// getAppID returns the App ID.
// Fallback: SteamServer.AppId -> GameDefinition.AppId -> 0
func (b *StatefulSetBuilder) getAppID() int32 {
	if b.server.Spec.AppId != nil {
		return *b.server.Spec.AppId
	}
	if b.gameDef != nil {
		return b.gameDef.Spec.AppId
	}
	return 0
}

// isAnonymous returns whether to use anonymous Steam login.
func (b *StatefulSetBuilder) isAnonymous() bool {
	if b.server.Spec.Anonymous == nil {
		return true
	}
	return *b.server.Spec.Anonymous
}

// shouldValidate returns whether to validate game files.
func (b *StatefulSetBuilder) shouldValidate() bool {
	if b.server.Spec.Validate == nil {
		return true
	}
	return *b.server.Spec.Validate
}

// buildInitEnvVars creates environment variables for the init container.
func (b *StatefulSetBuilder) buildInitEnvVars() []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "APP_ID",
			Value: fmt.Sprintf("%d", b.getAppID()),
		},
	}

	// Add Steam credentials from secret if not anonymous
	if !b.isAnonymous() && b.server.Spec.SteamCredentialsSecret != "" {
		envVars = append(envVars,
			corev1.EnvVar{
				Name: "STEAM_USERNAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: b.server.Spec.SteamCredentialsSecret,
						},
						Key: "username",
					},
				},
			},
			corev1.EnvVar{
				Name: "STEAM_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: b.server.Spec.SteamCredentialsSecret,
						},
						Key: "password",
					},
				},
			},
		)
	}

	return envVars
}

// buildMainEnvVars creates environment variables for the main container.
// Merges: GameDefinition.Env + SteamServer.Env + config secret refs
func (b *StatefulSetBuilder) buildMainEnvVars(configEnvVars []corev1.EnvVar) []corev1.EnvVar {
	var gameDefEnv []corev1.EnvVar
	if b.gameDef != nil {
		gameDefEnv = b.gameDef.Spec.Env
	}
	return config.MergeEnvVars(gameDefEnv, b.server.Spec.Env, configEnvVars)
}

// getPorts returns the ports to expose.
// Fallback: SteamServer.Ports -> GameDefinition.Ports -> empty
func (b *StatefulSetBuilder) getPorts() []boilerrv1alpha1.ServerPort {
	if len(b.server.Spec.Ports) > 0 {
		return b.server.Spec.Ports
	}
	if b.gameDef != nil {
		return b.gameDef.Spec.Ports
	}
	return nil
}

// buildContainerPorts creates the container port definitions.
func (b *StatefulSetBuilder) buildContainerPorts() []corev1.ContainerPort {
	ports := b.getPorts()
	result := make([]corev1.ContainerPort, len(ports))

	for i, port := range ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = corev1.ProtocolUDP
		}

		result[i] = corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.ContainerPort,
			Protocol:      protocol,
		}
	}

	return result
}

// buildVolumes creates the volume definitions for the pod.
func (b *StatefulSetBuilder) buildVolumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: ServerFilesVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: PVCName(b.server.Name),
				},
			},
		},
	}

	// Add config file volumes if specified
	if len(b.server.Spec.ConfigFiles) > 0 {
		volumes = append(volumes, corev1.Volume{
			Name: "config-files",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ConfigMapName(b.server.Name),
					},
				},
			},
		})
	}

	return volumes
}

// buildVolumeMounts creates the volume mounts for the main container.
func (b *StatefulSetBuilder) buildVolumeMounts() []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      ServerFilesVolumeName,
			MountPath: ServerFilesMountPath,
		},
	}

	// Add individual config file mounts
	for i, cf := range b.server.Spec.ConfigFiles {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "config-files",
			MountPath: cf.Path,
			SubPath:   fmt.Sprintf("config-%d", i),
			ReadOnly:  true,
		})
	}

	return mounts
}

// PVCName returns the PVC name for a SteamServer.
func PVCName(serverName string) string {
	return serverName + "-data"
}

// ConfigMapName returns the ConfigMap name for a SteamServer's config files.
func ConfigMapName(serverName string) string {
	return serverName + "-config"
}
