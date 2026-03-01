package cmdx

import (
	"chihqiang/deply/flagx"
	"chihqiang/deply/sshx"
	"chihqiang/deply/utilx"
	"context"
	"sort"
	"strconv"

	"github.com/chihqiang/logx"
	"github.com/urfave/cli/v3"
)

// History returns a CLI command for viewing deployment history on remote hosts
func History() *cli.Command {
	return &cli.Command{
		Name:  "history",
		Usage: "Peek into your app's past glory",
		Flags: []cli.Flag{},
		Action: func(ctx context.Context, command *cli.Command) error {
			// 1. Load remote host configuration
			hostConfig, err := sshx.Load(command)
			if err != nil {
				return err
			}
			// 2. all is used to store information of all hosts corresponding to each remote path
			all := make(map[string][]HostFileInfo)
			// 3. Iterate through all hosts
			for _, config := range hostConfig {
				// 3.1 Open SSH connection
				sshClient, err := sshx.Open(config)
				if err != nil {
					logx.Warn("[%s] Failed to open SSH connection: %v", config.Host, err)
					continue
				}
				defer sshClient.Close()
				sftpClient, err := sshx.OpenSftp(sshClient)
				if err != nil {
					logx.Warn("[%s] create sftp client: %v", config.Host, err)
					continue
				}
				defer sftpClient.Close()
				// 3.2 List remote directory file information
				// Parameters read from CLI command FlagRemoteRepo and FlagCurrentLink
				list, err := sshx.List(sftpClient, command.String(flagx.FlagRemoteRepo), command.String(flagx.FlagCurrentLink))
				if err != nil {
					logx.Warn("[%s] list failed: %v", config.Host, err)
					continue
				}
				// 3.3 Add each file information to the all map
				for _, fi := range list {
					all[fi.Path] = append(all[fi.Path], HostFileInfo{
						Host: config.Host,
						File: fi,
					})
				}
			}

			// 4. Find versions that exist on all hosts
			var common []HostFileInfo
			successHosts := len(hostConfig)
			for _, infos := range all {
				// If a path has records on all hosts, it is considered common
				if len(infos) == successHosts {
					common = append(common, infos[0])
				}
			}

			// 5. Print table display
			printTable(common)

			return nil
		},
	}
}

// printTable outputs deployment history information in table format
func printTable(list []HostFileInfo) {
	// 1. Sort by file modification time in descending order, newest version at top
	sort.Slice(list, func(i, j int) bool {
		return list[i].File.FileInfo.ModTime().
			After(list[j].File.FileInfo.ModTime())
	})

	// 2. Create table object
	tbl := utilx.NewTable()
	tbl.AddHeader("Host", "Path", "version", "Current", "ModTime")

	// 3. Iterate through list and add each record to the table
	for _, fi := range list {
		tbl.AddLine(
			fi.Host,                            // Host name
			fi.File.Path,                       // File path
			fi.File.Name,                       // File name or version name
			strconv.FormatBool(fi.File.IsLink), // Whether it is a soft link
			fi.File.FileInfo.ModTime().Format("2006-01-02 15:04:05"), // Modification time
		)
	}

	// 4. Output table
	tbl.Print()
}

// HostFileInfo contains host name and remote file information
type HostFileInfo struct {
	Host string
	File sshx.FileInfo
}
