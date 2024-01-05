package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
)

type IconParameter struct {
	url     string
	timeout int
	proxy   string
	file    string
	isClean bool
	thread  int
}

func init() {
	ew := &utils.EmptyWriter{}
	log.SetOutput(io.Writer(ew))
	rootCmd.AddCommand(iconCmd)
	iconCmd.Flags().IntP("thread", "r", 500, "线程数")
	iconCmd.Flags().StringP("url", "u", "", "目标url")
	iconCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	iconCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	iconCmd.Flags().StringP("proxy", "p", "", "代理地址")
	iconCmd.Flags().BoolP("clean", "c", false, "过滤者模式（目标相同title只保留一个）")
}

var iconCmd = &cobra.Command{
	Use:   "icon",
	Short: "目标icon提取",
	Run: func(cmd *cobra.Command, args []string) {
		var iconParameter IconParameter
		iconParameter.url, _ = cmd.Flags().GetString("url")
		iconParameter.file, _ = cmd.Flags().GetString("file")
		iconParameter.proxy, _ = cmd.Flags().GetString("proxy")
		iconParameter.timeout, _ = cmd.Flags().GetInt("timeout")
		iconParameter.thread, _ = cmd.Flags().GetInt("thread")
		iconParameter.isClean, _ = cmd.Flags().GetBool("clean")
		if iconParameter.url != "" {
			isSurvive, htmlDocument := SurviveCmd(Parameter(iconParameter))
			if isSurvive && htmlDocument.Icon != "" {
				GetIcon(iconParameter, htmlDocument)
			} else {
				color.Warn.Println(iconParameter.url + "未在此找到icon")
			}
		}
		//if parameter.file != "" {
		//	SurviveCmdByFile(parameter)
		//	return
		//}
	},
}

func GetIcon(iconParameter IconParameter, htmlDocument utils.HtmlDocument) {
	if htmlDocument.Icon != "" {
		// 第一种情况：icon href是一个直接的地址
		if utils.IsUrl(htmlDocument.Icon) {
			DownLoadIcon(iconParameter, htmlDocument)
			return
		}
		// 第二种情况：icon href是一个相对的路径
		if !utils.IsUrl(htmlDocument.Icon) {
			htmlDocument.Icon = iconParameter.url + string('/') + htmlDocument.Icon
			DownLoadIcon(iconParameter, htmlDocument)
			return
		}
		// TODO 第三种情况：icon href 是协议的格式

	} else {
		color.Warn.Println("未在此找到icon")
	}
}

func DownLoadIcon(iconParameter IconParameter, htmlDocument utils.HtmlDocument) {
	// 创建图片流
	iconDownloadPath := utils.ResultLogName + "/icon/" + utils.GetDomain(iconParameter.url) + string('.') + utils.GetSuffix(htmlDocument.Icon)

	// 创建目录
	err := os.MkdirAll(filepath.Dir(iconDownloadPath), os.ModePerm)
	if err != nil {
		color.Error.Println("目录创建失败")
		return
	}

	out, err := os.Create(iconDownloadPath) // 保存到本地的文件名
	if err != nil {
		color.Error.Println("日志文件创建失败")
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			color.Warn.Println("日志文件未正常关闭")
		}
	}(out)

	ask := utils.Ask{}
	ask.Url = htmlDocument.Icon
	ask.Proxy = iconParameter.proxy
	ask.Timeout = iconParameter.timeout
	resp := utils.Outsourcing(ask)

	// 写入图片流
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		color.Error.Println("icon下载失败")
	}
	color.Success.Println("icon已下载到：" + iconDownloadPath)
}
