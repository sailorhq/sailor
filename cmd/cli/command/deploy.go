package command

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func DeployCommand(cfg *CLIConfig) *cobra.Command {
	var version string
	var ns string
	var app string
	var kind string
	var name string
	var file string
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

	deployCmd.PersistentFlags().StringVarP(&file, "file", "", "", "resource data file")
	deployCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "resource from this namespace")
	deployCmd.PersistentFlags().StringVarP(&app, "app", "", "", "resource for this app")
	deployCmd.PersistentFlags().StringVarP(&kind, "kind", "", "", "kind of resource (config,secret,misc)")
	deployCmd.PersistentFlags().StringVarP(&name, "name", "", "", "name of the misc resource")

	create := &cobra.Command{
		Use:   "create",
		Short: "Helps you create a deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return errors.New("resource data file not provided")
			}

			if ns != "" && app != "" && kind != "" {
				if !slices.Contains([]string{"config", "misc", "secret"}, kind) {
					return errors.New("not a valid kind of resource")
				}

				if kind == "misc" && name == "" {
					return errors.New("resource name is required for misc kind")
				}

				reader := bufio.NewReader(os.Stdin)
				fmt.Print("description: ")
				desc, _ := reader.ReadString('\n')
				if desc == "" {
					return errors.New("deployment description is required")
				}

				// -- validations over --

				var data any
				b, err := os.ReadFile(file)
				if err != nil {
					return err
				}

				switch kind {
				case "config":
					var d map[string]any
					if err := json.Unmarshal(b, &d); err != nil {
						return err
					}
					data = d
				case "secret":
					var d map[string]string
					if err := json.Unmarshal(b, &d); err != nil {
						return err
					}
					data = d
				case "misc":
					data = string(b)
				}

				version, err := cfg.SailorClient.CreateDeployment(ns, app, kind, name, cfg.Token, strings.TrimSpace(desc), data)
				if err != nil {
					return err
				}

				fmt.Printf("Deployment: %v created!\n", *version)

				return nil
			}

			fmt.Println("missing ns, app or kind")

			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "Helps you create list sailor deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			if version == "" {
				version = "all"
			}

			if ns != "" && app != "" && version != "" && kind != "" {
				if !slices.Contains([]string{"config", "misc", "secret"}, kind) {
					return errors.New("not a valid kind of resource")
				}

				if kind == "misc" && name == "" {
					return errors.New("resource name is required for misc kind")
				}

				depResp, err := cfg.SailorClient.GetDeployment(ns, app, kind, name, cfg.Token, version)
				if err != nil {
					return err
				}

				if len(depResp.Deployments) == 0 {
					fmt.Println("no deployments in this project!")
					return nil
				}

				fmt.Printf("\n\nDeployments:\n\n")
				for _, dep := range depResp.Deployments {
					fmt.Println("Version:", dep.Version)
					fmt.Println("Date:", dep.CreatedAt)
					fmt.Println("	", dep.Description)
				}

				return nil
			}

			fmt.Println("missing ns, app, kind or name")

			return nil
		},
	}

	apply := &cobra.Command{
		Use:   "apply",
		Short: "Helps you create apply sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" && version != "" && kind != "" {
				if err := cfg.SailorClient.Deploy(ns, app, kind, name, cfg.Token, version); err != nil {
					return err
				}

				fmt.Printf("Deployed: ver.%s\n", version)

				return nil
			}

			fmt.Println("missing ns, app, version or kind")

			return nil
		},
	}
	apply.Flags().StringVarP(&version, "version", "v", "", "deploy this specific version")

	deployCmd.AddCommand(create)
	deployCmd.AddCommand(list)
	deployCmd.AddCommand(apply)
	return deployCmd
}
