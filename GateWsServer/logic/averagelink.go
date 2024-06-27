package logic

import (
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"encoding/binary"
	"google.golang.org/protobuf/proto"
)

type _AverageLink struct {
}

var AverageLink _AverageLink

func (this *_AverageLink) Init() {
	//注册逻辑业务事件
	//event.OnNet(network.Net_SyncEntityID, reflect.ValueOf(OnSyncEntityID))
	//event.OnNet(msg.Sys_HeartBeatRequest, reflect.ValueOf(OnHeartBeat))
	//event.OnNet(msg.Login_EnterGameRequest, reflect.ValueOf(OnEnterGameRequest))
}

func OnSyncEntityID(msgEV *network.MsgBodyEvent) {
	tEntityID := binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
	WsManager.ChangeTcpConnectID(msgEV.WsLink.GetID(), uint64(tEntityID))
	WsManager.SendMsgToClient(network.Net_SyncEntityID, msgEV.MsgBody, uint64(tEntityID))
}
func OnEnterGameRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EnterGameRequest{}
	err := proto.Unmarshal(msgEV.MsgBody, msgBody)
	if err != nil {
		log.Waring(err)
		log.Waring("反序列化失败 funName = ", msgEV.MsgTile)
		return
	}
	WsManager.ChangeTcpConnectID(msgEV.WsLink.GetID(), uint64(msgBody.EntityId))
	WsManager.SendMsgBodyPB(uint32(gmsg.MsgTile_Login_EnterGameRequest), msgBody)
}

func OnHeartBeat(msgEV *network.MsgBodyEvent) {
	// msgRequest := &msg.HeartBeatRequest{}
	// err := msgEV.Unmarshal(msgRequest)
	// if err != nil {
	// 	return
	// }
	tEntityID := binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
	tWsConnect := WsManager.GetWsConnectByID(uint64(tEntityID))
	tWsConnect.CheckHeartBeat()
	msgResponse := &gmsg.HeartBeatResponse{}
	msgResponse.Result = 0
	msgResponse.Code = 0
	msgResponse.EntityId = 1000001
	WsManager.SendMsgPbToClientAll(uint32(gmsg.MsgTile_Sys_HeartBeatResponse), msgResponse)
}
