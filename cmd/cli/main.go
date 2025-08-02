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

	cfg, err := loadConfig()
	if err != nil {
		panic(err.Error())
	}

	rootCmd.PersistentFlags().StringVarP(&env, "env", "", "", "Helps set explicit environment")

	if env != "" {
		cfg.Env = env
	} else {
		b, err := os.ReadFile(filepath.Join(cfg.SailorRoot, "env"))
		if err != nil {
			fmt.Println("no environment mentioned")
			return
		}
		cfg.Env = string(b)
	}

	for _, envConfig := range cfg.Manifest.Envs {
		if strings.EqualFold(envConfig.Name, cfg.Env) {
			cfg.SailorHost = envConfig.Host
		}
	}

	if cfg.SailorHost != "" {
		cfg.SailorClient = v1.CoreV1(cfg.SailorHost)
	}

	fmt.Println("🐧 working on: ", cfg.Env)

	rootCmd.AddCommand(command.ApplyCommand(cfg))
	rootCmd.AddCommand(command.LoginCommand(cfg))
	rootCmd.AddCommand(command.CreateCommand(cfg))
	rootCmd.AddCommand(command.DeployCommand())
	rootCmd.AddCommand(command.SchemaCommand())
	rootCmd.AddCommand(command.SettingCommand())
	rootCmd.AddCommand(command.GetCommand())
	rootCmd.AddCommand(command.RBACCommand(cfg))

	if err := rootCmd.Execute(); err != nil {
		panic(err.Error())
	}
}

func loadConfig() (*command.CLIConfig, error) {
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

	sailorManifestFilePath := filepath.Join(sailorRoot, "config")
	var manifest v1.SailorManifest
	if f, _ := os.Stat(sailorManifestFilePath); f != nil {

		b, err := os.ReadFile(sailorManifestFilePath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &manifest); err != nil {
			return nil, err
		}
	} else {
		fmt.Println("manifest not loaded do [ sailor apply --file ] or [ sailor apply --host ]")
	}

	sailorUserTokenPath := filepath.Join(sailorRoot, "tok")
	var token string
	if f, _ := os.Stat(sailorUserTokenPath); f != nil {
		b, err := os.ReadFile(sailorUserTokenPath)
		if err != nil {
			return nil, err
		}

		token = string(b)
	}

	return &command.CLIConfig{Manifest: manifest, SailorRoot: sailorRoot, Token: token}, nil
}
