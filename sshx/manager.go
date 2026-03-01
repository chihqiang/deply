package sshx

import (
	"github.com/chihqiang/logx"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClient is an alias for ssh.Client
type SSHClient = ssh.Client

// SFTPClient is an alias for sftp.Client
type SFTPClient = sftp.Client

// ExecuteFunc defines a function that executes with an SSH client
type ExecuteFunc func(*SSHClient, *Config) error

// ExecuteWithSFTPFunc defines a function that executes with both SSH and SFTP clients
type ExecuteWithSFTPFunc func(*SFTPClient, *SSHClient, *Config) error

// ForEachHost iterates over all hosts and executes the given function with SSH connection
// It handles connection opening/closing and error logging automatically
func ForEachHost(hostConfigs []*Config, fn ExecuteFunc) {
	for _, config := range hostConfigs {
		func() {
			sshClient, err := Open(config)
			if err != nil {
				logx.Warn("[%s] Failed to open SSH connection: %v", config.Host, err)
				return
			}
			defer sshClient.Close()

			if err := fn(sshClient, config); err != nil {
				logx.Warn("[%s] Operation failed: %v", config.Host, err)
			}
		}()
	}
}

// ForEachHostWithSFTP iterates over all hosts and executes the given function with both SSH and SFTP connections
// It handles connection opening/closing and error logging automatically
func ForEachHostWithSFTP(hostConfigs []*Config, fn ExecuteWithSFTPFunc) {
	for _, config := range hostConfigs {
		func() {
			sshClient, err := Open(config)
			if err != nil {
				logx.Warn("[%s] Failed to open SSH connection: %v", config.Host, err)
				return
			}
			defer sshClient.Close()

			sftpClient, err := OpenSftp(sshClient)
			if err != nil {
				logx.Warn("[%s] Failed to create SFTP client: %v", config.Host, err)
				return
			}
			defer sftpClient.Close()

			if err := fn(sftpClient, sshClient, config); err != nil {
				logx.Warn("[%s] Operation failed: %v", config.Host, err)
			}
		}()
	}
}
