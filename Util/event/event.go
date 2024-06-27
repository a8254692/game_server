package event

import (
	"BilliardServer/Util/log"
	"reflect"
)

// 事件对象
type Event struct {
	HadleFunc string      //事件处理函数名字
	Data      interface{} //事件
	FireFile  string      //事件派发信息，用于统计派发数量，派发失败后显示派发源
}

// 玩家消息路由
type PlayerMsgRouter struct {
	MsgMap map[string]reflect.Type
}

var playerMsgRouter PlayerMsgRouter

// 注册一个消息
func RegisterPlayerMsg(method string, msgType reflect.Type) {
	if playerMsgRouter.MsgMap[method] != nil {
		log.Error("重复注册消息 funname = ", method)
		return
	}

	playerMsgRouter.MsgMap[method] = msgType
}

// 根据函数名称获取参数类型
func GetMethodParam(method string) reflect.Type {
	return playerMsgRouter.MsgMap[method]
}
