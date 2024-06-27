// network
package network

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"encoding/binary"
	"encoding/json"
	"reflect"

	proto "google.golang.org/protobuf/proto"
)

//Connect_Manager的声明对象
//var ConnectManager Connect_Manager

// 初始化网络
func Init(addr string) error {

	//开启Tcp服务
	// if addr != "" {
	// 	ConnectManager.InitServer(addr, ServerType_Gate)
	// 	err := ConnectManager.Server.Start(addr)
	// 	if err != nil {
	// 		log.Error(err)
	// 		return err
	// 	}
	// 	go ConnectManager.Server.Run()
	// }
	return nil
}

// 处理一个新的连接
func OnAcceptLink(newLink Link) error {
	//ID计数器自增
	// ConnectManager.linkIDCounter = ConnectManager.linkIDCounter + 1
	// //设置连接ID
	// newLink.SetID(ConnectManager.linkIDCounter)

	// //开始运行连接
	// go newLink.Run()

	// ConnectManager.mapLink[newLink.GetID()] = newLink

	return nil
}

// 关闭一个连接
func OnCloseLink(link Link) {
	link.Close()
	//delete(ConnectManager.mapLink, link.GetID())
}

// 解析json消息吗
func MarshalJsonMsg(data []byte, link Link) {
	//取出函数长度
	funNameLen := binary.LittleEndian.Uint16(data[0:2])

	//取出函数名
	funName := string(data[2 : funNameLen+2])

	//取出改函数注册的
	paramType := event.GetMethodParam(funName)
	if paramType == nil {
		return
	}

	if len(data) > 10*1024 {
		log.Error("************消息长度超过10K funName = ", funName)
	}

	//解析消息结构
	msg := reflect.New(paramType).Interface()
	err := json.Unmarshal(data[funNameLen+2:], &msg)
	if err != nil {
		log.Error("客户端发来消息错误 funname = ", funName, " err = ", err)
		return
	}

	//事件派发
	if link.IsLinkDrive() {
		logicEv := new(BackEndLogicEvent)
		logicEv.CallMethod = funName
		logicEv.Msg = msg
		logicEv.BackEndLink = link

		event.Fire("OnLegacyBackEndMsg", logicEv)
	} else {
		logicEv := new(PlayerLogicEvent)
		logicEv.CallMathod = funName
		logicEv.Msg = msg
		logicEv.PlayerLink = link
		event.Fire("OnPlayerMsg", logicEv)
	}
}

// 解析PB消息吗
func MarshalProtobufMsg(data []byte, link Link) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	//判断基础长度
	if len(data) < 2 {
		return
	}

	//取出函数长度
	funNameLen := binary.LittleEndian.Uint16(data[0:2])

	//跳过自身两字节
	funNameLen += 2
	//判断后续长度
	if len(data) < int(funNameLen) {
		return
	}
	//取出函数名
	funName := string(data[2:funNameLen])

	//取出消息类型
	paramType := event.GetMethodParam(funName)
	if paramType == nil {
		log.Waring("获取消息类型失败 funName = ", funName)
		return
	}

	if len(data) > 10*1024 {
		log.Error("************消息长度超过10K funName = ", funName)
	}

	//跳过后续数据长度两个字节
	funNameLen += 2
	//对数据进行pb解码
	msg := reflect.New(paramType).Interface()
	err := proto.Unmarshal(data[funNameLen:], msg.(proto.Message))
	if err != nil {
		log.Waring(err)
		log.Waring("funName = ", funName)
	}

	logicEv := new(PlayerLogicEvent)

	logicEv.CallMathod = funName
	logicEv.Msg = msg
	logicEv.PlayerLink = link

	event.Fire("OnPlayerMsg", logicEv)
}

// 解析PB消息吗
func MarshalBackEndProtobufMsg(data []byte, link Link) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	//判断基础长度
	if len(data) < 2 {
		return
	}

	//取出函数长度
	funNameLen := binary.LittleEndian.Uint32(data[0:4])

	//跳过自身两字节
	funNameLen += 4
	//判断后续长度
	if len(data) < int(funNameLen) {
		return
	}
	//取出函数名
	funName := string(data[4:funNameLen])

	//取出消息类型
	paramType := event.GetMethodParam(funName)
	if paramType == nil {
		log.Error("获取消息类型失败 funName = ", funName)
		if link != nil {
			log.Error("远程IP：", link.RemoteAddr())
		}
		return
	}

	if len(data) > 10*2048 {
		log.Error("************消息长度超过20K funName = ", funName)
	}

	//跳过后续数据长度两个字节
	funNameLen += 4
	//对数据进行pb解码
	msg := reflect.New(paramType).Interface()
	err := proto.Unmarshal(data[funNameLen:], msg.(proto.Message))
	if err != nil {
		log.Waring(err)
		log.Waring("funName = ", funName)
	}

	logicEv := new(BackEndLogicEvent)
	logicEv.CallMethod = funName
	logicEv.Msg = msg
	logicEv.BackEndLink = link

	event.Fire("OnLegacyBackEndMsg", logicEv)
}

// 解析客户端PB消息码
func MarshalClientProtobufMsg(data []byte, link Link) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	//判断基础长度
	if len(data) < 2 {
		return
	}

	//取出函数长度
	funNameLen := binary.LittleEndian.Uint32(data[0:4])

	//跳过自身两字节
	funNameLen += 4
	//判断后续长度
	if len(data) < int(funNameLen) {
		return
	}
	//取出函数名
	funName := string(data[4:funNameLen])

	if len(data) > 10*2048 {
		log.Error("************消息长度超过20K funName = ", funName)
	}
	logicEv := new(BackEndLogicEvent)
	logicEv.CallMethod = funName
	logicEv.Msg = data
	logicEv.BackEndLink = link
	event.Fire("OnLegacyClientMsg", logicEv)
}
