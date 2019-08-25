package main

import (
	"AutoSS/lib"
	"log"
	"os/exec"
	"time"
)

func main() {
	log.Println("正在采集账号...")
	g, _ := lib.NewConfig("gui-config.json", "rule.json")
	g.GetURLs()
	g.Save()
	reload()
}

func reload() {
	log.Println("正在重新启动...")
	exec.Command("taskkill", "/im", "Shadowsocks.exe", "/f").Run()
	time.Sleep(time.Second * 2)
	exec.Command("Shadowsocks.exe").Start()
}
