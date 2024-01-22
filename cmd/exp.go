package cmd

import (
	"CharcoalFire/cmd/poc"
	"CharcoalFire/utils"
	"github.com/spf13/cobra"
	"io"
	"log"
	"sync"
)

var Lexp = utils.GetSlog("exp")
var wgs sync.WaitGroup
var mus sync.Mutex // 互斥锁
var vulUrls []string

func init() {
	ew := &utils.EmptyWriter{}
	log.SetOutput(io.Writer(ew))
	rootCmd.AddCommand(expCmd)
	expCmd.Flags().IntP("thread", "r", 500, "线程数（同时扫多少目标）")
	expCmd.Flags().StringP("url", "u", "", "目标url")
	expCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	expCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	expCmd.Flags().StringP("proxy", "p", "", "代理地址")
	expCmd.Flags().BoolP("isexp", "e", false, "执行EXP模块")
	expCmd.Flags().BoolP("ispoc", "o", false, "执行POC模块")
	expCmd.Flags().StringP("vulname", "v", "", "漏洞名称（直接输入CVE号，例如CVE-2023-42793）")
}

var expCmd = &cobra.Command{
	Use:   "exp",
	Short: "exp模块（肉鸡收集）",
	Run: func(cmd *cobra.Command, args []string) {
		var expParameter utils.ExpParameter
		expParameter.Url, _ = cmd.Flags().GetString("url")
		expParameter.File, _ = cmd.Flags().GetString("file")
		expParameter.Proxy, _ = cmd.Flags().GetString("proxy")
		expParameter.Timeout, _ = cmd.Flags().GetInt("timeout")
		expParameter.Thread, _ = cmd.Flags().GetInt("thread")
		expParameter.IsExp, _ = cmd.Flags().GetBool("isexp")
		expParameter.IsPoc, _ = cmd.Flags().GetBool("ispoc")
		expParameter.Vulname, _ = cmd.Flags().GetString("vulname")
		if expParameter.Url != "" && expParameter.IsPoc == true {
			ChoiceMethodPOC(expParameter)
			return
		}
		if expParameter.File != "" && expParameter.IsPoc == true {
			// 先对文件进行处理
			utils.ProcessSourceFile(expParameter.File)
			result, err := utils.ReadLinesFromFile(expParameter.File)
			if err != nil {
				Lexp.Fatal("exp列表文件解析失败")
				return
			}
			urlChan := make(chan string)
			for i := 0; i < expParameter.Thread; i++ {
				wgs.Add(1)
				go func() {
					defer wgs.Done()
					for url := range urlChan {
						expParameter.Url = url
						ChoiceMethodPOC(expParameter)
					}
				}()
			}
			for _, url := range result {
				urlChan <- url
			}
			close(urlChan)
			wgs.Wait()
			utils.WriteFile("poc", vulUrls, false)
			return
		}
	},
}

func ChoiceMethodPOC(expParameter utils.ExpParameter) {
	if expParameter.Vulname == "CVE-2023-42793" {
		isVul := poc.V202342793(expParameter)
		if isVul {
			Lexp.Info("发现TeamCity 任意代码执行漏洞 " + expParameter.Url)
			mus.Lock() // 加锁
			vulUrls = append(vulUrls, expParameter.Url)
			mus.Unlock() // 解锁
		} else {
			Lexp.Fatal("不存在TeamCity 任意代码执行漏洞 " + expParameter.Url)
		}
	} else {
		Lexp.Fatal("未找到此POC")
	}
}
