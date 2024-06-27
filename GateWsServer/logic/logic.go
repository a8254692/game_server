package logic

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"reflect"

	"github.com/go-ini/ini"
)

// 连接与服务管理器
var WsManager *network.Ws_Manager

// 连接与服务管理器
var TcpMananger *GateWsTcp_Mananger

// 服务器配置ini
var ConfigServer *ini.File

func Init() {
	//注册系统公共事件
	event.Register("ExecuteCommond", reflect.ValueOf(ExecuteCommond))
	TcpMananger = new(GateWsTcp_Mananger)
	TcpMananger.Init()

	//连接游戏服，所有的连接要在业务模块初始化之后执行
	ipGameServer := ConfigServer.Section("gameserver").Key("ip").String()
	portGameServer := ConfigServer.Section("gameserver").Key("port").String()
	ConnectGame := new(network.TcpConnect)
	ConnectGame.Init(ipGameServer+":"+portGameServer, TcpMananger.ServerTypeLocal, network.ServerType_Game)
	WsManager.ConnectGame = ConnectGame
}
func SetConnectManager(cm *network.Ws_Manager) {
	WsManager = cm
}
func SetConfig(cm *ini.File) {
	ConfigServer = cm
}

// 启动协程调用函数
func GoFunc(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
				stack.PrintCallStack()
			}
		}()

		f()
	}()
}

// 处理回调消息
func OnCallBackFun(f func(param interface{})) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()

	f(nil)
}

// 停止逻辑模块
func StopLogic() {
	WsManager.StopServer()

}

// 执行指令
func ExecuteCommond(pCommand *string) {
	command := *pCommand
	if command == "PrintMsg" {
		//PrintPlayerMsg(nil)
		//event.PrintEventMsg(nil)
	} else if command == "Startpprof" {
		//开启runtime/pprof
		//pprof.Startpprof()
	} else if command == "Endpprof" {
		//关闭runtime/pprof
		//pprof.Endpprof()
	} else if command == "ClearNotice" {
		//NoticeManager.Release()
		log.Info("清除公告信息成功")
	} else if command == "Line" {
		//显示在线人数
		log.Info("在线:", WsManager.SizeConnect, "/", WsManager.SizeConnect)
	} else if command == "PrintCall" {
		//打印函数性能
		//PrintPlayerCallData(&command)
	} else if command == "PrintTime" {
		//打印定时器性能
		//timer.PrintPerfromance()
	} else if command == "JJCJL" {
		//打印函数性能
		//SortManager.SendJingJiChangRewardByTime(time.Now())
	} else if command == "UpdateServerConfig" {
		//刷新服务器配置信息
		//GServerManager.Update()
	} else {
		log.Error("不存在指令：", command)
	}
}
