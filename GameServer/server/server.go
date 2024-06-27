// server
package server

import (
	"BilliardServer/Util/tools"
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"BilliardServer/Common"
	"BilliardServer/GameServer/logic"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/pprof"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"

	"github.com/go-ini/ini"
)

// 服务器开始
func Start(args []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()

	serverName := "gameserver"

	var mode string
	if len(args) > 0 {
		mode = tools.GetArgsMode(args)
	}
	if mode == "" {
		mode = Common.ModeLocal
	}
	cfgPath := tools.GetModeConfPath(mode)

	config, err := ini.Load(cfgPath + Common.CfgFileName)
	if err != nil {
		//失败
		fmt.Printf("-->Fail to read server.ini file:%v\n", err)
		os.Exit(1)
	}

	portServer := config.Section(serverName).Key("port").String()
	serverID := serverName + "-" + portServer

	log.Init(serverID, mode)
	log.Print("-->", serverName, " Start--------------------")
	if portServer == "" {
		log.Print(fmt.Sprintf("-->empty to read server.ini portServer:%v\n", portServer))
		os.Exit(1)
	}
	log.Print(fmt.Sprintf("-->success to read server.ini portServer:%v\n", portServer))

	//开启runtime/pprof性能统计
	intServerID, _ := strconv.Atoi(portServer)
	intServerID += 1
	if intServerID > 10000 {
		intServerID %= 10000
	}
	if intServerID < 1000 {
		intServerID += 7000
	}
	pprofPort := strconv.Itoa(intServerID)
	pprof.Init(pprofPort)
	log.Print("-->", serverName, " start pprof on ", pprofPort)

	//判断是否重复启动
	l, err1 := net.Listen("tcp", ":"+portServer)
	if err1 != nil {
		log.Error(err1)
		return
	} else {
		l.Close()
	}

	//初始化事件管理器
	event.Init(serverName, 2000, doStop)
	//初始化计时器
	timer.Init(serverName)
	timer.Star()

	//开始接收命令
	log.Print("-->Server System is ", runtime.GOOS, "-------------------")
	if runtime.GOOS == "windows" {
		StartCommond()
	}
	go Singallisten()

	//启动服务
	cm := new(network.Connect_Manager)
	cm.InitServer(":"+portServer, network.ServerType_Game)
	logic.SetConnectManager(cm)
	err = cm.StartServer()
	if err != nil {
		log.Waring("Connect_Manager err；", err)
		return
	}
	//初始化逻辑模块
	logic.SetConfig(config)
	logic.Init()
	//开始运行逻辑
	Run()
}

// 服务器运行
func Run() {
	log.Print("-->Server Is On Running-------------------")

	if !event.Run() {
		log.Error("-->事件发生错误，停止运行")
	}

	//逻辑停止
	doStop()
}

// 服务器关闭
func Stop() {
	pprof.Endpprof()

	//关闭事件通道，停止逻辑
	log.Print("-->关闭事件运行")
	event.Stop()
	logic.StopLogic()
	time.AfterFunc(time.Minute, doStop)

	log.Print("-->Server Stop---------------------")
}

var stopOnce sync.Once  // 退出逻辑只执行一次
var stopLock sync.Mutex // 退出锁(防止主线程比执行退出逻辑的线程先退出)

// 逻辑真正停止
func doStop() {
	stopLock.Lock()
	defer stopLock.Unlock()

	stopOnce.Do(logic.StopLogic)
}

// 监听停止信号
func Singallisten() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	for {
		sig := <-signalChan
		log.Print("get signal:", sig)
		Stop()
		break
	}
}

// 开始接收命令
func StartCommond() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()
		reader := bufio.NewReader(os.Stdin)
		for {
			data, _, _ := reader.ReadLine()
			command := string(data)
			if command == "s" {
				break
			}

			event.Fire("ExecuteCommond", &command)
		}
		Stop()
	}()
}
