package resources

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestStatefulSetBuilder_Build(t *testing.T) {
	tests := []struct {
		name   string
		server *boilerrv1alpha1.SteamServer
		checks func(t *testing.T, sts interface{})
	}{
		{
			name: "basic statefulset with defaults",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				if s.Name != "test-server" {
					t.Errorf("expected name 'test-server', got %s", s.Name)
				}
				if s.Namespace != "default" {
					t.Errorf("expected namespace 'default', got %s", s.Namespace)
				}
				if *s.Spec.Replicas != 1 {
					t.Errorf("expected replicas 1, got %d", *s.Spec.Replicas)
				}
				if s.Spec.ServiceName != "test-server" {
					t.Errorf("expected serviceName 'test-server', got %s", s.Spec.ServiceName)
				}
			},
		},
		{
			name: "labels include app-id",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				if s.Labels["boilerr.dev/app-id"] != "896660" {
					t.Errorf("expected app-id label '896660', got %s", s.Labels["boilerr.dev/app-id"])
				}
				if s.Labels["app.kubernetes.io/name"] != "steamserver" {
					t.Errorf("expected name label 'steamserver', got %s", s.Labels["app.kubernetes.io/name"])
				}
				if s.Labels["app.kubernetes.io/instance"] != "valheim" {
					t.Errorf("expected instance label 'valheim', got %s", s.Labels["app.kubernetes.io/instance"])
				}
			},
		},
		{
			name: "custom image override",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Image: "custom/steamcmd:v1.0",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				initContainer := s.Spec.Template.Spec.InitContainers[0]
				if initContainer.Image != "custom/steamcmd:v1.0" {
					t.Errorf("expected custom image, got %s", initContainer.Image)
				}
				mainContainer := s.Spec.Template.Spec.Containers[0]
				if mainContainer.Image != "custom/steamcmd:v1.0" {
					t.Errorf("expected custom image on main container, got %s", mainContainer.Image)
				}
			},
		},
		{
			name: "default image when not specified",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				initContainer := s.Spec.Template.Spec.InitContainers[0]
				if initContainer.Image != "steamcmd/steamcmd:latest" {
					t.Errorf("expected default image 'steamcmd/steamcmd:latest', got %s", initContainer.Image)
				}
			},
		},
		{
			name: "container ports with protocol defaults to UDP",
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
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				mainContainer := s.Spec.Template.Spec.Containers[0]
				if len(mainContainer.Ports) != 2 {
					t.Fatalf("expected 2 ports, got %d", len(mainContainer.Ports))
				}
				if mainContainer.Ports[0].Protocol != corev1.ProtocolUDP {
					t.Errorf("expected UDP protocol for first port, got %s", mainContainer.Ports[0].Protocol)
				}
				if mainContainer.Ports[1].Protocol != corev1.ProtocolTCP {
					t.Errorf("expected TCP protocol for second port, got %s", mainContainer.Ports[1].Protocol)
				}
			},
		},
		{
			name: "command and args passed to main container",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:   123456,
					Command: []string{"/bin/bash", "-c"},
					Args:    []string{"./start_server.sh", "-name", "MyServer"},
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				mainContainer := s.Spec.Template.Spec.Containers[0]
				if len(mainContainer.Command) != 2 {
					t.Errorf("expected 2 command elements, got %d", len(mainContainer.Command))
				}
				if len(mainContainer.Args) != 3 {
					t.Errorf("expected 3 args, got %d", len(mainContainer.Args))
				}
			},
		},
		{
			name: "environment variables passed through",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Env: []corev1.EnvVar{
						{Name: "SERVER_NAME", Value: "MyServer"},
						{Name: "MAX_PLAYERS", Value: "16"},
					},
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				mainContainer := s.Spec.Template.Spec.Containers[0]
				if len(mainContainer.Env) != 2 {
					t.Errorf("expected 2 env vars, got %d", len(mainContainer.Env))
				}
				if mainContainer.Env[0].Name != "SERVER_NAME" {
					t.Errorf("expected first env var 'SERVER_NAME', got %s", mainContainer.Env[0].Name)
				}
			},
		},
		{
			name: "resource requirements applied",
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
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
				},
			},
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				mainContainer := s.Spec.Template.Spec.Containers[0]
				cpuReq := mainContainer.Resources.Requests[corev1.ResourceCPU]
				if cpuReq.String() != "500m" {
					t.Errorf("expected CPU request '500m', got %s", cpuReq.String())
				}
				memLimit := mainContainer.Resources.Limits[corev1.ResourceMemory]
				if memLimit.String() != "4Gi" {
					t.Errorf("expected memory limit '4Gi', got %s", memLimit.String())
				}
			},
		},
		{
			name: "config files add volume and mounts",
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
					ConfigFiles: []boilerrv1alpha1.ConfigFile{
						{Path: "/config/server.cfg", Content: "hostname MyServer"},
						{Path: "/config/admins.txt", Content: "STEAM_0:1:12345"},
					},
				},
			},
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				volumes := s.Spec.Template.Spec.Volumes
				if len(volumes) != 2 {
					t.Fatalf("expected 2 volumes (serverfiles + config-files), got %d", len(volumes))
				}
				foundConfigVolume := false
				for _, v := range volumes {
					if v.Name == "config-files" {
						foundConfigVolume = true
						if v.ConfigMap == nil {
							t.Error("expected config-files volume to use ConfigMap")
						}
					}
				}
				if !foundConfigVolume {
					t.Error("config-files volume not found")
				}

				mainContainer := s.Spec.Template.Spec.Containers[0]
				// 1 for serverfiles + 2 for config files
				if len(mainContainer.VolumeMounts) != 3 {
					t.Errorf("expected 3 volume mounts, got %d", len(mainContainer.VolumeMounts))
				}
			},
		},
		{
			name: "no config files means no config volume",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				volumes := s.Spec.Template.Spec.Volumes
				if len(volumes) != 1 {
					t.Fatalf("expected 1 volume (serverfiles only), got %d", len(volumes))
				}
				if volumes[0].Name != ServerFilesVolumeName {
					t.Errorf("expected volume name '%s', got %s", ServerFilesVolumeName, volumes[0].Name)
				}
			},
		},
		{
			name: "PVC name derived from server name",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-game-server",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				volumes := s.Spec.Template.Spec.Volumes
				pvcName := volumes[0].PersistentVolumeClaim.ClaimName
				if pvcName != "my-game-server-data" {
					t.Errorf("expected PVC claim name 'my-game-server-data', got %s", pvcName)
				}
			},
		},
		{
			name: "init container has correct name and mount",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				initContainer := s.Spec.Template.Spec.InitContainers[0]
				if initContainer.Name != InitContainerName {
					t.Errorf("expected init container name '%s', got %s", InitContainerName, initContainer.Name)
				}
				if len(initContainer.VolumeMounts) != 1 {
					t.Fatalf("expected 1 volume mount on init container, got %d", len(initContainer.VolumeMounts))
				}
				if initContainer.VolumeMounts[0].MountPath != ServerFilesMountPath {
					t.Errorf("expected mount path '%s', got %s", ServerFilesMountPath, initContainer.VolumeMounts[0].MountPath)
				}
			},
		},
		{
			name: "main container has correct name",
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
			checks: func(t *testing.T, sts interface{}) {
				s := sts.(*StatefulSetBuilder).Build()
				mainContainer := s.Spec.Template.Spec.Containers[0]
				if mainContainer.Name != GameServerContainerName {
					t.Errorf("expected main container name '%s', got %s", GameServerContainerName, mainContainer.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewStatefulSetBuilder(tt.server)
			tt.checks(t, builder)
		})
	}
}

func TestStatefulSetBuilder_SteamCMDScript(t *testing.T) {
	tests := []struct {
		name             string
		server           *boilerrv1alpha1.SteamServer
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "anonymous login by default",
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
			shouldContain: []string{
				"+login anonymous",
				"+app_update 123456",
				"validate",
			},
			shouldNotContain: []string{
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
			},
		},
		{
			name: "anonymous explicitly true",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:     123456,
					Anonymous: boolPtr(true),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			shouldContain: []string{
				"+login anonymous",
			},
		},
		{
			name: "non-anonymous login uses credentials",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:                  123456,
					Anonymous:              boolPtr(false),
					SteamCredentialsSecret: "steam-creds",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			shouldContain: []string{
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
			},
			shouldNotContain: []string{
				"+login anonymous",
			},
		},
		{
			name: "validate enabled by default",
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
			shouldContain: []string{
				"validate",
			},
		},
		{
			name: "validate disabled",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:    123456,
					Validate: boolPtr(false),
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			shouldNotContain: []string{
				"validate",
			},
		},
		{
			name: "beta branch included",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId: 123456,
					Beta:  "experimental",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			shouldContain: []string{
				"-beta experimental",
			},
		},
		{
			name: "no beta when not specified",
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
			shouldNotContain: []string{
				"-beta",
			},
		},
		{
			name: "force_install_dir set correctly",
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
			shouldContain: []string{
				"+force_install_dir " + ServerFilesMountPath,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewStatefulSetBuilder(tt.server)
			sts := builder.Build()
			script := sts.Spec.Template.Spec.InitContainers[0].Command[2]

			for _, s := range tt.shouldContain {
				if !containsString(script, s) {
					t.Errorf("expected script to contain %q, got:\n%s", s, script)
				}
			}
			for _, s := range tt.shouldNotContain {
				if containsString(script, s) {
					t.Errorf("expected script to NOT contain %q, got:\n%s", s, script)
				}
			}
		})
	}
}

func TestStatefulSetBuilder_InitEnvVars(t *testing.T) {
	tests := []struct {
		name               string
		server             *boilerrv1alpha1.SteamServer
		expectedEnvCount   int
		checkCredentialEnv bool
	}{
		{
			name: "anonymous login has only APP_ID env",
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
			expectedEnvCount:   1,
			checkCredentialEnv: false,
		},
		{
			name: "non-anonymous login has credential env vars",
			server: &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					AppId:                  123456,
					Anonymous:              boolPtr(false),
					SteamCredentialsSecret: "steam-creds",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			},
			expectedEnvCount:   3, // APP_ID + STEAM_USERNAME + STEAM_PASSWORD
			checkCredentialEnv: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewStatefulSetBuilder(tt.server)
			sts := builder.Build()
			initContainer := sts.Spec.Template.Spec.InitContainers[0]

			if len(initContainer.Env) != tt.expectedEnvCount {
				t.Errorf("expected %d env vars, got %d", tt.expectedEnvCount, len(initContainer.Env))
			}

			if tt.checkCredentialEnv {
				foundUsername := false
				foundPassword := false
				for _, env := range initContainer.Env {
					if env.Name == "STEAM_USERNAME" {
						foundUsername = true
						if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
							t.Error("STEAM_USERNAME should reference a secret")
						} else if env.ValueFrom.SecretKeyRef.Name != "steam-creds" {
							t.Errorf("expected secret name 'steam-creds', got %s", env.ValueFrom.SecretKeyRef.Name)
						}
					}
					if env.Name == "STEAM_PASSWORD" {
						foundPassword = true
						if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
							t.Error("STEAM_PASSWORD should reference a secret")
						}
					}
				}
				if !foundUsername {
					t.Error("STEAM_USERNAME env var not found")
				}
				if !foundPassword {
					t.Error("STEAM_PASSWORD env var not found")
				}
			}
		})
	}
}

func TestPVCName(t *testing.T) {
	tests := []struct {
		serverName string
		expected   string
	}{
		{"my-server", "my-server-data"},
		{"valheim", "valheim-data"},
		{"test", "test-data"},
	}

	for _, tt := range tests {
		t.Run(tt.serverName, func(t *testing.T) {
			got := PVCName(tt.serverName)
			if got != tt.expected {
				t.Errorf("PVCName(%s) = %s, want %s", tt.serverName, got, tt.expected)
			}
		})
	}
}

func TestConfigMapName(t *testing.T) {
	tests := []struct {
		serverName string
		expected   string
	}{
		{"my-server", "my-server-config"},
		{"valheim", "valheim-config"},
		{"test", "test-config"},
	}

	for _, tt := range tests {
		t.Run(tt.serverName, func(t *testing.T) {
			got := ConfigMapName(tt.serverName)
			if got != tt.expected {
				t.Errorf("ConfigMapName(%s) = %s, want %s", tt.serverName, got, tt.expected)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
