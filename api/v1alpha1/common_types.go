package v1alpha1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

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

// ConfigValue represents a config value that can be a literal or secret reference.
//
// Supports clean syntax via custom unmarshaling:
//
//	config:
//	  serverName: "Vikings Only"       # direct string
//	  password:                         # object with secretKeyRef
//	    secretKeyRef: {...}
//
// The implementation supports both direct strings and structured objects.
// If only `value` is set, it's a literal. If `secretKeyRef` is set, the value
// comes from a Secret.
// +kubebuilder:validation:Type=""
// +kubebuilder:pruning:PreserveUnknownFields
type ConfigValue struct {
	// Value is a literal string value.
	// +optional
	Value string `json:"value,omitempty"`

	// SecretKeyRef references a key in a Secret.
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling to support both direct strings
// and structured objects.
func (cv *ConfigValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a simple string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		cv.Value = str
		cv.SecretKeyRef = nil
		return nil
	}

	// Fall back to unmarshaling as the full struct
	type configValueAlias ConfigValue
	var alias configValueAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	cv.Value = alias.Value
	cv.SecretKeyRef = alias.SecretKeyRef
	return nil
}
