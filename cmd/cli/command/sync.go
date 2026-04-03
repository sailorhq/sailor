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
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

func SyncCommand(cfg *CLIConfig) *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Helps you sync sailor resources for this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.CwdSailorFile.Project.App == "" {
				return errors.New("sailor file not found or invalid, do `sailor init --namespace [NS] --app [APP]`")
			}

			if err := syncResources(&cfg.CwdSailorFile, cfg); err != nil {
				return err
			}

			return nil
		},
	}

	return syncCmd
}

type SyncResource struct {
	Path string
	Name string
	Kind string
}

// syncResources orchestrates the synchronization of all resources defined in the sailor.json file.
// It iterates through Config, Secret, and Misc resources, fetches their latest data from the 
// server, and delegates the per-resource sync logic to syncResource. If any changes are 
// detected (files updated or lock metadata changed), it updates the local sailor.lock file.
func syncResources(sf *types.SailorFile, cfg *CLIConfig) error {
	var envResourceVersion types.ResourceVersion
	if rv, ok := cfg.CwdSailorLockFile.Environments[cfg.Env]; ok {
		envResourceVersion = rv
	}

	if envResourceVersion.Misc == nil {
		envResourceVersion.Misc = make(map[string]types.LockVersion)
	}

	resourcesToSync := []SyncResource{}
	if sf.Config.File != "" {
		resourcesToSync = append(resourcesToSync, SyncResource{
			Path: sf.Config.File,
			Name: "",
			Kind: "config",
		})
	}

	if sf.Secret.File != "" {
		resourcesToSync = append(resourcesToSync, SyncResource{
			Path: sf.Secret.File,
			Name: "",
			Kind: "secret",
		})
	}

	if len(sf.Misc) > 0 {
		for name, misc := range sf.Misc {
			resourcesToSync = append(resourcesToSync, SyncResource{
				Path: misc.File,
				Name: name,
				Kind: "misc",
			})
		}
	}

	var noticedAnUpdate = false
	for _, res := range resourcesToSync {
		resData, err := cfg.SailorClient.GetResource(sf.Project.Namespace,
			sf.Project.App, res.Kind, res.Name)
		if err != nil {
			fmt.Printf("Error fetching %s (%s): %v\n", res.Kind, res.Name, err)
			continue
		}

		updated, err := syncResource(cfg, res, resData, &envResourceVersion)
		if err != nil {
			return err
		}
		if updated {
			noticedAnUpdate = true
		}
	}

	if noticedAnUpdate {
		if cfg.CwdSailorLockFile.Environments == nil {
			cfg.CwdSailorLockFile.Environments = make(map[string]types.ResourceVersion)
		}
		cfg.CwdSailorLockFile.Environments[cfg.Env] = envResourceVersion
		lockbytes, err := json.MarshalIndent(cfg.CwdSailorLockFile, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile("./sailor.lock", lockbytes, 0644)
	}

	return nil
}

// syncResource handles the synchronization logic for a single resource. It performs 
// state comparison between local file, remote data, and lock file to decide 
// whether to overwrite the local file and whether the lock file needs an update.
func syncResource(cfg *CLIConfig, res SyncResource, resData *v1.ResourceData, envResourceVersion *types.ResourceVersion) (bool, error) {
	replacedFilePath := strings.ReplaceAll(res.Path, "$env", cfg.Env)
	remoteHash := fmt.Sprintf("%x", md5.Sum(resData.Data))

	// Get lock info
	var lockHash string
	switch res.Kind {
	case "config":
		if envResourceVersion.Config != nil {
			lockHash = envResourceVersion.Config.Hash
		}
	case "secret":
		if envResourceVersion.Secret != nil {
			lockHash = envResourceVersion.Secret.Hash
		}
	case "misc":
		if lv, ok := envResourceVersion.Misc[res.Name]; ok {
			lockHash = lv.Hash
		}
	}

	// Check local file
	localFileExists := true
	localBytes, err := os.ReadFile(replacedFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			localFileExists = false
		} else {
			return false, err
		}
	}

	var localHash string
	if localFileExists {
		localHash = fmt.Sprintf("%x", md5.Sum(localBytes))
	}

	shouldOverwrite := false
	noticedAnUpdate := false

	// The synchronization logic follows a "safe-by-default" approach.
	// We compare three states: the local file, the remote resource, and the last known 
	// state recorded in the sailor.lock file (lockHash).
	if !localFileExists {
		// If the file doesn't exist locally, it's always safe to download.
		shouldOverwrite = true
		fmt.Printf("File %s does not exist, downloading...\n", replacedFilePath)
	} else if localHash == remoteHash {
		// If local already matches remote, no file update is needed.
		// However, if the lock file is missing or has a different hash, 
		// we mark it for update to ensure consistency.
		if lockHash != remoteHash {
			noticedAnUpdate = true
		}
	} else {
		// Local and remote states have diverged. 
		// We use the lockHash to determine the nature of the divergence.
		if lockHash != "" {
			// If localHash != lockHash, the user has made local changes that 
			// haven't been synced or deployed.
			if localHash != lockHash {
				fmt.Printf("Warning: local changes in %s have not been deployed\n", replacedFilePath)
			}
			// If remoteHash != lockHash, someone else (or another process) 
			// has updated the resource upstream.
			if remoteHash != lockHash {
				fmt.Printf("Warning: upstream changes detected for %s, please pull latest and apply changes on top\n", replacedFilePath)
			}
		}

		// Decision: Should we overwrite the local file?
		// Overwriting is considered safe if:
		// 1. There is no local file.
		// 2. The local file matches the last known state (no local changes).
		// 3. There is no record of a previous state (fresh sync).
		if !localFileExists || localHash == lockHash || lockHash == "" {
			shouldOverwrite = true
			fmt.Printf("Updating %s to latest version...\n", replacedFilePath)
		} else {
			// If there are local changes (localHash != lockHash), we skip 
			// the overwrite to prevent data loss, requiring manual intervention.
			fmt.Printf("Skipping overwrite for %s to avoid losing local changes\n", replacedFilePath)
		}
	}

	if shouldOverwrite {
		dir := filepath.Dir(replacedFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return false, err
		}
		if err = os.WriteFile(replacedFilePath, resData.Data, 0644); err != nil {
			return false, err
		}
		noticedAnUpdate = true
		fmt.Println("synced", res.Kind, "file:", replacedFilePath)
		fmt.Println("version:", resData.Version)
	}

	// Update our in-memory lock version
	newLockVersion := types.LockVersion{
		Version: resData.Version,
		Hash:    remoteHash,
	}
	switch res.Kind {
	case "config":
		envResourceVersion.Config = &newLockVersion
	case "secret":
		envResourceVersion.Secret = &newLockVersion
	case "misc":
		envResourceVersion.Misc[res.Name] = newLockVersion
	}

	return noticedAnUpdate, nil
}
