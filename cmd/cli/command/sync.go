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
	"crypto/md5"
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
	type SyncResource struct {
		Path string
		Name string
		Kind string
	}

	var envResourceVersion types.ResourceVersion
	if rv, ok := cfg.CwdSailorLockFile.Environments[cfg.Env]; ok {
		envResourceVersion = rv
	}

	resourcesToSync := []SyncResource{}
	// TODO :: we need to do md5 checksum check
	var noticedAnUpdate = false
	if sf.Config.File != "" {
		resourcesToSync = append(resourcesToSync, SyncResource{
			Path: sf.Config.File,
			Name: "",
			Kind: "config",
		})
	} else {
		fmt.Println("config file path not specified, will not be synced...")
	}

	// secret file sync
	if sf.Secret.File != "" {
		resourcesToSync = append(resourcesToSync, SyncResource{
			Path: sf.Secret.File,
			Name: "",
			Kind: "secret",
		})
	} else {
		fmt.Println("secret file path not specified, will not be synced...")
	}

	// misc file sync
	if len(sf.Misc) > 0 {
		for name, misc := range sf.Misc {
			resourcesToSync = append(resourcesToSync, SyncResource{
				Path: misc.File,
				Name: name,
				Kind: "misc",
			})
		}
	} else {
		fmt.Println("misc file paths not specified, will not be synced...")
	}

	for _, res := range resourcesToSync {
		resData, err := cfg.SailorClient.GetResource(sf.Project.Namespace,
			sf.Project.App, res.Kind, res.Name, cfg.Token, cfg.KeyPairs[projectKey])
		if err != nil {
			return err
		}

		replacedFilePath := strings.ReplaceAll(res.Path, "$env", cfg.Env)
		if err = os.WriteFile(replacedFilePath, resData.Data, 0644); err != nil {
			return err
		}

		fmt.Println("synced", res.Kind, "file:", replacedFilePath)
		fmt.Println("version:", resData.Version)

		switch res.Kind {
		case "config":
			if envResourceVersion.Config == nil {
				envResourceVersion.Config = &types.LockVersion{
					Version: resData.Version,
					Hash:    fmt.Sprintf("%x", md5.Sum(resData.Data)),
				}
				noticedAnUpdate = true
			} else if envResourceVersion.Config.Version != resData.Version {
				envResourceVersion.Config.Version = resData.Version
				envResourceVersion.Config.Hash = fmt.Sprintf("%x", md5.Sum(resData.Data))
				noticedAnUpdate = true
			}
		case "secret":
			if envResourceVersion.Secret == nil {
				envResourceVersion.Secret = &types.LockVersion{
					Version: resData.Version,
					Hash:    fmt.Sprintf("%x", md5.Sum(resData.Data)),
				}
				noticedAnUpdate = true
			} else if envResourceVersion.Secret.Version != resData.Version {
				envResourceVersion.Secret.Version = resData.Version
				envResourceVersion.Secret.Hash = fmt.Sprintf("%x", md5.Sum(resData.Data))
				noticedAnUpdate = true
			}
		case "misc":
			if len(envResourceVersion.Misc) == 0 {
				envResourceVersion.Misc = make(map[string]types.LockVersion)
				envResourceVersion.Misc[res.Name] = types.LockVersion{
					Version: resData.Version,
					Hash:    fmt.Sprintf("%x", md5.Sum(resData.Data)),
				}
				noticedAnUpdate = true
			} else if envResourceVersion.Misc[res.Name].Version != resData.Version {
				envResourceVersion.Misc[res.Name] = types.LockVersion{
					Version: resData.Version,
					Hash:    fmt.Sprintf("%x", md5.Sum(resData.Data)),
				}
				noticedAnUpdate = true
			}
		}
	}

	if noticedAnUpdate {
		// update the lock file
		cfg.CwdSailorLockFile.Environments[cfg.Env] = envResourceVersion

		lockbytes, err := json.MarshalIndent(cfg.CwdSailorLockFile, "", "  ")
		if err != nil {
			return err
		}

		return os.WriteFile("./sailor.lock", lockbytes, 0644)
	}

	return nil
}
