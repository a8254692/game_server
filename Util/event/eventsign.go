package event

import (
	gmsg "BilliardServer/Proto/gmsg"
	"reflect"
	"strconv"
	"strings"
)

var EventSign struct {
	emName    string
	emPointer *EventManager
}

// 初始化 名称，通道长度，停止回调函数
func Init(nameEM string, enentChanLength int, stopFunc func()) {
	EventSign.emPointer = new(EventManager)
	EventSign.emName = nameEM
	EventSign.emPointer.Init(nameEM, enentChanLength, stopFunc)
}

// 开始运行
func Run() (result bool) {
	return EventSign.emPointer.Run()
}

// 停止
func Stop() {
	EventSign.emPointer.Stop()
}

// 是否注册了此事件
func IsExist(funcName string) bool {
	return EventSign.emPointer.IsExist(funcName)
}

// 是否注册了此网络消息码事件
func IsExistMsgTile(msgTile uint32) bool {
	return EventSign.emPointer.IsExist(ToEventMsgTile(msgTile))
}

// 侦听事件
func On(funcName string, funVal reflect.Value) {
	EventSign.emPointer.Register(funcName, funVal)
}

// 侦听事件 只针对网络消息码
func OnNet(msgTile gmsg.MsgTile, funVal reflect.Value) {
	EventSign.emPointer.Register(ToEventMsgTile((uint32(msgTile))), funVal)
}

// 发送事件
func Emit(funcName string, data interface{}) bool {
	return EventSign.emPointer.Fire(funcName, data)
}

// 发送事件 只会对网络消息码
func EmitNet(msgTile uint32, data interface{}) bool {
	return EventSign.emPointer.Fire(ToEventMsgTile(msgTile), data)
}

// 侦听事件
func Register(funcName string, funVal reflect.Value) {
	EventSign.emPointer.Register(funcName, funVal)
}

// 发送事件
func Fire(funcName string, data interface{}) bool {
	return EventSign.emPointer.Fire(funcName, data)
}

// 打印事件列表
func PrintEventMsg(commond *string) {
	EventSign.emPointer.PrintEventMsg(commond)
}

// 获得最后调用
func GetLastTimeEvent() string {
	return EventSign.emPointer.LastTimeEvent
}

// 设置最后调用
func SetLastTimeEvent(eventName string) {
	EventSign.emPointer.LastTimeEvent = eventName
}

// 获得注册的网络消息列表
func GetMsgTileList() []uint32 {
	return EventSign.emPointer.GetMsgTileList()
}

func ToEventMsgTile(msgTile uint32) string {
	return "Msg_" + strconv.Itoa(int(msgTile))
}
func ToNetMsgTile(eMsgTile string) uint32 {
	tileNameArgs := strings.Split(eMsgTile, "_")
	msgTile, err := strconv.Atoi(tileNameArgs[1])
	if err != nil {
		return 0
	}
	return uint32(msgTile)
}
