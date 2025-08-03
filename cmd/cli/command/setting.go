package command

import "github.com/spf13/cobra"

func SettingCommand() *cobra.Command {
	var input string
	var file string
	var ns string
	var app string
	settingCmd := &cobra.Command{
		Use:   "setting",
		Short: "Helps you update sailor or resource setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	settingCmd.Flags().StringVarP(&ns, "namespace", "", "", "create a resource in this namespace")
	settingCmd.Flags().StringVarP(&app, "app", "", "", "create a resource for this app")
	settingCmd.Flags().StringVarP(&file, "file", "f", "", "update setting from json file")
	settingCmd.Flags().StringVarP(&input, "input", "i", "", "update setting from input string")

	return settingCmd
}
