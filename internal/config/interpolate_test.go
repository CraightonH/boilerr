package config

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

func TestInterpolateArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		config   map[string]string
		expected []string
		wantErr  bool
	}{
		{
			name:     "no templates",
			args:     []string{"-name", "MyServer", "-port", "27015"},
			config:   map[string]string{},
			expected: []string{"-name", "MyServer", "-port", "27015"},
			wantErr:  false,
		},
		{
			name:     "single template substitution",
			args:     []string{"-name", "{{.Config.serverName}}"},
			config:   map[string]string{"serverName": "MyServer"},
			expected: []string{"-name", "MyServer"},
			wantErr:  false,
		},
		{
			name:     "multiple template substitutions",
			args:     []string{"-name", "{{.Config.serverName}}", "-world", "{{.Config.worldName}}", "-port", "2456"},
			config:   map[string]string{"serverName": "Vikings", "worldName": "Midgard"},
			expected: []string{"-name", "Vikings", "-world", "Midgard", "-port", "2456"},
			wantErr:  false,
		},
		{
			name:     "mixed literal and template",
			args:     []string{"-name", "Server-{{.Config.suffix}}"},
			config:   map[string]string{"suffix": "001"},
			expected: []string{"-name", "Server-001"},
			wantErr:  false,
		},
		{
			name:     "empty args",
			args:     []string{},
			config:   map[string]string{"key": "value"},
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "missing config key returns empty",
			args:     []string{"-name", "{{.Config.missing}}"},
			config:   map[string]string{},
			expected: []string{"-name", "<no value>"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InterpolateArgs(tt.args, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterpolateArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("InterpolateArgs() len = %d, want %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("InterpolateArgs()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestInterpolateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		config   map[string]string
		expected string
		wantErr  bool
	}{
		{
			name:     "no template",
			input:    "static content",
			config:   map[string]string{},
			expected: "static content",
		},
		{
			name:     "single replacement",
			input:    "Server: {{.Config.serverName}}",
			config:   map[string]string{"serverName": "Valheim"},
			expected: "Server: Valheim",
		},
		{
			name:     "multiple replacements",
			input:    "{{.Config.greeting}} {{.Config.name}}!",
			config:   map[string]string{"greeting": "Hello", "name": "Player"},
			expected: "Hello Player!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InterpolateString(tt.input, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterpolateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("InterpolateString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResolveConfigValues(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]boilerrv1alpha1.ConfigValue
		schema         map[string]boilerrv1alpha1.ConfigSchemaEntry
		expectedValues map[string]string
		expectedEnvLen int
	}{
		{
			name: "literal values only",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"serverName": {Value: "MyServer"},
				"maxPlayers": {Value: "16"},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {},
				"maxPlayers": {},
			},
			expectedValues: map[string]string{
				"serverName": "MyServer",
				"maxPlayers": "16",
			},
			expectedEnvLen: 0,
		},
		{
			name: "secret ref creates env var",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"password": {
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "my-secret"},
						Key:                  "password",
					},
				},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"password": {Secret: true},
			},
			expectedValues: map[string]string{
				"password": "$(CONFIG_PASSWORD)",
			},
			expectedEnvLen: 1,
		},
		{
			name:   "schema defaults applied",
			config: map[string]boilerrv1alpha1.ConfigValue{},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {Default: "DefaultServer"},
				"maxPlayers": {Default: "10"},
			},
			expectedValues: map[string]string{
				"serverName": "DefaultServer",
				"maxPlayers": "10",
			},
			expectedEnvLen: 0,
		},
		{
			name: "user config overrides defaults",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"serverName": {Value: "CustomServer"},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {Default: "DefaultServer"},
			},
			expectedValues: map[string]string{
				"serverName": "CustomServer",
			},
			expectedEnvLen: 0,
		},
		{
			name: "hyphen in key name converted to underscore",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"server-name": {
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
						Key:                  "name",
					},
				},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"server-name": {},
			},
			expectedValues: map[string]string{
				"server-name": "$(CONFIG_SERVER_NAME)",
			},
			expectedEnvLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, envVars := ResolveConfigValues(tt.config, tt.schema)

			for k, expected := range tt.expectedValues {
				if values[k] != expected {
					t.Errorf("values[%q] = %q, want %q", k, values[k], expected)
				}
			}

			if len(envVars) != tt.expectedEnvLen {
				t.Errorf("len(envVars) = %d, want %d", len(envVars), tt.expectedEnvLen)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]boilerrv1alpha1.ConfigValue
		schema  map[string]boilerrv1alpha1.ConfigSchemaEntry
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"serverName": {Value: "MyServer"},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {Required: true},
			},
			wantErr: false,
		},
		{
			name:   "missing required field",
			config: map[string]boilerrv1alpha1.ConfigValue{},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {Required: true},
			},
			wantErr: true,
			errMsg:  "required config key",
		},
		{
			name: "required field with empty value",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"serverName": {Value: ""},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {Required: true},
			},
			wantErr: true,
			errMsg:  "empty value",
		},
		{
			name: "required field with secret ref is valid",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"password": {
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
						Key:                  "pass",
					},
				},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"password": {Required: true},
			},
			wantErr: false,
		},
		{
			name: "unknown config key",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"unknownKey": {Value: "value"},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"serverName": {},
			},
			wantErr: true,
			errMsg:  "unknown config key",
		},
		{
			name: "valid enum value",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"difficulty": {Value: "hard"},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"difficulty": {Enum: []string{"easy", "normal", "hard"}},
			},
			wantErr: false,
		},
		{
			name: "invalid enum value",
			config: map[string]boilerrv1alpha1.ConfigValue{
				"difficulty": {Value: "extreme"},
			},
			schema: map[string]boilerrv1alpha1.ConfigSchemaEntry{
				"difficulty": {Enum: []string{"easy", "normal", "hard"}},
			},
			wantErr: true,
			errMsg:  "not in allowed values",
		},
		{
			name:    "empty config and schema is valid",
			config:  map[string]boilerrv1alpha1.ConfigValue{},
			schema:  map[string]boilerrv1alpha1.ConfigSchemaEntry{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config, tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %v, should contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestMergeEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		envSlices [][]corev1.EnvVar
		expected  []corev1.EnvVar
	}{
		{
			name:      "empty slices",
			envSlices: [][]corev1.EnvVar{},
			expected:  nil,
		},
		{
			name: "single slice",
			envSlices: [][]corev1.EnvVar{
				{{Name: "FOO", Value: "bar"}},
			},
			expected: []corev1.EnvVar{{Name: "FOO", Value: "bar"}},
		},
		{
			name: "no overlap merges all",
			envSlices: [][]corev1.EnvVar{
				{{Name: "FOO", Value: "1"}},
				{{Name: "BAR", Value: "2"}},
			},
			expected: []corev1.EnvVar{
				{Name: "FOO", Value: "1"},
				{Name: "BAR", Value: "2"},
			},
		},
		{
			name: "later slice overrides earlier",
			envSlices: [][]corev1.EnvVar{
				{{Name: "FOO", Value: "old"}},
				{{Name: "FOO", Value: "new"}},
			},
			expected: []corev1.EnvVar{{Name: "FOO", Value: "new"}},
		},
		{
			name: "complex merge scenario",
			envSlices: [][]corev1.EnvVar{
				{
					{Name: "A", Value: "1"},
					{Name: "B", Value: "2"},
				},
				{
					{Name: "B", Value: "override"},
					{Name: "C", Value: "3"},
				},
				{
					{Name: "D", Value: "4"},
				},
			},
			expected: []corev1.EnvVar{
				{Name: "A", Value: "1"},
				{Name: "B", Value: "override"},
				{Name: "C", Value: "3"},
				{Name: "D", Value: "4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeEnvVars(tt.envSlices...)
			if len(got) != len(tt.expected) {
				t.Errorf("MergeEnvVars() len = %d, want %d", len(got), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if got[i].Name != expected.Name || got[i].Value != expected.Value {
					t.Errorf("MergeEnvVars()[%d] = {%s, %s}, want {%s, %s}",
						i, got[i].Name, got[i].Value, expected.Name, expected.Value)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
