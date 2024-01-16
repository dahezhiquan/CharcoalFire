package utils

import (
	"net/url"
	"strings"
)

// 格式化器

// GetDomain 提取url的domain
func GetDomain(urlp string) string {
	parsedURL, err := url.Parse(urlp)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

// GetSuffix 提取文件后缀
func GetSuffix(path string) string {
	splitResult := strings.Split(path, "?")
	lastDotIndex := strings.LastIndex(splitResult[0], ".")
	substringAfterLastDot := path[lastDotIndex+1:]
	res := strings.Split(substringAfterLastDot, "?")
	return res[0]
}

// GetLanVersion X-Powered-By字段提取出脚本语言+版本号
func GetLanVersion(total string) []string {
	res := strings.Split(total, "/")
	return res
}
