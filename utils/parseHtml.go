package utils

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

// 解析HTML文档，返回各个节点内容

type HtmlDocument struct {
	Title string
}

func ParseHtml(resp *http.Response) (htmlDocument HtmlDocument) {

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return HtmlDocument{}
	}

	title := doc.Find("title").First().Text()

	htmlDocument.Title = title

	return htmlDocument
}
