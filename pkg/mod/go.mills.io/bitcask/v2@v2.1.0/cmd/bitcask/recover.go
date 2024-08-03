package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mills.io/bitcask/v2/internal/config"
	"go.mills.io/bitcask/v2/internal/data"
)

// RecoverCmd creates the recover command
func RecoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recover",
		Aliases: []string{"fix", "repair"},
		Short:   "Analyze and recover the index file for corruption scenarios",
		Long: `This analyses files to detect different forms of persistence corruption in 
persisted files. It also allows to recover the files to the latest point of integrity.
Recovered files have the .recovered extension`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			cfg, err := config.Load(filepath.Join(path, "config.json"))
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			if err := data.CheckAndRecover(path, cfg); err != nil {
				return fmt.Errorf("recovering database: %s", err)
			}

			return nil
		},
	}

	return cmd
}
