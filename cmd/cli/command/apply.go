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
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

func ApplyCommand(cfg *CLIConfig) *cobra.Command {
	var file string
	var host string
	var namespace string
	var appName string	
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Helps configure your sailor environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if host == "" && file == "" && len(args) == 0 {
				return errors.New("flags or env as argument is required")
			}

			if host == "" && file == "" {
				env := args[0]
				for _, envConfig := range cfg.Manifest.Envs {
					if strings.EqualFold(env, envConfig.Name) {
						if err := updateEnv(envConfig.Name, cfg.SailorRoot); err != nil {
							return err
						}
                        if namespace != "" {
                            if err := updateNamespace(namespace, cfg.SailorRoot); err != nil {
                                return err
                            }
                        }
                        if appName != "" {
                            if err := updateApp(appName, cfg.SailorRoot); err != nil {
                                return err
                            }
                        }
						fmt.Printf("switched to %s", envConfig.Name)
						if namespace != "" {
                            fmt.Printf(" | namespace: %s", namespace)
                        }
                        if appName != "" {
                            fmt.Printf(" | app: %s", appName)
                        }
						fmt.Println()
						return nil
					}
				}

				fmt.Println("no environment named: ", env)

				return nil
			}

			if host != "" {
				manifestURL := fmt.Sprintf("%s/api/v1/setting/manifest", strings.TrimRight(host, "/"))
				resp, err := http.Get(manifestURL)
				if err != nil {
					return err
				}

				if resp.StatusCode != 200 {
					return fmt.Errorf("error while fetching manifest, status: %d", resp.StatusCode)
				}

				defer resp.Body.Close()
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				var manifest v1.SailorManifest
				if err := json.Unmarshal(b, &manifest); err != nil {
					return err
				}

				if err := updateManifest(&manifest, cfg.SailorRoot); err != nil {
					return err
				}
				fmt.Println("sailor manifest loaded!")
				return nil
			}

			if file != "" {
				b, err := os.ReadFile(file)
				if err != nil {
					return err
				}

				var manifest v1.SailorManifest
				if err := json.Unmarshal(b, &manifest); err != nil {
					return err
				}

				if err := updateManifest(&manifest, cfg.SailorRoot); err != nil {
					return err
				}
				fmt.Println("sailor manifest loaded!")
				return nil
			}

			return nil
		},
	}

	applyCmd.Flags().StringVarP(&host, "host", "", "", "host where sailor is hosted")
	applyCmd.Flags().StringVarP(&file, "file", "f", "", "configuration file")
	applyCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace to target")
    applyCmd.Flags().StringVarP(&appName, "app", "a", "", "application name to target")

	return applyCmd
}

func updateManifest(manifest *v1.SailorManifest, basePath string) error {
	readCfg, _ := GetConfig(basePath)
	if readCfg == nil {
		return UpdateConfig(&CLIConfig{
			Manifest: *manifest,
		}, basePath)
	}

	readCfg.Manifest = *manifest
	return UpdateConfig(readCfg, basePath)
}

func updateTokenAndKeyPairs(token string, keyPairs map[string]v1.KeyPair, basePath string) error {
	readCfg, _ := GetConfig(basePath)
	if readCfg == nil {
		return UpdateConfig(&CLIConfig{
			Token:    token,
			KeyPairs: keyPairs,
		}, basePath)
	}

	readCfg.Token = token
	readCfg.KeyPairs = keyPairs
	return UpdateConfig(readCfg, basePath)
}

func updateEnv(env string, basePath string) error {
	readCfg, _ := GetConfig(basePath)
	if readCfg == nil {
		return UpdateConfig(&CLIConfig{
			Env: env,
		}, basePath)
	}

	readCfg.Env = env
	return UpdateConfig(readCfg, basePath)
}

func updateNamespace(ns string, basePath string) error {
    readCfg, _ := GetConfig(basePath)
    if readCfg == nil {
        return UpdateConfig(&CLIConfig{
            Namespace: ns,
        }, basePath)
    }
    readCfg.Namespace = ns
    return UpdateConfig(readCfg, basePath)
}

func updateApp(app string, basePath string) error {
    readCfg, _ := GetConfig(basePath)
    if readCfg == nil {
        return UpdateConfig(&CLIConfig{
            App: app,
        }, basePath)
    }
    readCfg.App = app
    return UpdateConfig(readCfg, basePath)
}

func GetConfig(basePath string) (*CLIConfig, error) {
	configPath := path.Join(basePath, "config")
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cliConfig CLIConfig
	if err := json.Unmarshal(b, &cliConfig); err != nil {
		return nil, err
	}

	return &cliConfig, nil
}

func UpdateConfig(config *CLIConfig, basePath string) error {
	configPath := path.Join(basePath, "config")

	b, err := json.Marshal(&config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, b, 0755)
}
