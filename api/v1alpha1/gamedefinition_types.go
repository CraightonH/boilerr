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
