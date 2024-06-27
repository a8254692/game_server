// tcp_connect
package network

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/timer"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
)

// TCP连接类，用于处理TCP的各种操作
type TcpLink struct {
	Connect   net.Conn     //TCP连接
	linkID    uint64       //连接ID
	WirteBuff chan MsgBody //写入缓冲区
	Exception bool         //是否发生了异常
	addr      string       //后端服务地址
	IsReconn  bool         //是否在重连
	BeginTime int64
	RecvTime  int64
	sync.RWMutex
	LinkType        uint16
	LstLogTime      time.Time
	ServerTypeLocal uint16 //本端服务器名称
	ServerTypeOther uint16 //另一端服务器名称
}

// 设置ID
func (this *TcpLink) SetID(linkID uint64) {
	this.linkID = linkID
}

// 获得连接ID
func (this *TcpLink) GetID() uint64 {
	return this.linkID
}
func (this *TcpLink) SetLinkType(localType uint16, otherType uint16) {
	this.ServerTypeLocal = localType
	this.ServerTypeOther = otherType
}

// 是否为主动连接
func (this *TcpLink) IsLinkDrive() bool {
	return this.LinkType == LinkType_Drive
}

// 开始连接
func (this *TcpLink) Start() bool {
	return true
}

// 运行逻辑
func (this *TcpLink) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			this.OnException()
		}
	}()

	if this.Connect == nil {
		log.Error("-->TcpLink Run Error:Connect is nil")
		return
	}

	//创建数据写入缓冲通道
	this.WirteBuff = make(chan MsgBody, 1500)
	this.BeginTime = time.Now().Unix()
	this.RecvTime = this.BeginTime

	// if !this.IsLinkDrive() {
	// 	//判断连接是否废弃
	// 	go func() {
	// 		for {
	// 			timeout := make(chan bool, 1)
	// 			go func() {
	// 				time.Sleep(100 * time.Millisecond)
	// 				timeout <- true
	// 			}()

	// 			select {
	// 			case tm := <-this.RecvChan:
	// 				this.RecvTime = tm
	// 			case <-timeout:
	// 			}

	// 			if time.Now().Unix()-this.RecvTime >= 300 {
	// 				log.Error("timeout>300")
	// 				this.OnException()
	// 				break
	// 			}
	// 		}
	// 	}()
	// }

	//开启数据写入协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()

		for {
			//非阻塞式协程通讯，效率低
			// netPackage, open := <-this.WirteBuff
			// if open == false {
			// 	return
			// }
			//阻塞式协程通讯
			netPackage := <-this.WirteBuff

			this.RLock()
			if this.Exception {
				this.RUnlock()
				break
			}
			this.RUnlock()

			data := netPackage.ConvertBytes()
			if len(data) > 1024*1024*10 {
				log.Error("错误的数据长度，dataLen=", len(data))
				continue
			}

			_, err := this.Connect.Write(data)
			if err != nil {
				log.Waring("-->Connect.Write err ", err)
				this.OnException()
				break
			}
		}
	}()

	//循环接收数据
	for {
		if this.Exception {
			return
		}

		len := 4
		dataBuffLen := make([]byte, len)
		//读取数据长度
		if _, err := io.ReadFull(this.Connect, dataBuffLen); err != nil {
			//发生异常（客户端断开链接或者网络不稳定）
			log.Waring("client disconnected or network is poor linkId:", this.linkID)
			this.OnException()
			return
		}

		//读取数据
		dataLen := 0
		dataLen = int(binary.LittleEndian.Uint32(dataBuffLen))
		log.Info("消息...，长度 dataLen:", dataLen, "| 类型 this.LenType:", this.LinkType)
		if dataLen >= 1000*1024 || dataLen == 0 {
			log.Error("消息错误，长度 dataLen:", dataLen, "| 类型 this.LenType:", this.LinkType)
			if this.Connect != nil {
				log.Error("远程IP：", this.Connect.RemoteAddr())
			}
			this.OnException()
			return
		}

		//读取数据
		data := make([]byte, dataLen)
		readLen, err := io.ReadFull(this.Connect, data)

		if err != nil {
			this.OnException()
			continue
		}

		//判断基础长度
		if readLen < 2 {
			continue
		}
		this.MarshalMsg(data)
	}
}

// 写入数据
func (this *TcpLink) Send(data []byte) error {
	if this.Exception {
		return nil
	}
	//发生异常后抛弃所有发送
	netPackage := MsgBody{}
	netPackage.Init(this.ServerTypeLocal, this.ServerTypeOther)
	netPackage.SetData(data)

	timeout := time.NewTimer(time.Millisecond * 10)
	select {
	case this.WirteBuff <- netPackage:
		//log.Info("-->TcpLink-->Send id:", this.linkID)
		timeout.Stop()
	case <-timeout.C:
		log.Waring("-->TcpLink Send Timeout form:", this.ServerTypeLocal, " goto:", this.ServerTypeOther)
		//如果缓冲区满了则主动断开链接
		//this.OnException()
	}
	return nil
}

// 连接的本地地址
func (this *TcpLink) LocalAddr() net.Addr {
	return this.Connect.LocalAddr()
}

// 连接的远程地址
func (this *TcpLink) RemoteAddr() net.Addr {
	return this.Connect.RemoteAddr()
}

// 关闭连接
func (this *TcpLink) Close() {
	this.Lock()
	defer this.Unlock()

	this.Exception = true
	if this.Connect == nil {
		return
	}

	//关闭通道
	close(this.WirteBuff)
	//关闭连接
	this.Connect.(*net.TCPConn).Close()
	this.Connect = nil
}

// 销毁连接
func (this *TcpLink) Destroy() {
	this.Close()
}

// 重连
func (this *TcpLink) ReLink() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	for {
		var err error
		this.Connect, err = net.Dial("tcp", this.addr)
		if err != nil {
			if time.Now().Unix()-this.LstLogTime.Unix() > 1 {
				this.LstLogTime = time.Now()
				log.Error("-->连接后端服务器%s失败，请检查该服务器状态,连接地址：", this.addr)
			}
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		this.Lock()
		this.Exception = false
		this.IsReconn = false
		this.Unlock()

		go this.Run() //此处可能后执行，导致重连事件发出后，后续消息发不出去

		//500毫秒后再发送重连功成功的事件消息
		timer.AddTimer(this, "SendReLinkEvent", 200, false)
		// linkEvent := new(LinkEvent)
		// linkEvent.NewLink = this
		// event.Fire(EK_ReLink, linkEvent)
		return
	}
}

// 发送重连功成功的事件消息
func (this *TcpLink) SendReLinkEvent() {
	linkEvent := new(LinkEvent)
	linkEvent.NewLink = this
	event.Fire(EK_ReLink, linkEvent)
}

// 链接异常
func (this *TcpLink) OnException() {
	this.Lock()
	this.Exception = true
	if this.Connect == nil {
		//主动关闭的则不需要重连
		this.Unlock()
		return
	}

	if this.IsLinkDrive() {
		//如果是主动连接 需要自动重连
		if !this.IsReconn {
			this.IsReconn = true
			this.Unlock()
			this.Close()
			this.ReLink()
		} else {
			this.Unlock()
		}
	} else {
		//派发连接断开消息
		this.IsReconn = false
		this.Unlock()
		this.Close()
		linkEvent := new(LinkEvent)
		linkEvent.NewLink = this
		event.Fire(EK_LinkOff, linkEvent)
	}
}

// 解析PB消息码
func (this *TcpLink) MarshalMsg(data []byte) {
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
	msgBodyEV.ServerTypeOther = binary.LittleEndian.Uint16(data[0:2])
	msgBodyEV.ServerTypeLocal = binary.LittleEndian.Uint16(data[2:4])
	//取出消息头名称长度
	msgTile := binary.LittleEndian.Uint32(data[4:8])
	//跳过上面8字节
	msgTileLen := 8
	msgBodyEV.MsgTile = msgTile
	msgBodyEV.MsgBody = data[msgTileLen+4:] //消息头长度不包含在消息体中
	msgBodyEV.TcpLink = this
	event.Fire(EK_ReceiveMsg, msgBodyEV)
}
