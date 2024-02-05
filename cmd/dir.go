package cmd

import (
	"CharcoalFire/utils"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type DirParameter struct {
	url          string
	timeout      int
	proxy        string
	file         string
	thread       int
	level        string
	scanMethod   string
	isPhpMyAdmin bool // phpmyadmin爆破
	isBackstage  bool // 后台爆破
	isBackUp     bool // 备份文件爆破
	isSqlBack    bool // 数据库备份快扫
	isTotal      bool // 综合扫描
	isWebShell   bool // webshell扫描
}

type TargetInfo struct {
	url   string
	code  string
	title string
	size  string // 返回包大小
}

var Ldir = utils.GetSlog("dir")
var dictionary []string // 字典
var targetInfo []TargetInfo
var connCntDir = 0 // 请求数
var totalCntDir = 0
var wgDir sync.WaitGroup
var muDir sync.Mutex

func incrementConnCntDir() {
	muDir.Lock()
	defer muDir.Unlock()
	connCntDir++
}

func appendToTargetInfo(info TargetInfo) {
	muDir.Lock()
	defer muDir.Unlock()
	targetInfo = append(targetInfo, info)
}

func init() {
	ew := &utils.EmptyWriter{}
	log.SetOutput(io.Writer(ew))
	rootCmd.AddCommand(dirCmd)
	dirCmd.Flags().IntP("thread", "r", 50, "单个目标线程数")
	dirCmd.Flags().StringP("url", "u", "", "目标url")
	dirCmd.Flags().StringP("file", "f", "", "目标url列表文件")
	dirCmd.Flags().IntP("timeout", "t", 10, "超时时间")
	dirCmd.Flags().StringP("proxy", "p", "", "代理地址")
	dirCmd.Flags().StringP("method", "m", "GET", "目录爆破请求的方法（GET可防止目录泛解析，HEAD更快）")
	dirCmd.Flags().StringP("level", "l", "1", "爆破等级（1为小字典爆破，2为中字典爆破，3为大字典爆破）")
	dirCmd.Flags().BoolP("admin", "a", false, "后台发现")
	dirCmd.Flags().BoolP("backup", "b", false, "备份文件发现（level:4，不使用字典，只做相关性扫描）")
	dirCmd.Flags().BoolP("sqlbackup", "s", false, "数据库备份快扫")
	dirCmd.Flags().BoolP("total", "o", false, "综合扫描")
	dirCmd.Flags().BoolP("webshell", "w", false, "webshell扫描")
	dirCmd.Flags().BoolP("phpmyadmin", "n", false, "phpmyadmin扫描")
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
		dirParameter.level, _ = cmd.Flags().GetString("level")
		dirParameter.scanMethod, _ = cmd.Flags().GetString("method")
		dirParameter.isBackstage, _ = cmd.Flags().GetBool("admin")
		dirParameter.isBackUp, _ = cmd.Flags().GetBool("backup")
		dirParameter.isSqlBack, _ = cmd.Flags().GetBool("sqlbackup")
		dirParameter.isTotal, _ = cmd.Flags().GetBool("total")
		dirParameter.isWebShell, _ = cmd.Flags().GetBool("webshell")
		dirParameter.isPhpMyAdmin, _ = cmd.Flags().GetBool("phpmyadmin")

		if dirParameter.isPhpMyAdmin {
			temp, err := utils.ReadLinesFromFile("dict/phpmyadmin" + dirParameter.level + ".txt")
			if err != nil {
				Ldir.Fatal("phpmyadmin字典解析失败")
			}
			dictionary = append(dictionary, temp...)
		}

		if dirParameter.isWebShell {
			temp, err := utils.ReadLinesFromFile("dict/webshell" + dirParameter.level + ".txt")
			if err != nil {
				Ldir.Fatal("webshell字典解析失败")
			}
			dictionary = append(dictionary, temp...)
		}

		if dirParameter.isTotal {
			temp, err := utils.ReadLinesFromFile("dict/total" + dirParameter.level + ".txt")
			if err != nil {
				Ldir.Fatal("综合字典解析失败")
			}
			dictionary = append(dictionary, temp...)
		}

		if dirParameter.isSqlBack {
			temp, err := utils.ReadLinesFromFile("dict/sqlbackup" + dirParameter.level + ".txt")
			if err != nil {
				Ldir.Fatal("数据库备份文件字典解析失败")
			}
			dictionary = append(dictionary, temp...)
		}

		if dirParameter.isBackstage {
			temp, err := utils.ReadLinesFromFile("dict/back" + dirParameter.level + ".txt")
			if err != nil {
				Ldir.Fatal("后台字典解析失败")
			}
			dictionary = append(dictionary, temp...)
		}
		if dirParameter.isBackUp {
			if dirParameter.level == "4" {
				Ldir.Debug("开始整合目标相关字典...")
			} else {
				temp, err := utils.ReadLinesFromFile("dict/backup" + dirParameter.level + ".txt")
				if err != nil {
					Ldir.Fatal("备份文件字典解析失败")
				}
				dictionary = append(dictionary, temp...)
			}
		}

		// 字典去重，防止扫描了多个一样的目标
		dictionary = utils.RemoveDuplicates(dictionary)

		if dirParameter.url != "" {
			if dirParameter.isBackUp {
				totalCntDir += len(dictionary) + 2*len(utils.BackUpFileList)
			} else {
				totalCntDir += len(dictionary)
			}

			Ldir.Info("成功加载字典： " + strconv.Itoa(totalCntDir) + " 条")
			CrackIt(dirParameter)
			if targetInfo != nil {
				SaveRes(utils.GetDomain(dirParameter.url))
			}
			return
		}
		if dirParameter.file != "" {
			CrackItsTarget(dirParameter)
			return
		}
	},
}

// ScanByTargetDict 扫描和域名相关的字典
func ScanByTargetDict(target string) []string {
	var addS []string
	for _, v := range utils.BackUpFileList {
		addS = append(addS, "/"+utils.GetDomain(target)+v)
		addS = append(addS, "/"+utils.GetDomainName(target)+v)
	}
	return addS
}

func CrackItsTarget(dirParameter DirParameter) {
	// 先对文件进行处理
	utils.ProcessSourceFile(dirParameter.file)
	result, err := utils.ReadLinesFromFile(dirParameter.file)
	if err != nil {
		Lf.Fatal("dir列表文件解析失败")
	}

	if dirParameter.isBackUp {
		totalCntDir += len(dictionary)*len(result) + len(result)*len(utils.BackUpFileList)*2
	} else {
		totalCntDir += len(dictionary) * len(result)
	}

	Ldir.Info("成功加载字典： " + strconv.Itoa(totalCntDir) + " 条")

	for _, url := range result {
		var dir2 DirParameter
		dir2 = dirParameter
		dir2.url = url
		CrackIt(dir2)
		// 有结果才保存文件
		if targetInfo != nil {
			SaveRes(utils.GetDomain(url))
		}
		utils.DelLine(dirParameter.file, url)
		targetInfo = nil // 重置targetInfo
	}
}

// CrackIt 爆破启动器
func CrackIt(dirParameter DirParameter) {
	var dicts []string

	if dirParameter.isBackUp {
		muDir.Lock()
		dicts = append(dictionary, ScanByTargetDict(dirParameter.url)...)
		muDir.Unlock()
	} else {
		muDir.Lock()
		dicts = dictionary
		muDir.Unlock()
	}

	targetTitle := "ahahahahahahahaha"

	// 先发一个探测title的包，防止目录泛解析
	ask := utils.Ask{}
	ask.Url = dirParameter.url
	ask.Proxy = dirParameter.proxy
	ask.Timeout = dirParameter.timeout
	resp := utils.OutsourcingByPwn(ask)
	var doc *goquery.Document
	var err error

	if resp != nil {
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if doc != nil {
			targetTitle = doc.Find("title").First().Text()
		}
		if err != nil {
			Ldir.Error(dirParameter.url + " 目标节点解析失败")
		}
	}

	urlChan := make(chan string, len(dicts))
	soldiers := 0 // 排雷兵，用来检测是不是被目标ban了
	isBan := false
	for i := 0; i < dirParameter.thread; i++ {
		wgDir.Add(1)
		go func() {
			defer wgDir.Done()
			for dict := range urlChan {
				// IP被ban了或者网络波动，停止扫描当前目标
				if isBan {
					incrementConnCntDir()
					continue
				}

				dirPage := utils.DelExtraSlash(dirParameter.url + dict)
				ask2 := utils.Ask{}
				ask2.Url = dirPage
				ask2.Proxy = dirParameter.proxy
				ask2.Timeout = dirParameter.timeout
				ask2.Method = dirParameter.scanMethod
				resp2 := utils.OutsourcingByPwn(ask2)

				if resp2 == nil {
					incrementConnCntDir()
					muDir.Lock()
					soldiers++
					muDir.Unlock()
					// 连续连接失败次数达到阈值
					if soldiers > utils.SoldiersThreshold {
						isBan = true
						Ldir.Error(dirParameter.url + " 已经被ban，停止扫描")
						return
					}
				} else {
					incrementConnCntDir()
					muDir.Lock()
					soldiers = 0 // 重置排雷兵
					muDir.Unlock()

					// 防止resp2指针移动到末尾导致无法读取的情况
					body, _ := io.ReadAll(resp2.Body)
					resBody := ioutil.NopCloser(bytes.NewReader(body))
					isValid, nowTitle := IsValid(dirParameter, resp2, targetTitle, resBody)

					if isValid {
						Ldir.Info(getProgress() + dirParameter.url + " 发现目录 " + dirPage + " 状态码 " + strconv.Itoa(resp2.StatusCode))

						htmlContent := string(body)

						// size = 0 ,排除
						if len(htmlContent) == 0 {
							return
						}

						info := TargetInfo{
							size:  strconv.Itoa(len(htmlContent)),
							url:   dirPage,
							code:  strconv.Itoa(resp2.StatusCode),
							title: nowTitle,
						}
						appendToTargetInfo(info)
					} else {
						Ldir.Fatal(getProgress() + dirParameter.url + " 不存在该目录 " + dirPage + " 状态码 " + strconv.Itoa(resp2.StatusCode))
					}

				}
			}
		}()
	}
	// 将url发送到urlChan供消费者goroutine处理
	for _, dict := range dicts {
		urlChan <- dict
	}
	close(urlChan)
	wgDir.Wait()
}

// 进度条前缀输出
func getProgress() string {
	return "[" + strconv.Itoa(connCntDir) + "/" + strconv.Itoa(totalCntDir) + "] "
}

func IsValid(dirParameter DirParameter, resp *http.Response, targetTitle string, respBody io.ReadCloser) (bool, string) {

	if resp.StatusCode != 200 {
		return false, ""
	}

	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return false, ""
	}

	nowTitle := doc.Find("title").First().Text()

	// 防止目录泛解析
	if nowTitle == targetTitle {
		return false, ""
	}

	for _, v := range utils.NotFoudList {
		if strings.Contains(nowTitle, v) {
			return false, ""
		}
	}

	// PHPmyadmin进一步检验
	if dirParameter.isPhpMyAdmin {
		nowTitleLower := strings.ToLower(nowTitle)
		if !strings.Contains(nowTitleLower, "phpmyadmin") {
			return false, ""
		}
	}

	return true, nowTitle
}

// SaveRes 保存结果到CSV文件中
func SaveRes(filename string) {
	data := make([][]string, len(targetInfo)+1)
	titles := []string{"URL", "Code", "Title", "Size"}
	data[0] = titles
	for i := 0; i < len(targetInfo); i++ {
		data[i+1] = []string{targetInfo[i].url, targetInfo[i].code, targetInfo[i].title, targetInfo[i].size}
	}
	utils.WriteCsvByName("dir", filename, data)
}
