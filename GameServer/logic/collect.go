package logic

import (
	"BilliardServer/Common/entity"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"reflect"
)

/***
 *@disc: 称号
 *@author: lsj
 *@date: 2023/10/11
 */

type _Collect struct {
}

var Collect _Collect

func (c *_Collect) Init() {
	event.OnNet(gmsg.MsgTile_Player_CollectApplyRequest, reflect.ValueOf(c.OnCollectApplyRequest))
	event.OnNet(gmsg.MsgTile_Player_CollectActivateRequest, reflect.ValueOf(c.OnCollectActivateRequest))
}

// 称号应用
func (c *_Collect) OnCollectApplyRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.CollectApplyRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnCollectApplyRequest-->", msgBody)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	msgResponse := &gmsg.CollectApplyResponse{}
	msgResponse.Code = 1

	if tEntityPlayer.GetCollect(msgBody.CollectID) == nil || tEntityPlayer.GetCollect(msgBody.CollectID).State <= 1 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectApplyResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	collect := tEntityPlayer.CollectApply(msgBody.CollectID, tEntityPlayer.CollectId)
	if collect == nil {
		log.Error("称号异常。", msgBody.EntityID, "-->collectid", msgBody.CollectID)
		msgResponse.Code = 2
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectApplyResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	msgResponse.Code = 0
	msgResponse.Collect = new(gmsg.CollectInfo)
	stack.SimpleCopyProperties(msgResponse.Collect, collect)
	// 同步称号
	Player.PlayerCollectIDSync(msgBody.EntityID, msgBody.CollectID)
	tEntityPlayer.SyncEntity(1)
	log.Info("-->OnCollectApplyRequest-->msgResponse->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectApplyResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 称号激活
func (c *_Collect) OnCollectActivateRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ActivateCollectRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}
	log.Info("-->OnCollectActivateRequest-->begin->", msgBody)
	msgResponse := &gmsg.ActivateCollectResponse{}
	msgResponse.Code = 1

	if tEntityPlayer.GetCollect(msgBody.CollectID) == nil || tEntityPlayer.GetCollect(msgBody.CollectID).State == 0 {
		log.Info("-->OnCollectActivateRequest-->end->", msgResponse)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectActivateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	collect := tEntityPlayer.CollectActivate(msgBody.CollectID)
	if collect == nil {
		log.Error("称号异常。", msgBody.EntityID, "-->collectid", msgBody.CollectID)
		log.Info("-->OnCollectActivateRequest-->end->", msgResponse)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectActivateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	msgResponse.Code = 0
	msgResponse.Collect = new(gmsg.CollectInfo)
	stack.SimpleCopyProperties(msgResponse.Collect, collect)
	tEntityPlayer.SyncEntity(1)
	log.Info("-->OnCollectActivateRequest-->end->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectActivateResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Collect) getCollectList(tEntityPlayer *entity.EntityPlayer) (collectList []*gmsg.CollectInfo) {
	for _, vl := range tEntityPlayer.CollectList {
		collect := new(gmsg.CollectInfo)
		stack.SimpleCopyProperties(collect, vl)
		collectList = append(collectList, collect)
	}
	return
}
