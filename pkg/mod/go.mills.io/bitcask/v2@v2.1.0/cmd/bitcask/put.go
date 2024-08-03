package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// PutCmd creates the put command
func PutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "put <key> [<value>]",
		Aliases: []string{"add", "set", "store", "write"},
		Short:   "Adds a new Key/Value pair",
		Long: `This adds a new key/value pair or modifies an existing one.

If the value is not specified as an argument it is read from standard input.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			key := args[0]

			var value io.Reader
			if len(args) > 1 {
				value = bytes.NewBufferString(args[1])
			} else {
				value = cmd.InOrStdin()
			}

			db, err := bitcask.Open(path, getBitcaskOptionsFromFlags(cmd)...)
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			data, err := io.ReadAll(value)
			if err != nil {
				return fmt.Errorf("error reading value: %w", err)
			}

			if err := db.Put([]byte(key), data); err != nil {
				return fmt.Errorf("error writing key %q: %w", key, err)
			}

			return nil
		},
	}

	addBitcaskFlagOptions(cmd)

	return cmd
}
