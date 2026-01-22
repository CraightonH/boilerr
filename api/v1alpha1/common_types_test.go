/*
Copyright 2026 CraightonH.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestConfigValueUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ConfigValue
		wantErr  bool
	}{
		{
			name:  "direct string value",
			input: `"Vikings Only"`,
			expected: ConfigValue{
				Value:        "Vikings Only",
				SecretKeyRef: nil,
			},
			wantErr: false,
		},
		{
			name:  "empty string value",
			input: `""`,
			expected: ConfigValue{
				Value:        "",
				SecretKeyRef: nil,
			},
			wantErr: false,
		},
		{
			name:  "numeric string value",
			input: `"2456"`,
			expected: ConfigValue{
				Value:        "2456",
				SecretKeyRef: nil,
			},
			wantErr: false,
		},
		{
			name:  "structured object with value",
			input: `{"value": "Vikings Valhalla"}`,
			expected: ConfigValue{
				Value:        "Vikings Valhalla",
				SecretKeyRef: nil,
			},
			wantErr: false,
		},
		{
			name:  "structured object with secretKeyRef",
			input: `{"secretKeyRef": {"name": "my-secret", "key": "password"}}`,
			expected: ConfigValue{
				Value: "",
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-secret",
					},
					Key: "password",
				},
			},
			wantErr: false,
		},
		{
			name:  "structured object with both value and secretKeyRef",
			input: `{"value": "fallback", "secretKeyRef": {"name": "my-secret", "key": "password"}}`,
			expected: ConfigValue{
				Value: "fallback",
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-secret",
					},
					Key: "password",
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cv ConfigValue
			err := json.Unmarshal([]byte(tt.input), &cv)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if cv.Value != tt.expected.Value {
				t.Errorf("Value = %q, expected %q", cv.Value, tt.expected.Value)
			}

			if (cv.SecretKeyRef == nil) != (tt.expected.SecretKeyRef == nil) {
				t.Errorf("SecretKeyRef presence mismatch: got %v, expected %v",
					cv.SecretKeyRef != nil, tt.expected.SecretKeyRef != nil)
				return
			}

			if cv.SecretKeyRef != nil && tt.expected.SecretKeyRef != nil {
				if cv.SecretKeyRef.Name != tt.expected.SecretKeyRef.Name {
					t.Errorf("SecretKeyRef.Name = %q, expected %q",
						cv.SecretKeyRef.Name, tt.expected.SecretKeyRef.Name)
				}
				if cv.SecretKeyRef.Key != tt.expected.SecretKeyRef.Key {
					t.Errorf("SecretKeyRef.Key = %q, expected %q",
						cv.SecretKeyRef.Key, tt.expected.SecretKeyRef.Key)
				}
			}
		})
	}
}

func TestConfigValueUnmarshalJSON_InStruct(t *testing.T) {
	// Test unmarshaling ConfigValue as part of a larger struct (like SteamServer.Spec.Config)
	type TestStruct struct {
		Config map[string]ConfigValue `json:"config"`
	}

	tests := []struct {
		name     string
		input    string
		expected map[string]ConfigValue
		wantErr  bool
	}{
		{
			name: "mixed config values",
			input: `{
				"config": {
					"serverName": "Vikings Only",
					"worldName": "Midgard",
					"password": {
						"secretKeyRef": {
							"name": "valheim-secrets",
							"key": "server-password"
						}
					},
					"public": "0"
				}
			}`,
			expected: map[string]ConfigValue{
				"serverName": {Value: "Vikings Only"},
				"worldName":  {Value: "Midgard"},
				"password": {
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "valheim-secrets"},
						Key:                  "server-password",
					},
				},
				"public": {Value: "0"},
			},
			wantErr: false,
		},
		{
			name: "all direct string values",
			input: `{
				"config": {
					"serverName": "My Server",
					"port": "2456",
					"maxPlayers": "10"
				}
			}`,
			expected: map[string]ConfigValue{
				"serverName": {Value: "My Server"},
				"port":       {Value: "2456"},
				"maxPlayers": {Value: "10"},
			},
			wantErr: false,
		},
		{
			name: "backward compatibility with structured syntax",
			input: `{
				"config": {
					"serverName": {
						"value": "My Server"
					},
					"password": {
						"secretKeyRef": {
							"name": "secrets",
							"key": "pass"
						}
					}
				}
			}`,
			expected: map[string]ConfigValue{
				"serverName": {Value: "My Server"},
				"password": {
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "secrets"},
						Key:                  "pass",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct
			err := json.Unmarshal([]byte(tt.input), &ts)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(ts.Config) != len(tt.expected) {
				t.Errorf("Config length = %d, expected %d", len(ts.Config), len(tt.expected))
				return
			}

			for key, expectedCV := range tt.expected {
				actualCV, ok := ts.Config[key]
				if !ok {
					t.Errorf("Missing config key %q", key)
					continue
				}

				if actualCV.Value != expectedCV.Value {
					t.Errorf("Config[%q].Value = %q, expected %q", key, actualCV.Value, expectedCV.Value)
				}

				if (actualCV.SecretKeyRef == nil) != (expectedCV.SecretKeyRef == nil) {
					t.Errorf("Config[%q].SecretKeyRef presence mismatch", key)
					continue
				}

				if actualCV.SecretKeyRef != nil && expectedCV.SecretKeyRef != nil {
					if actualCV.SecretKeyRef.Name != expectedCV.SecretKeyRef.Name {
						t.Errorf("Config[%q].SecretKeyRef.Name = %q, expected %q",
							key, actualCV.SecretKeyRef.Name, expectedCV.SecretKeyRef.Name)
					}
					if actualCV.SecretKeyRef.Key != expectedCV.SecretKeyRef.Key {
						t.Errorf("Config[%q].SecretKeyRef.Key = %q, expected %q",
							key, actualCV.SecretKeyRef.Key, expectedCV.SecretKeyRef.Key)
					}
				}
			}
		})
	}
}
