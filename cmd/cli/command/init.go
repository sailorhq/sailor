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
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/sailorhq/sailor/internal/types"
	"github.com/spf13/cobra"
)

func InitCommand(cfg *CLIConfig) *cobra.Command {
	var ns string
	var app string

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Helps you init a sailor project",
		RunE: func(cmd *cobra.Command, args []string) error {
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}

			sailorFilePath := filepath.Join(pwd, "sailor.json")
			if f, _ := os.Stat(sailorFilePath); f != nil {
				return errors.New("sailor file is already present")
			}

			sf := types.SailorFile{
				Project: types.ProjectDetails{
					Namespace: ns,
					App:       app,
				},
			}

			sfbytes, err := json.MarshalIndent(sf, "", "    ")
			if err != nil {
				return err
			}

			// incase we have gitignore file here read and update gitignore with `sailor-lock.json`
			if f, _ := os.Stat(".gitignore"); f != nil {
				gitignore, err := os.ReadFile(".gitignore")
				if err != nil {
					return err
				}
				if !bytes.Contains(gitignore, []byte("sailor-lock.json")) {
					gitignore = append(gitignore, []byte("\n\n# sailor version tracking file\nsailor-lock.json")...)
				}

				// update gitignore file
				if err := os.WriteFile(".gitignore", gitignore, 0644); err != nil {
					return err
				}
			}

			return os.WriteFile("sailor.json", sfbytes, 0755)
		},
	}

	initCmd.PersistentFlags().StringVarP(&ns, "namespace", "", "", "get resource from this namespace")
	initCmd.PersistentFlags().StringVarP(&app, "app", "", "", "get resource from this app")
	initCmd.MarkPersistentFlagRequired("namespace")
	initCmd.MarkPersistentFlagRequired("app")

	return initCmd
}
