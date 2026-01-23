package main

import (
	"github.com/cy77cc/k8s-manage/internal/cmd"
	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load(".env")
	cmd.Execute()
}
