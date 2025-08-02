package command

import (
	"errors"
	"fmt"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

func CreateCommand(cfg *CLIConfig) *cobra.Command {
	// var setting string
	var ns string
	var app string
	// var def bool
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

	// createCmd.PersistentFlags().BoolVarP(&def, "default", "", false, "use default setting for this resource")
	// TODO :: enable this once we have vim flow done and dusted!
	// createCmd.PersistentFlags().StringVarP(&setting, "setting", "s", "", "resource settings")
	createCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "create a resource in this namespace")
	createCmd.PersistentFlags().StringVarP(&app, "app", "", "", "create a resource for this app")

	config := &cobra.Command{
		Use:   "config",
		Short: "Helps you create config sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" {
				err := cfg.SailorClient.CreateResource(ns, app, cfg.Token, "", "config", v1.ResourceSetting{
					Deploy: v1.DeploySetting{K8s: false},
					Schema: v1.SchemaSetting{Strict: false},
				})
				if err != nil {
					return err
				}

				fmt.Println("resource created!")
				return nil
			}

			fmt.Println("namespace and app are required to create a resource.")
			return nil
		},
	}

	secret := &cobra.Command{
		Use:   "secret",
		Short: "Helps you create secret sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" {
				err := cfg.SailorClient.CreateResource(ns, app, cfg.Token, "", "secret", v1.ResourceSetting{
					Deploy: v1.DeploySetting{K8s: false},
					Schema: v1.SchemaSetting{Strict: false},
				})
				if err != nil {
					return err
				}

				fmt.Println("resource created!")
				return nil
			}

			fmt.Println("namespace and app are required to create a resource.")
			return nil
		},
	}

	misc := &cobra.Command{
		Use:   "misc",
		Short: "Helps you create misc sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			var resourceName string
			fmt.Print("resource name: ")
			fmt.Scan(&resourceName)
			if resourceName == "" {
				return errors.New("misc resource name is mandatory")
			}

			if ns != "" && app != "" {
				err := cfg.SailorClient.CreateResource(ns, app, cfg.Token, resourceName, "misc", v1.ResourceSetting{
					Deploy: v1.DeploySetting{K8s: false},
					Schema: v1.SchemaSetting{Strict: false},
				})
				if err != nil {
					return err
				}

				fmt.Println("resource created!")
				return nil
			}

			fmt.Println("namespace and app are required to create a resource.")

			return nil
		},
	}

	project := &cobra.Command{
		Use:   "project",
		Short: "Helps you create a sailor project | admin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" {
				projResp, err := cfg.SailorClient.CreateProject(ns, app, cfg.Token)
				if err != nil {
					return err
				}

				fmt.Println("created project: ", projResp.Key)
				return nil
			}

			fmt.Println("missing namespace or app flags")

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
