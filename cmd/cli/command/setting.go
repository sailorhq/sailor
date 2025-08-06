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

import "github.com/spf13/cobra"

func SettingCommand() *cobra.Command {
	var input string
	var file string
	var ns string
	var app string
	settingCmd := &cobra.Command{
		Use:   "setting",
		Short: "Helps you update sailor or resource setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	settingCmd.Flags().StringVarP(&ns, "namespace", "", "", "create a resource in this namespace")
	settingCmd.Flags().StringVarP(&app, "app", "", "", "create a resource for this app")
	settingCmd.Flags().StringVarP(&file, "file", "f", "", "update setting from json file")
	settingCmd.Flags().StringVarP(&input, "input", "i", "", "update setting from input string")

	return settingCmd
}
