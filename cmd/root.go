package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "CharcoalFire",
		Short: "炭火 - 综合渗透测试工具",
		Run: func(cmd *cobra.Command, args []string) {
			color.Info.Println(utils.Banner)
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// TODO 配置文件的支持
	cobra.OnInitialize(initConfig)
}

func initConfig() {

}
