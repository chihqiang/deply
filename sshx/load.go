package sshx

import (
	"chihqiang/deply/flagx"
	"github.com/urfave/cli/v3"
	"strings"
)

func Load(cmd *cli.Command) ([]*Config, error) {
	hosts := cmd.StringSlice(flagx.FlagHosts)
	var configs []*Config
	for _, h := range hosts {
		cfg, err := ParseSSHURL(strings.TrimSpace(h))
		if err != nil {
			return nil, err
		}
		cfg.KeyPath = cmd.String(flagx.FlagKey)
		cfg.Timeout = cmd.Duration(flagx.FlagTimeout)
		cfg.Passphrase = cmd.String(flagx.FlagPassphrase)
		configs = append(configs, cfg)
	}
	return configs, nil
}
