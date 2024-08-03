package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// InitCmd creates the init command
func InitCmd() *cobra.Command {
	var cmd *cobra.Command

	cmd = &cobra.Command{
		Use:     "init",
		Aliases: []string{"create", "new"},
		Short:   "Initialize a new database",
		Long:    `This initializes a new database with persisted options`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			db, err := bitcask.Open(path, getBitcaskOptionsFromFlags(cmd)...)
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			return nil
		},
	}

	addBitcaskFlagOptions(cmd)

	return cmd
}
