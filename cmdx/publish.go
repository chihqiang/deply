package cmdx

import (
	"chihqiang/deply/depx"
	"chihqiang/deply/flagx"
	"chihqiang/deply/sshx"
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/ssh"
)

func Publish() *cli.Command {
	return &cli.Command{
		Name:  "publish",
		Usage: "Fire up your app remotely",
		Flags: append(flagx.PublishFlags(), flagx.VersionFlags()...),
		Action: func(ctx context.Context, command *cli.Command) error {
			// 1. Load remote host configuration
			hostConfig, err := sshx.Load(command)
			if err != nil {
				return fmt.Errorf("failed to load host config: %v", err)
			}

			// 2. Load deployment configuration
			deployConfig := depx.Load(command)

			// 3. Pack local directory as tar.gz file
			localTarGz, err := depx.PackDir(deployConfig)
			if err != nil {
				return fmt.Errorf("failed to pack directory: %v", err)
			}
			defer func() {
				_ = os.Remove(localTarGz)
			}()

			// 4. Iterate through all hosts and execute deployment
			sshx.ForEachHost(hostConfig, func(sshClient *ssh.Client, config *sshx.Config) error {
				return depx.PostDeployHost(sshClient, localTarGz, deployConfig)
			})

			return nil
		},
	}
}
