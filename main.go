package main

import (
	"CharcoalFire/cmd"
	"CharcoalFire/utils"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
	l := utils.GetSlog("icon")
	l.Info("你好啊")
}
