package network

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/websocket"
)

// Tcp服务对象，用于监听端口并构建连接
type WebSocketServer struct {
	addr string // 监听地址
	run  bool   //是否运行
}

type MyHandler func(*websocket.Conn)

func GameHandler(ws *websocket.Conn) {
	//构建Tcp连接
	newWsLink := new(WsLink)
	newWsLink.LinkType = LinkType_Passive
	newWsLink.Connect = ws
	newWsLink.RemoteAddress.Addr = ws.Request().RemoteAddr

	wsLinkEvent := new(WsLinkEvent)
	wsLinkEvent.NewLink = newWsLink
	event.Fire(EK_WsLinkSuccessPassive, wsLinkEvent)

	//必须在这里启动websocket的for循环,否则就会不停的断开，尚不知道为什么
	//但是这样会导致无法主动知道websocket断开了，需要配合心跳包处理，否则会导至wslink的for循环空转占用CPU时间
	newWsLink.Loop()
}

func GameShakeHandler(config *websocket.Config, request *http.Request) error {
	return nil
}

func (h MyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s := websocket.Server{Handler: GameHandler, Handshake: GameShakeHandler}
	s.ServeHTTP(w, req)
}

// 开启一个WebSocket服务
func (this *WebSocketServer) Start(addr string) error {
	log.Print("WebSocket服务启动", addr)

	addr = strings.Replace(addr, "ws://", "", 1)
	index := strings.Index(addr, "/")
	var listenAddr, path string
	if index < 0 {
		listenAddr = addr
		path = "/"
	} else {
		if index == 0 {
			log.Error("监听地址配置错误:", addr)
			return errors.New("监听地址配置错误:" + addr)
		}
		listenAddr = addr[0:index]
		path = addr[index:len(addr)]
	}
	http.Handle(path, MyHandler(GameHandler))
	this.run = false
	this.addr = listenAddr
	return nil
}

// TCP服务器主循环
func (this *WebSocketServer) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()
	this.run = true
	err := http.ListenAndServe(this.addr, nil)
	if err != nil {
		log.Error(err)
	}
}

// 关闭监听
func (this *WebSocketServer) Stop() {
	this.run = false
}
