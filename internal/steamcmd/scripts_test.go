package steamcmd

import (
	"strings"
	"testing"
)

func TestScriptBuilder_Build(t *testing.T) {
	tests := []struct {
		name             string
		config           ScriptConfig
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "anonymous login with defaults",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: true,
				Validate:  true,
			},
			shouldContain: []string{
				"set -e",
				"+login anonymous",
				"+app_update 123456",
				"validate",
				"+quit",
				"SteamCMD complete",
			},
			shouldNotContain: []string{
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
				"-beta",
			},
		},
		{
			name: "authenticated login",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: false,
				Validate:  true,
			},
			shouldContain: []string{
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
				"+login",
			},
			shouldNotContain: []string{
				"+login anonymous",
			},
		},
		{
			name: "validation disabled",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: true,
				Validate:  false,
			},
			shouldContain: []string{
				"+app_update 123456",
			},
			shouldNotContain: []string{
				"validate",
			},
		},
		{
			name: "beta branch specified",
			config: ScriptConfig{
				AppID:     896660,
				Anonymous: true,
				Beta:      "public-test",
				Validate:  true,
			},
			shouldContain: []string{
				"-beta public-test",
				"+app_update 896660",
			},
		},
		{
			name: "beta branch with password",
			config: ScriptConfig{
				AppID:        123456,
				Anonymous:    true,
				Beta:         "private-beta",
				BetaPassword: "secret123",
				Validate:     true,
			},
			shouldContain: []string{
				"-beta private-beta",
				"-betapassword \"$STEAM_BETA_PASSWORD\"",
			},
		},
		{
			name: "custom install directory",
			config: ScriptConfig{
				AppID:      123456,
				InstallDir: "/custom/path",
				Anonymous:  true,
				Validate:   true,
			},
			shouldContain: []string{
				"+force_install_dir /custom/path",
			},
			shouldNotContain: []string{
				"+force_install_dir /serverfiles",
			},
		},
		{
			name: "default install directory when empty",
			config: ScriptConfig{
				AppID:      123456,
				InstallDir: "",
				Anonymous:  true,
				Validate:   true,
			},
			shouldContain: []string{
				"+force_install_dir /serverfiles",
			},
		},
		{
			name: "valheim server config",
			config: ScriptConfig{
				AppID:     896660,
				Anonymous: true,
				Validate:  true,
			},
			shouldContain: []string{
				"+app_update 896660",
				"+login anonymous",
				"validate",
			},
		},
		{
			name: "satisfactory server config",
			config: ScriptConfig{
				AppID:     1690800,
				Anonymous: true,
				Beta:      "experimental",
				Validate:  true,
			},
			shouldContain: []string{
				"+app_update 1690800",
				"-beta experimental",
				"validate",
			},
		},
		{
			name: "all options combined",
			config: ScriptConfig{
				AppID:        123456,
				InstallDir:   "/game",
				Anonymous:    false,
				Beta:         "staging",
				BetaPassword: "pw",
				Validate:     true,
			},
			shouldContain: []string{
				"set -e",
				"+force_install_dir /game",
				"$STEAM_USERNAME",
				"$STEAM_PASSWORD",
				"+app_update 123456",
				"-beta staging",
				"-betapassword \"$STEAM_BETA_PASSWORD\"",
				"validate",
				"+quit",
			},
			shouldNotContain: []string{
				"+login anonymous",
			},
		},
		{
			name: "minimal config - no validation no beta",
			config: ScriptConfig{
				AppID:     999999,
				Anonymous: true,
				Validate:  false,
			},
			shouldContain: []string{
				"+app_update 999999",
				"+login anonymous",
			},
			shouldNotContain: []string{
				"validate",
				"-beta",
				"-betapassword",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewScriptBuilder(tt.config)
			script := builder.Build()

			for _, s := range tt.shouldContain {
				if !strings.Contains(script, s) {
					t.Errorf("expected script to contain %q\nscript:\n%s", s, script)
				}
			}

			for _, s := range tt.shouldNotContain {
				if strings.Contains(script, s) {
					t.Errorf("expected script to NOT contain %q\nscript:\n%s", s, script)
				}
			}
		})
	}
}

func TestScriptBuilder_GetLoginCommand(t *testing.T) {
	tests := []struct {
		name     string
		config   ScriptConfig
		expected string
	}{
		{
			name: "anonymous login",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: true,
			},
			expected: "+login anonymous",
		},
		{
			name: "authenticated login",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: false,
			},
			expected: "+login \"$STEAM_USERNAME\" \"$STEAM_PASSWORD\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewScriptBuilder(tt.config)
			got := builder.GetLoginCommand()
			if got != tt.expected {
				t.Errorf("GetLoginCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestScriptBuilder_GetAppUpdateCommand(t *testing.T) {
	tests := []struct {
		name     string
		config   ScriptConfig
		expected string
	}{
		{
			name: "basic app update",
			config: ScriptConfig{
				AppID:    123456,
				Validate: false,
			},
			expected: "+app_update 123456",
		},
		{
			name: "app update with validation",
			config: ScriptConfig{
				AppID:    123456,
				Validate: true,
			},
			expected: "+app_update 123456 validate",
		},
		{
			name: "app update with beta",
			config: ScriptConfig{
				AppID:    123456,
				Beta:     "experimental",
				Validate: false,
			},
			expected: "+app_update 123456 -beta experimental",
		},
		{
			name: "app update with beta and validation",
			config: ScriptConfig{
				AppID:    123456,
				Beta:     "experimental",
				Validate: true,
			},
			expected: "+app_update 123456 -beta experimental validate",
		},
		{
			name: "app update with beta password",
			config: ScriptConfig{
				AppID:        123456,
				Beta:         "private",
				BetaPassword: "secret",
				Validate:     false,
			},
			expected: "+app_update 123456 -beta private -betapassword \"$STEAM_BETA_PASSWORD\"",
		},
		{
			name: "full app update command",
			config: ScriptConfig{
				AppID:        896660,
				Beta:         "test",
				BetaPassword: "pw",
				Validate:     true,
			},
			expected: "+app_update 896660 -beta test -betapassword \"$STEAM_BETA_PASSWORD\" validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewScriptBuilder(tt.config)
			got := builder.GetAppUpdateCommand()
			if got != tt.expected {
				t.Errorf("GetAppUpdateCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestScriptBuilder_RequiresCredentials(t *testing.T) {
	tests := []struct {
		name     string
		config   ScriptConfig
		expected bool
	}{
		{
			name: "anonymous does not require credentials",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: true,
			},
			expected: false,
		},
		{
			name: "non-anonymous requires credentials",
			config: ScriptConfig{
				AppID:     123456,
				Anonymous: false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewScriptBuilder(tt.config)
			got := builder.RequiresCredentials()
			if got != tt.expected {
				t.Errorf("RequiresCredentials() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestScriptBuilder_RequiresBetaPassword(t *testing.T) {
	tests := []struct {
		name     string
		config   ScriptConfig
		expected bool
	}{
		{
			name: "no beta branch",
			config: ScriptConfig{
				AppID: 123456,
			},
			expected: false,
		},
		{
			name: "beta without password",
			config: ScriptConfig{
				AppID: 123456,
				Beta:  "experimental",
			},
			expected: false,
		},
		{
			name: "beta with password",
			config: ScriptConfig{
				AppID:        123456,
				Beta:         "private",
				BetaPassword: "secret",
			},
			expected: true,
		},
		{
			name: "password without beta is false",
			config: ScriptConfig{
				AppID:        123456,
				BetaPassword: "secret",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewScriptBuilder(tt.config)
			got := builder.RequiresBetaPassword()
			if got != tt.expected {
				t.Errorf("RequiresBetaPassword() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInstallDirConstant(t *testing.T) {
	if InstallDir != "/serverfiles" {
		t.Errorf("InstallDir = %q, want /serverfiles", InstallDir)
	}
}

// TestScriptFormat verifies the overall structure of generated scripts.
func TestScriptFormat(t *testing.T) {
	builder := NewScriptBuilder(ScriptConfig{
		AppID:     123456,
		Anonymous: true,
		Validate:  true,
	})
	script := builder.Build()

	// Script should start with strict mode
	if !strings.HasPrefix(script, "set -e\n") {
		t.Error("script should start with 'set -e'")
	}

	// Script should end with success message
	if !strings.HasSuffix(script, "echo \"SteamCMD complete, game files ready.\"\n") {
		t.Error("script should end with success message")
	}

	// Commands should use line continuations for readability
	if !strings.Contains(script, " \\\n") {
		t.Error("script should use line continuations")
	}

	// Should contain +quit
	if !strings.Contains(script, "+quit") {
		t.Error("script should contain +quit command")
	}
}
