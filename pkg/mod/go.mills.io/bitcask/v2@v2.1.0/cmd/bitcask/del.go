package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// DelCmd creates the del command
func DelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del <key>",
		Aliases: []string{"delete", "remove", "rm"},
		Short:   "Delete a key and its value",
		Long:    `This deletes a key and its value`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			key := args[0]

			db, err := bitcask.Open(path, getBitcaskOptionsFromFlags(cmd)...)
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			err = db.Delete([]byte(key))
			if err != nil {
				return fmt.Errorf("error deleting key %q: %w", key, err)
			}

			return nil
		},
	}

	addBitcaskFlagOptions(cmd)

	return cmd
}
