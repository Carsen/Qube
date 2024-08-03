package main

import (
	"github.com/spf13/cobra"

	"go.mills.io/bitcask/v2"
)

const (
	maxFileSizeOption  = "max-file-size"
	maxKeySizeOption   = "max-key-size"
	maxValueSizeOption = "max-value-size"
)

func addBitcaskFlagOptions(cmd *cobra.Command) {
	cmd.Flags().Uint32P(
		maxKeySizeOption, "k", bitcask.DefaultMaxKeySize,
		"Maximum size of each key",
	)
	cmd.Flags().IntP(
		maxFileSizeOption, "f", bitcask.DefaultMaxDatafileSize,
		"Maximum size of each datafile",
	)
	cmd.Flags().Uint64P(
		maxValueSizeOption, "v", bitcask.DefaultMaxValueSize,
		"Maximum size of each value",
	)
}

func getBitcaskOptionsFromFlags(cmd *cobra.Command) []bitcask.Option {
	options := []bitcask.Option{}

	if val, err := cmd.Flags().GetUint64(maxFileSizeOption); err == nil {
		options = append(options, bitcask.WithMaxDatafileSize(val))
	}

	if val, err := cmd.Flags().GetUint32(maxKeySizeOption); err == nil {
		options = append(options, bitcask.WithMaxKeySize(val))
	}

	if val, err := cmd.Flags().GetBool(openReadonlyOption); err == nil {
		options = append(options, bitcask.WithOpenReadonly(val))
	}

	if val, err := cmd.Flags().GetUint64(maxValueSizeOption); err == nil {
		options = append(options, bitcask.WithMaxValueSize(val))
	}

	return options
}
