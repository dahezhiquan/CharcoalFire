package utils

import (
	"bufio"
	"github.com/gookit/color"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteFile 向文件中写入
func WriteFile(path string, content []string) {
	err := os.MkdirAll(ResultLogName+"/"+path, 0755)
	if err != nil {
		color.Error.Println("日志文件创建失败")
	}

	// 构建文件名
	currentTime := time.Now().Format("20060102150405")
	filePath := filepath.Join(ResultLogName, path, currentTime+".txt")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		color.Error.Println("日志文件打开失败")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			color.Warn.Println("日志文件未正常关闭")
		}
	}(file)

	for _, line := range content {
		_, err = file.WriteString(line + "\n")
		if err != nil {
			color.Error.Println("结果写入日志失败")
		}
	}
}

// ClearFile 清空文件内容
func ClearFile(path string) {
	currentTime := time.Now().Format("20060102")
	filePath := filepath.Join(ResultLogName, path, currentTime+".txt")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		color.Error.Println("日志文件打开失败")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			color.Warn.Println("日志文件未正常关闭")
		}
	}(file)

	err = file.Truncate(0)
	if err != nil {
		color.Error.Println("清空文件失败")
	}
}

// ReadLinesFromFile 读取一个文件，返回该文件内容的切片
func ReadLinesFromFile(filename string) ([]string, error) {
	var result []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// ProcessSourceFile 对目标文件的内容提炼，目前实现
// 1.去重
// 2.去除空行
// domain转url
func ProcessSourceFile(path string) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		color.Error.Println("文件打开失败")
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			color.Warn.Println("目标文件未正常关闭")
		}
	}(file)

	uniqueLines := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// 保证line未url格式，加入前缀
			if IsDoamin(line) {
				httpUrl, httpsUrl := AddPrefix(line)
				uniqueLines[httpUrl] = true
				uniqueLines[httpsUrl] = true
			} else {
				uniqueLines[line] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		color.Error.Println("读取文件失败")
		return
	}

	if err := file.Truncate(0); err != nil {
		color.Error.Println("清空文件失败")
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		color.Error.Println("重定位文件指针失败")
		return
	}

	writer := bufio.NewWriter(file)
	for line := range uniqueLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			color.Error.Println("文件写入失败")
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		return
	}
}

// AddPrefix domain加前缀
func AddPrefix(url string) (string, string) {
	httpURL := "http://" + url
	httpsURL := "https://" + url
	return httpURL, httpsURL
}
