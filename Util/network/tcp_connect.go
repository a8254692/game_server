// Tcp连接器
package network

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/timer"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"
)

type TcpConnect struct {
	TcpLink         *TcpLink
	mapSubscribeMsg map[uint32]uint32 //注册消息对应管理器
	addr            string
	ServerTypeLocal uint16 //本端服务器名称
	ServerTypeOther uint16 //另一端服务器名称

	TimeHeartBeat int64 //上一次收到心跳包的时间
}

// 初始化 建立主动连接
func (this *TcpConnect) Init(addr string, localType uint16, otherType uint16) {
	this.addr = addr
	this.ServerTypeLocal = localType
	this.ServerTypeOther = otherType
	this.mapSubscribeMsg = make(map[uint32]uint32)
	this.ReLink(addr)
	return
}

// 初始化 建立被动连接
func (this *TcpConnect) InitByLink(tcpLink *TcpLink, localType uint16, otherType uint16) {
	this.ServerTypeLocal = localType
	this.ServerTypeOther = otherType
	this.TcpLink = tcpLink
	this.TcpLink.SetLinkType(localType, otherType)
	this.mapSubscribeMsg = make(map[uint32]uint32)

	if localType == ServerType_Gate {
		this.TimeHeartBeat = time.Now().Unix()
		timer.AddTimer(this, "CheckHeartBeat", 1000*5, true)
	}

	go this.TcpLink.Run()
	return
}

// 设置连接器连接类型 本地类型 另一端类型
func (this *TcpConnect) SetConnectType(localType uint16, otherType uint16) {
	this.ServerTypeLocal = localType
	this.ServerTypeOther = otherType
	if this.TcpLink != nil {
		this.TcpLink.SetLinkType(localType, otherType)
	}
	return
}

// 重新请求建立
func (this *TcpConnect) ReLink(addr string) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()
	this.addr = addr
	newLink := new(TcpLink)
	var err error
	newLink.Connect, err = net.Dial("tcp", addr)
	if err != nil {
		log.Error("连接服务器%s失败，请检查该服务器状态 连接地址:", addr)
		//十秒后重试
		timer.DellTimer(this, "Retry")
		timer.AddTimer(this, "Retry", 5*1000, false)
	} else {
		newLink.linkID = 0
		newLink.addr = addr
		newLink.LinkType = LinkType_Drive
		newLink.SetLinkType(this.ServerTypeLocal, this.ServerTypeOther)
		this.TcpLink = newLink
		go newLink.Run()

		log.Info("-->连接后端服务成功：", addr)
		timer.DellTimer(this, "Retry")
		timer.AddTimer(this, "SendLinkSuccessDriveEvent", 200, false)
		//event.Fire(EK_LinkSuccessDrive, this)
	}
	return
}

// 连接服务器失败，重试到成功为止
func (this *TcpConnect) Retry() {
	this.ReLink(this.addr)
	return
}

// 发送主动连接成功的事件消息
func (this *TcpConnect) SendLinkSuccessDriveEvent() {
	event.Fire(EK_LinkSuccessDrive, this)
	return
}

// 关闭连接器
func (this *TcpConnect) CloseConnect() {
	this.TcpLink.Close()
	timer.DellTimer(this, "Retry")
	timer.DellTimer(this, "CheckHeartBeat")
	log.Info("-->TcpConnect 关闭 connectID：", strconv.FormatUint(this.TcpLink.linkID, 10))
	return
}

// 向服务器发送数据
func (this *TcpConnect) Send(data []byte) error {
	if this.TcpLink == nil {
		//this.ReLink(this.addr)
		return errors.New("对应的后端服务不存在,无法发送消息......")
	}

	return this.TcpLink.Send(data)
}

// 发送PB消息
func (this *TcpConnect) SendMsgBodyPB(msgTile uint32, param interface{}) {
	buff := new(MyBuff)
	buff.WriteUint32(msgTile)

	data, err := proto.Marshal(param.(proto.Message))
	if err != nil {
		log.Error(err, msgTile)
		return
	}

	buff.WriteBytes(data)
	err = this.Send(buff.GetBytes())
	if err != nil {
		log.Error(err, msgTile)
	}
	log.Info("-->SendMsgBodyPB:", msgTile, " form:", ServerName(this.ServerTypeLocal), " goto:", ServerName(this.ServerTypeOther), " data length:", len(data))
	return
}

// 发送[]byte消息
func (this *TcpConnect) SendMsgBody(msgTile uint32, body []byte) {
	buff := new(MyBuff)
	buff.WriteUint32(msgTile)
	buff.WriteBytes(body)
	err := this.Send(buff.GetBytes())
	if err != nil {
		log.Error(err, msgTile)
	}
	log.Info("-->SendMsgBody:", msgTile, " form:", ServerName(this.ServerTypeLocal), " goto:", ServerName(this.ServerTypeOther), " data length:", len(body))
	return
}

// 广播PB消息,只针对网关服
func (this *TcpConnect) SendMsgPbToBroadCast(msgTile uint32, idList []uint32, param interface{}) {
	buff := new(MyBuff)
	//写入需要广播的消息码
	buff.WriteUint32(msgTile)
	//写入广播ID数组长度
	buff.WriteInt(len(idList))
	//写入所有需要广播的ID
	for i := 0; i < len(idList); i++ {
		buff.WriteUint32(idList[i])
	}
	//先发送广播请求消息
	this.SendMsgBody(Net_Broadcast, buff.GetBytes())
	//再发送需要广播的消息
	this.SendMsgBodyPB(msgTile, param)
	return
}

// 广播PB消息给全部连接器,只针对网关服
func (this *TcpConnect) SendMsgPbToBroadCastAll(msgTile uint32, param interface{}) {
	buff := new(MyBuff)
	//写入需要广播的消息码
	buff.WriteUint32(msgTile)
	//写入广播ID数组长度
	buff.WriteInt(0) //长度为0表示向全部连接器广播
	//先发送广播请求消息
	this.SendMsgBody(Net_Broadcast, buff.GetBytes())
	//再发送需要广播的消息
	this.SendMsgBodyPB(msgTile, param)
	return
}

// 获取消息码对应回调对象
func (this *TcpConnect) GetMsgFunVal(msgTile uint32) uint32 {
	if !this.IsSubscribeMsg(msgTile) {
		return msgTile
	}
	return this.mapSubscribeMsg[msgTile]
}

// 是否订阅此消息码
func (this *TcpConnect) IsSubscribeMsg(msgTile uint32) bool {
	funVal := this.mapSubscribeMsg[msgTile]
	if funVal == 0 {
		return false
	}
	return true
}

// 订阅消息码 添加
func (this *TcpConnect) SubscribeMsg(msgTile uint32) {
	this.mapSubscribeMsg[msgTile] = msgTile
	return
}

// 反订阅消息码 删除
func (this *TcpConnect) UnSubscribeMsg(msgTile uint32) {
	delete(this.mapSubscribeMsg, msgTile)
	return
}

// 发送身份确认消息
func (this *TcpConnect) SendIdentity() {
	buff := make([]byte, 2)
	//写入数据长度
	binary.LittleEndian.PutUint16(buff, this.ServerTypeLocal)
	this.SendMsgBody(Net_Identity, buff)
	return
}

// 发送消息码定阅消息
func (this *TcpConnect) SendSubscribeMsg(msgTileArgs []uint32) {
	buff := new(MyBuff)
	for i := 0; i < len(msgTileArgs); i++ {
		buff.WriteUint32(msgTileArgs[i])
	}
	this.SendMsgBody(Net_Subscribemsg, buff.GetBytes())
	return
}

// 发送消息码定阅消息
func (this *TcpConnect) SendHeartBeat() {
	buff := new(MyBuff)
	buff.WriteUint32(uint32(this.TcpLink.linkID))
	this.SendMsgBody(Net_HeartBeat, buff.GetBytes())
	return
}

// 发送消息码定阅消息
func (this *TcpConnect) ReceivingHeartBeat() {
	this.TimeHeartBeat = time.Now().Unix()
	return
}

// 检测心跳包
func (this *TcpConnect) CheckHeartBeat() {
	//处理心跳包时间问题
	timeHeartBeatNow := time.Now().Unix()
	timeDis := timeHeartBeatNow - this.TimeHeartBeat
	if timeDis >= 30 {
		//派发连接断开消息
		log.Waring("-->logic--TcpConnect--CheckHeartBeat--timeDis >= 30--", this.TcpLink.linkID)

		linkEvent := new(LinkEvent)
		linkEvent.NewLink = this.TcpLink
		event.Fire(EK_LinkOff, linkEvent)
	} else {
		this.SendHeartBeat()
	}
	return
}
