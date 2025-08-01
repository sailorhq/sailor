package command

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func CreateCommand(cfg *CLIConfig) *cobra.Command {
	var setting string
	var ns string
	var app string
	var def bool
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Helps you create a sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("flags or env as argument is required")
			}

			return nil
		},
	}

	createCmd.PersistentFlags().BoolVarP(&def, "default", "", false, "use default setting for this resource")
	createCmd.PersistentFlags().StringVarP(&setting, "setting", "s", "", "resource settings")
	createCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "create a resource in this namespace")
	createCmd.PersistentFlags().StringVarP(&app, "app", "", "", "create a resource for this app")

	config := &cobra.Command{
		Use:   "config",
		Short: "Helps you create config sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	secret := &cobra.Command{
		Use:   "secret",
		Short: "Helps you create secret sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	misc := &cobra.Command{
		Use:   "misc",
		Short: "Helps you create misc sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	project := &cobra.Command{
		Use:   "project",
		Short: "Helps you create a sailor project | admin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	user := &cobra.Command{
		Use:   "user",
		Short: "Helps you create a sailor user | admin",
		RunE: func(cmd *cobra.Command, args []string) error {
			var user string
			fmt.Print("sailor user: ")
			fmt.Scan(&user)

			cur, err := cfg.SailorClient.CreateUser(user, cfg.Token)
			if err != nil {
				return err
			}

			fmt.Println(cur.Note)
			fmt.Println("password: ", cur.Pass)
			return nil
		},
	}

	createCmd.AddCommand(config)
	createCmd.AddCommand(secret)
	createCmd.AddCommand(misc)
	createCmd.AddCommand(project)
	createCmd.AddCommand(user)
	return createCmd
}
