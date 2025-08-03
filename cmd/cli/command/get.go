package command

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

func GetCommand(cfg *CLIConfig) *cobra.Command {
	var ns string
	var app string
	var kind string
	var name string
	var output string
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Helps fetching resource, schema or setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	getCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "get resource from this namespace")
	getCmd.PersistentFlags().StringVarP(&app, "app", "", "", "get resource from this app")
	getCmd.PersistentFlags().StringVarP(&kind, "kind", "", "", "kind of resource")
	getCmd.PersistentFlags().StringVarP(&name, "name", "", "", "name of the misc resource")
	getCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output the resource data into a file")

	schema := &cobra.Command{
		Use:   "schema",
		Short: "Helps you fetch schema for a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	setting := &cobra.Command{
		Use:   "setting",
		Short: "Helps you fetch setting for a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	resource := &cobra.Command{
		Use:   "resource",
		Short: "Helps you fetch a specific resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" && kind != "" {
				if !slices.Contains([]string{"config", "misc", "secret"}, kind) {
					return errors.New("not a valid kind of resource")
				}

				if kind == "misc" && name == "" {
					return errors.New("resource name is required for misc kind")
				}

				dataBytes, err := cfg.SailorClient.GetResource(ns, app, kind, name, cfg.Token)
				if err != nil {
					return err
				}

				if output != "" {
					fmt.Println("resource written to:", output)
					return os.WriteFile(output, dataBytes, 0755)
				}

				fmt.Println(string(dataBytes))
				return nil
			}

			fmt.Println("missing ns, app, kind or name")
			return nil
		},
	}

	getCmd.AddCommand(schema)
	getCmd.AddCommand(setting)
	getCmd.AddCommand(resource)
	return getCmd
}
