package resources

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

func TestServiceBuilder_Build(t *testing.T) {
	tests := []struct {
		name   string
		server *boilerrv1alpha1.SteamServer
		checks func(t *testing.T, svc *corev1.Service)
	}{
		{
			name: "basic service with defaults",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if svc.Name != "test-server" {
					t.Errorf("expected name 'test-server', got %s", svc.Name)
				}
				if svc.Namespace != "default" {
					t.Errorf("expected namespace 'default', got %s", svc.Namespace)
				}
			},
		},
		{
			name: "labels are correct",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valheim",
					Namespace: "games",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 896660,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 2456},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("20Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if svc.Labels["app.kubernetes.io/name"] != "steamserver" {
					t.Errorf("expected name label 'steamserver', got %s", svc.Labels["app.kubernetes.io/name"])
				}
				if svc.Labels["app.kubernetes.io/instance"] != "valheim" {
					t.Errorf("expected instance label 'valheim', got %s", svc.Labels["app.kubernetes.io/instance"])
				}
				if svc.Labels["app.kubernetes.io/managed-by"] != "boilerr" {
					t.Errorf("expected managed-by label 'boilerr', got %s", svc.Labels["app.kubernetes.io/managed-by"])
				}
			},
		},
		{
			name: "default service type is LoadBalancer",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
					t.Errorf("expected service type LoadBalancer, got %s", svc.Spec.Type)
				}
			},
		},
		{
			name: "NodePort service type",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:       123456,
					ServiceType: corev1.ServiceTypeNodePort,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if svc.Spec.Type != corev1.ServiceTypeNodePort {
					t.Errorf("expected service type NodePort, got %s", svc.Spec.Type)
				}
			},
		},
		{
			name: "ClusterIP service type",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:       123456,
					ServiceType: corev1.ServiceTypeClusterIP,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if svc.Spec.Type != corev1.ServiceTypeClusterIP {
					t.Errorf("expected service type ClusterIP, got %s", svc.Spec.Type)
				}
			},
		},
		{
			name: "selector matches labels",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if svc.Spec.Selector["app.kubernetes.io/instance"] != "test-server" {
					t.Errorf("expected selector instance 'test-server', got %s", svc.Spec.Selector["app.kubernetes.io/instance"])
				}
			},
		},
		{
			name: "single port with UDP default",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if len(svc.Spec.Ports) != 1 {
					t.Fatalf("expected 1 port, got %d", len(svc.Spec.Ports))
				}
				port := svc.Spec.Ports[0]
				if port.Name != "game" {
					t.Errorf("expected port name 'game', got %s", port.Name)
				}
				if port.Port != 27015 {
					t.Errorf("expected port 27015, got %d", port.Port)
				}
				if port.TargetPort.IntVal != 27015 {
					t.Errorf("expected target port 27015, got %d", port.TargetPort.IntVal)
				}
				if port.Protocol != corev1.ProtocolUDP {
					t.Errorf("expected protocol UDP, got %s", port.Protocol)
				}
			},
		},
		{
			name: "multiple ports with mixed protocols",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
						{Name: "query", ContainerPort: 27016, Protocol: corev1.ProtocolTCP},
						{Name: "rcon", ContainerPort: 27017, Protocol: corev1.ProtocolTCP},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if len(svc.Spec.Ports) != 3 {
					t.Fatalf("expected 3 ports, got %d", len(svc.Spec.Ports))
				}
				if svc.Spec.Ports[0].Protocol != corev1.ProtocolUDP {
					t.Errorf("expected first port UDP, got %s", svc.Spec.Ports[0].Protocol)
				}
				if svc.Spec.Ports[1].Protocol != corev1.ProtocolTCP {
					t.Errorf("expected second port TCP, got %s", svc.Spec.Ports[1].Protocol)
				}
				if svc.Spec.Ports[2].Protocol != corev1.ProtocolTCP {
					t.Errorf("expected third port TCP, got %s", svc.Spec.Ports[2].Protocol)
				}
			},
		},
		{
			name: "service port different from container port",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015, ServicePort: 7777},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if len(svc.Spec.Ports) != 1 {
					t.Fatalf("expected 1 port, got %d", len(svc.Spec.Ports))
				}
				port := svc.Spec.Ports[0]
				if port.Port != 7777 {
					t.Errorf("expected service port 7777, got %d", port.Port)
				}
				if port.TargetPort.IntVal != 27015 {
					t.Errorf("expected target port 27015, got %d", port.TargetPort.IntVal)
				}
			},
		},
		{
			name: "service port defaults to container port when zero",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015, ServicePort: 0},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				port := svc.Spec.Ports[0]
				if port.Port != 27015 {
					t.Errorf("expected service port to default to 27015, got %d", port.Port)
				}
			},
		},
		{
			name: "valheim server ports",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valheim-server",
					Namespace: "games",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 896660,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 2456},
						{Name: "query", ContainerPort: 2457},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("20Gi"),
					},
				},
			},
			checks: func(t *testing.T, svc *corev1.Service) {
				if len(svc.Spec.Ports) != 2 {
					t.Fatalf("expected 2 ports, got %d", len(svc.Spec.Ports))
				}
				if svc.Spec.Ports[0].Port != 2456 {
					t.Errorf("expected first port 2456, got %d", svc.Spec.Ports[0].Port)
				}
				if svc.Spec.Ports[1].Port != 2457 {
					t.Errorf("expected second port 2457, got %d", svc.Spec.Ports[1].Port)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewServiceBuilder(tt.server)
			svc := builder.Build()
			tt.checks(t, svc)
		})
	}
}

func TestServiceName(t *testing.T) {
	tests := []struct {
		serverName string
		expected   string
	}{
		{"my-server", "my-server"},
		{"valheim", "valheim"},
		{"test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.serverName, func(t *testing.T) {
			got := ServiceName(tt.serverName)
			if got != tt.expected {
				t.Errorf("ServiceName(%s) = %s, want %s", tt.serverName, got, tt.expected)
			}
		})
	}
}
