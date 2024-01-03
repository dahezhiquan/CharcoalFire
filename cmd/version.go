package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看当前版本",
	Run: func(cmd *cobra.Command, args []string) {
		color.Info.Println(utils.Version)
	},
}
