package command

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/pkg/browser"
)

func LoginCommand(cfg *CLIConfig) *cobra.Command {
	var role string
	var oidc bool
	login := &cobra.Command{
		Use:   "login",
		Short: "Helps you login to a Sailor Instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if oidc {
				var user string
				fmt.Print("sailor user: ")
				fmt.Scan(&user)

				url := fmt.Sprintf("%s/api/v1/auth/oidc", cfg.SailorHost)
				if err := browser.OpenURL(url); err != nil {
					return err
				}

				// TODO :: ideally the client should create a tiny server and the client should
				// be redirected to the server created by client
				fmt.Println("\nsailor will wait for 20s to finish logging in...")
				time.Sleep(20 * time.Second)

				// TODO (Security) :: right now token can be fetched by anyone who knows the proper sailor
				// user name and can gain access to the sailor APIs .. this is a threat and should be solved
				// by passing OIDC API a secure key which can identify the token created, the token
				// should not be provided if this secure key is not given during request
				tokenResp, err := cfg.SailorClient.GetToken(user)
				if err != nil {
					return err
				}

				if err = os.WriteFile(filepath.Join(cfg.SailorRoot, "tok"), []byte(tokenResp.Token), 0755); err != nil {
					return err
				}

				fmt.Println("logged in as:", user)
			}

			err := basicLoginFlow(cfg)
			if err != nil {
				return err
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

func basicLoginFlow(cfg *CLIConfig) error {
	var user string
	fmt.Print("sailor user: ")
	fmt.Scan(&user)

	fmt.Print("sailor pass: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	pass := string(bytePassword)
	loginResp, err := cfg.SailorClient.LoginBasic(user, pass)
	if err != nil {
		return err
	}

	if err := updateTokenAndKeyPairs(loginResp.Token, loginResp.KeyPairs, cfg.SailorRoot); err != nil {
		return err
	}

	fmt.Println("logged in as: ", user)
	return nil
}
