package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.mills.io/bitcask/v2"
	"go.mills.io/bitcask/v2/internal"
)

const (
	pathOption         = "path"
	debugOption        = "debug"
	openReadonlyOption = "open-read-only"
)

// RootCmd represents the base command when called without any subcommands
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bitcask",
		Version: internal.FullVersion(),
		Short:   "Command-line tools for bitcask",
		Long: `This is the command-line tool to interact with a bitcask database.

This lets you get, set and delete key/value pairs as well as perform merge
(or compaction) operations. This tool serves as an example implementation
however is also intended to be useful in shell scripts.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// set logging level
			if viper.GetBool(debugOption) {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.InfoLevel)
			}
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolP(
		debugOption, "d", false,
		"Enable debug logging",
	)

	cmd.PersistentFlags().StringP(
		pathOption, "p", "/tmp/bitcask",
		"Path to Bitcask database",
	)

	cmd.PersistentFlags().BoolP(
		openReadonlyOption, "r", bitcask.DefaultOpenReadonly,
		"Open database in readonly mode",
	)

	if err := viper.BindPFlag(debugOption, cmd.PersistentFlags().Lookup(debugOption)); err != nil {
		log.Fatalf("error binding %s flag", debugOption)
	}
	viper.SetDefault(debugOption, bitcask.DefaultDebug)

	if err := viper.BindPFlag(pathOption, cmd.PersistentFlags().Lookup(pathOption)); err != nil {
		log.Fatalf("error binding %s flag", pathOption)
	}
	viper.SetDefault(pathOption, bitcask.DefaultPath)

	if err := viper.BindPFlag(openReadonlyOption, cmd.PersistentFlags().Lookup(openReadonlyOption)); err != nil {
		log.Fatalf("error binding %s flag", openReadonlyOption)
	}
	viper.SetDefault(openReadonlyOption, bitcask.DefaultOpenReadonly)

	return cmd
}

// Execute adds all child commands to the root command
// and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("bitcask")
		viper.AutomaticEnv()
	})

	root := RootCmd()

	// Sub-commands
	root.AddCommand(DelCmd())
	root.AddCommand(ExportCmd())
	root.AddCommand(GetCmd())
	root.AddCommand(ImportCmd())
	root.AddCommand(InitCmd())
	root.AddCommand(KeysCmd())
	root.AddCommand(MergeCmd())
	root.AddCommand(PutCmd())
	root.AddCommand(RangeCmd())
	root.AddCommand(RecoverCmd())
	root.AddCommand(ScanCmd())
	root.AddCommand(StatsCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(root.ErrOrStderr(), "%s\n", err)
		os.Exit(1)
	}
}
