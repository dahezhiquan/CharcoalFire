package utils

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

// 解析HTML文档，返回各个节点内容

type HtmlDocument struct {
	Title string
	Icon  string
}

var Lparsehtml = GetSlog("parse")

func GetHtmlDocument(url string, resp *http.Response) (htmlDocument HtmlDocument) {

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		Lparsehtml.Error(url + " 目标节点解析失败")
		return HtmlDocument{}
	}

	title := doc.Find("title").First().Text()
	htmlDocument.Title = title

	// ICON提取
	iconURL := ""
	found := false
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		if !found && strings.Contains(strings.ToLower(rel), "icon") {
			href, _ := s.Attr("href")
			iconURL = href
			found = true
		}
	})
	htmlDocument.Icon = iconURL

	return htmlDocument
}
