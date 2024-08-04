package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// RangeCmd creates the range command
func RangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "range <start> <end>",
		Aliases: []string{},
		Short:   "Perform a range scan for keys from a start to end key",
		Long: `This perform a range scan for keys  starting with the given start
key and ending with the end key. This uses a Trie to search for matching keys
within the range and returns all matched keys.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			start := args[0]
			end := args[1]

			db, err := bitcask.Open(path, bitcask.WithOpenReadonly(viper.GetBool(openReadonlyOption)))
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			if err := db.Range(bitcask.Key(start), bitcask.Key(end), func(key bitcask.Key) error {
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
