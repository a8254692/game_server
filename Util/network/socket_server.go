// tcp_server
package network

//Tcp服务对象，用于监听端口并构建连接
type SocketServer interface {
	Start(addr string) error
	Run()
	Stop()
}