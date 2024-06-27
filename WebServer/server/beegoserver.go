package server

import (
	"BilliardServer/WebServer/routers"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"os"
	"runtime"

	"BilliardServer/WebServer/utils"
	"BilliardServer/WebServer/utils/path"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func StartWebServer(args []string) {
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"access-token", "a-auth-token", "x-auth-token", "Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"access-token", "Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	//设置静态文件下载地址
	web.SetStaticPath("/download", path.Download_Path)

	if runtime.GOOS == "windows" {
		os.Setenv("ZONEINFO", "./date.zip")
	}

	//初始化数据库
	dbPath, _ := web.AppConfig.String("mongodb::dbPath")
	err := utils.InitMongoDB(dbPath)
	if err != nil {
		fmt.Printf("-->err InitMongoDB :%s\n", err.Error())
		return
	}

	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	err = logs.SetLogger(logs.AdapterFile, `{"filename":"logdata/WebServer.log","level":7,"maxlines":100000,"maxsize":0,"daily":true,"maxdays":90,"color":true}`)
	if err != nil {
		fmt.Printf("-->err to logs SetLogger :%s\n", err.Error())
		return
	}

	//授权登录中间件
	//middleware.AuthMiddle()

	web.AddNamespace(routers.GetNsRouter())

	web.Run()
}
