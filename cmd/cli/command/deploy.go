package command

import (
	"errors"

	"github.com/spf13/cobra"
)

func DeployCommand() *cobra.Command {
	var version string
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Helps you create and deploy a sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("flags or env as argument is required")
			}

			return nil
		},
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Helps you create a deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "Helps you create list sailor deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	apply := &cobra.Command{
		Use:   "apply",
		Short: "Helps you create apply sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	apply.Flags().StringVarP(&version, "version", "v", "", "deploy this specific version")

	deployCmd.AddCommand(create)
	deployCmd.AddCommand(list)
	deployCmd.AddCommand(apply)
	return deployCmd
}
