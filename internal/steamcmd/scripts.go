// Package steamcmd provides SteamCMD script generation for game server installation.
package steamcmd

import (
	"fmt"
	"strings"
)

// InstallDir is the default installation directory for game server files.
const InstallDir = "/serverfiles"

// ScriptConfig holds configuration for generating SteamCMD installation scripts.
type ScriptConfig struct {
	// AppID is the Steam application ID for the dedicated server.
	AppID int32

	// InstallDir is the directory where game files will be installed.
	// Defaults to InstallDir constant if empty.
	InstallDir string

	// Anonymous indicates whether to use anonymous Steam login.
	// Most dedicated servers support anonymous login.
	Anonymous bool

	// Beta specifies an optional beta branch to install (e.g., "experimental", "public-test").
	// Leave empty for the default/stable branch.
	Beta string

	// BetaPassword is an optional password for accessing private beta branches.
	// Some games require a password to access certain beta branches.
	BetaPassword string

	// Validate enables file validation after download.
	// This verifies all game files and re-downloads corrupted ones.
	// Recommended for production use but adds time to startup.
	Validate bool
}

// ScriptBuilder generates SteamCMD installation scripts.
type ScriptBuilder struct {
	config ScriptConfig
}

// NewScriptBuilder creates a new ScriptBuilder with the given configuration.
func NewScriptBuilder(config ScriptConfig) *ScriptBuilder {
	return &ScriptBuilder{config: config}
}

// Build generates the SteamCMD installation script.
//
// The generated script:
//   - Sets bash strict mode (set -e) to fail fast on errors
//   - Configures the installation directory
//   - Handles anonymous or authenticated login
//   - Downloads/updates the specified app with optional beta branch
//   - Optionally validates game files
//
// Steam Guard Limitations:
// Authenticated (non-anonymous) logins may require Steam Guard verification.
// Steam Guard is a two-factor authentication system that sends a code via email
// or mobile app. This presents challenges for automated/containerized deployments:
//
//   - Email codes: Not practical for automated deployments as they require
//     manual intervention to retrieve and enter the code.
//
//   - Mobile authenticator: Codes can be generated offline but require integration
//     with a Steam authenticator library or shared secret.
//
//   - Trusted device: Once a device is trusted, Steam Guard won't prompt again
//     for that device. However, containers are ephemeral and won't retain trust.
//
// Workarounds for authenticated access:
//  1. Use a dedicated Steam account with Steam Guard disabled (if possible)
//  2. Pre-authenticate on a persistent volume and mount ~/.steam credentials
//  3. Use anonymous login when the game supports it (most dedicated servers do)
//  4. For games requiring authentication, consider running steamcmd manually once
//     to cache credentials, then mounting that cache into containers
//
// For production deployments, anonymous login is strongly recommended when supported.
func (b *ScriptBuilder) Build() string {
	var sb strings.Builder

	// Bash strict mode - exit on any error
	sb.WriteString("set -e\n\n")

	// Build the steamcmd command
	sb.WriteString("steamcmd \\\n")

	// Set installation directory
	installDir := b.config.InstallDir
	if installDir == "" {
		installDir = InstallDir
	}
	sb.WriteString(fmt.Sprintf("  +force_install_dir %s \\\n", installDir))

	// Login - anonymous or with credentials
	if b.config.Anonymous {
		sb.WriteString("  +login anonymous \\\n")
	} else {
		// Credentials are expected in environment variables for security.
		// STEAM_USERNAME and STEAM_PASSWORD should be injected from a Kubernetes Secret.
		//
		// Note: Steam Guard may still require manual verification on first login.
		// See package documentation for Steam Guard limitations and workarounds.
		sb.WriteString("  +login \"$STEAM_USERNAME\" \"$STEAM_PASSWORD\" \\\n")
	}

	// App update command with optional beta branch and validation
	sb.WriteString(fmt.Sprintf("  +app_update %d", b.config.AppID))

	if b.config.Beta != "" {
		sb.WriteString(fmt.Sprintf(" -beta %s", b.config.Beta))
		if b.config.BetaPassword != "" {
			// Beta password is also expected in an environment variable for security
			sb.WriteString(" -betapassword \"$STEAM_BETA_PASSWORD\"")
		}
	}

	if b.config.Validate {
		sb.WriteString(" validate")
	}

	sb.WriteString(" \\\n")
	sb.WriteString("  +quit\n\n")

	// Success message
	sb.WriteString("echo \"SteamCMD complete, game files ready.\"\n")

	return sb.String()
}

// GetLoginCommand returns just the login portion of the steamcmd command.
// Useful for testing or debugging authentication issues.
func (b *ScriptBuilder) GetLoginCommand() string {
	if b.config.Anonymous {
		return "+login anonymous"
	}
	return "+login \"$STEAM_USERNAME\" \"$STEAM_PASSWORD\""
}

// GetAppUpdateCommand returns just the app_update portion of the steamcmd command.
// Useful for testing or debugging update issues.
func (b *ScriptBuilder) GetAppUpdateCommand() string {
	cmd := fmt.Sprintf("+app_update %d", b.config.AppID)

	if b.config.Beta != "" {
		cmd += fmt.Sprintf(" -beta %s", b.config.Beta)
		if b.config.BetaPassword != "" {
			cmd += " -betapassword \"$STEAM_BETA_PASSWORD\""
		}
	}

	if b.config.Validate {
		cmd += " validate"
	}

	return cmd
}

// RequiresCredentials returns true if the script requires Steam credentials.
func (b *ScriptBuilder) RequiresCredentials() bool {
	return !b.config.Anonymous
}

// RequiresBetaPassword returns true if the script requires a beta password.
func (b *ScriptBuilder) RequiresBetaPassword() bool {
	return b.config.Beta != "" && b.config.BetaPassword != ""
}
