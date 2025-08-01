package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func LoginCommand(cfg *CLIConfig) *cobra.Command {
	var role string
	var oidc bool
	login := &cobra.Command{
		Use:   "login",
		Short: "Helps you login to a Sailor Instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if role == "" {
				return errors.New("role is required")
			} else if role != "admin" && role != "user" {
				return errors.New("not a valid role")
			}

			switch role {
			case "admin":
				return adminLoginFlow(cfg)
			case "user":
				return userLoginFlow()
			}

			return nil
		},
	}

	// TODO :: use permitted flag feature to define an enum type and let cobra
	// validate for allowed roles
	login.Flags().StringVarP(&role, "role", "", "", "sailor user can be admin/user")
	login.Flags().BoolVarP(&oidc, "oidc", "", false, "use oidc flow for user authentication")

	return login
}

func adminLoginFlow(cfg *CLIConfig) error {
	var user string
	fmt.Print("sailor username: ")
	fmt.Scan(&user)

	fmt.Print("sailor pass: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	pass := string(bytePassword)
	tokenResp, err := cfg.SailorClient.LoginBasic(user, pass)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filepath.Join(cfg.SailorRoot, "tok"), []byte(tokenResp.Token), 0755); err != nil {
		return err
	}

	fmt.Println("logged in as: ", user)
	return nil
}

func userLoginFlow() error {
	return nil
}
