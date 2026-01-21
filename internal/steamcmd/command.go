// Package steamcmd provides SteamCMD command argument generation for game server installation.
package steamcmd

import (
	"fmt"
)

// DefaultInstallDir is the default installation directory for game server files.
const DefaultInstallDir = "/data/server"

// CommandConfig holds configuration for building SteamCMD arguments.
type CommandConfig struct {
	// AppID is the Steam application ID for the dedicated server.
	AppID int32

	// InstallDir is the directory where game files will be installed.
	// Defaults to DefaultInstallDir if empty.
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

// CommandBuilder builds SteamCMD command arguments.
type CommandBuilder struct {
	config CommandConfig
}

// NewCommandBuilder creates a new CommandBuilder with the given configuration.
func NewCommandBuilder(config CommandConfig) *CommandBuilder {
	return &CommandBuilder{config: config}
}

// Build returns the SteamCMD arguments as a string slice.
// These arguments are intended to be passed directly to the steamcmd binary.
//
// Steam Guard Limitations:
// Authenticated (non-anonymous) logins may require Steam Guard verification.
// For production deployments, anonymous login is strongly recommended when supported.
func (b *CommandBuilder) Build() []string {
	args := []string{}

	// Set installation directory
	installDir := b.config.InstallDir
	if installDir == "" {
		installDir = DefaultInstallDir
	}
	args = append(args, "+force_install_dir", installDir)

	// Login - anonymous or with credentials
	if b.config.Anonymous {
		args = append(args, "+login", "anonymous")
	} else {
		// Credentials are expected in environment variables for security.
		// STEAM_USERNAME and STEAM_PASSWORD should be injected from a Kubernetes Secret.
		args = append(args, "+login", "$STEAM_USERNAME", "$STEAM_PASSWORD")
	}

	// App update command with optional beta branch and validation
	args = append(args, "+app_update", fmt.Sprintf("%d", b.config.AppID))

	if b.config.Beta != "" {
		args = append(args, "-beta", b.config.Beta)
		if b.config.BetaPassword != "" {
			// Beta password is also expected in an environment variable for security
			args = append(args, "-betapassword", "$STEAM_BETA_PASSWORD")
		}
	}

	if b.config.Validate {
		args = append(args, "validate")
	}

	// Quit
	args = append(args, "+quit")

	return args
}

// RequiresCredentials returns true if the command requires Steam credentials.
func (b *CommandBuilder) RequiresCredentials() bool {
	return !b.config.Anonymous
}

// RequiresBetaPassword returns true if the command requires a beta password.
func (b *CommandBuilder) RequiresBetaPassword() bool {
	return b.config.Beta != "" && b.config.BetaPassword != ""
}
