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

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Helps update RBAC of sailor user",
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
	addCmd.Flags().StringVarP(&role, "role", "", "", "update role for sailor user")
	addCmd.Flags().StringVarP(&perm, "perm", "", "", "update permission for sailor user")
	addCmd.Flags().StringVarP(&app, "app", "", "", "update allowed apps for sailor user")

	rbacCmd.AddCommand(addCmd)

	return rbacCmd
}
