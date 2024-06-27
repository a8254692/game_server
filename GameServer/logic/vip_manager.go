package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
	"time"
)

var VipMgr _VipMgr

type _VipMgr struct {
}

func (s *_VipMgr) Init() {

	//注册逻辑业务事件
	event.OnNet(gmsg.MsgTile_Vip_GetVipDailyBoxRequest, reflect.ValueOf(s.OnGetVipDailyBoxRequest))
	event.OnNet(gmsg.MsgTile_Vip_GetVipLvRewardRequest, reflect.ValueOf(s.OnGetVipLvRewardRequest))
}

func (s *_VipMgr) isGetDailyBox(tEntityPlayer *entity.EntityPlayer) gmsg.ReceiveStatus {
	resp := gmsg.ReceiveStatus_Receive_Status_Yes

	if tEntityPlayer == nil {
		log.Waring("-->logic--_VipMgr--isGetDailyBox--tEntityPlayer == nil")
		return resp
	}

	var isInHis bool
	now := time.Now()
	for _, v := range tEntityPlayer.SignInRewardList {
		if v.ID == uint64(gmsg.SystemActivityId_Sys_Activity_VipDailySign) {
			isInHis = true

			if v.LastSignInUnixSec < 0 {
				continue
			}

			lastTime := time.Unix(v.LastSignInUnixSec, 0)
			if now.Year() <= lastTime.Year() && now.Month() <= lastTime.Month() && now.Day() <= lastTime.Day() {
				continue
			}

			resp = gmsg.ReceiveStatus_Receive_Status_No
		}
	}

	if !isInHis {
		resp = gmsg.ReceiveStatus_Receive_Status_No
	}

	return resp
}

func (s *_VipMgr) setDailyBox(tEntityPlayer *entity.EntityPlayer) {
	if tEntityPlayer == nil {
		log.Waring("-->logic--_VipMgr--setDailyBox--tEntityPlayer == nil")
		return
	}

	var isInHis bool
	now := time.Now()
	for k, v := range tEntityPlayer.SignInRewardList {
		if v.ID == uint64(gmsg.SystemActivityId_Sys_Activity_VipDailySign) {
			isInHis = true

			if v.LastSignInUnixSec < 0 {
				continue
			}

			lastTime := time.Unix(v.LastSignInUnixSec, 0)
			if now.Year() <= lastTime.Year() && now.Month() <= lastTime.Month() && now.Day() <= lastTime.Day() {
				continue
			}

			tEntityPlayer.SignInRewardList[k].LastSignInUnixSec = now.Unix()
			tEntityPlayer.SyncEntity(1)
		}
	}

	if !isInHis {
		info := entity.SignInReward{
			ID:                 uint64(gmsg.SystemActivityId_Sys_Activity_VipDailySign),
			SignLog:            nil,
			LastSignInUnixSec:  now.Unix(),
			FirstSignInUnixSec: now.Unix(),
		}
		tEntityPlayer.SignInRewardList = append(tEntityPlayer.SignInRewardList, info)
		tEntityPlayer.SyncEntity(1)
	}

	return
}

// OnGetVipList 获取vip列表
func (s *_VipMgr) OnGetVipList(EntityID uint32) []*gmsg.VipInfo {
	respList := make([]*gmsg.VipInfo, 0)
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		log.Waring("-->logic--_VipMgr--isGetDailyBox--GetEntityByID--tEntity == nil")
		return respList
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	list := VipExp.GetExpConfList()

	if len(list) > 0 {
		for _, v := range list {
			isGetLvReward := gmsg.ReceiveStatus_Receive_Status_No
			if len(tEntityPlayer.VipLvReward) > 0 {
				for _, lv := range tEntityPlayer.VipLvReward {
					if lv == v.Level {
						isGetLvReward = gmsg.ReceiveStatus_Receive_Status_Yes
					}
				}
			}

			info := &gmsg.VipInfo{
				Level:           v.Level,
				GetRewardStatus: isGetLvReward,
			}
			respList = append(respList, info)
		}
	}

	return respList
}

// OnGetVipDailyBoxRequest 获取VIP每日礼包请求
func (s *_VipMgr) OnGetVipDailyBoxRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetVipDailyBoxRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_VipMgr--OnGetVipDailyBoxRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 {
		log.Waring("-->logic--_VipMgr--OnGetVipDailyBoxRequest--req.EntityID <= 0")
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
	if tEntity == nil {
		log.Waring("-->logic--_VipMgr--OnGetVipDailyBoxRequest--GetEntityByID--tEntity == nil")
		return
	}

	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		log.Waring("-->logic--_VipMgr--OnGetVipDailyBoxRequest--tEntityPlayer == nil")
		return
	}

	if s.isGetDailyBox(tEntityPlayer) == gmsg.ReceiveStatus_Receive_Status_No {
		s.setDailyBox(tEntityPlayer)

		resParam := GetResParam(consts.SYSTEM_ID_VIP, consts.Reward)
		conf := VipExp.GetExpConf(tEntityPlayer.VipLv)
		//发奖
		RewardManager.AddRewardByRegularList(req.EntityID, conf.Box, *resParam)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetVipDailyBoxResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.Status = gmsg.ReceiveStatus_Receive_Status_Yes

	log.Info("-->logic--_VipMgr--OnGetVipDailyBoxRequest--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Vip_GetVipDailyBoxResponse, resp, []uint32{req.EntityID})
	return
}

// OnGetVipLvRewardRequest 获取VIP等级礼包返回
func (s *_VipMgr) OnGetVipLvRewardRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetVipLvRewardRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_VipMgr--OnGetVipLvRewardRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.VipLv <= 0 {
		log.Waring("-->logic--_VipMgr--OnGetVipLvRewardRequest--req.EntityID <= 0")
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
	if tEntity == nil {
		log.Waring("-->logic--_VipMgr--OnGetVipLvRewardRequest--GetEntityByID--tEntity == nil")
		return
	}

	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		log.Waring("-->logic--_VipMgr--OnGetVipLvRewardRequest--tEntityPlayer == nil")
		return
	}

	if req.VipLv > tEntityPlayer.VipLv {
		log.Waring("-->logic--_VipMgr--OnGetVipLvRewardRequest--req.VipLv > tEntityPlayer.VipLv")
		return
	}

	var isGetLvReward bool
	if len(tEntityPlayer.VipLvReward) > 0 {
		for _, v := range tEntityPlayer.VipLvReward {
			if v == req.VipLv {
				isGetLvReward = true
			}
		}
	}

	//未购买则进行购买逻辑
	if !isGetLvReward {
		conf := VipExp.GetExpConf(req.VipLv)

		var canBuy bool
		var consumeGoldNum uint32
		var consumeStoneNum uint32
		if len(conf.Price) > 0 {
			for k, v := range conf.Price {
				//TODO:货币类型需要修改
				if k == consts.Gold {
					consumeGoldNum += v
					canBuy = true
				} else if k == consts.Diamond {
					consumeStoneNum += v
					canBuy = true
				}
			}
		}

		if canBuy {
			var isSuccessDeduct bool
			if tEntityPlayer.NumGold >= consumeGoldNum {
				isSuccessDeduct = true
			}

			if tEntityPlayer.NumStone >= consumeStoneNum {
				isSuccessDeduct = true
			}

			//发放物品
			if len(conf.Reward) > 0 && isSuccessDeduct {
				buySource := GetResParam(consts.SYSTEM_ID_VIP, consts.BuyVipGift)
				if consumeGoldNum > 0 {
					Player.UpdatePlayerPropertyItem(req.EntityID, consts.Gold, int32(-consumeGoldNum), *buySource)
				}
				if consumeStoneNum > 0 {
					Player.UpdatePlayerPropertyItem(req.EntityID, consts.Gold, int32(-consumeStoneNum), *buySource)
				}

				RewardManager.AddRewardByRegularList(req.EntityID, conf.Reward, *buySource)
			}
		}

		tEntityPlayer.VipLvReward = append(tEntityPlayer.VipLvReward, req.VipLv)
		tEntityPlayer.SyncEntity(1)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetVipLvRewardResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.Status = gmsg.ReceiveStatus_Receive_Status_Yes

	log.Info("-->logic--_VipMgr--OnGetVipLvRewardRequest--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Vip_GetVipLvRewardResponse, resp, []uint32{req.EntityID})
	return
}
