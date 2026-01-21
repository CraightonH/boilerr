package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SteamServerSpec defines the desired state of a Steam dedicated game server.
type SteamServerSpec struct {
	// AppId is the Steam application ID for the dedicated server.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	AppId int32 `json:"appId"`

	// Image is the container image to use for SteamCMD.
	// +kubebuilder:default="steamcmd/steamcmd:latest"
	// +optional
	Image string `json:"image,omitempty"`

	// Ports defines the ports to expose for the game server.
	// +kubebuilder:validation:MinItems=1
	Ports []ServerPort `json:"ports"`

	// Command is the entrypoint command for the game server container.
	// +optional
	Command []string `json:"command,omitempty"`

	// Args are the arguments to pass to the game server command.
	// +optional
	Args []string `json:"args,omitempty"`

	// Env defines environment variables for the game server container.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// ConfigFiles defines configuration files to mount into the container.
	// +optional
	ConfigFiles []ConfigFile `json:"configFiles,omitempty"`

	// Storage defines the persistent storage configuration.
	// +kubebuilder:validation:Required
	Storage StorageSpec `json:"storage"`

	// Resources defines the compute resource requirements.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// ServiceType specifies the type of Service to create.
	// +kubebuilder:validation:Enum=LoadBalancer;NodePort;ClusterIP
	// +kubebuilder:default="LoadBalancer"
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`

	// Beta specifies the Steam beta branch to install.
	// +optional
	Beta string `json:"beta,omitempty"`

	// Validate enables SteamCMD validation on startup.
	// +kubebuilder:default=true
	// +optional
	Validate *bool `json:"validate,omitempty"`

	// Anonymous specifies whether to use anonymous Steam login.
	// If false, SteamCredentialsSecret must be provided.
	// +kubebuilder:default=true
	// +optional
	Anonymous *bool `json:"anonymous,omitempty"`

	// SteamCredentialsSecret references a Secret containing Steam login credentials.
	// Required if Anonymous is false. The Secret must contain 'username' and 'password' keys.
	// +optional
	SteamCredentialsSecret string `json:"steamCredentialsSecret,omitempty"`
}

// ServerPort defines a port to expose for the game server.
type ServerPort struct {
	// Name is a unique identifier for this port.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// ContainerPort is the port number on the container.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ContainerPort int32 `json:"containerPort"`

	// ServicePort is the port number exposed on the Service.
	// Defaults to ContainerPort if not specified.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`

	// Protocol is the network protocol for this port.
	// +kubebuilder:validation:Enum=TCP;UDP
	// +kubebuilder:default="UDP"
	// +optional
	Protocol corev1.Protocol `json:"protocol,omitempty"`
}

// ConfigFile defines a configuration file to mount into the container.
type ConfigFile struct {
	// Path is the absolute path where the file should be mounted.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`

	// Content is the content of the configuration file.
	// +kubebuilder:validation:Required
	Content string `json:"content"`
}

// StorageSpec defines the persistent storage configuration.
type StorageSpec struct {
	// Size is the requested storage size.
	// +kubebuilder:validation:Required
	Size resource.Quantity `json:"size"`

	// StorageClassName is the name of the StorageClass to use.
	// If not specified, the default StorageClass will be used.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
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

// PortStatus contains information about an exposed port.
type PortStatus struct {
	// Name is the identifier for this port.
	Name string `json:"name"`

	// Port is the exposed port number.
	Port int32 `json:"port"`

	// Protocol is the network protocol for this port.
	// +optional
	Protocol corev1.Protocol `json:"protocol,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ss
// +kubebuilder:printcolumn:name="App ID",type="integer",JSONPath=".spec.appId",description="Steam App ID"
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
