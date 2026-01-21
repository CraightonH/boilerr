package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
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
)

// StatefulSetBuilder builds a StatefulSet for a SteamServer.
type StatefulSetBuilder struct {
	server *boilerrv1alpha1.SteamServer
}

// NewStatefulSetBuilder creates a new StatefulSetBuilder.
func NewStatefulSetBuilder(server *boilerrv1alpha1.SteamServer) *StatefulSetBuilder {
	return &StatefulSetBuilder{server: server}
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
	return map[string]string{
		"app.kubernetes.io/name":       "steamserver",
		"app.kubernetes.io/instance":   b.server.Name,
		"app.kubernetes.io/managed-by": "boilerr",
		"boilerr.dev/app-id":           fmt.Sprintf("%d", b.server.Spec.AppId),
	}
}

// buildInitContainer creates the SteamCMD init container.
func (b *StatefulSetBuilder) buildInitContainer() corev1.Container {
	container := corev1.Container{
		Name:  InitContainerName,
		Image: b.getImage(),
		Command: []string{
			"/bin/bash",
			"-c",
			b.buildSteamCMDScript(),
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      ServerFilesVolumeName,
				MountPath: ServerFilesMountPath,
			},
		},
		Env: b.buildInitEnvVars(),
	}

	return container
}

// buildMainContainer creates the main game server container.
func (b *StatefulSetBuilder) buildMainContainer() corev1.Container {
	container := corev1.Container{
		Name:         GameServerContainerName,
		Image:        b.getImage(),
		Command:      b.server.Spec.Command,
		Args:         b.server.Spec.Args,
		Ports:        b.buildContainerPorts(),
		Env:          b.buildEnvVars(),
		Resources:    b.server.Spec.Resources,
		VolumeMounts: b.buildVolumeMounts(),
	}

	return container
}

// getImage returns the container image to use.
func (b *StatefulSetBuilder) getImage() string {
	if b.server.Spec.Image != "" {
		return b.server.Spec.Image
	}
	return "steamcmd/steamcmd:latest"
}

// buildSteamCMDScript generates the SteamCMD installation script using the steamcmd package.
func (b *StatefulSetBuilder) buildSteamCMDScript() string {
	config := steamcmd.ScriptConfig{
		AppID:      b.server.Spec.AppId,
		InstallDir: ServerFilesMountPath,
		Anonymous:  b.isAnonymous(),
		Beta:       b.server.Spec.Beta,
		Validate:   b.shouldValidate(),
	}

	builder := steamcmd.NewScriptBuilder(config)
	return builder.Build()
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
			Value: fmt.Sprintf("%d", b.server.Spec.AppId),
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

// buildEnvVars creates environment variables for the main container.
func (b *StatefulSetBuilder) buildEnvVars() []corev1.EnvVar {
	return b.server.Spec.Env
}

// buildContainerPorts creates the container port definitions.
func (b *StatefulSetBuilder) buildContainerPorts() []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, len(b.server.Spec.Ports))

	for i, port := range b.server.Spec.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = corev1.ProtocolUDP
		}

		ports[i] = corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.ContainerPort,
			Protocol:      protocol,
		}
	}

	return ports
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
