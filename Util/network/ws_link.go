// tcp_connect
package network

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type WsAddr struct {
	Addr string
}

func (wa WsAddr) Network() string {
	return "WebSocket"
}
func (wa WsAddr) String() string {
	return wa.Addr
}

// TCP连接类，用于处理TCP的各种操作
type WsLink struct {
	Connect   *websocket.Conn //TCP连接
	linkID    uint64          //连接ID
	WirteBuff chan MsgBody    //写入缓冲区
	Exception bool            //是否发生了异常
	addr      string          //后端服务地址
	BeginTime int64
	RecvTime  int64
	RecvChan  chan int64
	sync.RWMutex
	RemoteAddress   WsAddr
	LinkType        uint16
	ServerTypeLocal uint16 //本端服务器名称
	ServerTypeOther uint16 //另一端服务器名称
}

// 设置ID
func (this *WsLink) SetID(linkID uint64) {
	this.linkID = linkID
}

// 获得连接ID
func (this *WsLink) GetID() uint64 {
	return this.linkID
}
func (this *WsLink) SetLinkType(localType uint16, otherType uint16) {
	this.ServerTypeLocal = localType
	this.ServerTypeOther = otherType
}

// 是否为主动连接
func (this *WsLink) IsLinkDrive() bool {
	return this.LinkType == LinkType_Drive
}

// 运行逻辑
func (this *WsLink) Loop() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			this.OnException()
		}
	}()

	if this.Connect == nil {
		log.Error("Connect is nil")
		return
	}

	//创建数据写入缓冲通道
	this.WirteBuff = make(chan MsgBody, 1500)
	this.BeginTime = time.Now().Unix()
	this.RecvTime = this.BeginTime
	this.RecvChan = make(chan int64, 10)

	//判断连接是否废弃
	go func() {
		for {
			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(100 * time.Millisecond)
				timeout <- true
			}()

			select {
			case tm := <-this.RecvChan:
				this.RecvTime = tm
			case <-timeout:
			}

			if time.Now().Unix()-this.RecvTime >= 300 {
				this.OnException()
				break
			}
		}
	}()

	//开启数据写入协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()

		for {
			netPackage, open := <-this.WirteBuff
			if open == false {
				return
			}
			this.RLock()
			if this.Exception {
				this.RUnlock()
				break
			}

			this.RUnlock()
			err := websocket.Message.Send(this.Connect, netPackage.ConvertBytes())
			if err != nil {
				this.OnException()
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	//循环接收数据
	for {
		this.RLock()
		if this.Exception {
			this.RUnlock()
			close(this.RecvChan)
			break
		}
		this.RUnlock()
		var dataBuffLen []byte
		var err error
		err = websocket.Message.Receive(this.Connect, &dataBuffLen)
		//读取数据长度
		if err != nil && err != io.EOF {
			//发生异常（客户端断开链接或者网络不稳定）
			continue
		}

		if err != nil && err == io.EOF {
			//发生异常（客户端断开链接）
			close(this.RecvChan)
			this.OnException()
			return
		}
		// readLength:= len(dataBuffLen)
		// len := 2
		// if this.LinkType == LinkType_Drive {
		// 	len = 2
		// } else if this.LinkType == LinkType_Passive {
		// 	len = 4
		// }
		// MarshalProtobufMsg(dataBuffLen[len:readLength], this)
		this.MarshalMsg(dataBuffLen)

		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(10 * time.Millisecond)
			timeout <- true
		}()

		select {
		case this.RecvChan <- time.Now().Unix():
		case <-timeout:
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (this *WsLink) Run() {

}

// 写入数据
func (this *WsLink) Send(data []byte) error {
	//写入超时处理
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Millisecond)
		timeout <- true
	}()

	//发生异常后抛弃所有发送
	netPackage := MsgBody{}
	netPackage.Init(this.ServerTypeLocal, this.ServerTypeOther)
	netPackage.SetData(data)
	select {
	case this.WirteBuff <- netPackage:
	case <-timeout:
		log.Waring("Send Timeout")
		//如果缓冲区满了则主动断开链接
		this.OnException()
		return nil
	}

	return nil
}

// 连接的本地地址
func (this *WsLink) LocalAddr() net.Addr {
	return this.Connect.LocalAddr()
}

// 连接的远程地址
func (this *WsLink) RemoteAddr() net.Addr {
	return this.RemoteAddress
}

// 关闭连接
func (this *WsLink) Close() {
	this.Lock()
	defer this.Unlock()

	this.Exception = true
	if this.Connect == nil {
		return
	}

	//关闭通道
	close(this.WirteBuff)
	//关闭连接
	this.Connect.Close()
}

// 销毁连接
func (this *WsLink) Destroy() {
	this.Close()
}

// 链接异常
func (this *WsLink) OnException() {
	if this.Exception {
		return
	}
	this.Lock()
	this.Exception = true
	this.Unlock()

	//派发连接断开消息
	wsLinkEvent := new(WsLinkEvent)
	wsLinkEvent.NewLink = this
	event.Fire(EK_WsLinkOff, wsLinkEvent)
}

// 解析PB消息码
func (this *WsLink) MarshalMsg(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	//判断基础长度
	if len(data) < 2 {
		return
	}
	msgBodyEV := new(MsgBodyEvent)
	binary.LittleEndian.Uint32(data[0:])
	msgBodyEV.ServerTypeOther = binary.LittleEndian.Uint16(data[4:])
	msgBodyEV.ServerTypeLocal = binary.LittleEndian.Uint16(data[6:])
	//取出消息头名称长度
	msgTile := binary.LittleEndian.Uint32(data[8:])
	//跳过上面8字节
	msgTileLen := 12
	msgBodyEV.MsgTile = msgTile
	msgBodyEV.MsgBody = data[msgTileLen+4:] //消息体长度不包含在消息体中
	msgBodyEV.WsLink = this
	event.Fire(EK_WsReceiveMsg, msgBodyEV)
}
