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
		return "\033[35m" + type2 + "\033[0m" // 紫色
	}
	return type2
}
