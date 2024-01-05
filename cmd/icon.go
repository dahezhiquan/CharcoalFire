package cmd

import (
	"CharcoalFire/utils"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
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
	// TODO 过滤者模式 相同icon hash只保留一个
	// iconCmd.Flags().BoolP("clean", "c", false, "过滤者模式（目标相同icon hash只保留一个）")
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
			return
		}
		if iconParameter.file != "" {
			GetIconByFile(iconParameter)
		}
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

// GetIconByFile 批量提取icon
func GetIconByFile(iconParameter IconParameter) {
	// 先对文件进行处理
	utils.ProcessSourceFile(iconParameter.file)
	result, err := utils.ReadLinesFromFile(iconParameter.file)
	if err != nil {
		color.Error.Println("文件解析失败")
	}

	threadNum := iconParameter.thread

	var wg sync.WaitGroup
	var mu sync.Mutex // 互斥锁
	urlChan := make(chan string)

	for i := 0; i < threadNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChan {
				var iconParameter2 IconParameter
				iconParameter2 = iconParameter
				iconParameter2.url = url
				isSurvive, htmlDocument := SurviveCmd(Parameter(iconParameter2))
				if isSurvive && htmlDocument.Icon != "" {
					mu.Lock() // 加锁
					GetIcon(iconParameter2, htmlDocument)
					mu.Unlock() // 解锁
				} else {
					color.Warn.Println(iconParameter.url + "未在此找到icon")
				}
			}
		}()
	}

	// 将url发送到urlChan供消费者goroutine处理
	for _, url := range result {
		urlChan <- url
	}
	close(urlChan)
	wg.Wait()
}
