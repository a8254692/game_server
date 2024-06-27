package network

import (
	gmsg "BilliardServer/Proto/gmsg"
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
type Ws_Manager struct {
	Server          SocketServer          //连接服务
	mapConnect      map[uint64]*WsConnect //连接器Connect数组
	SizeConnect     uint64                //ID计数器
	AddrServer      string                //本服务地址
	ServerTypeLocal uint16                //本端服务器类型

	ConnectGame *TcpConnect
}

// 初始化连接器管理器，服务地址，服务类型
func (this *Ws_Manager) InitServer(addr string, serverType uint16) {
	if strings.Contains(addr, "ws://") {
		this.Server = new(WebSocketServer)
	} else {
		this.Server = new(TcpServer)
	}
	this.SizeConnect = 0
	this.AddrServer = addr
	this.ServerTypeLocal = serverType
	this.mapConnect = make(map[uint64]*WsConnect, 0)

	//注册系统公共事件
	event.Register(EK_WsLinkSuccessDrive, reflect.ValueOf(this.OnLinkSuccessDrive))
	event.Register(EK_WsLinkSuccessPassive, reflect.ValueOf(this.OnLinkSuccessPassive))
	event.Register(EK_WsReLink, reflect.ValueOf(this.OnReLink))
	event.Register(EK_WsLinkOff, reflect.ValueOf(this.OnLinkOff))
	event.Register(EK_WsReceiveMsg, reflect.ValueOf(this.OnReceiveMsgBody))
}

// 开启连接器管理器，及相关服务
func (this *Ws_Manager) StartServer() error {
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
func (this *Ws_Manager) StopServer() {
	//通知所有玩家服务器关闭
	for _, connect := range this.mapConnect {
		connect.CloseConnect()
	}
	this.Server.Stop()
	return
}

// 处理一个新的被动连接
func (this *Ws_Manager) AcceptLink(wsLink *WsLink) error {
	//ID计数器自增
	this.SizeConnect = this.SizeConnect + 1
	//设置连接ID
	connectID := this.getSoleRandomID(100000, 999999)
	wsLink.SetID(connectID)
	tc := new(WsConnect)
	tc.InitByLink(wsLink, this.ServerTypeLocal, ServerType_Client)
	this.mapConnect[connectID] = tc
	log.Info("-->新的被动连接接入 ID："+strconv.FormatUint(connectID, 10), " 连接器总数:-----------------------", this.SizeConnect)
	//tc.SendIdentity() //马上发送身份确认消息码
	return nil
}

// 处理一个新的主动连接
func (this *Ws_Manager) AcceptConnect(wsConnet *WsConnect) error {
	//ID计数器自增
	this.SizeConnect = this.SizeConnect + 1
	//设置连接ID
	connectID := this.getSoleRandomID(100000, 999999)
	wsConnet.WsLink.SetID(connectID)
	wsConnet.SetConnectType(this.ServerTypeLocal, wsConnet.WsLink.ServerTypeOther)
	this.mapConnect[connectID] = wsConnet
	log.Info("-->新的主动连接接入 ID："+strconv.FormatUint(connectID, 10), " 连接器总数:-----------------------", this.SizeConnect)
	//发送身份确认消息码
	wsConnet.SendIdentity()
	//发送消息码订阅
	wsConnet.SendSubscribeMsg(event.GetMsgTileList())
	return nil
}

// 关闭一个连接
func (this *Ws_Manager) CloseConnect(wsLink *WsLink) {
	linkID := wsLink.GetID()
	wsConnect := this.GetWsConnectByID(linkID)
	delete(this.mapConnect, linkID)
	wsConnect.CloseConnect()
	this.SizeConnect = this.SizeConnect - 1
	log.Info("-->Ws--关闭一条连接 ID："+strconv.FormatUint(linkID, 10), " 连接器总数:-----------------------", this.SizeConnect)
	return
}

// 获取TcpConnect连接器 按连接ID
func (this *Ws_Manager) ChangeTcpConnectID(connectID uint64, accID uint64) bool {
	tc := this.GetWsConnectByID(connectID)
	if tc == nil {
		log.Error("没有此ID的TcpConnect连接器:", strconv.FormatUint(connectID, 10), "by ChangeTcpConnectID")
		return false
	}
	tc.WsLink.SetID(accID)
	delete(this.mapConnect, connectID)
	this.mapConnect[accID] = tc
	log.Info("-->同步AccID ", connectID, " 转换为 ", accID)
	return true
}

// 获取mapConnect
func (this *Ws_Manager) GetMapConnect() map[uint64]*WsConnect {
	return this.mapConnect
}

// 获取TcpConnect连接器 按连接Link
func (this *Ws_Manager) GetWsConnectByLink(wsLink *WsLink) *WsConnect {
	for _, value := range this.mapConnect {
		if value.WsLink == wsLink {
			return value
		}
	}
	return nil
}

// 获取TcpConnect连接器 按连接ID
func (this *Ws_Manager) GetWsConnectByID(connectID uint64) *WsConnect {
	if !this.IsExistConnectByID(connectID) {
		log.Error("没有此ID的TcpConnect连接器:" + strconv.FormatUint(connectID, 10))
		return nil
	}
	return this.mapConnect[connectID]
}

// 获取TcpConnect连接器，按连接另一端的服务器类型
func (this *Ws_Manager) GetWsConnectByType(serverType uint16) *WsConnect {
	for _, value := range this.mapConnect {
		if value.ServerTypeOther == serverType {
			return value
		}
	}
	return nil
}

// 是否存在此ID的连接器
func (this *Ws_Manager) IsExistConnectByID(connectID uint64) bool {
	tc := this.mapConnect[connectID]
	if tc == nil {
		return false
	}
	return true
}

// 是否存连接到此类型服务器的连接器
func (this *Ws_Manager) IsExistConnectByType(serverType uint16) bool {
	for _, value := range this.mapConnect {
		if value.ServerTypeOther == serverType {
			return true
		}
	}
	return false
}

// 获取订阅了此消息码的所有连接器
func (this *Ws_Manager) GetConnectByMsgTile(msgTile uint32) []uint64 {
	var keyArgs []uint64
	for key, value := range this.mapConnect {
		if value.IsSubscribeMsg(msgTile) {
			keyArgs = append(keyArgs, key)
		}
	}
	return keyArgs
}

// 生成不重复的连接ID
func (this *Ws_Manager) getSoleRandomID(start int, end int) uint64 {
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
func (this *Ws_Manager) OnLinkSuccessDrive(ev *WsConnect) {
	//创建一个Connect
	this.AcceptConnect(ev)
	return
}

// 处理一个被动连接 WsLink
func (this *Ws_Manager) OnLinkSuccessPassive(ev *WsLinkEvent) {
	//创建一个Connect
	this.AcceptLink(ev.NewLink)
	return
}

// 处理主动连接重连
func (this *Ws_Manager) OnReLink(ev *WsLinkEvent) {
	tc := this.GetWsConnectByID(ev.NewLink.GetID())
	if tc == nil {
		log.Waring("-->重连时未找到指定ID的TcpConnect对象 connectID:", strconv.FormatUint(ev.NewLink.GetID(), 10))
		return
	}
	//发送身份确认消息码
	tc.SendIdentity()
	//发送消息码订阅
	tc.SendSubscribeMsg(event.GetMsgTileList())
	return
}

// 处理连接断开
func (this *Ws_Manager) OnLinkOff(ev *WsLinkEvent) {
	this.CloseConnect(ev.NewLink)

	//断开连接后通知游戏服
	req := &gmsg.EntityOfflineToGameRequest{
		EntityID: uint32(ev.NewLink.linkID),
	}
	this.ConnectGame.SendMsgBodyPB(uint32(gmsg.MsgTile_Sys_EntityOfflineToGameRequest), req)
	return
}

// 处理客户端过来的逻辑消息 判断分发
func (this *Ws_Manager) OnReceiveMsgBody(msgEV *MsgBodyEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()
	log.Info("-->收到消息码 MsgTile:", msgEV.MsgTile, " form:", ServerName(msgEV.ServerTypeOther), " to ", ServerName(msgEV.ServerTypeLocal))
	if msgEV.MsgTile == Net_Identity {
		tc := this.GetWsConnectByID(msgEV.WsLink.GetID())
		if tc == nil {
			log.Error("TcpConnect 对象不存在，无法认证连接身份 connectID:" + strconv.FormatUint(msgEV.WsLink.GetID(), 10))
			return
		}
		serverTypeOther := binary.LittleEndian.Uint16(msgEV.MsgBody)
		tc.SetConnectType(tc.ServerTypeLocal, serverTypeOther)
		log.Info("-->TcpConnect 认证连接身份 connectID:" + strconv.FormatUint(msgEV.WsLink.GetID(), 10) + " 身份类型：" + ServerName(serverTypeOther))
	} else if msgEV.MsgTile == Net_SyncEntityID {
		tc := this.GetWsConnectByID(msgEV.WsLink.GetID())
		if tc == nil {
			log.Error("TcpConnect 对象不存在，无法同步实体ID connectID:" + strconv.FormatUint(msgEV.WsLink.GetID(), 10))
			return
		}
		tEntityID := binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
		tcNow := this.GetWsConnectByID(uint64(tEntityID))
		if tcNow == nil {
			this.ChangeTcpConnectID(msgEV.WsLink.GetID(), uint64(tEntityID))
			tc.SendMsgBody(Net_SyncEntityID, msgEV.MsgBody)
		} else {
			//log.Waring("已有相同EntityID 的TcpConnect 关闭此连接器"+strconv.FormatUint(msgEV.WsLink.GetID(), 10), " 连接器总数:-----------------------", this.SizeConnect)
			//tc.SendMsgBody(Net_HadConnect, nil)
			this.CloseConnect(tcNow.WsLink)

			//存在链接则踢人并且发送客户端消息
			this.ChangeTcpConnectID(msgEV.WsLink.GetID(), uint64(tEntityID))
			tc.SendMsgBody(Net_SyncEntityID, msgEV.MsgBody)
		}
	} else if msgEV.MsgTile == Net_HeartBeat {
		tc := this.GetWsConnectByID(msgEV.WsLink.GetID())
		if tc == nil {
			log.Error("TcpConnect 对象不存在，无法处理心跳包逻辑 connectID:" + strconv.FormatUint(msgEV.WsLink.GetID(), 10))
			return
		}
		tc.RevceHeartBeat()
	} else {
		if event.IsExistMsgTile(msgEV.MsgTile) {
			//如果本服有订阅，则派发事件给上层逻辑模块处理
			event.EmitNet(msgEV.MsgTile, msgEV)
		} else {
			//如果本服没有订阅，则搜索有订阅此消息码的连接做转发
			this.AnalyseTransmitBySubscribe(msgEV)
		}
	}
	return
}

// 分析转发 按消息订阅
func (this *Ws_Manager) AnalyseTransmitBySubscribe(msgEV *MsgBodyEvent) {
	if this.ConnectGame == nil {
		return
	}
	this.ConnectGame.SendMsgBody(msgEV.MsgTile, msgEV.MsgBody)
	return
}

// 发送[]byte消息码，按订阅需求
func (this *Ws_Manager) SendMsgBody(msgTile uint32, body []byte) {
	tcArgs := this.GetConnectByMsgTile(msgTile)
	if len(tcArgs) < 1 {
		log.Info("-->此消息没有任何一个TcpConnect订阅 MsgTile：", msgTile)
		return
	}
	for i := 0; i < len(tcArgs); i++ {
		tc := this.GetWsConnectByID(tcArgs[i])
		tc.SendMsgBody(msgTile, body)
	}
	return
}

// 发送PB消息码，按订阅需求
func (this *Ws_Manager) SendMsgBodyPB(msgTile uint32, param interface{}) {
	tcArgs := this.GetConnectByMsgTile(msgTile)
	if len(tcArgs) < 1 {
		log.Info("-->此消息没有任何一个TcpConnect订阅 MsgTile：", msgTile)
		return
	}
	for i := 0; i < len(tcArgs); i++ {
		tc := this.GetWsConnectByID(tcArgs[i])
		tc.SendMsgBodyPB(msgTile, param)
	}
	return
}

// 发送消息码给网关服，并让网关服向client广播PB消息
func (this *Ws_Manager) SendMsgPbToGateBroadCast(msgTile uint32, param interface{}, idList []uint32) {
	tc := this.GetWsConnectByType(ServerType_Gate)
	if tc == nil {
		log.Info("-->没有找到对应类型的TcpConnect对象 by SendMsgPbToGateBroadCast")
		return
	}
	tc.SendMsgPbToBroadCast(msgTile, idList, param)
	return
}

// 发送消息码给网关服，并让网关服向所有client广播PB消息
func (this *Ws_Manager) SendMsgPbToGateBroadCastAll(msgTile uint32, param interface{}) {
	tc := this.GetWsConnectByType(ServerType_Gate)
	if tc == nil {
		log.Info("-->-->没有找到对应类型的TcpConnect对象 by SendMsgPbToGateBroadCastAll")
		return
	}
	tc.SendMsgPbToBroadCastAll(msgTile, param)
	return
}

// 发送消息码给指定的client，只在网关服使用
func (this *Ws_Manager) SendMsgToClient(msgTile uint32, body []byte, clientID uint64) {
	tc := this.GetWsConnectByID(clientID)
	if tc == nil {
		log.Info("-->-->没有找到对应ID的TcpConnect对象 by SendMsgToClient")
		return
	}
	tc.SendMsgBody(msgTile, body)
	return
}

// 发送PB消息码给指定的client，只在网关服使用
func (this *Ws_Manager) SendMsgPbToClient(msgTile uint32, param interface{}, clientID uint64) {
	tc := this.GetWsConnectByID(clientID)
	if tc == nil {
		log.Info("-->-->没有找到对应ID的TcpConnect对象 by SendMsgPbToClient")
		return
	}
	tc.SendMsgBodyPB(msgTile, param)
	return
}

// 发送消息码给所有的client，只在网关服使用
func (this *Ws_Manager) SendMsgPbToClientAll(msgTile uint32, param interface{}) {
	for _, tc := range this.mapConnect {
		tc.SendMsgBodyPB(msgTile, param)
	}
	return
}
