package command

import (
	"errors"

	"github.com/spf13/cobra"
)

func LoginCommand() *cobra.Command {
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
				return adminLoginFlow()
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

func adminLoginFlow() error {
	return nil
}

func userLoginFlow() error {
	return nil
}
