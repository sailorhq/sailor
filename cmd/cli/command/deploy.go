// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
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

// deployResourceCommand creates a command for deploying a specific resource type
func deployResourceCommand(cfg *CLIConfig, kind string, ns, app, name, file *string) *cobra.Command {
	return &cobra.Command{
		Use:   kind,
		Short: "Helps you deploy a sailor " + kind,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployResource(cfg, kind, *ns, *app, *name, *file)
		},
	}
}

func DeployCommand(cfg *CLIConfig) *cobra.Command {
	var version string
	var ns string
	var app string
	var name string
	var file string
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Helps you create and deploy a sailor resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return ErrKindArgumentMissing
			}

			return nil
		},
	}

	deployCmd.PersistentFlags().StringVarP(&file, "file", "", "", "resource data file")
	deployCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "resource from this namespace")
	deployCmd.PersistentFlags().StringVarP(&app, "app", "", "", "resource for this app")
	deployCmd.PersistentFlags().StringVarP(&name, "name", "", "", "name of the misc resource")

	// Config subcommand
	config := deployResourceCommand(cfg, "config", &ns, &app, &name, &file)

	// Secret subcommand
	secret := deployResourceCommand(cfg, "secret", &ns, &app, &name, &file)

	// Misc subcommand
	misc := deployResourceCommand(cfg, "misc", &ns, &app, &name, &file)

	create := &cobra.Command{
		Use:   "create",
		Short: "Helps you create a deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return errors.New("resource data file not provided")
			}

			if ns != "" && app != "" {
				// Get kind from args
				if len(args) == 0 {
					return ErrKindArgumentMissing
				}

				kind := args[0]
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

			if ns != "" && app != "" && version != "" {
				// Get kind from args
				if len(args) == 0 {
					return ErrKindArgumentMissing
				}

				kind := args[0]
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
			if ns != "" && app != "" && version != "" {
				// Get kind from args
				if len(args) == 0 {
					return ErrKindArgumentMissing
				}

				kind := args[0]
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

	deployCmd.AddCommand(config)
	deployCmd.AddCommand(secret)
	deployCmd.AddCommand(misc)
	deployCmd.AddCommand(create)
	deployCmd.AddCommand(list)
	deployCmd.AddCommand(apply)
	return deployCmd
}

// Helper function for deploying resources
func deployResource(cfg *CLIConfig, kind, ns, app, name, file string) error {
	if file == "" {
		return errors.New("resource data file not provided")
	}

	if ns == "" || app == "" {
		return errors.New("namespace and app are required")
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
