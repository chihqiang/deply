package depx

import (
	"chihqiang/deply/sshx"
	"fmt"
	"github.com/pkg/sftp"
	"path"
	"path/filepath"

	"github.com/chihqiang/logx"
	"golang.org/x/crypto/ssh"
)

// PostDeployHost executes deployment on remote server
func PostDeployHost(sshClient *ssh.Client, localTarGz string, config *Config) error {
	// Validate configuration parameters
	if err := config.Validate(); err != nil {
		return err
	}
	sftpClient, err := sshx.OpenSftp(sshClient)
	if err != nil {
		return fmt.Errorf("create sftp client: %w", err)
	}
	defer sftpClient.Close()
	// Execute pre-deployment checks
	if err := preDeployChecks(sftpClient, config); err != nil {
		return err
	}
	// Ensure parent directory of currentLink exists
	// path.Dir returns the parent directory of a path, for example /data/app/current → /data/app
	baseLinkDir := path.Dir(config.GetCurrentLink())
	if err := sshx.Mkdir(sftpClient, baseLinkDir); err != nil {
		return fmt.Errorf("failed to create base directory %s: %w", baseLinkDir, err)
	}

	// Ensure version directory exists (for example /data/app/releases/v1.0.0)
	if err := sshx.Mkdir(sftpClient, config.GetVersionRemoteDir()); err != nil {
		return fmt.Errorf("version directory creation failed %s: %w", config.GetVersionRemoteDir(), err)
	}
	// Upload archive to remote version directory
	remoteTar := path.Join(config.GetVersionRemoteDir(), filepath.Base(localTarGz))
	if err := sshx.UploadFile(sftpClient, localTarGz, remoteTar); err != nil {
		return fmt.Errorf("file upload failed : %w", err)
	}
	// Extract uploaded tar.gz file and delete archive
	// Use %q to automatically add quotes, preventing errors with spaces or special characters in paths
	tarCmd := fmt.Sprintf(
		"cd %q && tar xf %q && rm -f %q",
		config.GetVersionRemoteDir(), // Enter version directory
		remoteTar,                    // Extract remote archive
		remoteTar,                    // Delete archive after extraction
	)
	if _, err := sshx.Command(sshClient, tarCmd); err != nil {
		return fmt.Errorf("decompression failed: %w", err)
	}

	// Execute deployment hooks (pre-hook / post-hook) and update currentLink
	if err := ExecuteDeployHooks(sshClient, config); err != nil {
		return fmt.Errorf("hook deployment failed: %w", err)
	}

	//Post-deployment verification (ensure currentLink correctly points to new version)
	if err := postDeployVerification(sftpClient, config); err != nil {
		return fmt.Errorf("post-deployment verification failed.: %w", err)
	}

	return nil
}

// ExecuteDeployHooks executes pre-hook / update currentLink / post-hook
func ExecuteDeployHooks(sshClient *ssh.Client, config *Config) error {
	// Pre-deployment hook
	if hookPre := config.GetHookPre(); hookPre != "" {
		// cd to version directory to execute pre-hook
		if _, err := sshx.Command(sshClient, fmt.Sprintf("cd %s && %s", config.GetVersionRemoteDir(), hookPre)); err != nil {
			// Failure only warns, does not block deployment
			logx.Warn("pre-hook failed: %v", err)
			return fmt.Errorf("pre-hook failed: %v", err)
		}
	}

	// Update currentLink to point to new version (atomic operation ln -sfn)
	deployCmd := fmt.Sprintf("ln -sfn %s %s", config.GetVersionRemoteDir(), config.GetCurrentLink())
	if _, err := sshx.Command(sshClient, deployCmd); err != nil {
		return fmt.Errorf("deploy cmdx failed: %w", err)
	}

	// Post-deployment hook
	if hookPost := config.GetHookPost(); hookPost != "" {
		if _, err := sshx.Command(sshClient, fmt.Sprintf("cd %s && %s", config.GetVersionRemoteDir(), hookPost)); err != nil {
			// Failure only warns, does not block deployment
			logx.Warn("post-hook: %v", err)
			return fmt.Errorf("post-hook: %v", err)
		}
	}
	return nil
}

// preDeployChecks executes pre-deployment checks
func preDeployChecks(sftpClient *sftp.Client, config *Config) error {
	// Check if currentLink exists and is a symbolic link
	isLink, err := sshx.IsSymlink(sftpClient, config.GetCurrentLink())
	if err != nil {
		return fmt.Errorf("symbolic link check failed %s: %w", config.GetCurrentLink(), err)
	}
	if !isLink {
		// If currentLink already exists and is not a symbolic link, manual handling is required
		return fmt.Errorf("the deployment directory %s already exists and is not a symbolic link. For data security, please manually back it up and delete it", config.GetCurrentLink())
	}

	// Check if version directory already exists to prevent overwriting existing versions
	versionDir := config.GetVersionRemoteDir()
	if sshx.RemoteExists(sftpClient, versionDir) {
		return fmt.Errorf("version %s already exists. Please use a different version number or clear the old version first", config.Version)
	}

	return nil
}

// postDeployVerification post-deployment verification
func postDeployVerification(sftpClient *sftp.Client, config *Config) error {
	// Read the symbolic link target of currentLink
	actualTarget, err := sshx.ReadLink(sftpClient, config.GetCurrentLink())
	if err != nil {
		return fmt.Errorf("failed to read symbolic link: %w", err)
	}
	// Ensure currentLink correctly points to the new version directory
	if actualTarget != config.GetVersionRemoteDir() {
		return fmt.Errorf("the symbolic link is pointing to the wrong address: Expected %s, actual %s", config.GetVersionRemoteDir(), actualTarget)
	}
	return nil
}
