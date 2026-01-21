// Package config provides configuration interpolation and validation utilities.
package config

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
)

// TemplateData holds data for template interpolation.
type TemplateData struct {
	Config map[string]string
}

// InterpolateArgs replaces {{.Config.key}} in args with actual values.
func InterpolateArgs(args []string, config map[string]string) ([]string, error) {
	result := make([]string, len(args))
	data := TemplateData{Config: config}

	for i, arg := range args {
		tmpl, err := template.New("arg").Parse(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse arg template %q: %w", arg, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("failed to execute arg template %q: %w", arg, err)
		}
		result[i] = buf.String()
	}

	return result, nil
}

// InterpolateString replaces {{.Config.key}} in a string with actual values.
func InterpolateString(s string, config map[string]string) (string, error) {
	data := TemplateData{Config: config}

	tmpl, err := template.New("str").Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ResolveConfigValues converts ConfigValue map to string map and generates env vars for secrets.
// Returns:
//   - values: resolved config values (literals or env var references for secrets)
//   - envVars: env vars that need to be added to the container for secret references
func ResolveConfigValues(
	config map[string]boilerrv1alpha1.ConfigValue,
	schema map[string]boilerrv1alpha1.ConfigSchemaEntry,
) (map[string]string, []corev1.EnvVar) {
	values := make(map[string]string)
	envVars := []corev1.EnvVar{}

	// Apply defaults from schema
	for key, entry := range schema {
		if entry.Default != "" {
			values[key] = entry.Default
		}
	}

	// Override with user config
	for key, cv := range config {
		if cv.SecretKeyRef != nil {
			// Create env var and reference it
			envName := "CONFIG_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
			envVars = append(envVars, corev1.EnvVar{
				Name: envName,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: cv.SecretKeyRef,
				},
			})
			// Use shell variable expansion syntax for the value
			values[key] = "$(" + envName + ")"
		} else {
			values[key] = cv.Value
		}
	}

	return values, envVars
}

// ValidateConfig checks config against schema.
func ValidateConfig(
	config map[string]boilerrv1alpha1.ConfigValue,
	schema map[string]boilerrv1alpha1.ConfigSchemaEntry,
) error {
	// Check required fields
	for key, entry := range schema {
		if entry.Required {
			cv, ok := config[key]
			if !ok {
				return fmt.Errorf("required config key %q not provided", key)
			}
			// Check that the value is not empty
			if cv.Value == "" && cv.SecretKeyRef == nil {
				return fmt.Errorf("required config key %q has empty value", key)
			}
		}
	}

	// Check for unknown keys
	for key := range config {
		if _, ok := schema[key]; !ok {
			return fmt.Errorf("unknown config key %q", key)
		}
	}

	// Check enum values
	for key, cv := range config {
		entry := schema[key]
		if len(entry.Enum) > 0 && cv.Value != "" {
			valid := false
			for _, allowed := range entry.Enum {
				if cv.Value == allowed {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("config key %q value %q not in allowed values %v",
					key, cv.Value, entry.Enum)
			}
		}
	}

	return nil
}

// MergeEnvVars merges multiple env var slices, with later slices overriding earlier ones.
func MergeEnvVars(envSlices ...[]corev1.EnvVar) []corev1.EnvVar {
	seen := make(map[string]int)
	var result []corev1.EnvVar

	for _, envs := range envSlices {
		for _, env := range envs {
			if idx, exists := seen[env.Name]; exists {
				// Override existing
				result[idx] = env
			} else {
				// Add new
				seen[env.Name] = len(result)
				result = append(result, env)
			}
		}
	}

	return result
}
