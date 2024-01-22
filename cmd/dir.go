package cmd

import (
	"CharcoalFire/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type DirParameter struct {
	url         string
	timeout     int
	proxy       string
	file        string
	thread      int
	threadOnly  int
	level       string
	isBackstage bool // 后台爆破
	isBackUp    bool // 备份文件爆破
}

type TargetInfo struct {
	code  string
	title string
	size  int // 返回包大小
}

var Ldir = utils.GetSlog("dir")
var dictionary []string // 字典
var wg sync.WaitGroup
var targetInfo []TargetInfo

func init() {
	ew := &utils.EmptyWriter{}
	log.SetOutput(io.Writer(ew))
	rootCmd.AddCommand(dirCmd)
	dirCmd.Flags().IntP("thread", "r", 500, "线程数（同时扫多少目标）")
	dirCmd.Flags().IntP("threadonly", "y", 20, "单个目标线程数")
	dirCmd.Flags().StringP("url", "u", "", "目标url")
	dirCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	dirCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	dirCmd.Flags().StringP("proxy", "p", "", "代理地址")
	dirCmd.Flags().StringP("level", "l", "1", "爆破等级（1为小字典爆破，2为中字典爆破，3为大字典爆破）")
	dirCmd.Flags().BoolP("admin", "a", false, "后台发现")
	dirCmd.Flags().BoolP("backup", "b", false, "备份文件发现")
}

var dirCmd = &cobra.Command{
	Use:   "dir",
	Short: "目录爆破",
	Run: func(cmd *cobra.Command, args []string) {
		var dirParameter DirParameter
		dirParameter.url, _ = cmd.Flags().GetString("url")
		dirParameter.file, _ = cmd.Flags().GetString("file")
		dirParameter.proxy, _ = cmd.Flags().GetString("proxy")
		dirParameter.timeout, _ = cmd.Flags().GetInt("timeout")
		dirParameter.thread, _ = cmd.Flags().GetInt("thread")
		dirParameter.threadOnly, _ = cmd.Flags().GetInt("threadonly")
		dirParameter.level, _ = cmd.Flags().GetString("level")
		dirParameter.isBackstage, _ = cmd.Flags().GetBool("admin")
		dirParameter.isBackUp, _ = cmd.Flags().GetBool("backup")

		if dirParameter.isBackstage {

		}
		if dirParameter.isBackUp {
			temp, err := utils.ReadLinesFromFile("dict/backup" + dirParameter.level + ".txt")
			if err != nil {
				Ldir.Fatal("备份文件字典解析失败")
			}
			dictionary = append(dictionary, temp...)
		}

		if dirParameter.url != "" {
			CrackIt(dirParameter)
			return
		}
	},
}

// CrackIt 爆破启动器
func CrackIt(dirParameter DirParameter) {

	targetTitle := "ahahahahahahahaha"
	ask := utils.Ask{}
	ask.Url = dirParameter.url
	ask.Proxy = dirParameter.proxy
	ask.Timeout = dirParameter.timeout
	resp := utils.Outsourcing(ask)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		Ldir.Error(dirParameter.url + " 目标节点解析失败")
	}
	targetTitle = doc.Find("title").First().Text()

	urlChan := make(chan string)
	soldiers := 0 // 排雷兵，用来检测是不是被目标ban了
	for i := 0; i < dirParameter.threadOnly; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dict := range urlChan {
				dirPage := utils.DelExtraSlash(dirParameter.url + dict)
				ask := utils.Ask{}
				ask.Url = dirPage
				ask.Proxy = dirParameter.proxy
				ask.Timeout = dirParameter.timeout
				resp := utils.Outsourcing(ask)
				if resp == nil {
					soldiers++
					// 连续连接失败次数达到阈值
					if soldiers > utils.SoldiersThreshold {
						Ldir.Error(dirParameter.url + " 已经被ban，停止扫描")
						return
					}
				} else {
					soldiers = 0 // 重置排雷兵
					isValid, nowTitle := IsValid(resp, targetTitle)
					if isValid {
						Ldir.Info(dirParameter.url + " 发现目录 " + dirPage + " 状态码 " + strconv.Itoa(resp.StatusCode))
						info := TargetInfo{
							code:  strconv.Itoa(resp.StatusCode),
							title: nowTitle,
						}
						targetInfo = append(targetInfo, info)
					} else {
						Ldir.Fatal(dirParameter.url + " 不存在该目录 " + dirPage + " 状态码 " + strconv.Itoa(resp.StatusCode))
					}
				}
			}
		}()
	}

	// 将url发送到urlChan供消费者goroutine处理
	for _, dict := range dictionary {
		urlChan <- dict
	}
	close(urlChan)
	wg.Wait()
}

func IsValid(resp *http.Response, targetTitle string) (bool, string) {
	if resp.StatusCode != 200 {
		return false, ""
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, ""
	}
	nowTitle := doc.Find("title").First().Text()
	// 防止目录泛解析
	if nowTitle == targetTitle {
		return false, ""
	}

	for _, v := range utils.NotFoudList {
		if nowTitle == v {
			return false, ""
		}
	}
	return true, nowTitle
}
