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
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func SchemaCommand(cfg *CLIConfig) *cobra.Command {
	var ns string
	var app string
	var kind string
	var name string
	var file string
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Helps you define a schema for a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	schemaCmd.PersistentFlags().StringVarP(&ns, "namespace", "", cfg.CwdSailorFile.Project.Namespace, "get resource from this namespace")
	schemaCmd.PersistentFlags().StringVarP(&app, "app", "", cfg.CwdSailorFile.Project.App, "get resource from this app")
	schemaCmd.PersistentFlags().StringVarP(&kind, "kind", "", "", "kind of resource")
	schemaCmd.PersistentFlags().StringVarP(&name, "name", "", "", "name of the misc resource")
	schemaCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "update schema from json file")

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Helps apply an update to resource schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" && kind != "" {
				if kind == "misc" && name == "" {
					return errors.New("misc resource requires a name")
				}

				if file != "" {
					b, err := os.ReadFile(file)
					if err != nil {
						return nil
					}
					var schema map[string]any

					if err := json.Unmarshal(b, &schema); err != nil {
						return errors.New("schema should be a valid json and value for each key must be a string")
					}
					if err := cfg.SailorClient.UpdateSchema(schema, ns, app, kind, name, cfg.Token); err != nil {
						return err
					}
					fmt.Println("updated resource schema!")
				}

				switch kind {
				case "config":
					if cfg.CwdSailorFile.Config.Schema == nil {
						return errors.New("config.schema is required in sailor file")
					}
					if err := cfg.SailorClient.UpdateSchema(*cfg.CwdSailorFile.Config.Schema, ns, app, kind, name, cfg.Token); err != nil {
						return err
					}
				case "secret":
					if cfg.CwdSailorFile.Secret.Schema == nil {
						return errors.New("secret.schema is required in sailor file")
					}
					if err := cfg.SailorClient.UpdateSchema(*cfg.CwdSailorFile.Secret.Schema, ns, app, kind, name, cfg.Token); err != nil {
						return err
					}
				case "misc":
					if rfile, ok := cfg.CwdSailorFile.Misc[name]; ok && rfile.Schema == nil {
						if err := cfg.SailorClient.UpdateSchema(*rfile.Schema, ns, app, kind, name, cfg.Token); err != nil {
							return err
						}
					}
					return errors.New("misc.schema is required in sailor file")
				default:
					return errors.New("unknown kind")
				}

				fmt.Println("updated resource schema!")
				return nil
			}

			fmt.Println("missing ns, app, kind or name")

			return nil
		},
	}

	schemaCmd.AddCommand(applyCmd)

	return schemaCmd
}
