package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

func ApplyCommand(cfg *CLIConfig) *cobra.Command {
	var file string
	var host string
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Helps configure your sailor environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if host == "" && file == "" && len(args) == 0 {
				return errors.New("flags or env as argument is required")
			}

			// TODO :: separate logic into nice functions!

			if host == "" && file == "" {
				env := args[0]
				for _, envConfig := range cfg.Manifest.Envs {
					if strings.EqualFold(env, envConfig.Name) {
						if err := os.WriteFile(filepath.Join(cfg.SailorRoot, "env"), []byte(envConfig.Name), 0655); err != nil {
							return err
						}
						fmt.Printf("switched to %s\n", envConfig.Name)
						return nil
					}
				}

				fmt.Println("no environment named: ", env)

				return nil
			}

			if host != "" {
				manifestURL := fmt.Sprintf("%s/api/v1/setting/manifest", host)
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

				// since it is parsable as sailor manifest, we save it!
				if err := os.WriteFile(filepath.Join(cfg.SailorRoot, "config"), b, 0755); err != nil {
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

				// since it is parsable as sailor manifest, we save it!
				if err := os.WriteFile(filepath.Join(cfg.SailorRoot, "config"), b, 0755); err != nil {
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

	return applyCmd
}
