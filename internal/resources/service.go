package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

// ServiceBuilder builds a Service for a SteamServer.
type ServiceBuilder struct {
	server  *boilerrv1alpha1.SteamServer
	gameDef *boilerrv1alpha1.GameDefinition
}

// NewServiceBuilder creates a new ServiceBuilder.
// gameDef can be nil for backwards compatibility (fallback mode).
func NewServiceBuilder(server *boilerrv1alpha1.SteamServer, gameDef *boilerrv1alpha1.GameDefinition) *ServiceBuilder {
	return &ServiceBuilder{server: server, gameDef: gameDef}
}

// Build creates the Service for the SteamServer.
func (b *ServiceBuilder) Build() *corev1.Service {
	labels := b.labels()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.server.Name,
			Namespace: b.server.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     b.getServiceType(),
			Selector: labels,
			Ports:    b.buildServicePorts(),
		},
	}
}

// labels returns the common labels for the Service.
func (b *ServiceBuilder) labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "steamserver",
		"app.kubernetes.io/instance":   b.server.Name,
		"app.kubernetes.io/managed-by": "boilerr",
		"boilerr.dev/game":             b.server.Spec.GameDefinition,
	}
}

// getServiceType returns the service type to use.
func (b *ServiceBuilder) getServiceType() corev1.ServiceType {
	if b.server.Spec.ServiceType != "" {
		return b.server.Spec.ServiceType
	}
	return corev1.ServiceTypeLoadBalancer
}

// getPorts returns the ports to expose.
// Fallback: SteamServer.Ports -> GameDefinition.Ports -> empty
func (b *ServiceBuilder) getPorts() []boilerrv1alpha1.ServerPort {
	if len(b.server.Spec.Ports) > 0 {
		return b.server.Spec.Ports
	}
	if b.gameDef != nil {
		return b.gameDef.Spec.Ports
	}
	return nil
}

// buildServicePorts creates the service port definitions.
func (b *ServiceBuilder) buildServicePorts() []corev1.ServicePort {
	ports := b.getPorts()
	result := make([]corev1.ServicePort, len(ports))

	for i, port := range ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = corev1.ProtocolUDP
		}

		servicePort := port.ServicePort
		if servicePort == 0 {
			servicePort = port.ContainerPort
		}

		result[i] = corev1.ServicePort{
			Name:       port.Name,
			Port:       servicePort,
			TargetPort: intstr.FromInt32(port.ContainerPort),
			Protocol:   protocol,
		}
	}

	return result
}

// ServiceName returns the Service name for a SteamServer.
func ServiceName(serverName string) string {
	return serverName
}
