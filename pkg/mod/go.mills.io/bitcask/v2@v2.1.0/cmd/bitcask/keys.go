package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// KeysCmd creates the keys command
func KeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"list", "ls"},
		Short:   "Display all keys in Database",
		Long:    `This displays all known keys in the Database"`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			db, err := bitcask.Open(path, bitcask.WithOpenReadonly(viper.GetBool(openReadonlyOption)))
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			if err := db.ForEach(func(key bitcask.Key) error {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(key)); err != nil {
					return fmt.Errorf("error writing key %q: %w", key, err)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("error reading keys: %w", err)
			}

			return nil
		},
	}

	return cmd
}
