package main

import (
	gmsg "BilliardServer/Proto/gmsg"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"BilliardServer/ClientTest/model"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/pprof"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
)

// 临时连接器
var MapTc map[uint64]*network.TcpConnect

func main() {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Error(r)
	// 	}
	// }()
	pprofPort := strconv.Itoa(101)
	pprof.Init(pprofPort)
	log.Init("101", "dev")
	event.Init("action", 1000, doStop)
	event.Register("ExecuteCommond", reflect.ValueOf(ExecuteCommond))
	event.Register(network.EK_LinkSuccessDrive, reflect.ValueOf(OnLinkSuccessDrive))
	event.Register(network.EK_ReLink, reflect.ValueOf(OnReLink))
	event.Register(network.EK_LinkOff, reflect.ValueOf(OnLinkOff))
	event.Register(network.EK_ReceiveMsg, reflect.ValueOf(OnReceiveMsgBody)) // 新的事件消息码测试
	event.OnNet(gmsg.MsgTile_Login_EnterGameResponse, reflect.ValueOf(OnEnterGameResponse))

	//event.On("MultiNinjaPointWarEnemyTeamRequest", reflect.ValueOf(TeamRequest))

	//开始接收命令
	log.Info("-----------------Server System is ", runtime.GOOS, "-------------------")
	if runtime.GOOS == "windows" {
		StartCommond()
	}
	go Singallisten()

	timer.Init("action")
	timer.Star()

	MapTc = make(map[uint64]*network.TcpConnect, 0)

	TimerCreaterTcpConnect()

	model.ConnectMananger = MapTc
	//TimerSendMsg()
	//timer.AddTimer(CreaterTcpConnectOne, "CreaterTcpConnectOne", 3*1000, true)

	model.SendTest()
	// timeout := time.After(time.Second * 5)
	// finish := make(chan bool)
	// count := 1
	// go func() {
	// 	for {
	// 		select {
	// 		case <-timeout:
	// 			fmt.Println("timeout")
	// 			finish <- true
	// 			return
	// 		default:
	// 			//fmt.Printf("haha %d\n", count)
	// 			//e1 := "sendmsg1"
	// 			//ExecuteCommond(&e1)
	// 			count++
	// 		}
	// 		time.Sleep(time.Second * 1)
	// 	}
	// }()
	// <-finish
	// fmt.Println("Timer Finish")
	Run()

	// var input string
	// fmt.Scanln(&input)
}

// 监听停止信号
func Singallisten() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	for {
		sig := <-signalChan
		log.Info("get signal:", sig)
		Stop()
		break
	}
}

// 开始接收命令
func StartCommond() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()
		reader := bufio.NewReader(os.Stdin)
		for {
			data, _, _ := reader.ReadLine()
			command := string(data)
			if command == "s" {
				break
			}
			log.Info("-->Execute Commond : ", command)
			event.Fire("ExecuteCommond", &command)
		}
		Stop()
	}()
}

// 服务器运行
func Run() {
	log.Info("-----------------Main Action Is On Running-------------------")

	if !event.Run() {
		log.Error("事件发生错误，停止运行")
	}

	//逻辑停止
	doStop()
}

// 服务器关闭
func Stop() {
	pprof.Endpprof()
	//测试连接也关闭
	for key, value := range MapTc {
		value.CloseConnect()
		delete(MapTc, key)
	}
	//关闭事件通道，停止逻辑
	log.Info("关闭事件运行")
	event.Stop()

	time.AfterFunc(time.Minute, doStop)

	log.Info("-----------------Server Stop---------------------")
}

var stopOnce sync.Once  // 退出逻辑只执行一次
var stopLock sync.Mutex // 退出锁(防止主线程比执行退出逻辑的线程先退出)

// 逻辑真正停止
func doStop() {
	stopLock.Lock()
	defer stopLock.Unlock()

	//stopOnce.Do(logic.StopLogic)
}

// 执行指令
func ExecuteCommond(pCommand *string) {
	command := *pCommand
	log.Info("-->ExecuteCommond：", command)
	if command == "PrintMsg" {
		//PrintPlayerMsg(nil)
		//event.PrintEventMsg(nil)
	} else if command == "c" {
		TimerCreaterTcpConnect()
	} else if command == "msgone" {
		TimerSendMsg()
		//SendMsgBodyPBTest()
	} else if command == "msgloop" {
		TimerSendMsgLoop()
	} else if command == "enter" {
		model.SendEnterGame()
		//SendMsgBodyPBTest
		//()
	} else if command == "http" {
		model.SendHttp()
	} else if command == "test" {
		model.SendTest()
	} else {
		log.Error("不存在指令：", command)
	}
}

// 是否存在此ID的连接器
func IsExistConnectByID(connectID uint64) bool {
	tc := MapTc[connectID]
	if tc == nil {
		return false
	}
	return true
}

// 生成不重复的连接ID
func GetSoleRandomID(start int, end int) uint64 {
	//范围检查
	if end < start {
		return 0
	}
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//生成随机数
	num := uint64(r.Intn((end - start)) + start)
	if IsExistConnectByID(uint64(num)) {
		num = GetSoleRandomID(100000, 999999)
	}
	return num
}
func OnLinkSuccessDrive(tcpConnet *network.TcpConnect) {
	//设置连接ID
	connectID := GetSoleRandomID(100000, 999999)
	tcpConnet.TcpLink.SetID(connectID)
	tcpConnet.SetConnectType(network.ServerType_Client, network.ServerType_Gate)
	MapTc[connectID] = tcpConnet
	log.Info("-->新的主动连接接入 ID："+strconv.FormatUint(connectID, 10), " 连接器总数:", len(MapTc))
	tcpConnet.SendIdentity() //马上发送身份确认消息码
	//同步一下AccID 测试用
	//SendMsgBodyPBSyncAccID(MapTc[tcpConnet.TcpLink.GetID()])
}

// 处理主动连接的重连
func OnReLink(ev *network.LinkEvent) {
	tc := MapTc[ev.NewLink.GetID()]
	if tc == nil {
		log.Waring("-->重连时未找到指定ID的TcpConnect对象 connectID:", strconv.FormatUint(ev.NewLink.GetID(), 10))
		return
	}
	tc.SendIdentity() //马上发送身份确认消息码
	//同步一下AccID 测试用
	//SendMsgBodyPBSyncAccID(MapTc[tc.TcpLink.GetID()])
}
func OnLinkOff(ev *network.LinkEvent) {
	tc := MapTc[ev.NewLink.GetID()]
	if tc == nil {
		log.Waring("-->关闭TcpConnect连接时未找到指定的ID:", ev.NewLink.GetID())
		return
	}
	tc.CloseConnect()
	delete(MapTc, ev.NewLink.GetID())
	log.Info("-->关闭一条连接 ID："+strconv.FormatUint(ev.NewLink.GetID(), 10), " 连接器总数:", len(MapTc))
}
func OnReceiveMsgBody(msgEV *network.MsgBodyEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()
	if msgEV.MsgTile == network.Net_Identity {
		log.Info("-->测试前端收到身份认证消息")
	} else if msgEV.MsgTile == network.Net_Subscribemsg {
		log.Info("-->测试前端收到消息订阅：", msgEV.MsgTile)
	} else {
		log.Info("-->测试前端收到消息 MsgTile：", msgEV.MsgTile)
		event.EmitNet(msgEV.MsgTile, msgEV)
	}
}
func TeamRequest(msgEV *network.MsgBodyEvent) {
	log.Info("-->测试前端收到消息：", msgEV.MsgTile)
}

func TimerSendMsgLoop() {
	//timer.AddTimer(SendMsgBodyPBTest, "SendMsgBodyPBTest", 10*1000, false)
}

// 创新一批新的连接器
func CreaterTcpConnect(num uint16) {
	for i := 0; i < int(num); i++ {
		tc := new(network.TcpConnect)
		tc.Init("127.0.0.1:7060", network.ServerType_Client, network.ServerType_Gate)
	}
}
func TimerCreaterTcpConnect() {
	tim := time.NewTicker(100 * time.Millisecond)
	<-tim.C
	go CreaterTcpConnectOne()
}

func CreaterTcpConnectOne() {
	if len(MapTc) >= 1 {
		return
	}
	tc := new(network.TcpConnect)
	tc.Init("127.0.0.1:7060", network.ServerType_Client, network.ServerType_Gate)
	TimerCreaterTcpConnect()
}
func TimerSendMsg() {
	tim := time.NewTicker(1 * time.Second)
	<-tim.C
	go SendMsgOne()
}

var Idx int = 0

func SendMsgOne() {
	//request := new(gmsg.LS2CS_MultiNinjaPointWarEnemyTeamRequest)
	//request.ZoneID = 1001
	//request.ObjID = "player01"
	//request.EnemyObjID = "10001"
	//request.EnemyServerID = "101"
	//request.Index = 1001
	//request.TeamType = proto.Int32(1002)
	//mapKeyArgs := make([]uint64, len(MapTc))
	//for key, _ := range MapTc {
	//	mapKeyArgs = append(mapKeyArgs, key)
	//	request.ZoneID = int32(MapTc[key].TcpLink.GetID())
	//	MapTc[key].SendMsgBodyPB(msg.Sys_Test, request)
	//}
	// MapTc[mapKeyArgs[Idx]].SendMsgBodyPB("Msg_MultiNinjaPointWarEnemyTeam", request)
	// Idx++
	// if Idx > len(MapTc) {
	// 	Idx = 0
	// }
	TimerSendMsg()
}

var i int

func OnEnterGameResponse() {
	i++

	fmt.Printf("OnEnterGameResponse--%d", i)
	return
}
