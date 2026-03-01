package cmdx

import (
	"chihqiang/deply/depx"
	"chihqiang/deply/flagx"
	"chihqiang/deply/sshx"
	"context"
	"fmt"

	"github.com/chihqiang/logx"
	"github.com/urfave/cli/v3"
)

func Rollback() *cli.Command {
	return &cli.Command{
		Name:  "rollback",
		Usage: "Revert your app to a previous version",
		Flags: flagx.VersionFlags(),
		Action: func(ctx context.Context, command *cli.Command) error {
			// 1. Load remote host configuration
			// hostConfig is a slice containing information of all hosts to be deployed (Host, Port, User, Key, etc.)
			hostConfig, err := sshx.Load(command)
			if err != nil {
				return fmt.Errorf("failed to load host config: %v", err)
			}

			// 2. Load deployment configuration
			// deployConfig contains version number, directory, hook commands and other information
			deployConfig := depx.Load(command)

			// 3. Iterate through all remote hosts to perform operations
			for _, config := range hostConfig {
				// 3.1 Open SSH connection
				sshClient, err := sshx.Open(config)
				if err != nil {
					// Connection failed, only log and continue to next host
					logx.Warn("[%s] Failed to open SSH connection: %v", config.Host, err)
					continue
				}
				// Ensure connection is closed when function returns to avoid resource leakage
				defer sshClient.Close()
				sftpClient, err := sshx.OpenSftp(sshClient)
				if err != nil {
					logx.Warn("[%s] create sftp client: %v", config.Host, err)
					continue
				}
				defer sftpClient.Close()
				// 3.2 Check if remote version directory exists
				// If version directory does not exist, rollback is not possible
				if !sshx.RemoteExists(sftpClient, deployConfig.GetVersionRemoteDir()) {
					logx.Warn("[%s] version not found: %s", config.Host, deployConfig.GetVersionRemoteDir())
					continue
				}

				// 3.3 Execute deployment hooks (pre/post hooks)
				// This can be understood as "rollback operation" or redirecting to specified version
				if err := depx.ExecuteDeployHooks(sshClient, deployConfig); err != nil {
					// Hook execution failure only logs, continue to next host
					logx.Warn("[%s] rollback failed: %v", config.Host, err)
					continue
				}
			}

			// 4. All hosts processing completed
			return nil

		},
	}
}
