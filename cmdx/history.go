package cmdx

import (
	"chihqiang/deply/flagx"
	"chihqiang/deply/sshx"
	"chihqiang/deply/utilx"
	"context"
	"sort"
	"strconv"

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

			// 2. Collect information from all hosts
			all := make(map[string][]HostFileInfo)

			// 3. Iterate through all hosts using common helper
			sshx.ForEachHostWithSFTP(hostConfig, func(sftpClient *sshx.SFTPClient, sshClient *sshx.SSHClient, config *sshx.Config) error {
				list, err := sshx.List(sftpClient, command.String(flagx.FlagRemoteRepo), command.String(flagx.FlagCurrentLink))
				if err != nil {
					return err
				}
				for _, fi := range list {
					all[fi.Path] = append(all[fi.Path], HostFileInfo{
						Host: config.Host,
						File: fi,
					})
				}
				return nil
			})

			// 4. Find versions that exist on all hosts
			var common []HostFileInfo
			successHosts := len(hostConfig)
			for _, infos := range all {
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
			fi.Host,
			fi.File.Path,
			fi.File.Name,
			strconv.FormatBool(fi.File.IsLink),
			fi.File.FileInfo.ModTime().Format("2006-01-02 15:04:05"),
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
