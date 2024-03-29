package utils

import (
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"time"
)

// 空写入器，用于日志报错屏蔽

type EmptyWriter struct{}

func (ew *EmptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// GetSlog 自定义模板日志写入
func GetSlog(type2 string) *slog.Logger {
	// 获取主题皮肤
	type2 = GetSkin(type2)
	currentTime := time.Now()
	currentTimeStr := "\033[34m" + currentTime.Format("15:04:05") + "\033[0m" // 蓝色
	myTemplate := "[{{" + currentTimeStr + "}}] [{{" + type2 + "}}] [{{level}}] {{message}}\n"
	h := handler.NewConsoleHandler(slog.AllLevels)
	h.Formatter().(*slog.TextFormatter).SetTemplate(myTemplate)
	l := slog.NewWithHandlers(h)
	return l
}

func GetSkin(type2 string) string {
	if type2 == "icon" {
		return "\033[34m" + type2 + "\033[0m" // 蓝色
	}
	if type2 == "survive" {
		return "\033[95m" + type2 + "\033[0m" // 粉紫色（洋红色）
	}
	if type2 == "net" {
		return "\033[33m" + type2 + "\033[0m" // 黄色
	}
	if type2 == "file" {
		return "\033[90m" + type2 + "\033[0m" // 灰色
	}
	if type2 == "parse" {
		return "\033[97m" + type2 + "\033[0m" // 亮灰色
	}
	if type2 == "fisher" {
		return "\033[35m" + type2 + "\033[0m" // 紫色
	}
	if type2 == "cdncheck" {
		return "\033[36m" + type2 + "\033[0m" // 青色
	}
	if type2 == "dir" {
		return "\033[37m" + type2 + "\033[0m" // 白色
	}
	if type2 == "exp" {
		return "\033[31m" + type2 + "\033[0m" // 红色
	}
	return type2
}
