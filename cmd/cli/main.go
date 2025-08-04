package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codekidx/sailor/cmd/cli/command"
	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func main() {
	var env string
	var rootCmd = &cobra.Command{
		Use:   "sailor",
		Short: "Sailor is a resource management and delivery system for cloud-native apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cfg, err := loadConfig(env)
	if err != nil {
		panic(err.Error())
	}

	rootCmd.PersistentFlags().StringVarP(&env, "env", "", "", "Helps set explicit environment")

	fmt.Println("🐧 working on: ", cfg.Env)

	rootCmd.AddCommand(command.ApplyCommand(cfg))
	rootCmd.AddCommand(command.LoginCommand(cfg))
	rootCmd.AddCommand(command.CreateCommand(cfg))
	rootCmd.AddCommand(command.DeployCommand(cfg))
	rootCmd.AddCommand(command.SchemaCommand())
	rootCmd.AddCommand(command.SettingCommand())
	rootCmd.AddCommand(command.GetCommand(cfg))
	rootCmd.AddCommand(command.RBACCommand(cfg))

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
			return nil, errors.New("error while creating sailor workspace")
		}
	}

	sailorConfig := filepath.Join(sailorRoot, "config")
	var config command.CLIConfig
	if f, _ := os.Stat(sailorConfig); f != nil {

		b, err := os.ReadFile(sailorConfig)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &config); err != nil {
			return nil, err
		}
	} else {
		fmt.Println("manifest not loaded do [ sailor apply --file ] or [ sailor apply --host ]")
	}

	config.SailorRoot = sailorRoot

	if config.Env != "" {
		for _, envConfig := range config.Manifest.Envs {
			if strings.EqualFold(envConfig.Name, config.Env) {
				config.SailorHost = envConfig.Host
			}
		}
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

	if config.SailorHost != "" {
		config.SailorClient = v1.CoreV1(config.SailorHost)
	}

	return &config, nil
}
