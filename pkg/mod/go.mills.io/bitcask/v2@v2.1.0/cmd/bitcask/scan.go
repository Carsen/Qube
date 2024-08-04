package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// ScanCmd creates the scan command
func ScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scan <prefix>",
		Aliases: []string{"search", "find"},
		Short:   "Perform a prefix scan for keys",
		Long: `This perform a prefix scan for keys  starting with the given
prefix. This uses a Trie to search for matching keys and returns all matched
keys.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			prefix := args[0]

			db, err := bitcask.Open(path, bitcask.WithOpenReadonly(viper.GetBool(openReadonlyOption)))
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			if err := db.Scan(bitcask.Key(prefix), func(key bitcask.Key) error {
				value, err := db.Get(key)
				if err != nil {
					return fmt.Errorf("error reading key %q: %w", key, err)
				}

				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(value)); err != nil {
					return fmt.Errorf("error writing key %q: %w", key, err)
				}

				return nil
			}); err != nil {
				return fmt.Errorf("error ranging over keys: %w", err)
			}

			return nil
		},
	}

	return cmd
}
