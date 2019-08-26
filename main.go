package main

import (
	"AutoSS/collection"
	"log"
	"os/exec"
	"time"
)

func main() {

	log.Println("正在采集账号...")
	g, _ := collection.NewConfig("gui-config.json", "rule.json")
	n, err := g.GetURLs()
	if err != nil {
		log.Println("采集出错", err)
		return
	}
	log.Printf("采集了 %d 个账号", n)
	g.Save()
	reload()
}

func reload() {
	log.Println("正在重新启动...")
	exec.Command("taskkill", "/im", "Shadowsocks.exe", "/f").Run()
	time.Sleep(time.Second * 2)
	exec.Command("Shadowsocks.exe").Start()
}
