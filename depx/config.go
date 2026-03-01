package depx

import (
	"errors"
	"path"
	"strings"
)

type Config struct {
	Dir         string   `yaml:"dir"`         // Deployment root directory (local or remote path)
	Version     string   `yaml:"version"`     // Deployment version number, for example v1.0.0 or 20260102153000
	Include     []string `yaml:"include"`     // List of files or directories to include
	Exclude     []string `yaml:"exclude"`     // List of files or directories to exclude
	RemoteRepo  string   `yaml:"remoteRepo"`  // Directory for storing remote versions, for example /data/app/releases
	CurrentLink string   `yaml:"currentLink"` // Current symbolic link path, for example /data/app/current
	HookPre     string   `yaml:"hookPre"`     // Hook command to execute before deployment
	HookPost    string   `yaml:"hookPost"`    // Hook command to execute after deployment
}

// Validate validates configuration parameters
// Check required fields are not empty to avoid deployment failure
func (c *Config) Validate() error {
	if c.Dir == "" {
		return errors.New("dir must not be empty")
	}
	if c.RemoteRepo == "" {
		return errors.New("remoteRepo must not be empty")
	}
	if c.CurrentLink == "" {
		return errors.New("currentLink must not be empty")
	}
	if c.Version == "" {
		return errors.New("version must not be empty")
	}
	return nil
}

// GetVersionRemoteDir gets the complete path of remote version directory
// For example /data/app/releases/v1.0.0
func (c *Config) GetVersionRemoteDir() string {
	return path.Join(c.GetRemoteRepo(), c.GetVersion())
}

// GetVersion gets the version number
func (c *Config) GetVersion() string {
	return c.Version
}

// GetRemoteRepo gets the directory path for storing remote versions
// If the path does not start with /, it will be automatically added
// For example remoteRepo = "data/app/releases" → "/data/app/releases"
func (c *Config) GetRemoteRepo() string {
	remoteRepo := c.RemoteRepo
	if !strings.HasPrefix(remoteRepo, "/") {
		remoteRepo = "/" + remoteRepo
	}
	return remoteRepo
}

// GetCurrentLink gets the current symbolic link path
// If the path does not start with /, it will be automatically added
// For example currentLink = "current" → "/current"
func (c *Config) GetCurrentLink() string {
	currentLink := c.CurrentLink
	if !strings.HasPrefix(currentLink, "/") {
		currentLink = "/" + currentLink
	}
	return currentLink
}

// GetHookPre gets the pre-deployment hook command
func (c *Config) GetHookPre() string {
	return c.HookPre
}

// GetHookPost gets the post-deployment hook command
func (c *Config) GetHookPost() string {
	return c.HookPost
}
