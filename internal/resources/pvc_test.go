package resources

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

func stringPtr(s string) *string {
	return &s
}

func TestPVCBuilder_Build(t *testing.T) {
	tests := []struct {
		name    string
		server  *boilerrv1alpha1.SteamServer
		gameDef *boilerrv1alpha1.GameDefinition
		checks  func(t *testing.T, pvc *corev1.PersistentVolumeClaim)
	}{
		{
			name: "basic PVC with defaults",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Name != "test-server-data" {
					t.Errorf("expected name 'test-server-data', got %s", pvc.Name)
				}
				if pvc.Namespace != "default" {
					t.Errorf("expected namespace 'default', got %s", pvc.Namespace)
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
					Game:  "valheim",
					AppId: int32Ptr(896660),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 2456},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("20Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Labels["app.kubernetes.io/name"] != "steamserver" {
					t.Errorf("expected name label 'steamserver', got %s", pvc.Labels["app.kubernetes.io/name"])
				}
				if pvc.Labels["app.kubernetes.io/instance"] != "valheim" {
					t.Errorf("expected instance label 'valheim', got %s", pvc.Labels["app.kubernetes.io/instance"])
				}
				if pvc.Labels["app.kubernetes.io/managed-by"] != "boilerr" {
					t.Errorf("expected managed-by label 'boilerr', got %s", pvc.Labels["app.kubernetes.io/managed-by"])
				}
			},
		},
		{
			name: "access mode is ReadWriteOnce",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if len(pvc.Spec.AccessModes) != 1 {
					t.Fatalf("expected 1 access mode, got %d", len(pvc.Spec.AccessModes))
				}
				if pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
					t.Errorf("expected access mode ReadWriteOnce, got %s", pvc.Spec.AccessModes[0])
				}
			},
		},
		{
			name: "storage size is correct",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("50Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				expectedSize := resource.MustParse("50Gi")
				actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				if !actualSize.Equal(expectedSize) {
					t.Errorf("expected storage size 50Gi, got %s", actualSize.String())
				}
			},
		},
		{
			name: "small storage size",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("1Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				expectedSize := resource.MustParse("1Gi")
				actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				if !actualSize.Equal(expectedSize) {
					t.Errorf("expected storage size 1Gi, got %s", actualSize.String())
				}
			},
		},
		{
			name: "large storage size",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("500Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				expectedSize := resource.MustParse("500Gi")
				actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				if !actualSize.Equal(expectedSize) {
					t.Errorf("expected storage size 500Gi, got %s", actualSize.String())
				}
			},
		},
		{
			name: "no storage class when not specified",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Spec.StorageClassName != nil {
					t.Errorf("expected nil storage class, got %s", *pvc.Spec.StorageClassName)
				}
			},
		},
		{
			name: "storage class set when specified",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size:             resource.MustParse("10Gi"),
						StorageClassName: stringPtr("fast-storage"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Spec.StorageClassName == nil {
					t.Fatal("expected storage class to be set")
				}
				if *pvc.Spec.StorageClassName != "fast-storage" {
					t.Errorf("expected storage class 'fast-storage', got %s", *pvc.Spec.StorageClassName)
				}
			},
		},
		{
			name: "different storage class",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size:             resource.MustParse("10Gi"),
						StorageClassName: stringPtr("longhorn"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Spec.StorageClassName == nil {
					t.Fatal("expected storage class to be set")
				}
				if *pvc.Spec.StorageClassName != "longhorn" {
					t.Errorf("expected storage class 'longhorn', got %s", *pvc.Spec.StorageClassName)
				}
			},
		},
		{
			name: "PVC name derived from server name",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-awesome-game",
					Namespace: "production",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Name != "my-awesome-game-data" {
					t.Errorf("expected name 'my-awesome-game-data', got %s", pvc.Name)
				}
				if pvc.Namespace != "production" {
					t.Errorf("expected namespace 'production', got %s", pvc.Namespace)
				}
			},
		},
		{
			name: "valheim server PVC",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valheim-server",
					Namespace: "games",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "valheim",
					AppId: int32Ptr(896660),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 2456},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size:             resource.MustParse("20Gi"),
						StorageClassName: stringPtr("local-path"),
					},
				},
			},
			checks: func(t *testing.T, pvc *corev1.PersistentVolumeClaim) {
				if pvc.Name != "valheim-server-data" {
					t.Errorf("expected name 'valheim-server-data', got %s", pvc.Name)
				}
				expectedSize := resource.MustParse("20Gi")
				actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				if !actualSize.Equal(expectedSize) {
					t.Errorf("expected storage size 20Gi, got %s", actualSize.String())
				}
				if *pvc.Spec.StorageClassName != "local-path" {
					t.Errorf("expected storage class 'local-path', got %s", *pvc.Spec.StorageClassName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPVCBuilder(tt.server, tt.gameDef)
			pvc := builder.Build()
			tt.checks(t, pvc)
		})
	}
}

func TestPVCBuilder_StorageSizeFormats(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		expected string
	}{
		{"gigabytes", "10Gi", "10Gi"},
		{"megabytes", "500Mi", "500Mi"},
		{"terabytes", "1Ti", "1Ti"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					Game:  "test-game",
					AppId: int32Ptr(123456),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse(tt.size),
					},
				},
			}

			builder := NewPVCBuilder(server, nil)
			pvc := builder.Build()

			expectedSize := resource.MustParse(tt.expected)
			actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			if !actualSize.Equal(expectedSize) {
				t.Errorf("expected storage size %s, got %s", tt.expected, actualSize.String())
			}
		})
	}
}
