package command

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func ApplyCommand() *cobra.Command {
	var file string
	var url string
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Helps configure your sailor environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" && file == "" && len(args) == 0 {
				return errors.New("flags or env as argument is required")
			}

			if url == "" && file == "" {
				env := args[0]
				// TODO :: we need to check if sailor manifest file has this environment
				fmt.Println("selected: ", env)
			}

			return nil
		},
	}

	applyCmd.Flags().StringVarP(&url, "url", "", "", "configuration fetch url")
	applyCmd.Flags().StringVarP(&file, "file", "f", "", "configuration file")

	return applyCmd
}
