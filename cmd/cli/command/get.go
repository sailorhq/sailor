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

// getResourceCommand creates a command for getting a specific resource type
func getResourceCommand(cfg *CLIConfig, kind string, ns, app, name, output *string) *cobra.Command {
	return &cobra.Command{
		Use:   kind,
		Short: "Helps you get a sailor " + kind,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getResource(cfg, kind, *ns, *app, *name, *output)
		},
	}
}

func GetCommand(cfg *CLIConfig) *cobra.Command {
	var ns string
	var app string
	var name string
	var output string
	var limit int
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Helps fetching resource, schema or setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	getCmd.PersistentFlags().StringVarP(&ns, "namespace", "", cfg.CwdSailorFile.Project.Namespace, "get resource from this namespace")
	getCmd.PersistentFlags().StringVarP(&app, "app", "", cfg.CwdSailorFile.Project.App, "get resource from this app")
	getCmd.PersistentFlags().StringVarP(&name, "name", "", "", "name of the misc resource")
	getCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output the resource data into a file")

	// Config subcommand
	config := getResourceCommand(cfg, "config", &ns, &app, &name, &output)

	// Secret subcommand
	secret := getResourceCommand(cfg, "secret", &ns, &app, &name, &output)

	// Misc subcommand
	misc := getResourceCommand(cfg, "misc", &ns, &app, &name, &output)

	schema := &cobra.Command{
		Use:   "schema",
		Short: "Helps you fetch schema for a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if ns != "" && app != "" {
				// Get kind from args
				if len(args) == 0 {
					return ErrKindArgumentMissing
				}

				kind := args[0]
				if kind == "misc" && name == "" {
					return errors.New("misc resource requires a name")
				}

				schema, err := cfg.SailorClient.GetSchema(ns, app, kind, name)
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
			if ns != "" && app != "" && len(args) > 0 {
				kind := args[0]
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

				return getResource(cfg, kind, ns, app, name, output)
			}

			fmt.Println("missing ns, app, kind or name")
			return nil
		},
	}

	// audit command
	audit := &cobra.Command{
		Use:   "audit",
		Short: "Helps you fetch audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := cfg.SailorClient.GetAuditLogEvents(limit)
			if err != nil {
				return err
			}

			fmt.Printf("\nEvents (n=%d):\n\n", limit)

			for _, e := range *events {
				eventjson, _ := json.Marshal(&e)
				fmt.Println(string(eventjson))
			}
			return nil
		},
	}
	audit.PersistentFlags().IntVarP(&limit, "limit", "", 10, "")

	getCmd.AddCommand(config)
	getCmd.AddCommand(secret)
	getCmd.AddCommand(misc)
	getCmd.AddCommand(schema)
	getCmd.AddCommand(setting)
	getCmd.AddCommand(resource)
	getCmd.AddCommand(audit)
	return getCmd
}

// Helper function for getting resources
func getResource(cfg *CLIConfig, kind, ns, app, name, output string) error {
	if ns == "" || app == "" {
		return errors.New("namespace and app are required")
	}

	if kind == "misc" && name == "" {
		return errors.New("resource name is required for misc kind")
	}

	resData, err := cfg.SailorClient.GetResource(ns, app, kind, name)
	if err != nil {
		return err
	}

	if kind == "secret" {
		var encSecrets map[string]vault.SecretRecord
		if err := json.Unmarshal(resData.Data, &encSecrets); err != nil {
			return err
		}

		kp, err := cfg.SailorClient.GetKeyPair(ns, app)
		if err != nil {
			return err
		}

		kek, err := vault.DeriveKEK(kp.SecretKey, []byte(kp.AccessKey))
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

		resData.Data, err = json.Marshal(&secrets)
		if err != nil {
			return err
		}
	}

	if output != "" {
		if err := os.WriteFile(output, resData.Data, 0755); err != nil {
			return err
		}
		fmt.Println("resource written to:", output)

		// After getting a resource and writing to a file, we should sync to update the lock file
		if cfg.CwdSailorFile.Project.App != "" {
			if err := syncResources(&cfg.CwdSailorFile, cfg); err != nil {
				fmt.Printf("Warning: failed to sync resources after get: %v\n", err)
			}
		}
		return nil
	}

	fmt.Println(string(resData.Data))
	return nil
}
