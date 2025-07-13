package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "Apartment",
		Short: "Apartment is a service for managing apartment listings",
	}
)

func init() {
	rootCmd.AddCommand(startCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
