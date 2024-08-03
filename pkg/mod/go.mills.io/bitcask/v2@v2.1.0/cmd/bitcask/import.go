package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// ImportCmd creates the import command
func ImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "import",
		Aliases: []string{"restore", "read"},
		Short:   "Import a database",
		Long: `This command allows you to import or restore a database from a
previous export/dump using the export command either creating a new database
or adding additional key/value pairs to an existing one.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string

			path := viper.GetString("path")

			if len(args) == 1 {
				input = args[0]
			} else {
				input = "-"
			}

			db, err := bitcask.Open(path, getBitcaskOptionsFromFlags(cmd)...)
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			r := cmd.InOrStdin()
			if input != "-" {
				f, err := os.Open(input)
				if err != nil {
					return fmt.Errorf("error opening input file %q for reading: %w", input, err)
				}
				r = f
				defer f.Close()
			}

			var kv kvPair

			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				if err := json.Unmarshal(scanner.Bytes(), &kv); err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}

				key, err := base64.StdEncoding.DecodeString(kv.Key)
				if err != nil {
					return fmt.Errorf("error decoding key %q: %w", kv.Key, err)
				}

				value, err := base64.StdEncoding.DecodeString(kv.Value)
				if err != nil {
					return fmt.Errorf("error decoding value for %q: %w", kv.Key, err)
				}

				if err := db.Put(key, value); err != nil {
					return fmt.Errorf("error writing key %q: %w", key, err)
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading input: %w", err)

			}

			return nil
		},
	}

	addBitcaskFlagOptions(cmd)

	return cmd
}
