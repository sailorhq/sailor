package command

import "github.com/spf13/cobra"

func GetCommand() *cobra.Command {
	var ns string
	var app string
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Helps fetching resource, schema or setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	getCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "create a resource in this namespace")
	getCmd.PersistentFlags().StringVarP(&app, "app", "", "", "create a resource for this app")

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

	getCmd.AddCommand(schema)
	getCmd.AddCommand(setting)
	return getCmd
}
