package main

import (
	"CharcoalFire/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
