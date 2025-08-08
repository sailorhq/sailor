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

	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

func SettingCommand(cfg *CLIConfig) *cobra.Command {
	var file string
	var ns string
	var app string
	var kind string
	var name string
	settingCmd := &cobra.Command{
		Use:   "setting",
		Short: "Helps you update sailor or resource setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return errors.New("require setting file as input")
			}
			return nil
		},
	}

	settingCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "create a resource in this namespace")
	settingCmd.PersistentFlags().StringVarP(&app, "app", "", "", "create a resource for this app")
	settingCmd.PersistentFlags().StringVarP(&kind, "kind", "", "", "kind of this resource")
	settingCmd.PersistentFlags().StringVarP(&name, "name", "", "", "name of this misc resource")
	settingCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "update setting from json file")

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Helps apply an update to sailor or resource setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			// we will update resource setting in this case
			if ns != "" && app != "" && kind != "" {
				if kind == "misc" && name == "" {
					return errors.New("misc resource requires a name")
				}
				return nil
			}

			// we will update sailor settings here
			b, err := os.ReadFile(file)
			if err != nil {
				return nil
			}
			var ss v1.SailorSetting
			if err := json.Unmarshal(b, &ss); err != nil {
				return err
			}

			if err := cfg.SailorClient.UpdateSailorSetting(ss, cfg.Token); err != nil {
				return err
			}

			fmt.Println("updated sailor settings!")

			return nil
		},
	}

	settingCmd.AddCommand(applyCmd)

	return settingCmd
}
