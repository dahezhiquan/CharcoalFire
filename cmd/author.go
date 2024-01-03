package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"

	"github.com/spf13/cobra"
)

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
