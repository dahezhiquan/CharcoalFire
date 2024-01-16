package utils

import (
	"bufio"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var Lw = GetSlog("file")

// WriteFile 向文件中写入
func WriteFile(path string, content []string, isDoamin bool) {
	err := os.MkdirAll(ResultLogName+"/"+path, 0755)
	if err != nil {
		Lw.Fatal(path + " 文件创建失败")
	}

	// 构建文件名
	currentTime := time.Now().Format("20060102150405")
	filePath := filepath.Join(ResultLogName, path, currentTime+".txt")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Lw.Fatal(filePath + " 文件打开失败")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			Lw.Warning(filePath + " 文件未正常关闭")
		}
	}(file)

	for _, line := range content {
		// 判断是否输出为domain格式
		if !isDoamin {
			_, err = file.WriteString(line + "\n")
		} else {
			u, _ := url.Parse(line)
			domain := u.Hostname()
			_, err = file.WriteString(domain + "\n")
		}
		if err != nil {
			Lw.Fatal(filePath + " 结果写入文件失败")
		}
	}
}

// WriteFileBySuffix 向文件中写入 包含后缀信息
func WriteFileBySuffix(path string, content []string, suffix []string) {
	err := os.MkdirAll(ResultLogName+"/"+path, 0755)
	if err != nil {
		Lw.Fatal(path + " 文件创建失败")
	}

	// 构建文件名
	currentTime := time.Now().Format("20060102150405")
	filePath := filepath.Join(ResultLogName, path, currentTime+".txt")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Lw.Fatal(filePath + " 文件打开失败")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			Lw.Warning(filePath + " 文件未正常关闭")
		}
	}(file)

	for i, line := range content {
		// 判断是否输出为domain格式
		_, err = file.WriteString(line + " " + suffix[i] + "\n")

		if err != nil {
			Lw.Fatal(filePath + " 结果写入文件失败")
		}
	}
	Lw.Info("结果已保存到：" + filePath)
}

// ClearFile 清空文件内容
func ClearFile(path string) {
	currentTime := time.Now().Format("20060102")
	filePath := filepath.Join(ResultLogName, path, currentTime+".txt")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		Lw.Fatal(filePath + " 文件打开失败")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			Lw.Warning(filePath + " 文件未正常关闭")
		}
	}(file)

	err = file.Truncate(0)
	if err != nil {
		Lw.Fatal(filePath + " 清空文件失败")
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
// 3.domain转url
// 4.CIDR格式的支持
func ProcessSourceFile(path string) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		Lw.Fatal(path + " 文件打开失败")
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			Lw.Warning(path + " 文件未正常关闭")
		}
	}(file)

	uniqueLines := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// 保证line为url格式，加入前缀
			if IsDoamin(line) {
				httpUrl, httpsUrl := AddPrefix(line)
				uniqueLines[httpUrl] = true
				uniqueLines[httpsUrl] = true
			} else if IsCIDR(line) {
				// CIDR转换
				ip, ipnet, _ := net.ParseCIDR(line)
				var ips []string
				for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
					ips = append(ips, ip.String())
				}
				for _, ip := range ips {
					httpUrl, httpsUrl := AddPrefix(ip)
					uniqueLines[httpUrl] = true
					uniqueLines[httpsUrl] = true
				}
			} else if IsIpAddr(line) {
				httpUrl, httpsUrl := AddPrefix(line)
				uniqueLines[httpUrl] = true
				uniqueLines[httpsUrl] = true
			} else {
				uniqueLines[line] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		Lw.Fatal(path + " 读取文件失败")
		return
	}

	if err := file.Truncate(0); err != nil {
		Lw.Fatal(path + " 清空文件失败")
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		Lw.Fatal(path + " 重定位文件指针失败")
		return
	}

	writer := bufio.NewWriter(file)
	for line := range uniqueLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			Lw.Fatal(path + " 文件写入失败")
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		return
	}
}

// CIDR解析的支持方法
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// AddPrefix domain加前缀
func AddPrefix(url string) (string, string) {
	httpURL := "http://" + url
	httpsURL := "https://" + url
	return httpURL, httpsURL
}

// DelExtraSlash 去除多余的斜杠
func DelExtraSlash(url string) string {
	re := regexp.MustCompile(`/{2,}`)
	result := re.ReplaceAllString(url, "/")
	prefixes := []string{"http:/", "https:/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(result, prefix) {
			result = strings.Replace(result, prefix, prefix+"/", 1)
			break
		}
	}
	return result
}
