package main

import (
	"os"
	"runtime"

	"BilliardServer/WebServer/server"
)

func main() {
	//设置核心数
	num := runtime.NumCPU()
	if runtime.GOOS == "windows" {
		num = runtime.NumCPU() / 2
	}
	if num <= 0 {
		num = 1
	}
	runtime.GOMAXPROCS(num)

	//开启服务
	server.StartWebServer(os.Args)
}
