package flagx

import (
	"fmt"
	"github.com/urfave/cli/v3"
	"os"
	"path"
	"time"
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
	dir     string
	baseDir string
)

func init() {
	dir, _ = os.Getwd()
	baseDir = path.Base(dir)
}
func PublishFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  FlagDir,
			Usage: "Local directory or packaged file directory",
			Value: dir,
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
			Value: fmt.Sprintf(DefaultRemoteRepoPattern, baseDir),
		},
		&cli.StringFlag{
			Name:  FlagCurrentLink,
			Usage: "Symbolic link path pointing to the current version",
			Value: fmt.Sprintf(DefaultCurrentLinkPattern, baseDir),
		},
	}
}
