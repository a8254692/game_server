package logic

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"encoding/binary"
	"reflect"
	"strconv"
)

type GateWsTcp_Mananger struct {
	state           uint16
	tcpConnect      *network.TcpConnect //游戏服的TcpConnect连接器
	AddrServer      string              //本服务地址
	ServerTypeLocal uint16              //本端服务器类型

	//TODO：如果同时到达多个消息是否会存在消息覆盖问题(将此处改为缓存channel接收)
	broadcastMsgTile uint32   //需要广播的消息码
	broadcastIdList  []uint32 //需要广播的ID数组
}

func (this *GateWsTcp_Mananger) Init() {
	this.state = 0
	this.broadcastMsgTile = 0
	this.broadcastIdList = nil
	this.ServerTypeLocal = network.ServerType_Gate
	//注册系统公共事件
	event.Register(network.EK_LinkSuccessDrive, reflect.ValueOf(this.OnLinkSuccessDrive))
	event.Register(network.EK_ReLink, reflect.ValueOf(this.OnReLink))
	event.Register(network.EK_LinkOff, reflect.ValueOf(this.OnLinkOff))
	event.Register(network.EK_ReceiveMsg, reflect.ValueOf(this.OnReceiveMsgBody))
}

// 处理一个新的主动连接
func (this *GateWsTcp_Mananger) OnLinkSuccessDrive(tcpConnet *network.TcpConnect) error {
	this.state = 1
	this.tcpConnect = tcpConnet
	tcpConnet.TcpLink.SetID(100001)
	tcpConnet.SetConnectType(this.ServerTypeLocal, tcpConnet.ServerTypeOther)
	log.Info("-->新的主动连接接入 ID：", tcpConnet.TcpLink.GetID())
	//发送身份确认消息码
	tcpConnet.SendIdentity()
	//发送消息码订阅
	//tcpConnet.SendSubscribeMsg(event.GetMsgTileList())
	return nil
}

// 处理主动连接的重连
func (this *GateWsTcp_Mananger) OnReLink(ev *network.LinkEvent) error {
	log.Info("-->主动连接重新连接 ID：", ev.NewLink.GetID())
	//发送身份确认消息码
	this.tcpConnect.SendIdentity()
	//发送消息码订阅
	//ConnectGame.SendSubscribeMsg(event.GetMsgTileList())
	return nil
}

// 处理连接断开
func (this *GateWsTcp_Mananger) OnLinkOff(ev *network.LinkEvent) {
	log.Info("-->Tcp连接器已断开 ID：", this.tcpConnect.TcpLink.GetID())
	this.tcpConnect.CloseConnect()
}

// 处理客户端过来的逻辑消息 判断分发
func (this *GateWsTcp_Mananger) OnReceiveMsgBody(msgEV *network.MsgBodyEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()
	log.Info("-->收到消息码 MsgTile:", msgEV.MsgTile, " form:", network.ServerName(msgEV.ServerTypeOther), " to ", network.ServerName(msgEV.ServerTypeLocal))
	if msgEV.MsgTile == network.Net_Identity {
		serverTypeOther := binary.LittleEndian.Uint16(msgEV.MsgBody)
		this.tcpConnect.SetConnectType(this.tcpConnect.ServerTypeLocal, serverTypeOther)
		log.Info("-->TcpConnect 认证连接身份 connectID:" + strconv.FormatUint(msgEV.TcpLink.GetID(), 10) + " 身份类型：" + network.ServerName(serverTypeOther))
	} else if msgEV.MsgTile == network.Net_Subscribemsg {
		var tS uint32
		i := 0
		for {
			if i >= len(msgEV.MsgBody) {
				break
			}
			tS = binary.LittleEndian.Uint32(msgEV.MsgBody[i:])
			this.tcpConnect.SubscribeMsg(tS)
			log.Info("-->添加订阅消息码：", tS, " connectID:"+strconv.FormatUint(msgEV.TcpLink.GetID(), 10))
			i = i + 4
		}
	} else if msgEV.MsgTile == network.Net_Unsubscribemsg {
		var tS uint32
		i := 0
		for {
			if i >= len(msgEV.MsgBody) {
				break
			}
			tS = binary.LittleEndian.Uint32(msgEV.MsgBody[i:])
			this.tcpConnect.UnSubscribeMsg(tS)
			i = i + 4
		}
	} else if msgEV.MsgTile == network.Net_Broadcast {
		this.broadcastMsgTile = binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
		tileLen := 4
		length := binary.LittleEndian.Uint32(msgEV.MsgBody[tileLen:])
		if length > 0 {
			this.broadcastIdList = make([]uint32, length)
			for i := 0; i < int(length); i++ {
				this.broadcastIdList[i] = binary.LittleEndian.Uint32(msgEV.MsgBody[tileLen+4+i*4:])
			}
		}
	} else {
		if event.IsExistMsgTile(msgEV.MsgTile) {
			//如果本服有订阅，则派发事件给上层逻辑模块处理
			event.EmitNet(msgEV.MsgTile, msgEV)
		} else {

			//TODO：是否是只有客户端发送无意义消息才会到达此次执行逻辑

			//如果需要广播，则广播此消息码，广播只向client广播
			if this.broadcastIdList == nil {
				//broadcastIdList为nil则向所有的连接器广播
				tMapConnect := WsManager.GetMapConnect()
				for _, value := range tMapConnect {
					value.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
				}
			} else {
				//broadcastIdList不为nil则向broadcastIdList的连接器广播
				for i := 0; i < len(this.broadcastIdList); i++ {
					tc := WsManager.GetWsConnectByID(uint64(this.broadcastIdList[i]))
					if tc != nil {
						tc.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
					}
				}
			}
			//广播需求是一次性的，此处要将标识重置
			this.broadcastMsgTile = 0
			this.broadcastIdList = nil
		}
	}
}
