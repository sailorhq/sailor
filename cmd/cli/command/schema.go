package command

import "github.com/spf13/cobra"

func SchemaCommand() *cobra.Command {
	var file string
	var input string
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Helps you define a schema for a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	schemaCmd.Flags().StringVarP(&file, "file", "f", "", "update schema from json file")
	schemaCmd.Flags().StringVarP(&input, "input", "i", "", "update schema from input string")

	return schemaCmd
}
