package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

// init
// @Description 	查看当前版本信息
// @Create          2025-09-01 17:32:18
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
