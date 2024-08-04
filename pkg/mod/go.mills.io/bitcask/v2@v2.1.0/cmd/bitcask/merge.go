package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// MergeCmd creates the merge command
func MergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "merge",
		Aliases: []string{"clean", "compact", "defrag"},
		Short:   "Merges the Datafiles in the Database",
		Long: `This merges all non-active Datafiles in the Database and
compacts the data stored on disk. Old values are removed as well as deleted
keys.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			db, err := bitcask.Open(path, getBitcaskOptionsFromFlags(cmd)...)
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}

			if err := db.Merge(); err != nil {
				return fmt.Errorf("error merging database: %w", err)
			}

			return nil
		},
	}

	addBitcaskFlagOptions(cmd)

	return cmd
}
