package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

var errNotAllDataWritten = errors.New("error: not all data written")

type kvPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func exportKey(db *bitcask.Bitcask, w io.Writer) bitcask.KeyFunc {
	return func(key bitcask.Key) error {
		value, err := db.Get(key)
		if err != nil {
			return fmt.Errorf("error reading key %q: %w", key, err)
		}

		kv := kvPair{
			Key:   base64.StdEncoding.EncodeToString([]byte(key)),
			Value: base64.StdEncoding.EncodeToString(value),
		}

		data, err := json.Marshal(&kv)
		if err != nil {
			return fmt.Errorf("error serializing key %q: %w", key, err)
		}

		if n, err := w.Write(data); err != nil || n != len(data) {
			if err == nil && n != len(data) {
				err = errNotAllDataWritten
			}
			return fmt.Errorf("error exporting key %q: %w", key, err)
		}

		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("error writing newline separator: %w", err)
		}

		return nil
	}
}

// ExportCmd creates the export command
func ExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"backup", "dump"},
		Short:   "Export a database",
		Long: `This command allows you to export or dump/backup a database's
key/values into a long-term portable archival format suitable for backup and
restore purposes or migrating from older on-disk formats of Bitcask.

All key/value pairs are base64 encoded and serialized as JSON one pair per
line to form an output stream to either standard output or a file. You can
optionally compress the output with standard compression tools such as gzip.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var output string

			path := viper.GetString("path")

			if len(args) == 1 {
				output = args[0]
			} else {
				output = "-"
			}

			db, err := bitcask.Open(path, bitcask.WithOpenReadonly(viper.GetBool(openReadonlyOption)))
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			w := cmd.OutOrStdout()
			if output != "-" {
				f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, os.FileMode(644))
				if err != nil {
					return fmt.Errorf("error opening output file %q for writing: %w", output, err)
				}
				w = f
				defer f.Close()
			}

			if err := db.ForEach(exportKey(db, w)); err != nil {
				return fmt.Errorf("error exporting keys: %w", err)
			}

			return nil
		},
	}

	return cmd
}
