package utils

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Ask struct {
	Url     string
	Timeout int
	Proxy   string
	Data    map[string]string
	Header  map[string]string
	Method  string
}

var Ln = GetSlog("net")

// Outsourcing 发包工具
func Outsourcing(ask Ask) *http.Response {
	var client *http.Client
	// 不管是否使用了代理，都先按不使用代理发包
	client = &http.Client{
		Timeout:   time.Duration(ask.Timeout) * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	resp, err := client.Get(ask.Url)
	var flag = false // 代理是否能否访问目标的标记
	if err != nil {

		// 如果设置了代理，则再尝试使用代理访问
		if ask.Proxy != "" {
			proxyURL, err2 := url.Parse(ask.Proxy)
			if err2 != nil {
				Ln.Fatal(ask.Proxy + " 代理解析失败")
				return nil
			}
			client = &http.Client{
				Timeout: time.Duration(ask.Timeout) * time.Second,
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(proxyURL),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			resp, err2 = client.Get(ask.Url)
			if err2 == nil {
				flag = true
			}
		}
		if !flag {
			Ln.Fatal(ask.Url + " 目标连接失败")
			return nil
		}
	}
	return resp
}

func OutsourcingByPwn(ask Ask) *http.Response {
	var client *http.Client
	// 不管是否使用了代理，都先按不使用代理发包
	client = &http.Client{
		Timeout:   time.Duration(ask.Timeout) * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	// 构建要发送的数据data
	form := url.Values{}
	for key, value := range ask.Data {
		form.Add(key, value)
	}
	payload := strings.NewReader(form.Encode())

	// 默认使用GET请求方法
	if ask.Method == "" {
		ask.Method = "GET"
	}

	req, err := http.NewRequest(ask.Method, ask.Url, payload)
	if err != nil {
		Ln.Fatal(ask.Url + " 构建请求失败")
		return nil
	}

	// 随机UA头
	req.Header.Set("User-Agent", GetUA())

	// 添加自定义的Header
	for key, value := range ask.Header {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	var flag = false // 代理是否能否访问目标的标记
	if err != nil {

		// 如果设置了代理，则再尝试使用代理访问
		if ask.Proxy != "" {
			proxyURL, err2 := url.Parse(ask.Proxy)
			if err2 != nil {
				Ln.Fatal(ask.Proxy + " 代理解析失败")
				return nil
			}
			client = &http.Client{
				Timeout: time.Duration(ask.Timeout) * time.Second,
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(proxyURL),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			resp, err2 = client.Do(req)
			if err2 == nil {
				flag = true
			}
		}
		if !flag {
			Ln.Fatal(ask.Url + " 目标连接失败")
			return nil
		}
	}
	return resp
}

// GetUA UA生产池
func GetUA() string {
	firstNum := rand.Intn(8) + 55
	thirdNum := rand.Intn(3201)
	fourthNum := rand.Intn(141)
	osType := []string{
		"(Windows NT 6.1; WOW64)",
		"(Windows NT 10.0; WOW64)",
		"(Macintosh; Intel Mac OS X 10_12_6)",
	}
	chromeVersion := fmt.Sprintf("Chrome/%d.0.%d.%d", firstNum, thirdNum, fourthNum)

	ua := fmt.Sprintf("Mozilla/5.0 %s AppleWebKit/537.36 (KHTML, like Gecko) %s Safari/537.36", randomChoice(osType), chromeVersion)

	return ua
}

func randomChoice(options []string) string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(options))
	return options[index]
}
