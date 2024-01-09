package cmd

import (
	"CharcoalFire/utils"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type IconParameter struct {
	url      string
	timeout  int
	proxy    string
	file     string
	isClean  bool
	thread   int
	isDoamin bool
}

var Lc = utils.GetSlog("icon")

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
		// iconParameter.isClean, _ = cmd.Flags().GetBool("clean")
		if iconParameter.url != "" {
			isSurvive, htmlDocument := SurviveCmd(Parameter(iconParameter))
			if isSurvive && htmlDocument.Icon != "" {
				GetIcon(iconParameter, htmlDocument)
			} else {
				Lc.Error(iconParameter.url + " 未在此找到icon")
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
		// 特殊情况，不包含图片
		flag := false
		for _, fileType := range utils.ImgFileTypes {
			if strings.Contains(htmlDocument.Icon, fileType) {
				flag = true
			}
		}
		if !flag {
			Lc.Error(iconParameter.url + " 未在此找到icon")
			return
		}

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
		Lc.Error(iconParameter.url + " 未在此找到icon")
		return
	}
}

func DownLoadIcon(iconParameter IconParameter, htmlDocument utils.HtmlDocument) {
	// 去除多余的斜杠
	htmlDocument.Icon = utils.DelExtraSlash(htmlDocument.Icon)
	ask := utils.Ask{}
	ask.Url = htmlDocument.Icon
	ask.Proxy = iconParameter.proxy
	ask.Timeout = iconParameter.timeout
	resp := utils.Outsourcing(ask)

	// TODO 修复icon路径访问302跳转200误判为icon的bug
	//respTmp := resp
	//
	//// 修复icon路径访问失效的bug
	//body, _ := io.ReadAll(respTmp.Body)
	//sourceCode := string(body)
	//if strings.Contains(sourceCode, "html>") {
	//	color.Warn.Println(iconParameter.url + " 未在此找到icon")
	//	return
	//}

	if resp != nil && resp.StatusCode != 200 {
		Lc.Error(iconParameter.url + " 找到icon标识，但是图标已经破损")
		return
	}

	// 创建图片流
	iconDownloadPath := utils.ResultLogName + "/icon/" + utils.GetDomain(iconParameter.url) + string('.') + utils.GetSuffix(htmlDocument.Icon)

	// 创建目录
	err := os.MkdirAll(filepath.Dir(iconDownloadPath), os.ModePerm)
	if err != nil {
		Lc.Fatal("icon保存目录创建失败")
		return
	}

	out, err := os.Create(iconDownloadPath) // 保存到本地的文件名
	if err != nil {
		Lc.Fatal(iconParameter.url + " icon保存文件创建失败")
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			Lc.Warning(iconParameter.url + " icon保存文件未正常关闭")
		}
	}(out)

	if resp != nil {
		// 写入图片流
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			Lc.Fatal(iconParameter.url + " icon下载失败")
		}
		Lc.Info(iconParameter.url + " icon已下载到：" + iconDownloadPath)
	}

	if resp != nil {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}
}

// GetIconByFile 批量提取icon
func GetIconByFile(iconParameter IconParameter) {
	// 先对文件进行处理
	utils.ProcessSourceFile(iconParameter.file)
	result, err := utils.ReadLinesFromFile(iconParameter.file)
	if err != nil {
		Lc.Fatal("icon列表文件解析失败")
	}

	threadNum := iconParameter.thread

	var wg sync.WaitGroup
	//var mu sync.Mutex // 互斥锁
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
					//mu.Lock() // 加锁
					GetIcon(iconParameter2, htmlDocument)
					//mu.Unlock() // 解锁
				} else {
					Lc.Error(url + " 未在此找到icon")
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
