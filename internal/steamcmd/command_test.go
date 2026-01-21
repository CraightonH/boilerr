package steamcmd

import (
	"strings"
	"testing"
)

func TestCommandBuilder_Build(t *testing.T) {
	tests := []struct {
		name             string
		config           CommandConfig
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "anonymous login with defaults",
			config: CommandConfig{
				AppID:     123456,
				Anonymous: true,
				Validate:  true,
			},
			shouldContain: []string{
				"+login", "anonymous",
				"+app_update", "123456",
				"validate",
				"+quit",
			},
			shouldNotContain: []string{
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
				"-beta",
			},
		},
		{
			name: "authenticated login",
			config: CommandConfig{
				AppID:     123456,
				Anonymous: false,
				Validate:  true,
			},
			shouldContain: []string{
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
				"+login",
			},
			shouldNotContain: []string{},
		},
		{
			name: "validation disabled",
			config: CommandConfig{
				AppID:     123456,
				Anonymous: true,
				Validate:  false,
			},
			shouldContain: []string{
				"+app_update", "123456",
			},
			shouldNotContain: []string{
				"validate",
			},
		},
		{
			name: "beta branch specified",
			config: CommandConfig{
				AppID:     896660,
				Anonymous: true,
				Beta:      "public-test",
				Validate:  true,
			},
			shouldContain: []string{
				"-beta", "public-test",
				"+app_update", "896660",
			},
		},
		{
			name: "beta branch with password",
			config: CommandConfig{
				AppID:        123456,
				Anonymous:    true,
				Beta:         "private-beta",
				BetaPassword: "secret123",
				Validate:     true,
			},
			shouldContain: []string{
				"-beta", "private-beta",
				"-betapassword", "$STEAM_BETA_PASSWORD",
			},
		},
		{
			name: "custom install directory",
			config: CommandConfig{
				AppID:      123456,
				InstallDir: "/custom/path",
				Anonymous:  true,
				Validate:   true,
			},
			shouldContain: []string{
				"+force_install_dir", "/custom/path",
			},
			shouldNotContain: []string{
				"/data/server",
			},
		},
		{
			name: "default install directory when empty",
			config: CommandConfig{
				AppID:      123456,
				InstallDir: "",
				Anonymous:  true,
				Validate:   true,
			},
			shouldContain: []string{
				"+force_install_dir", "/data/server",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewCommandBuilder(tt.config)
			args := builder.Build()
			argsStr := strings.Join(args, " ")

			for _, s := range tt.shouldContain {
				if !strings.Contains(argsStr, s) {
					t.Errorf("expected args to contain %q\nargs: %v", s, args)
				}
			}

			for _, s := range tt.shouldNotContain {
				if strings.Contains(argsStr, s) {
					t.Errorf("expected args to NOT contain %q\nargs: %v", s, args)
				}
			}
		})
	}
}

func TestCommandBuilder_Build_ArgsOrder(t *testing.T) {
	builder := NewCommandBuilder(CommandConfig{
		AppID:      123456,
		InstallDir: "/data/server",
		Anonymous:  true,
		Validate:   true,
	})
	args := builder.Build()

	// Verify the order of arguments is correct
	// +force_install_dir should come before +login
	forceInstallIdx := indexOf(args, "+force_install_dir")
	loginIdx := indexOf(args, "+login")
	appUpdateIdx := indexOf(args, "+app_update")
	quitIdx := indexOf(args, "+quit")

	if forceInstallIdx == -1 || loginIdx == -1 || appUpdateIdx == -1 || quitIdx == -1 {
		t.Fatalf("missing required argument, args: %v", args)
	}

	if forceInstallIdx >= loginIdx || loginIdx >= appUpdateIdx || appUpdateIdx >= quitIdx {
		t.Errorf("arguments in wrong order, expected: force_install_dir < login < app_update < quit, args: %v", args)
	}
}

func TestCommandBuilder_RequiresCredentials(t *testing.T) {
	tests := []struct {
		name     string
		config   CommandConfig
		expected bool
	}{
		{
			name: "anonymous does not require credentials",
			config: CommandConfig{
				AppID:     123456,
				Anonymous: true,
			},
			expected: false,
		},
		{
			name: "non-anonymous requires credentials",
			config: CommandConfig{
				AppID:     123456,
				Anonymous: false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewCommandBuilder(tt.config)
			got := builder.RequiresCredentials()
			if got != tt.expected {
				t.Errorf("RequiresCredentials() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_RequiresBetaPassword(t *testing.T) {
	tests := []struct {
		name     string
		config   CommandConfig
		expected bool
	}{
		{
			name: "no beta branch",
			config: CommandConfig{
				AppID: 123456,
			},
			expected: false,
		},
		{
			name: "beta without password",
			config: CommandConfig{
				AppID: 123456,
				Beta:  "experimental",
			},
			expected: false,
		},
		{
			name: "beta with password",
			config: CommandConfig{
				AppID:        123456,
				Beta:         "private",
				BetaPassword: "secret",
			},
			expected: true,
		},
		{
			name: "password without beta is false",
			config: CommandConfig{
				AppID:        123456,
				BetaPassword: "secret",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewCommandBuilder(tt.config)
			got := builder.RequiresBetaPassword()
			if got != tt.expected {
				t.Errorf("RequiresBetaPassword() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultInstallDirConstant(t *testing.T) {
	if DefaultInstallDir != "/data/server" {
		t.Errorf("DefaultInstallDir = %q, want /data/server", DefaultInstallDir)
	}
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
