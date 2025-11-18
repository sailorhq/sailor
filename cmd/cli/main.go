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
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sailorhq/sailor/cmd/cli/command"
	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

var Version string

func main() {
	var env string
	var version bool
	var rootCmd = &cobra.Command{
		Use:   "sailor",
		Short: "Sailor is a resource management and delivery system for cloud-native apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			if version {
				fmt.Println(Version)
				return nil
			}
			return nil
		},
	}

	cfg, err := loadConfig(env)
	if err != nil {
		panic(err.Error())
	}

	rootCmd.PersistentFlags().StringVarP(&env, "env", "", "", "Helps set explicit environment")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "", false, "Your Sailor CLI version")

	fmt.Println("🐧 working on: ", cfg.Env)

	rootCmd.AddCommand(command.ApplyCommand(cfg))
	rootCmd.AddCommand(command.LoginCommand(cfg))
	rootCmd.AddCommand(command.CreateCommand(cfg))
	rootCmd.AddCommand(command.DeployCommand(cfg))
	rootCmd.AddCommand(command.SchemaCommand(cfg))
	rootCmd.AddCommand(command.SettingCommand(cfg))
	rootCmd.AddCommand(command.GetCommand(cfg))
	rootCmd.AddCommand(command.RBACCommand(cfg))
	rootCmd.AddCommand(command.InitCommand(cfg))
	rootCmd.AddCommand(command.SyncCommand(cfg))

	if err := rootCmd.Execute(); err != nil {
		panic(err.Error())
	}
}

func loadConfig(overrideEnv string) (*command.CLIConfig, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	sailorRoot := filepath.Join(home, ".sailor")
	if f, _ := os.Stat(sailorRoot); f == nil {
		if err := os.Mkdir(sailorRoot, 0755); err != nil {
			return nil, errors.New("error while loading sailor workspace")
		}
	}

	var config command.CLIConfig
	config.SailorRoot = sailorRoot

	sailorConfig := filepath.Join(sailorRoot, "config")
	if f, _ := os.Stat(sailorConfig); f != nil {

		b, err := os.ReadFile(sailorConfig)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &config); err != nil {
			return nil, err
		}

		if config.Env != "" {
			for _, envConfig := range config.Manifest.Envs {
				if strings.EqualFold(envConfig.Name, config.Env) {
					config.SailorHost = envConfig.Host
				}
			}
		}
	} else {
		fmt.Println("manifest not loaded do [ sailor apply --file ] or [ sailor apply --host ]")
		fmt.Println("using default environment: local")
		config.SailorHost = "http://localhost:7766"
		config.Env = "local"
	}

	// this means that user has overriden the env with flag
	if overrideEnv != "" {
		config.Env = overrideEnv
		for _, envConfig := range config.Manifest.Envs {
			if strings.EqualFold(envConfig.Name, config.Env) {
				config.SailorHost = envConfig.Host
			}
		}
	}

	if config.Manifest.Envs == nil {
		fmt.Println("manifest not loaded do [ sailor apply --file ] or [ sailor apply --host ]")
		fmt.Println("using default environment: local")
		config.SailorHost = "http://localhost:7766"
		config.Env = "local"
	}

	if config.SailorHost != "" {
		config.SailorClient = v1.CoreV1(config.SailorHost)
	}

	// check if sailor.json is there in current path
	if f, _ := os.Stat("./sailor.json"); f != nil {
		b, err := os.ReadFile("./sailor.json")
		if err != nil {
			return nil, err
		}

		var sf types.SailorFile
		if err := json.Unmarshal(b, &sf); err != nil {
			return nil, err
		}

		fmt.Println("🐧 namespace: ", sf.Project.Namespace)
		fmt.Println("🐧 app: ", sf.Project.App)

		config.CwdSailorFile = sf
	} else {
		config.CwdSailorFile = types.SailorFile{}
	}

	// check if sailor.lock.json is there in current path
	if f, _ := os.Stat("./sailor.lock"); f != nil {
		b, err := os.ReadFile("./sailor.lock")
		if err != nil {
			return nil, err
		}

		var sf types.SailorLockFile
		if err := json.Unmarshal(b, &sf); err != nil {
			return nil, err
		}

		config.CwdSailorLockFile = sf
	} else {
		config.CwdSailorLockFile = types.SailorLockFile{
			Environments: make(map[string]types.ResourceVersion),
		}
	}

	return &config, nil
}
