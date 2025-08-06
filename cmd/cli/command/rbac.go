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

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/spf13/cobra"
)

func RBACCommand(cfg *CLIConfig) *cobra.Command {
	var user string
	var role string
	var perm string
	var app string
	rbacCmd := &cobra.Command{
		Use:   "rbac",
		Short: "Helps manage sailor user RBAC",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	rbacCmd.PersistentFlags().StringVarP(&user, "user", "", "", "target user")
	rbacCmd.PersistentFlags().StringVarP(&role, "role", "", "", "update role for sailor user")
	rbacCmd.PersistentFlags().StringVarP(&perm, "perm", "", "", "update permission for sailor user")
	rbacCmd.PersistentFlags().StringVarP(&app, "app", "", "", "update allowed apps for sailor user")

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Helps add constraints to a sailor user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if user != "" {
				var domains = []string{}
				rbacReq := v1.RBACRequest{}
				if role != "" {
					domains = append(domains, "role")
					rbacReq.Addition.Roles = strings.Split(role, ",")
				}

				if perm != "" {
					domains = append(domains, "permission")
					rbacReq.Addition.Permissions = strings.Split(perm, ",")
				}

				if app != "" {
					domains = append(domains, "app")
					rbacReq.Addition.AllowedApps = strings.Split(app, ",")
				}

				if len(rbacReq.Addition.Roles) == 0 && len(rbacReq.Addition.Permissions) == 0 &&
					len(rbacReq.Addition.AllowedApps) == 0 {
					return errors.New("nothing to update.")
				}

				if err := cfg.SailorClient.UpdateRBAC(rbacReq, user, cfg.Token); err != nil {
					return err
				}

				fmt.Printf("updated %s\n", strings.Join(domains, "|"))

				return nil
			}

			fmt.Println("user not provided")
			return nil
		},
	}

	delCmd := &cobra.Command{
		Use:   "del",
		Short: "Helps delete constraints from a sailor user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if user != "" {
				var domains = []string{}
				rbacReq := v1.RBACRequest{}
				if role != "" {
					domains = append(domains, "role")
					rbacReq.Deletion.Roles = strings.Split(role, ",")
				}

				if perm != "" {
					domains = append(domains, "permission")
					rbacReq.Deletion.Permissions = strings.Split(perm, ",")
				}

				if app != "" {
					domains = append(domains, "app")
					rbacReq.Deletion.AllowedApps = strings.Split(app, ",")
				}

				if len(rbacReq.Deletion.Roles) == 0 && len(rbacReq.Deletion.Permissions) == 0 &&
					len(rbacReq.Deletion.AllowedApps) == 0 {
					return errors.New("nothing to update.")
				}

				if err := cfg.SailorClient.UpdateRBAC(rbacReq, user, cfg.Token); err != nil {
					return err
				}

				fmt.Printf("updated %s\n", strings.Join(domains, "|"))

				return nil
			}

			fmt.Println("user not provided")
			return nil
		},
	}

	rbacCmd.AddCommand(addCmd)
	rbacCmd.AddCommand(delCmd)

	return rbacCmd
}
