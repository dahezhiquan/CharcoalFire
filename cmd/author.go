package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"

	"github.com/spf13/cobra"
)

// init
// @Description 	作者信息
// @Create          2025-09-01 17:31:16

func init() {
	rootCmd.AddCommand(authorCmd)
}

var authorCmd = &cobra.Command{
	Use:   "author",
	Short: "查看作者信息",
	Run: func(cmd *cobra.Command, args []string) {
		color.Info.Println(utils.Author)
	},
}
