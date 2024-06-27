package logic

import (
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/network"
	"time"
)

// 全服广播跑马灯消息
func SendMarqueeMsgSync(mType uint32, context string) {
	resp := &gmsg.MarqueeMsgSync{
		MarqueeType: mType,
		Context:     context,
	}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCastAll(gmsg.MsgTile_Hall_GetMarqueeMsgSync, resp)
	return
}

func SendConsumeResourceLogToDb(uuid string, entityID uint32, resType uint32, resSubType uint32, resID uint32, incrType uint32, count uint64, afterModifyNum uint32, systemID uint32, actionID uint32) {
	now := time.Now().Unix()
	logResp := &gmsg.InGameLogResourceRequest{
		EntityID:       entityID,
		Time:           now,
		Uuid:           uuid,
		ResType:        resType,
		ResSubType:     resSubType,
		ResID:          resID,
		IncrType:       incrType,
		Count:          count,
		AfterModifyNum: afterModifyNum,
		SystemID:       systemID,
		ActionID:       actionID,
		DeviceID:       "",
		ChannelID:      0,
		BundleID:       "",
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Game_Log_Consume_Resource_Request), logResp, network.ServerType_DB)
	return
}

func SendProductionResourceLogToDb(uuid string, entityID uint32, resType uint32, resSubType uint32, resID uint32, incrType uint32, count uint64, afterModifyNum uint32, systemID uint32, actionID uint32) {
	now := time.Now().Unix()
	logResp := &gmsg.InGameLogResourceRequest{
		EntityID:       entityID,
		Time:           now,
		Uuid:           uuid,
		ResType:        resType,
		ResSubType:     resSubType,
		ResID:          resID,
		IncrType:       incrType,
		Count:          count,
		AfterModifyNum: afterModifyNum,
		SystemID:       systemID,
		ActionID:       actionID,
		DeviceID:       "",
		ChannelID:      0,
		BundleID:       "",
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Game_Log_Production_Resource_Request), logResp, network.ServerType_DB)
	return
}

func SendRechargeLogToDb(entityID uint32, channelID, orderID, beforeRecharge, afterRecharge uint32, orderAmount, discount, eventGifts, deduction, actualReceipt float32, createOrderTime, payTime int64, typeN uint32, rewardItems []*gmsg.InRewardInfo) {
	now := time.Now().Unix()
	logResp := &gmsg.InGameLogRechargeRequest{
		EntityID:        entityID,
		Time:            now,
		Channel:         channelID,
		OrderId:         orderID,
		BeforeRecharge:  beforeRecharge,
		AfterRecharge:   afterRecharge,
		OrderAmount:     orderAmount,
		Discount:        discount,
		EventGifts:      eventGifts,
		Deduction:       deduction,
		ActualReceipt:   actualReceipt,
		CreateOrderTime: createOrderTime,
		PayTime:         payTime,
		TypeN:           typeN,
		RewardItems:     rewardItems,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Game_Log_RechargeLog_Request), logResp, network.ServerType_DB)
	return
}
