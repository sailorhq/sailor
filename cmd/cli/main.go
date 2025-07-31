package main

import (
	"github.com/codekidx/sailor/cmd/cli/command"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "sailor",
		Short: "Sailor is a resource management and delivery system for cloud-native apps",
	}

	rootCmd.AddCommand(command.LoginCommand())
	rootCmd.AddCommand(command.ApplyCommand())
	rootCmd.AddCommand(command.CreateCommand())
	rootCmd.AddCommand(command.DeployCommand())
	rootCmd.AddCommand(command.SchemaCommand())
	rootCmd.AddCommand(command.SettingCommand())
	rootCmd.AddCommand(command.GetCommand())

	if err := rootCmd.Execute(); err != nil {
		panic(err.Error())
	}
}
