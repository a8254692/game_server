// main
package main

import (
	"BilliardServer/DBServer/server"
	"os"
	"runtime"
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

	//开启服务器
	server.Start(os.Args)
}
