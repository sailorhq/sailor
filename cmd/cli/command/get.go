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
	"slices"

	"github.com/sailorhq/sailor/pkg/vault"
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
			if ns != "" && app != "" && kind != "" {
				if kind == "misc" && name == "" {
					return errors.New("misc resource requires a name")
				}

				schema, err := cfg.SailorClient.GetSchema(ns, app, kind, name, cfg.Token)
				if err != nil {
					return err
				}

				b, err := json.MarshalIndent(schema, "", "	")
				if err != nil {
					return err
				}

				if output != "" {
					if err = os.WriteFile(output, b, 0755); err != nil {
						return err
					}

					fmt.Println("saved to:", output)
				}

				fmt.Println(string(b))
				return nil
			}

			fmt.Println("missing ns, app, kind or name")
			return nil
		},
	}

	setting := &cobra.Command{
		Use:   "setting",
		Short: "Helps you fetch resource or sailor setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" && kind != "" {
				if kind == "misc" && name == "" {
					return errors.New("misc resource requires a name")
				}

				rs, err := cfg.SailorClient.GetResourceSetting(ns, app, kind, name, cfg.Token)
				if err != nil {
					return err
				}

				b, err := json.MarshalIndent(&rs, "", "    ")
				if err != nil {
					return nil
				}

				if output != "" {
					if err := os.WriteFile(output, b, 0755); err != nil {
						return err
					}
					fmt.Println("saved to", output)
					return nil
				}

				fmt.Println(string(b))
				return nil
			}

			// get sailor settings
			ss, err := cfg.SailorClient.GetSailorSetting(cfg.Token)
			if err != nil {
				return err
			}

			b, err := json.MarshalIndent(&ss, "", "    ")
			if err != nil {
				return nil
			}

			if output != "" {
				if err := os.WriteFile(output, b, 0755); err != nil {
					return err
				}
				fmt.Println("saved to", output)
				return nil
			}

			fmt.Println(string(b))
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

				projectKey := fmt.Sprintf("%s-%s", ns, app)
				if _, ok := cfg.KeyPairs[projectKey]; !ok {
					return errors.New("access not available, re-login to get fresh access")
				}

				dataBytes, err := cfg.SailorClient.GetResource(ns, app, kind, name, cfg.Token, cfg.KeyPairs[projectKey])
				if err != nil {
					return err
				}

				if kind == "secret" {
					var encSecrets map[string]vault.SecretRecord
					if err := json.Unmarshal(dataBytes, &encSecrets); err != nil {
						return err
					}

					kek, err := vault.DeriveKEK(cfg.KeyPairs[projectKey].SecretKey, []byte(cfg.KeyPairs[projectKey].AccessKey))
					if err != nil {
						return err
					}

					var secrets = make(map[string]string, len(encSecrets))
					for k, ev := range encSecrets {
						dek, err := vault.DecryptDEK(ev.EncryptedDEK, kek)
						if err != nil {
							return err
						}
						v, err := vault.DecryptWithDEK(ev.EncryptedSecret, dek)
						if err != nil {
							return err
						}

						secrets[k] = v
					}

					dataBytes, err = json.Marshal(&secrets)
					if err != nil {
						return err
					}
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
