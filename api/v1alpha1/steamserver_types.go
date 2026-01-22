package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SteamServerSpec defines the desired state of a Steam dedicated game server.
type SteamServerSpec struct {
	// GameDefinition references a GameDefinition by name.
	// +kubebuilder:validation:Required
	GameDefinition string `json:"gameDefinition"`

	// Config provides values for GameDefinition.configSchema keys.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:XPreserveUnknownFields
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

// SteamServerStatus defines the observed state of a Steam dedicated game server.
type SteamServerStatus struct {
	// State is the current state of the game server.
	// +kubebuilder:validation:Enum=Pending;Installing;Starting;Running;Error
	// +optional
	State ServerState `json:"state,omitempty"`

	// Address is the external IP or hostname for the game server.
	// +optional
	Address string `json:"address,omitempty"`

	// Ports contains the exposed port information.
	// +optional
	Ports []PortStatus `json:"ports,omitempty"`

	// LastUpdated is the timestamp of the last successful reconciliation.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// AppBuildId is the current Steam build ID of the installed game.
	// +optional
	AppBuildId string `json:"appBuildId,omitempty"`

	// Message provides a human-readable status message or error.
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations of the server's state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ServerState represents the current state of a game server.
// +kubebuilder:validation:Enum=Pending;Installing;Starting;Running;Error
type ServerState string

const (
	// ServerStatePending indicates the server is waiting to be scheduled.
	ServerStatePending ServerState = "Pending"

	// ServerStateInstalling indicates SteamCMD is downloading/updating game files.
	ServerStateInstalling ServerState = "Installing"

	// ServerStateStarting indicates the game server process is starting.
	ServerStateStarting ServerState = "Starting"

	// ServerStateRunning indicates the game server is running and ready.
	ServerStateRunning ServerState = "Running"

	// ServerStateError indicates an error occurred.
	ServerStateError ServerState = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ss
// +kubebuilder:printcolumn:name="Game",type="string",JSONPath=".spec.gameDefinition",description="Game definition"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Server state"
// +kubebuilder:printcolumn:name="Address",type="string",JSONPath=".status.address",description="External address"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// SteamServer is the Schema for the steamservers API.
type SteamServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SteamServerSpec   `json:"spec,omitempty"`
	Status SteamServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SteamServerList contains a list of SteamServer.
type SteamServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SteamServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SteamServer{}, &SteamServerList{})
}
