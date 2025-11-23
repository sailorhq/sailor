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
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/pkg/browser"
)

var AdminUsername = "admin@super.sailor"

func LoginCommand(cfg *CLIConfig) *cobra.Command {
	var oidc bool
	var isAdmin bool

	login := &cobra.Command{
		Use:   "login",
		Short: "Helps you login to a Sailor Instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Env == "" {
				return errors.New("env not set. do [ sailor apply $env ]")
			}

			if oidc {
				var user string
				if isAdmin {
					user = AdminUsername
				} else if cfg.User != "" {
					user = cfg.User
				} else {
					fmt.Print("login as: ")
					fmt.Scan(&user)
				}

				if strings.TrimSpace(user) == "" {
					return errors.New("user cannot be empty")
				}

				fp := time.Now().UnixNano()
				url := fmt.Sprintf("%s/api/v1/auth/oidc?fp=%d", cfg.SailorHost, fp)
				if err := browser.OpenURL(url); err != nil {
					return err
				}

				// TODO :: ideally the client should create a tiny server and the client should
				// be redirected to the server created by client
				fmt.Println("\nsailor will wait for 15s to finish logging in...")
				time.Sleep(15 * time.Second)

				loginResp, err := cfg.SailorClient.GetToken(user, fmt.Sprint(fp))
				if err != nil {
					return err
				}

				if err = updateToken(loginResp.Token, cfg.SailorRoot); err != nil {
					return err
				}

				fmt.Println("logged in as:", user)
				return nil
			}

			err := basicLoginFlow(cfg, isAdmin)
			if err != nil {
				return err
			}

			return nil
		},
	}

	login.Flags().BoolVarP(&oidc, "oidc", "", false, "use oidc flow for user authentication")
	login.Flags().BoolVarP(&isAdmin, "admin", "", false, "use admin user as login username")

	return login
}

func basicLoginFlow(cfg *CLIConfig, isAdmin bool) error {
	var user string
	if isAdmin {
		user = AdminUsername
	} else if cfg.User != "" {
		user = cfg.User
	} else {
		fmt.Print("sailor user: ")
		fmt.Scan(&user)
	}

	if strings.TrimSpace(user) == "" {
		return errors.New("user cannot be empty")
	}

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

	if err := updateToken(loginResp.Token, cfg.SailorRoot); err != nil {
		return err
	}

	fmt.Println("\n\nlogged in as: ", user)
	return nil
}
