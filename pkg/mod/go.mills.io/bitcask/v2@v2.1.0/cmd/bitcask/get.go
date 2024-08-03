package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// GetCmd creates the get command
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <key>",
		Aliases: []string{"view"},
		Short:   "Get a new Key and display its Value",
		Long:    `This retrieves a key and display its value`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			key := args[0]

			db, err := bitcask.Open(path, bitcask.WithOpenReadonly(viper.GetBool(openReadonlyOption)))
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			value, err := db.Get([]byte(key))
			if err != nil {
				return fmt.Errorf("error reading key %q: %w", key, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s", string(value))

			return nil
		},
	}

	return cmd
}
