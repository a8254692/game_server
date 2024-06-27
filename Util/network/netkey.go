package network

import (
	"strconv"
	"strings"
)

const (
	ServerType_No = iota
	ServerType_Client
	ServerType_Gate
	ServerType_Game
	ServerType_Center
	ServerType_DB
	ServerType_Other
)

// 依据服务器类型获得服务器名称
func ServerName(serverType uint16) string {
	nameArgs := map[uint16]string{
		ServerType_No:     "no",
		ServerType_Client: "client",
		ServerType_Gate:   "gateserver",
		ServerType_Game:   "gameserver",
		ServerType_Center: "centerserver",
		ServerType_DB:     "dbserver",
		ServerType_Other:  "otherserver",
	}

	name := nameArgs[serverType]
	if name == "" {
		name = "no"
	}

	return name
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

const (
	//-----------系统消息码定义部份----------------------------------------------

	Net_Identity       uint32 = 100001 //连接身份认证
	Net_Subscribemsg   uint32 = 100002 //订阅消息码
	Net_Unsubscribemsg uint32 = 100003 //消息码反订阅
	Net_Broadcast      uint32 = 100004 //广播消息码
	Net_SyncEntityID   uint32 = 100005 //同步更新连接ID
	Net_HeartBeat      uint32 = 100006 //心跳包
	Net_HadConnect     uint32 = 100007 //已经登录
)
const (
	//连接相关事件关键字
	EK_LinkSuccessDrive   string = "LinkSuccessDrive"   //主动连接成功
	EK_LinkSuccessPassive string = "LinkSuccessPassive" //被动连接成功
	EK_ReLink             string = "ReLink"             //重连成功
	EK_LinkOff            string = "LinkOff"            //连接断开
	EK_LinkFail           string = "LinkFail"           //连接失败
	EK_LinkError          string = "LinkError"          //连接出错
	EK_ReceiveMsg         string = "ReceiveMsg"         //接收消息

	EK_WsLinkSuccessDrive   string = "WsLinkSuccessDrive"   //websocket主动连接成功
	EK_WsLinkSuccessPassive string = "WsLinkSuccessPassive" //websocket被动连接成功
	EK_WsReLink             string = "WsReLink"             //重连成功
	EK_WsLinkOff            string = "WsLinkOff"            //websocket连接断开
	EK_WsReceiveMsg         string = "WsReceiveMsg"         //接收消息

	//系统消息头关键字 赋值部份全部小写
	MT_SubscribeMsg   string = "subscribemsg"   //消息头：订阅消息码
	MT_UnSubscribeMsg string = "unsubscribemsg" //消息头：反订阅消息码(退订)
	MT_Identity       string = "identity"       //消息头:身份确认
)
