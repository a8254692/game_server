package model

import (
	"BilliardServer/WebServer/controllers/response"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
)

var ConnectMananger map[uint64]*network.TcpConnect

func SendEnterGame() {
	request := new(gmsg.EnterGameRequest)
	request.EntityId = 100100011
	request.Token = "test"
	for key, _ := range ConnectMananger {
		ConnectMananger[key].SendMsgBodyPB(uint32(gmsg.MsgTile_Login_EnterGameRequest), request)
	}
}

func SendTest() {
	res, _ := http.Get("http://127.0.0.1:7120/v1/login?user_name=zq1&is_iphone=false&platform=1&login_platform=1&channel=1&device_id=xxx&machine=xxx&remote_addr=xxx&package_name=xxx&language=1")
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	a := new(response.Login)
	_ = json.Unmarshal(body, a)
	log.Info(a)

	tc := new(network.TcpConnect)
	tc.Init("127.0.0.1:7060", network.ServerType_Client, network.ServerType_Gate)

	request := new(gmsg.EnterGameRequest)
	request.EntityId = a.EntityID
	request.Token = a.Token
	tc.SendMsgBodyPB(uint32(gmsg.MsgTile_Login_EnterGameRequest), request)

	fmt.Println("111")
}

func SendHttp() {
	res, _ := http.Get("http://127.0.0.1:7120/loginer?funname=login&username=aa&password=bb")
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	a := new(response.Login)
	json.Unmarshal(body, a)
	log.Info(a)
}
