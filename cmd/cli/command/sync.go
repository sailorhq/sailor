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
	"path/filepath"
	"strings"

	"github.com/sailorhq/sailor/internal/types"
	"github.com/spf13/cobra"
)

func SyncCommand(cfg *CLIConfig) *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Helps you sync sailor resources for this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}

			sailorFilePath := filepath.Join(pwd, "sailor.json")
			if f, _ := os.Stat(sailorFilePath); f == nil {
				return errors.New("sailor file not found, do `sailor init --namespace [NS] --app [APP]`")
			}

			sfbytes, err := os.ReadFile(sailorFilePath)
			if err != nil {
				return err
			}

			var sf types.SailorFile
			err = json.Unmarshal(sfbytes, &sf)
			if err != nil {
				return err
			}

			projectKey := fmt.Sprintf("%s-%s", sf.Project.Namespace, sf.Project.App)

			if err := syncResources(&sf, cfg, projectKey); err != nil {
				return err
			}

			return nil
		},
	}

	return syncCmd
}

func syncResources(sf *types.SailorFile, cfg *CLIConfig, projectKey string) error {
	if sf.Config.File != "" {
		resData, err := cfg.SailorClient.GetResource(sf.Project.Namespace,
			sf.Project.App, "config", "", cfg.Token, cfg.KeyPairs[projectKey])
		if err != nil {
			return err
		}

		replacedFilePath := strings.ReplaceAll(sf.Config.File, "$env", cfg.Env)
		if err = os.WriteFile(replacedFilePath, resData.Data, 0644); err != nil {
			return err
		}

		fmt.Println("synced config file:", replacedFilePath)
		fmt.Println("version:", resData.Version)
	} else {
		fmt.Println("config file path not specified, will not be synced...")
	}

	// secret file sync
	if sf.Secret.File != "" {
		resData, err := cfg.SailorClient.GetResource(sf.Project.Namespace,
			sf.Project.App, "secret", "", cfg.Token, cfg.KeyPairs[projectKey])
		if err != nil {
			return err
		}
		replacedFilePath := strings.ReplaceAll(sf.Secret.File, "$env", cfg.Env)
		if err := os.WriteFile(replacedFilePath, resData.Data, 0644); err != nil {
			return err
		}
		fmt.Println("synced secret file:", replacedFilePath)
		fmt.Println("version:", resData.Version)
	} else {
		fmt.Println("secret file path not specified, will not be synced...")
	}

	// misc file sync
	if len(sf.Misc) > 0 {
		for _, misc := range sf.Misc {
			resData, err := cfg.SailorClient.GetResource(sf.Project.Namespace,
				sf.Project.App, "misc", misc.Name, cfg.Token, cfg.KeyPairs[projectKey])
			if err != nil {
				return err
			}
			replacedFilePath := strings.ReplaceAll(misc.File, "$env", cfg.Env)
			if err := os.WriteFile(replacedFilePath, resData.Data, 0644); err != nil {
				return err
			}
			fmt.Println("synced misc file:", replacedFilePath)
			fmt.Println("version:", resData.Version)
		}
	} else {
		fmt.Println("misc file paths not specified, will not be synced...")
	}

	return nil
}
