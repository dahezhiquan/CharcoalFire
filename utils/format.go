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
	lastDotIndex := strings.LastIndex(path, ".")
	if lastDotIndex != -1 && lastDotIndex < len(path)-1 {
		substringAfterLastDot := path[lastDotIndex+1:]
		return substringAfterLastDot
	} else {
		return ""
	}
}
