package flagx

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	DefaultTimeout            = 30 * time.Second
	DefaultRemoteRepoPattern  = "/data/wwwroot/%s/releases"
	DefaultCurrentLinkPattern = "/data/wwwroot/%s/current"
)

const (
	FlagDir     = "dir"
	FlagVersion = "version"
	FlagInclude = "include"
	FlagExclude = "exclude"

	FlagHosts      = "hosts"
	FlagKey        = "key"
	FlagPassphrase = "passphrase"
	FlagTimeout    = "timeout"

	FlagRemoteRepo  = "remote-repo"
	FlagCurrentLink = "current-link"
	FlagHookPre     = "hook-pre-host"
	FlagHookPost    = "hook-post-host"
)

const (
	EnvHosts      = "DEPLY_HOSTS"
	EnvKey        = "DEPLY_KEY"
	EnvPassphrase = "DEPLY_PASSPHRASE"
	EnvTimeout    = "DEPLY_TIMEOUT"

	EnvHookPre  = "DEPLY_HOOK_PRE"
	EnvHookPost = "DEPLY_HOOK_POST"
)

var (
	workDirOnce sync.Once
	workDir     string
	baseDir     string
)

// getWorkDir returns the current working directory, initialized only once
func getWorkDir() string {
	workDirOnce.Do(func() {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			workDir = "."
		}
		baseDir = path.Base(workDir)
	})
	return workDir
}

// getBaseDir returns the base name of the current working directory
func getBaseDir() string {
	getWorkDir() // ensure initialized
	return baseDir
}

func PublishFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  FlagDir,
			Usage: "Local directory or packaged file directory",
			Value: getWorkDir(),
		},
		&cli.StringSliceFlag{
			Name:  FlagInclude,
			Usage: "Files or directories to include when packaging, relative to --dir",
		},
		&cli.StringSliceFlag{
			Name:  FlagExclude,
			Usage: "Files or directories to exclude when packaging, relative to --dir",
		},
	}
}

func VersionFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagVersion,
			Aliases: []string{"V"},
			Usage:   "Specify the version or branch to deploy, default is main",
			Value:   time.Now().Format("20060102150405"),
		},
	}
}

func SSHFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:     FlagHosts,
			Usage:    "List of remote hosts, format: user[:password]@host[:port]",
			Required: true,
			Sources:  cli.EnvVars(EnvHosts),
		},
		&cli.StringFlag{
			Name:    FlagKey,
			Usage:   "Path to SSH private key (optional)",
			Sources: cli.EnvVars(EnvKey),
		},
		&cli.StringFlag{
			Name:    FlagPassphrase,
			Usage:   "Passphrase for SSH private key (optional)",
			Sources: cli.EnvVars(EnvPassphrase),
		},
		&cli.DurationFlag{
			Name:    FlagTimeout,
			Value:   DefaultTimeout,
			Usage:   "SSH connection timeout",
			Sources: cli.EnvVars(EnvTimeout),
		},
		&cli.StringFlag{
			Name:    FlagHookPre,
			Usage:   "Remote command to run before deployment (optional)",
			Sources: cli.EnvVars(EnvHookPre),
		},
		&cli.StringFlag{
			Name:    FlagHookPost,
			Usage:   "Remote command to run after deployment (optional)",
			Sources: cli.EnvVars(EnvHookPost),
		},
		&cli.StringFlag{
			Name:  FlagRemoteRepo,
			Usage: "Remote deployment repository path",
			Value: fmt.Sprintf(DefaultRemoteRepoPattern, getBaseDir()),
		},
		&cli.StringFlag{
			Name:  FlagCurrentLink,
			Usage: "Symbolic link path pointing to the current version",
			Value: fmt.Sprintf(DefaultCurrentLinkPattern, getBaseDir()),
		},
	}
}
