package cmd

import (
	"CharcoalFire/utils"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type Parameter struct {
	url      string
	timeout  int
	proxy    string
	file     string
	isClean  bool
	thread   int
	isDoamin bool
	isClear  bool
}

var Ls = utils.GetSlog("survive")

func init() {
	ew := &utils.EmptyWriter{}
	log.SetOutput(io.Writer(ew))
	rootCmd.AddCommand(surviveCmd)
	surviveCmd.Flags().IntP("thread", "r", 500, "线程数")
	surviveCmd.Flags().StringP("url", "u", "", "目标url")
	surviveCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	surviveCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	surviveCmd.Flags().StringP("proxy", "p", "", "代理地址")
	surviveCmd.Flags().BoolP("clean", "c", false, "过滤者模式（目标相同title只保留一个）")
	surviveCmd.Flags().BoolP("domain", "d", false, "以domain格式输出结果")
	surviveCmd.Flags().BoolP("clear", "k", false, "智能去除解析多个IP地址的目标")
}

var surviveCmd = &cobra.Command{
	Use:   "survive",
	Short: "目标存活探测",
	Run: func(cmd *cobra.Command, args []string) {
		var parameter Parameter
		parameter.url, _ = cmd.Flags().GetString("url")
		parameter.file, _ = cmd.Flags().GetString("file")
		parameter.proxy, _ = cmd.Flags().GetString("proxy")
		parameter.timeout, _ = cmd.Flags().GetInt("timeout")
		parameter.thread, _ = cmd.Flags().GetInt("thread")
		parameter.isClean, _ = cmd.Flags().GetBool("clean")
		parameter.isDoamin, _ = cmd.Flags().GetBool("domain")
		parameter.isClear, _ = cmd.Flags().GetBool("clear")
		if parameter.url != "" {
			SurviveCmd(parameter)
			return
		}
		if parameter.file != "" {
			SurviveCmdByFile(parameter)
			return
		}
	},
}

func SurviveCmd(parameter Parameter) (bool, utils.HtmlDocument) {

	var isUrl = utils.IsUrl(parameter.url)
	if !isUrl {
		Ls.Error(parameter.url + " 目标不是URL")
		return false, utils.HtmlDocument{}
	}

	ask := utils.Ask{}
	ask.Url = parameter.url
	ask.Proxy = parameter.proxy
	ask.Timeout = parameter.timeout
	resp := utils.Outsourcing(ask)

	// 防止空指针问题
	if resp != nil {
		if resp.StatusCode == http.StatusOK {
			Ls.Info(parameter.url + " 目标存活")
			return true, utils.GetHtmlDocument(parameter.url, resp)
		} else {
			Ls.Error(parameter.url + " 目标不存活，状态码：" + strconv.Itoa(resp.StatusCode))
			return false, utils.HtmlDocument{}
		}
	} else {
		return false, utils.HtmlDocument{}
	}
}

func SurviveCmdByFile(parameter Parameter) []string {
	// 先对文件进行处理
	utils.ProcessSourceFile(parameter.file)
	result, err := utils.ReadLinesFromFile(parameter.file)
	if err != nil {
		Ls.Fatal("存活探测列表解析失败")
		return nil
	}

	// 存活的目标
	var surviveUrls []string

	// 存放存活目标的title字典
	surviveUrlsInfo := make(map[string]string)

	threadNum := parameter.thread

	var wg sync.WaitGroup
	var mu sync.Mutex // 互斥锁
	urlChan := make(chan string)

	for i := 0; i < threadNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range urlChan {
				var parameter2 Parameter
				parameter2 = parameter
				parameter2.url = target
				isSurvive, htmlDocument := SurviveCmd(parameter2)
				if isSurvive {
					mu.Lock() // 加锁
					surviveUrls = append(surviveUrls, parameter2.url)
					// 目标无标题，取随机数当标题
					if htmlDocument.Title == "" {
						surviveUrlsInfo[parameter2.url] = utils.RandomString(16)
					} else {
						surviveUrlsInfo[parameter2.url] = htmlDocument.Title
					}
					mu.Unlock() // 解锁
				}
			}
		}()
	}

	// 将url发送到urlChan供消费者goroutine处理
	for _, target := range result {
		urlChan <- target
	}
	close(urlChan)
	wg.Wait()

	if parameter.isClean {
		Ls.Debug("过滤者模式已开启，正在去重...")
		surviveUrls = DeduplicateDictValues(surviveUrlsInfo)
		Ls.Debug("去重已完成，共去除条数：")
	}

	if parameter.isClear {

		Ls.Debug("智能去除多IP目标已开启，正在去除...")

		urlChan2 := make(chan string)

		for i := 0; i < parameter.thread; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for target := range urlChan2 {
					if IsMultipleIPs(target) {
						mu.Lock() // 加锁
						Ls.Info("发现多IP解析目标：" + target)
						surviveUrls = utils.RemoveElementSlice(surviveUrls, target)
						mu.Unlock() // 解锁
					}
				}
			}()
		}

		// 将url发送到urlChan供消费者goroutine处理
		for _, target := range surviveUrls {
			urlChan2 <- target
		}
		close(urlChan2)
		wg.Wait()

		Ls.Debug("去除已完成")
	}
	println(len(surviveUrls))
	utils.WriteFile("survive", surviveUrls, parameter.isDoamin)
	return surviveUrls
}

// DeduplicateDictValues 去重字典的值，即相同的值只保留一个，并将去重后的键生成一个新的切片
func DeduplicateDictValues(surviveUrlsInfo map[string]string) []string {
	seen := make(map[string]bool)
	uniqueKeys := make([]string, 0)
	for key, value := range surviveUrlsInfo {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = true
		uniqueKeys = append(uniqueKeys, key)
	}

	return uniqueKeys
}

// IsMultipleIPs 智能去除多IP
func IsMultipleIPs(target string) bool {
	parsedURL, err := url.Parse(target)
	if err != nil {
		return false
	}
	domain := parsedURL.Hostname()
	ips, err := net.LookupIP(domain)
	if err != nil {
		Ls.Error(target + " DNS解析失败")
		return false
	}

	if len(ips) > 1 {
		return true
	} else {
		return false
	}
}
