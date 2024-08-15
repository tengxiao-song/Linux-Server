package main

import (
	"os"

	"main/command"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "./main",
		Short: "Linux Job Dispatcher Service application",
	}

	// Add subcommands
	rootCmd.AddCommand(command.ListCommand())
	rootCmd.AddCommand(command.QueryCommand())
	rootCmd.AddCommand(command.StopCommand())
	rootCmd.AddCommand(command.StartCommand())

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
