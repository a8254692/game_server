package logic

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/rabbitMQ"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
	"reflect"

	"github.com/go-ini/ini"
	"google.golang.org/protobuf/proto"
)

// 连接与服务管理器
var ConnectManager network.Connect_Manager

// 服务器配置ini
var ConfigServer *ini.File

var Mq *rabbitMQ.RabbitMQ

func Init() {
	//注册系统公共事件
	event.Register("ExecuteCommond", reflect.ValueOf(ExecuteCommond))

	mqHost := ConfigServer.Section("rabbitmq").Key("host").String()
	mqCli, err := rabbitMQ.NewRabbitMQSimple(mqHost, "admin")
	if err != nil || mqCli == nil {
		log.Waring("-->logic--Init--rabbitmq--start--err:", err)
	}
	if mqCli != nil {
		Mq = mqCli
	}

	AdminMsg.Init()
}

func SetConnectManager(cm *network.Connect_Manager) {
	ConnectManager = *cm
}

func SetConfig(cm *ini.File) {
	ConfigServer = cm
}

func UnmarshalMsgBody(buff []byte, paramType reflect.Type) (interface{}, error) {
	msg := reflect.New(paramType).Interface()
	return msg, proto.Unmarshal(buff, msg.(proto.Message))
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
	ConnectManager.StopServer()

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
		log.Info("在线:", ConnectManager.SizeConnect, "/", ConnectManager.SizeConnect)
	} else if command == "PrintCall" {
		//打印函数性能
		//PrintPlayerCallData(&command)
	} else if command == "PrintTime" {
		//打印定时器性能
		timer.PrintPerfromance()
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
