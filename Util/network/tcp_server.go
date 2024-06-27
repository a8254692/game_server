// tcp_server
package network

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"net"
)

// Tcp服务对象，用于监听端口并构建连接
type TcpServer struct {
	listener net.Listener //tcp监听对象
	addr     string       //监听地址
	run      bool         //是否运行
}

// 开启一个TCP服务
func (this *TcpServer) Start(addr string) error {
	log.Print("-->TCP服务启动", addr)
	var err error
	this.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	this.addr = addr
	this.run = false
	return nil
}

// TCP服务器主循环
func (this *TcpServer) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	this.run = true

	for this.run {
		tcpConnet, err := this.listener.Accept()
		if err != nil {
			log.Error(err)
		}

		//构建Tcp连接
		newLink := new(TcpLink)
		newLink.LinkType = LinkType_Passive
		newLink.Connect = tcpConnet

		linkEvent := new(LinkEvent)
		linkEvent.NewLink = newLink
		event.Fire(EK_LinkSuccessPassive, linkEvent)
	}
}

// 关闭监听
func (this *TcpServer) Stop() {
	this.run = false
}
