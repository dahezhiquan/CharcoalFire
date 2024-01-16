package cmd

import (
	"CharcoalFire/utils"
	"github.com/spf13/cobra"
	"io"
	"log"
	"sort"
	"strings"
	"sync"
)

type FisherParameter struct {
	timeout int
	proxy   string
	file    string
	thread  int
	isPHP   bool
}

var Lf = utils.GetSlog("fisher")

func init() {
	ew := &utils.EmptyWriter{}
	log.SetOutput(io.Writer(ew))
	rootCmd.AddCommand(fisherCmd)
	fisherCmd.Flags().IntP("thread", "r", 500, "线程数")
	fisherCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	fisherCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	fisherCmd.Flags().StringP("proxy", "p", "", "代理地址")
	fisherCmd.Flags().BoolP("php", "q", false, "导出所有的PHP资产（按照PHP版本号从小到大）")

}

var fisherCmd = &cobra.Command{
	Use:   "fisher",
	Short: "指定资产提取",
	Run: func(cmd *cobra.Command, args []string) {
		var fisherParameter FisherParameter
		fisherParameter.file, _ = cmd.Flags().GetString("file")
		fisherParameter.proxy, _ = cmd.Flags().GetString("proxy")
		fisherParameter.timeout, _ = cmd.Flags().GetInt("timeout")
		fisherParameter.thread, _ = cmd.Flags().GetInt("thread")
		fisherParameter.isPHP, _ = cmd.Flags().GetBool("php")
		if fisherParameter.file != "" && fisherParameter.isPHP {
			GetPhpByFile(fisherParameter)
			return
		}
	},
}

// 1.目标响应包的X-Powered-By字段判断

func GetPhpByFile(fisherParameter FisherParameter) {

	// 先对文件进行处理
	utils.ProcessSourceFile(fisherParameter.file)
	result, err := utils.ReadLinesFromFile(fisherParameter.file)
	if err != nil {
		Lf.Fatal("fisherman列表文件解析失败")
	}
	threadNum := fisherParameter.thread

	phpUrlsInfo := make(map[string]string)
	var wg sync.WaitGroup
	urlChan := make(chan string)

	for i := 0; i < threadNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChan {
				ask := utils.Ask{}
				ask.Url = url
				ask.Proxy = fisherParameter.proxy
				ask.Timeout = fisherParameter.timeout
				resp := utils.Outsourcing(ask)
				version := ""
				if resp != nil {
					// 目标响应包的X-Powered-By字段判断
					phpVersion := resp.Header.Get("X-Powered-By")
					if strings.Contains(phpVersion, "PHP") {
						lanVersion := utils.GetLanVersion(phpVersion)
						if len(lanVersion) > 1 {
							version = lanVersion[1]
						} else {
							version = "未知"
						}
						Lf.Info("发现PHP资产：" + url + " PHP Version：" + version)
						phpUrlsInfo[url] = version
					}
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

	sortVersion(phpUrlsInfo)

}

type someVersion struct {
	Url     string
	Version string
}

func sortVersion(urlsInfo map[string]string) {

	var someVersions []someVersion
	for k, v := range urlsInfo {
		someVersions = append(someVersions, someVersion{k, v})
	}

	sort.Slice(someVersions, func(i, j int) bool {
		return someVersions[i].Version < someVersions[j].Version // 降序
	})

	urls := make([]string, 0)
	versions := make([]string, 0)

	for _, v := range someVersions {
		urls = append(urls, v.Url)
		versions = append(versions, "PHP/"+v.Version)
	}

	utils.WriteFileBySuffix("fisher", urls, versions)
}
