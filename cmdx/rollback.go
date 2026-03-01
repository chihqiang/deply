package cmdx

import (
	"chihqiang/deply/depx"
	"chihqiang/deply/flagx"
	"chihqiang/deply/sshx"
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func Rollback() *cli.Command {
	return &cli.Command{
		Name:  "rollback",
		Usage: "Revert your app to a previous version",
		Flags: flagx.VersionFlags(),
		Action: func(ctx context.Context, command *cli.Command) error {
			// 1. Load remote host configuration
			hostConfig, err := sshx.Load(command)
			if err != nil {
				return fmt.Errorf("failed to load host config: %v", err)
			}

			// 2. Load deployment configuration
			deployConfig := depx.Load(command)

			// 3. Iterate through all remote hosts using common helper
			sshx.ForEachHostWithSFTP(hostConfig, func(sftpClient *sshx.SFTPClient, sshClient *sshx.SSHClient, config *sshx.Config) error {
				// 3.1 Check if remote version directory exists
				if !sshx.RemoteExists(sftpClient, deployConfig.GetVersionRemoteDir()) {
					return fmt.Errorf("version not found: %s", deployConfig.GetVersionRemoteDir())
				}

				// 3.2 Execute deployment hooks (pre/post hooks)
				if err := depx.ExecuteDeployHooks(sshClient, deployConfig); err != nil {
					return fmt.Errorf("rollback failed: %v", err)
				}
				return nil
			})

			return nil
		},
	}
}
