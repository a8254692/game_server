package network

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
)

// tcp client 用于连接后端服务器
type TcpClient struct {
	endpoint        string       //地址
	id              int          //id
	connect         net.Conn     //连接
	writeBuff       chan MsgBody //写缓存
	stop            bool         //停止标记
	reconnect       bool         //重连标记
	sync.RWMutex                 //读写锁
	LinkType        uint16
	ServerTypeLocal uint16 //本端服务器名称
	ServerTypeOther uint16 //另一端服务器名称
}

func (client *TcpClient) Init(endpoint string, id int, localType uint16, otherType uint16) error {
	var err error
	client.connect, err = net.DialTimeout("tcp", endpoint, time.Second)
	if nil != err {
		log.Error("tcpclient init error:(", endpoint, ",", err.Error(), ")")
		return err
	}

	client.id = id
	client.ServerTypeLocal = localType
	client.ServerTypeOther = otherType
	client.endpoint = endpoint
	client.reconnect = false
	client.writeBuff = make(chan MsgBody, 100)
	return nil
}

func (client *TcpClient) Run() {
	go client.handleRead()
	go client.handleWrite()
}

// 发送消息
func (client *TcpClient) Send(bytes []byte) error {
	netPackage := MsgBody{}
	netPackage.Init(client.ServerTypeLocal, client.ServerTypeOther)
	netPackage.SetData(bytes)
	select {
	case client.writeBuff <- netPackage:
	case <-time.After(time.Millisecond * 10):
		log.Waring("send timeout")
		return errors.New("send timeout")
	}

	return nil
}

// 处理写
func (client *TcpClient) handleWrite() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	for {
		client.RLock()
		if client.stop {
			client.RUnlock()
			return
		}

		client.RUnlock()
		pkg, open := <-client.writeBuff
		if !open || client.connect == nil {
			return
		}

		if _, err := client.connect.Write(pkg.ConvertBytes()); nil != err {
			client.OnException()
			return
		}
	}
}

// 处理读
func (client *TcpClient) handleRead() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	for {
		client.RLock()
		if client.stop {
			client.RUnlock()
			return
		}

		client.RUnlock()
		len := 2
		if client.LinkType == LinkType_Drive {
			len = 2
		} else if client.LinkType == LinkType_Passive {
			len = 4
		}
		lenBuff := make([]byte, len)
		if _, err := io.ReadFull(client.connect, lenBuff); nil != err {
			client.OnException()
			return
		}

		dataLen := 0
		if client.LinkType == LinkType_Drive {
			dataLen = int(binary.LittleEndian.Uint16(lenBuff))
		} else if client.LinkType == LinkType_Passive {
			dataLen = int(binary.LittleEndian.Uint32(lenBuff))
		}

		if dataLen == 0 || dataLen >= 1000*1024 {
			log.Error("消息错误，长度 dataLen:", dataLen)
			if client.connect != nil {
				log.Error("远程IP：", client.connect.RemoteAddr())
			}
			return
		}

		data := make([]byte, dataLen)
		if _, err := io.ReadFull(client.connect, data); nil != err {
			client.OnException()
			return
		}

		ev := new(BackendEvent)
		ev.LinkID = client.id
		ev.Msg = data
		event.Fire("OnBackendMsg", ev)
	}
}

// 重连
func (client *TcpClient) handleReconnect() {
	for {
		var err error
		client.connect, err = net.Dial("tcp", client.endpoint)
		if nil != err {
			time.Sleep(time.Duration(1) * time.Second)
			log.Info("reconnect :", client.endpoint)
			continue
		}

		client.Lock()
		client.reconnect = false
		client.stop = false
		client.Unlock()

		log.Info("reconnect success:", client.endpoint)
		go client.handleRead()
		go client.handleWrite()

		event.Fire("OnBackendReconnect", client)
		break
	}
}

// 异常处理
func (client *TcpClient) OnException() {
	client.Lock()
	if client.reconnect {
		client.Unlock()
		return
	}

	client.stop = true
	client.reconnect = true
	client.Unlock()
	go client.handleReconnect()
}

// 关闭
func (client *TcpClient) Close() {
	client.Lock()
	defer client.Unlock()

	client.stop = true
	close(client.writeBuff)
	client.connect.(*net.TCPConn).Close()
	client.connect = nil
}
