package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
)

// StatsCmd creates the stats command
func StatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{},
		Short:   "Display statistics about the Database",
		Long:    `This displays statistics about the Database"`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := viper.GetString("path")

			db, err := bitcask.Open(path, bitcask.WithOpenReadonly(viper.GetBool(openReadonlyOption)))
			if err != nil {
				return fmt.Errorf("error opening database: %w", err)
			}
			defer db.Close()

			stats, err := db.Stats()
			if err != nil {
				return fmt.Errorf("error getting stats: %w", err)
			}

			data, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				return fmt.Errorf("error serializing stats: %w", err)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s", string(data)); err != nil {
				return fmt.Errorf("error writing stats: %w", err)
			}

			return nil
		},
	}

	return cmd
}
