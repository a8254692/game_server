package logic

import (
	"reflect"

	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/game_log"
	"BilliardServer/Util/network"
)

type _GameLog struct {
}

// 游戏日志
var GameLog _GameLog

func (s *_GameLog) Init() {
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Statistics_User_Oline_Num_Request), reflect.ValueOf(s.OnUserOlineNumRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Game_Log_Consume_Resource_Request), reflect.ValueOf(s.OnConsumeResourceRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Game_Log_Production_Resource_Request), reflect.ValueOf(s.OnProductionResourceRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Game_Log_RechargeLog_Request), reflect.ValueOf(s.OnRechargeLogRequest))
}

func (s *_GameLog) OnUserOlineNumRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InStatisticsUserOlineNumRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		return
	}

	game_log.CreateHighOlineLog(req.Num)
	return
}

func (s *_GameLog) OnConsumeResourceRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InGameLogResourceRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		return
	}

	if req.EntityID <= 0 {
		return
	}

	game_log.SaveConsumeLog(req.Uuid, req.EntityID, req.ResType, req.ResSubType, req.ResID, req.IncrType, req.Count, req.AfterModifyNum, req.SystemID, req.ActionID)
	return
}

func (s *_GameLog) OnProductionResourceRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InGameLogResourceRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		return
	}

	if req.EntityID <= 0 {
		return
	}

	game_log.SaveProductionLog(req.Uuid, req.EntityID, req.ResType, req.ResSubType, req.ResID, req.IncrType, req.Count, req.AfterModifyNum, req.SystemID, req.ActionID)
	return
}

func (s *_GameLog) OnRechargeLogRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InGameLogRechargeRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		return
	}

	if req.EntityID <= 0 {
		return
	}

	game_log.SaveRechargeLog(req.EntityID, req.Channel, req.OrderId, req.BeforeRecharge, req.AfterRecharge, req.OrderAmount, req.Discount, req.EventGifts, req.Deduction, req.ActualReceipt, req.CreateOrderTime, req.PayTime, req.RewardItems)
	return
}
