package utils

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

type Ask struct {
	Url     string
	Timeout int
	Proxy   string
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
