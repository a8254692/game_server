package network

import (
	"BilliardServer/Util/log"
	"net"

	"google.golang.org/protobuf/proto"
)

// 网络连接的抽象接口
type Link interface {
	//设置ID
	SetID(linkID uint64)

	//获得连接ID
	GetID() uint64

	//是否为主动连接
	IsLinkDrive() bool

	//设置连接类型
	SetLinkType(localType uint16, otherType uint16)
	//运行连接
	Run()

	//写入数据
	Send(data []byte) error

	//连接的本地地址
	LocalAddr() net.Addr

	//连接的远程地址
	RemoteAddr() net.Addr

	//关闭连接
	Close()

	//销毁连接
	Destroy()
}

// 连接事件
type LinkEvent struct {
	NewLink *TcpLink //新连接
}

// 玩家连接事件
type WsLinkEvent struct {
	NewLink *WsLink //新连接
}

// 玩家连接事件
type PlayerOffLinkEvent struct {
	OffLink Link //新连接
}

// 玩家连接事件
type PlayerLogicEvent struct {
	CallMathod string      //调用函数
	Msg        interface{} //信息结构
	PlayerLink Link        //玩家连接
}

// 后端服务事件
type BackEndLogicEvent struct {
	CallMethod  string      //调用函数
	Msg         interface{} //消息结构
	BackEndLink Link        //后端连接
}

// 后端服务事件(HubSvr)
type BackendEvent struct {
	Msg    []byte
	LinkID int //连接
}

// 消息码事件
type MsgBodyEvent struct {
	ServerTypeLocal uint16   //本端服务器名称
	ServerTypeOther uint16   //另一端服务器名称
	MsgTile         uint32   //消息头
	MsgBody         []byte   //消息结构
	TcpLink         *TcpLink //TcpLink连接
	WsLink          *WsLink  //WsLink连接
}

// 反序列化
func (this *MsgBodyEvent) Unmarshal(msg proto.Message) error {
	err := proto.Unmarshal(this.MsgBody, msg.(proto.Message))
	if err != nil {
		log.Waring("-->MsgBodyEvent Unmarshal Error MsgTile:", this.MsgTile, " error:", err)
	}
	return err
}

// 序列化
func (this *MsgBodyEvent) Marshal(param interface{}) error {
	data, err := proto.Marshal(param.(proto.Message))
	if err == nil {
		this.MsgBody = data
	} else {
		log.Waring("-->MsgBodyEvent Marshal Error MsgTile:", this.MsgTile, " error:", err)
	}
	return err
}
