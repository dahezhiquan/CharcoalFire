package utils

import (
	"regexp"
)

// 正则判断器

type RegexpTool struct {
	pattern string
}

func NewRegexpTool(pattern string) *RegexpTool {
	return &RegexpTool{
		pattern: pattern,
	}
}

// IsMatch 判断字符串是否匹配正则表达式
func (rt *RegexpTool) IsMatch(input string) bool {
	match, _ := regexp.MatchString(rt.pattern, input)
	return match
}

// FindString 查找匹配的子字符串
func (rt *RegexpTool) FindString(input string) string {
	re := regexp.MustCompile(rt.pattern)
	return re.FindString(input)
}

// IsUrl 判断是否为url
func IsUrl(url string) bool {
	tool := NewRegexpTool(`^(http|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?$`)
	return tool.IsMatch(url)
}

// IsDoamin 判断是否为域名
func IsDoamin(url string) bool {
	tool := NewRegexpTool(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$`)
	return tool.IsMatch(url)
}
