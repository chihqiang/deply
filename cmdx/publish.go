package cmdx

import (
	"chihqiang/deply/depx"
	"chihqiang/deply/flagx"
	"chihqiang/deply/sshx"
	"context"
	"fmt"
	"os"

	"github.com/chihqiang/logx"
	"github.com/urfave/cli/v3"
)

func Publish() *cli.Command {
	return &cli.Command{
		Name:  "publish",
		Usage: "Fire up your app remotely",
		Flags: append(flagx.PublishFlags(), flagx.VersionFlags()...),
		Action: func(ctx context.Context, command *cli.Command) error {
			// 1. Load remote host configuration
			// Returns a slice, each element contains host, port, user, key and other information
			hostConfig, err := sshx.Load(command)
			if err != nil {
				return fmt.Errorf("failed to load host config: %v", err)
			}

			// 2. Load deployment configuration
			// deploy.Load reads yaml or command line parameters and returns deploy.Config
			deployConfig := depx.Load(command)

			// 3. Pack local directory as tar.gz file
			// Returns the temporary file path after packing
			localTarGz, err := depx.PackDir(deployConfig)
			if err != nil {
				return fmt.Errorf("failed to pack directory: %v", err)
			}
			// Delete temporary file after deployment to avoid occupying space
			defer func() {
				_ = os.Remove(localTarGz)
			}()

			// 4. Iterate through all hosts and execute deployment sequentially
			for _, config := range hostConfig {
				// Open SSH connection
				sshClient, err := sshx.Open(config)
				if err != nil {
					logx.Warn("[%s] Failed to open SSH connection: %v", config.Host, err)
					continue // Current host failed, continue to next host
				}
				// Ensure connection is closed to avoid resource leakage
				defer sshClient.Close()
				// 5. Execute deployment
				// Including uploading archive, extracting, executing hooks, updating currentLink
				if err := depx.PostDeployHost(sshClient, localTarGz, deployConfig); err != nil {
					logx.Warn("[%s] Deploy failed: %v", config.Host, err)
					continue // Current host failed, continue to next host
				}
			}
			// 6. All hosts deployment completed
			return nil
		},
	}
}
