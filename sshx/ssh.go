package sshx

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config SSH connection configuration
type Config struct {
	User       string        `yaml:"user"`       // Login username
	Password   string        `yaml:"password"`   // Password, optional
	Host       string        `yaml:"host"`       // Remote host address
	Port       int           `yaml:"port"`       // Port, default 22
	KeyPath    string        `yaml:"keyPath"`    // Private key path, optional
	Passphrase string        `yaml:"passphrase"` // Private key password, optional
	Timeout    time.Duration `yaml:"timeout"`    // SSH connection timeout
}

// Open establishes an SSH connection
func Open(cfg *Config) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// 1. If private key is provided, prioritize private key authentication
	if cfg.KeyPath != "" {
		key, err := os.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("read key error: %w", err)
		}

		var signer ssh.Signer
		if cfg.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(cfg.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, fmt.Errorf("parse key error: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 2. If password is provided, add password authentication
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	// 3. SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Ignore host key verification
		Timeout:         cfg.Timeout,
	}

	// 4. Combine host:port
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// 5. Establish TCP + SSH connection
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh dial error: %w", err)
	}

	return client, nil
}

// Command executes a command on the remote host and returns the output
func Command(ssh *ssh.Client, cmd string) (string, error) {
	// 1. Create new session
	session, err := ssh.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session error: %w", err)
	}
	defer session.Close()

	// 2. Execute command and get stdout + stderr
	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

// ParseSSHURL parses simplified SSH URL format
// Format: user[:password]@host[:port]
// Returns SSH configuration object
func ParseSSHURL(raw string) (*Config, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty ssh target")
	}

	user := ""
	password := ""
	port := 22 // Default port

	at := strings.Split(raw, "@")
	if len(at) == 2 {
		// user[:password]@host[:port]
		userPart := at[0]
		hostPart := at[1]

		// Parse user information
		up := strings.SplitN(userPart, ":", 2)
		user = up[0]
		if len(up) == 2 {
			password = up[1]
		}

		// Parse host[:port]
		hp := strings.SplitN(hostPart, ":", 2)
		host := hp[0]
		if len(hp) == 2 {
			p, err := strconv.Atoi(hp[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", hp[1])
			}
			port = p
		}

		// If user is not specified, use current system user
		if user == "" {
			user = os.Getenv("USER")
		}

		return &Config{
			User:     user,
			Password: password,
			Host:     host,
			Port:     port,
			Timeout:  10 * time.Second,
		}, nil
	}

	// No @, only host or host:port
	hp := strings.SplitN(raw, ":", 2)
	host := hp[0]
	if len(hp) == 2 {
		p, err := strconv.Atoi(hp[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", hp[1])
		}
		port = p
	}

	user = os.Getenv("USER")
	return &Config{
		User:    user,
		Host:    host,
		Port:    port,
		Timeout: 10 * time.Second,
	}, nil
}
