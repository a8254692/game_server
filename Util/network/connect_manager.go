// connect_manager
package network

import (
	msg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"
	"encoding/binary"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 连接管理器，管理tcp，udp，webscoket等连接
type Connect_Manager struct {
	Server          SocketServer           //连接服务
	mapConnect      map[uint64]*TcpConnect //连接器Connect数组
	SizeConnect     uint64                 //ID计数器
	AddrServer      string                 //本服务地址
	ServerTypeLocal uint16                 //本端服务器类型

	ConnectGame *TcpConnect

	broadcastMsgTile uint32   //需要广播的消息码
	broadcastIdList  []uint32 //需要广播的ID数组
}

// 初始化连接器管理器，服务地址，服务类型
func (this *Connect_Manager) InitServer(addr string, serverType uint16) {
	if strings.Contains(addr, "ws://") {
		this.Server = new(WebSocketServer)
	} else {
		this.Server = new(TcpServer)
	}
	this.AddrServer = addr
	this.ServerTypeLocal = serverType
	this.mapConnect = make(map[uint64]*TcpConnect)
	this.broadcastMsgTile = 0
	this.broadcastIdList = nil

	//注册系统公共事件
	event.Register(EK_LinkSuccessDrive, reflect.ValueOf(this.OnLinkSuccessDrive))
	event.Register(EK_LinkSuccessPassive, reflect.ValueOf(this.OnLinkSuccessPassive))
	event.Register(EK_ReLink, reflect.ValueOf(this.OnReLink))
	event.Register(EK_LinkOff, reflect.ValueOf(this.OnLinkOff))
	event.Register(EK_ReceiveMsg, reflect.ValueOf(this.OnReceiveMsgBody))
}

// 开启连接器管理器，及相关服务
func (this *Connect_Manager) StartServer() error {
	//开启Tcp服务
	if this.AddrServer != "" {
		err := this.Server.Start(this.AddrServer)
		if err != nil {
			log.Error(err)
			return err
		}
		go this.Server.Run()
	}
	return nil
}

// 关闭连接器管理器，及相关服务
func (this *Connect_Manager) StopServer() {
	//通知所有玩家服务器关闭
	for _, connect := range this.mapConnect {
		connect.CloseConnect()
	}
	this.Server.Stop()
}

// 处理一个新的被动连接
func (this *Connect_Manager) AcceptLink(tcpLink *TcpLink) error {
	//ID计数器自增
	this.SizeConnect = this.SizeConnect + 1
	//设置连接ID
	connectID := this.getSoleRandomID(100000, 999999)
	tcpLink.SetID(connectID)
	tc := new(TcpConnect)

	tc.InitByLink(tcpLink, this.ServerTypeLocal, tcpLink.ServerTypeOther)
	this.mapConnect[connectID] = tc
	log.Info("-->新的被动连接接入 ID："+strconv.FormatUint(connectID, 10), " 连接器总数:-----------------------", this.SizeConnect)
	//tc.SendIdentity() //马上发送身份确认消息码
	return nil
}

// 处理一个新的主动连接
func (this *Connect_Manager) AcceptConnect(tcpConnet *TcpConnect) error {
	//ID计数器自增
	this.SizeConnect = this.SizeConnect + 1
	//设置连接ID
	connectID := this.getSoleRandomID(100000, 999999)
	tcpConnet.TcpLink.SetID(connectID)
	tcpConnet.SetConnectType(this.ServerTypeLocal, tcpConnet.TcpLink.ServerTypeOther)
	this.mapConnect[connectID] = tcpConnet
	log.Info("-->新的主动连接接入 ID："+strconv.FormatUint(connectID, 10), " 连接器总数:-----------------------", this.SizeConnect)
	//发送身份确认消息码
	tcpConnet.SendIdentity()
	//发送消息码订阅
	tcpConnet.SendSubscribeMsg(event.GetMsgTileList())
	return nil
}

// 关闭一个连接
func (this *Connect_Manager) CloseConnect(tcpLink *TcpLink) {
	linkID := tcpLink.GetID()
	tc := this.GetTcpConnectByID(linkID)
	if tc == nil {
		log.Error("没有此ID的TcpConnect连接器:", strconv.FormatUint(linkID, 10), "by CloseConnect")
		return
	}

	tc.CloseConnect()

	delete(this.mapConnect, linkID)
	if this.SizeConnect > 0 {
		this.SizeConnect = this.SizeConnect - 1
	}
	log.Info("关闭一条连接 ID："+strconv.FormatUint(linkID, 10), " 连接器总数:-----------------------", this.SizeConnect)
}

// 获取TcpConnect连接器 按连接ID
func (this *Connect_Manager) ChangeTcpConnectID(connectID uint64, accID uint64) bool {
	tc := this.GetTcpConnectByID(connectID)
	if tc == nil {
		log.Error("没有此ID的TcpConnect连接器:", strconv.FormatUint(connectID, 10), "by ChangeTcpConnectID")
		return false
	}
	tc.TcpLink.SetID(accID)
	delete(this.mapConnect, connectID)
	this.mapConnect[accID] = tc
	log.Info("-->同步AccID ", connectID, " 转换为 ", accID)
	return true
}

// 获取TcpConnect连接器 按连接ID
func (this *Connect_Manager) GetTcpConnectByID(connectID uint64) *TcpConnect {
	if !this.IsExistConnectByID(connectID) {
		log.Error("没有此ID的TcpConnect连接器:" + strconv.FormatUint(connectID, 10))
		return nil
	}
	return this.mapConnect[connectID]
}

// 获取TcpConnect连接器，按连接另一端的服务器类型
func (this *Connect_Manager) GetTcpConnectByType(serverType uint16) *TcpConnect {
	for _, value := range this.mapConnect {
		if value.ServerTypeOther == serverType {
			return value
		}
	}
	return nil
}

// 是否存在此ID的连接器
func (this *Connect_Manager) IsExistConnectByID(connectID uint64) bool {
	tc := this.mapConnect[connectID]
	if tc == nil {
		return false
	}
	return true
}

// 是否存连接到此类型服务器的连接器
func (this *Connect_Manager) IsExistConnectByType(serverType uint16) bool {
	for _, value := range this.mapConnect {
		if value.ServerTypeOther == serverType {
			return true
		}
	}
	return false
}

// 获取订阅了此消息码的所有连接器
func (this *Connect_Manager) GetConnectByMsgTile(msgTile uint32) []uint64 {
	var keyArgs []uint64
	for key, value := range this.mapConnect {
		if value.IsSubscribeMsg(msgTile) {
			keyArgs = append(keyArgs, key)
		}
	}
	return keyArgs
}

// 生成不重复的连接ID
func (this *Connect_Manager) getSoleRandomID(start int, end int) uint64 {
	//范围检查
	if end < start {
		return 0
	}
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//生成随机数
	num := uint64(r.Intn((end - start)) + start)
	if this.IsExistConnectByID(uint64(num)) {
		num = this.getSoleRandomID(100000, 999999)
	}
	return num
}

// 处理一个主动连接 TcpConnect
func (this *Connect_Manager) OnLinkSuccessDrive(ev *TcpConnect) {
	//创建一个Connect
	this.AcceptConnect(ev)
}

// 处理一个被动连接 TcpLink
func (this *Connect_Manager) OnLinkSuccessPassive(ev *LinkEvent) {
	//创建一个Connect
	this.AcceptLink(ev.NewLink)
}

// 处理主动连接重连
func (this *Connect_Manager) OnReLink(ev *LinkEvent) {
	tc := this.GetTcpConnectByID(ev.NewLink.GetID())
	if tc == nil {
		log.Waring("-->重连时未找到指定ID的TcpConnect对象 connectID:", strconv.FormatUint(ev.NewLink.GetID(), 10))
		return
	}
	//发送身份确认消息码
	tc.SendIdentity()
	//发送消息码订阅
	tc.SendSubscribeMsg(event.GetMsgTileList())
}

// 处理连接断开
func (this *Connect_Manager) OnLinkOff(ev *LinkEvent) {
	this.CloseConnect(ev.NewLink)
}

// 处理客户端过来的逻辑消息 判断分发
func (this *Connect_Manager) OnReceiveMsgBody(msgEV *MsgBodyEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()
	log.Info("-->收到消息码 MsgTile:", msgEV.MsgTile, " form:", ServerName(msgEV.ServerTypeOther), " to ", ServerName(msgEV.ServerTypeLocal))
	if msgEV.MsgTile == Net_Identity {
		tc := this.GetTcpConnectByID(msgEV.TcpLink.GetID())
		if tc == nil {
			log.Error("TcpConnect 对象不存在，无法认证连接身份 connectID:" + strconv.FormatUint(msgEV.TcpLink.GetID(), 10))
			return
		}
		serverTypeOther := binary.LittleEndian.Uint16(msgEV.MsgBody)
		tc.SetConnectType(tc.ServerTypeLocal, serverTypeOther)
		log.Info("-->TcpConnect 认证连接身份 connectID:" + strconv.FormatUint(msgEV.TcpLink.GetID(), 10) + " 身份类型：" + ServerName(serverTypeOther))
	} else if msgEV.MsgTile == Net_Subscribemsg {
		tc := this.GetTcpConnectByID(msgEV.TcpLink.GetID())
		var tS uint32
		i := 0
		for {
			if i >= len(msgEV.MsgBody) {
				break
			}
			tS = binary.LittleEndian.Uint32(msgEV.MsgBody[i:])
			tc.SubscribeMsg(tS)
			log.Info("-->添加订阅消息码：", tS, " connectID:"+strconv.FormatUint(msgEV.TcpLink.GetID(), 10))
			i = i + 4
		}
		//如果对方的主动链接向本服订阅消息码，则向对方订阅消息码,网关服不向其它服订阅消息码
		if tc.TcpLink.LinkType == LinkType_Passive && this.ServerTypeLocal != ServerType_Gate {
			tMsgTileArgs := event.GetMsgTileList()
			if len(tMsgTileArgs) > 0 {
				tc.SendSubscribeMsg(tMsgTileArgs)
			}
		}

	} else if msgEV.MsgTile == Net_Unsubscribemsg {
		tc := this.GetTcpConnectByID(msgEV.TcpLink.GetID())
		var tS uint32
		i := 0
		for {
			if i >= len(msgEV.MsgBody) {
				break
			}
			tS = binary.LittleEndian.Uint32(msgEV.MsgBody[i:])
			tc.UnSubscribeMsg(tS)
			i = i + 4
		}
	} else if msgEV.MsgTile == Net_Broadcast {
		this.broadcastMsgTile = binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
		tileLen := 4
		length := binary.LittleEndian.Uint32(msgEV.MsgBody[tileLen:])
		if length > 0 {
			this.broadcastIdList = make([]uint32, length)
			for i := 0; i < int(length); i++ {
				this.broadcastIdList[i] = binary.LittleEndian.Uint32(msgEV.MsgBody[tileLen+4+i*4:])
			}
		}
	} else if msgEV.MsgTile == Net_SyncEntityID {
		tc := this.GetTcpConnectByID(msgEV.TcpLink.GetID())
		if tc == nil {
			log.Error("TcpConnect 对象不存在，无法同步实体ID connectID:" + strconv.FormatUint(msgEV.TcpLink.GetID(), 10))
			return
		}
		tEntityID := binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
		tcNow := this.GetTcpConnectByID(uint64(tEntityID))
		if tcNow == nil {
			this.ChangeTcpConnectID(msgEV.TcpLink.GetID(), uint64(tEntityID))
			tc.SendMsgBody(Net_SyncEntityID, msgEV.MsgBody)
		} else {
			//log.Waring("已有相同EntityID 的TcpConnect 关闭此连接器"+strconv.FormatUint(msgEV.TcpLink.GetID(), 10), " 连接器总数:-----------------------", this.SizeConnect)
			//tc.SendMsgBody(Net_HadConnect, nil)
			this.CloseConnect(tcNow.TcpLink)

			//存在链接则踢人并且发送客户端消息
			this.ChangeTcpConnectID(msgEV.TcpLink.GetID(), uint64(tEntityID))
			tc.SendMsgBody(Net_SyncEntityID, msgEV.MsgBody)
		}
	} else if msgEV.MsgTile == Net_HeartBeat {
		tc := this.GetTcpConnectByID(msgEV.TcpLink.GetID())
		if tc == nil {
			log.Error("TcpConnect 对象不存在，无法处理心跳包逻辑 connectID:" + strconv.FormatUint(msgEV.TcpLink.GetID(), 10))
			return
		}
		tc.ReceivingHeartBeat()
	} else {
		if event.IsExistMsgTile(msgEV.MsgTile) {
			//如果本服有订阅，则派发事件给上层逻辑模块处理
			event.EmitNet(msgEV.MsgTile, msgEV)
		} else {
			//如果本服没有订阅，则搜索有订阅此消息码的连接做转发
			this.AnalyseTransmitBySubscribe(msgEV)
		}
	}
}

// 分析转发 按消息订阅
func (this *Connect_Manager) AnalyseTransmitBySubscribe(msgEV *MsgBodyEvent) {
	if this.broadcastMsgTile == msgEV.MsgTile && this.ServerTypeLocal == ServerType_Gate {
		//如果本服是网关服，且需要广播，则广播此消息码，广播只向client广播
		if this.broadcastIdList == nil {
			//broadcastIdList为nil则向所有的连接器广播
			for _, value := range this.mapConnect {
				value.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
			}
		} else {
			if len(this.broadcastIdList) <= 0 {
				return
			}

			//broadcastIdList不为nil则向broadcastIdList的连接器广播
			for i := 0; i < len(this.broadcastIdList); i++ {
				tc := this.GetTcpConnectByID(uint64(this.broadcastIdList[i]))
				if tc != nil {
					tc.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
				}
			}
		}
		//广播需求是一次性的，此处要将标识重置
		this.broadcastMsgTile = 0
		this.broadcastIdList = nil
	} else {
		//如果不需要广播则按订阅需求处理，网关服发向其它服，其它服之间通讯均走此通道

		//tcArgs := this.GetConnectByMsgTile(msgEV.MsgTile)
		//if len(tcArgs) < 1 {
		//	log.Info("-->此消息没有任何一个TcpConnect订阅 MsgTile：", msgEV.MsgTile, " by AnalyseTransmitBySubscribe")
		//	return
		//}
		//for i := 0; i < len(tcArgs); i++ {
		//	tc := this.GetTcpConnectByID(tcArgs[i])
		//	if tc != nil {
		//		tc.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
		//	}
		//}
		if this.ConnectGame == nil {
			return
		}
		this.ConnectGame.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
		return
	}
}

// 发送[]byte消息码，按订阅需求
func (this *Connect_Manager) SendMsgBody(msgTile msg.MsgTile, body []byte) {
	uint32MsgTile := uint32(msgTile)
	tcArgs := this.GetConnectByMsgTile(uint32MsgTile)
	if len(tcArgs) < 1 {
		log.Info("-->此消息没有任何一个TcpConnect订阅 MsgTile：", uint32MsgTile)
		return
	}
	for i := 0; i < len(tcArgs); i++ {
		tc := this.GetTcpConnectByID(tcArgs[i])
		tc.SendMsgBody(uint32MsgTile, body)
	}
}

// 发送PB消息码，按订阅需求
func (this *Connect_Manager) SendMsgBodyPB(msgTile msg.MsgTile, param interface{}) {
	uint32MsgTile := uint32(msgTile)

	tcArgs := this.GetConnectByMsgTile(uint32MsgTile)
	if len(tcArgs) < 1 {
		log.Info("-->此消息没有任何一个TcpConnect订阅 MsgTile：", msgTile)
		return
	}
	for i := 0; i < len(tcArgs); i++ {
		tc := this.GetTcpConnectByID(tcArgs[i])
		tc.SendMsgBodyPB(uint32MsgTile, param)
	}
}

// 发送消息码给指定的服务器
func (this *Connect_Manager) SendMsgPbToOtherServer(msgTile msg.MsgTile, param interface{}, serverType uint16) {
	tc := this.GetTcpConnectByType(serverType)
	if tc == nil {
		log.Info("-->没有找到对应类型的TcpConnect对象 by SendMsgPbToOtherServer")
		return
	}
	tc.SendMsgBodyPB(uint32(msgTile), param)
}

// 发送消息码给指定的服务器
func (this *Connect_Manager) SendMsgToOtherServer(msgTile msg.MsgTile, body []byte, serverType uint16) {
	tc := this.GetTcpConnectByType(serverType)
	if tc == nil {
		log.Info("-->没有找到对应类型的TcpConnect对象 by SendMsgPbToOtherServer")
		return
	}
	tc.SendMsgBody(uint32(msgTile), body)
}

// 发送消息码给网关服
func (this *Connect_Manager) SendMsgPbToGate(msgTile msg.MsgTile, param interface{}) {
	tc := this.GetTcpConnectByType(ServerType_Gate)
	if tc == nil {
		log.Info("-->没有找到对应类型的TcpConnect对象 by SendMsgPbToGate")
		return
	}
	tc.SendMsgBodyPB(uint32(msgTile), param)
}

// 发送消息码给网关服，并让网关服向client广播PB消息
func (this *Connect_Manager) SendMsgPbToGateBroadCast(msgTile msg.MsgTile, param interface{}, idList []uint32) {
	tc := this.GetTcpConnectByType(ServerType_Gate)
	if tc == nil {
		log.Info("-->没有找到对应类型的TcpConnect对象 by SendMsgPbToGateBroadCast")
		return
	}
	tc.SendMsgPbToBroadCast(uint32(msgTile), idList, param)
}

// 发送消息码给网关服，并让网关服向所有client广播PB消息
func (this *Connect_Manager) SendMsgPbToGateBroadCastAll(msgTile msg.MsgTile, param interface{}) {
	tc := this.GetTcpConnectByType(ServerType_Gate)
	if tc == nil {
		log.Info("-->-->没有找到对应类型的TcpConnect对象 by SendMsgPbToGateBroadCastAll")
		return
	}
	tc.SendMsgPbToBroadCastAll(uint32(msgTile), param)
}

// 发送消息码给指定的client，只在网关服使用
func (this *Connect_Manager) SendMsgPbToClient(msgTile msg.MsgTile, param interface{}, clientID uint64) {
	tc := this.GetTcpConnectByID(clientID)
	if tc == nil {
		log.Info("-->-->没有找到对应ID的TcpConnect对象 by SendMsgPbToClient")
		return
	}
	tc.SendMsgBodyPB(uint32(msgTile), param)

}

// 发送消息码给所有的client，只在网关服使用
func (this *Connect_Manager) SendMsgPbToClientAll(msgTile msg.MsgTile, param interface{}) {
	for _, tc := range this.mapConnect {
		tc.SendMsgBodyPB(uint32(msgTile), param)
	}

}

// 获取TcpConnect连接器
func (this *Connect_Manager) GetTcpMapConnect() map[uint64]*TcpConnect {
	return this.mapConnect
}
