package cmd

import (
	"CharcoalFire/utils"
	"crypto/tls"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Parameter struct {
	url     string
	timeout int
	proxy   string
	file    string
	isClean bool
}

func init() {
	rootCmd.AddCommand(surviveCmd)
	surviveCmd.Flags().StringP("url", "u", "", "目标url")
	surviveCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	surviveCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	surviveCmd.Flags().StringP("proxy", "p", "", "代理地址")
	surviveCmd.Flags().BoolP("clean", "c", false, "过滤者模式（目标相同title只保留一个）")
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
		parameter.isClean, _ = cmd.Flags().GetBool("clean")
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
	var client *http.Client
	var isUrl = utils.IsUrl(parameter.url)
	if !isUrl {
		color.Error.Println(parameter.url + " 目标不是URL")
		return false, utils.HtmlDocument{}
	}

	if parameter.proxy != "" {
		proxyURL, err := url.Parse(parameter.proxy)
		if err != nil {
			color.Error.Println("代理解析失败")
		}
		client = &http.Client{
			Timeout: time.Duration(parameter.timeout) * time.Second,
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(proxyURL),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		client = &http.Client{
			Timeout:   time.Duration(parameter.timeout) * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	}

	resp, err := client.Get(parameter.url)
	if err != nil {
		color.Error.Println(parameter.url + " 目标连接失败")
		return false, utils.HtmlDocument{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			color.Warn.Println(parameter.url + " 目标连接未关闭")
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		color.Success.Println(parameter.url + " 目标存活")
		return true, utils.ParseHtml(resp)
	} else {
		color.Danger.Println(parameter.url + " 目标不存活，状态码：" + strconv.Itoa(resp.StatusCode))
		return false, utils.HtmlDocument{}
	}
}

func SurviveCmdByFile(parameter Parameter) {
	// 先对文件进行处理，去重和去空行
	utils.ProcessSourceFile(parameter.file)
	result, err := utils.ReadLinesFromFile(parameter.file)
	if err != nil {
		color.Error.Println("文件解析失败")
		return
	}

	// 产生本次结果前先将原来的结果清空
	utils.ClearFile("survive")

	// 存活的目标
	var surviveUrls []string

	// 存放存活目标的title字典
	surviveUrlsInfo := make(map[string]string)

	numThreads := runtime.NumCPU() // 获取当前系统的逻辑 CPU 数量
	runtime.GOMAXPROCS(numThreads) // 设置 Goroutine 可以并行执行的最大 CPU 数量
	color.Success.Println(parameter.file + "文件解析成功，目标总数：" + strconv.Itoa(len(result)) + " 当前线程数：" + strconv.Itoa(numThreads))

	var wg sync.WaitGroup

	for _, url := range result {
		var parameter2 Parameter
		parameter2 = parameter
		parameter2.url = url
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			isSurvive, htmlDocument := SurviveCmd(parameter2)
			if isSurvive {
				surviveUrls = append(surviveUrls, parameter2.url)
				// 目标无标题，取随机数当标题
				if htmlDocument.Title == "" {
					surviveUrlsInfo[parameter2.url] = utils.RandomString(16)
				} else {
					surviveUrlsInfo[parameter2.url] = htmlDocument.Title
				}
			}
		}(url)
	}

	wg.Wait()

	currentTime := time.Now().Format("20060102")
	filePath := filepath.Join(utils.ResultLogName, "survive", currentTime+".txt")
	if parameter.isClean {
		color.Info.Println("过滤者模式已开启，正在去重...")
		utils.WriteFile("survive", DeduplicateDictValues(surviveUrlsInfo))
		color.Success.Println("去重已完成")
		color.Success.Println("结果已保存到：" + filePath)
	} else {
		utils.WriteFile("survive", surviveUrls)
		color.Success.Println("结果已保存到：" + filePath)
	}
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
