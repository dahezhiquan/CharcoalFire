package utils

import (
	"bufio"
	"github.com/gookit/color"
	"os"
	"path/filepath"
	"time"
)

// WriteFile 向文件中写入
func WriteFile(path string, content []string) {
	err := os.MkdirAll(ResultLogName+"/"+path, 0755)
	if err != nil {
		color.Error.Println("日志文件创建失败")
	}

	// 构建文件名
	currentTime := time.Now().Format("20060102")
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
