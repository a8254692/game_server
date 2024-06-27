package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"reflect"
)

/***
 *@disc: 成就
 *@author: lsj
 *@date: 2023/10/12
 */

type _Achievement struct {
}

var Achievement _Achievement

func (c *_Achievement) Init() {
	event.OnNet(gmsg.MsgTile_Player_AchievementLvClaimRewardRequest, reflect.ValueOf(c.OnPlayerAchievementLVClaimRewardRequest))
	event.OnNet(gmsg.MsgTile_Player_AchievementLvClaimRewardListRequest, reflect.ValueOf(c.OnPlayerAchievementLVClaimRewardListRequest))
	event.OnNet(gmsg.MsgTile_Player_AchievementListRequest, reflect.ValueOf(c.OnPlayerAchievementListRequest))
	event.OnNet(gmsg.MsgTile_Player_AchievementChildListRequest, reflect.ValueOf(c.OnPlayerAchievementChildListRequest))
	event.OnNet(gmsg.MsgTile_Player_GameAchievementListRequest, reflect.ValueOf(c.OnGameAchievementListRequest))
}

// 获取成就等级，最高30,超出后就返回30
func (c *_Achievement) getAchievementLVRewardID(tEntityPlayer *entity.EntityPlayer) (resID uint32, isCanClaim bool) {
	for _, val := range tEntityPlayer.AchievementLVRewardList {
		if val.StateReward == 0 {
			resID = val.AchievementLvID
			break
		}
	}
	if resID <= tEntityPlayer.AchievementLV {
		isCanClaim = true
	}
	if resID == 0 && tEntityPlayer.AchievementLV == conf.MaxAchievementLV {
		resID = conf.MaxAchievementLV
		isCanClaim = false
	}
	return resID, isCanClaim
}

// 领取成就等级奖励
func (c *_Achievement) OnPlayerAchievementLVClaimRewardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.AchievementLvClaimRewardRequest{}
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

	log.Info("-->OnPlayerAchievementLVClaimRewardRequest--begin-->", msgBody)
	if !c.isAchievementLVid(msgBody.AchievementLvID) {
		return
	}
	msgResponse := &gmsg.AchievementLvClaimRewardResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = 1

	if tEntityPlayer.AchievementLV >= msgBody.AchievementLvID && !tEntityPlayer.IsInAchievementLVRewardList(msgBody.AchievementLvID) {
		tEntityPlayer.AchievementLVClaimReward(msgBody.AchievementLvID)
		tEntityPlayer.SaveNextRewardAchievementLV(msgBody.AchievementLvID)
		tEntityPlayer.SyncEntity(1)
		msgResponse.Code = 0
		resAchievementLVid, isCanClaim := c.getAchievementLVRewardID(tEntityPlayer)
		msgResponse.NextAchievementLvID = resAchievementLVid
		msgResponse.IsCanClaim = isCanClaim
		target := Table.GetAchievementLvCfg(msgBody.AchievementLvID)
		if target == nil {
			return
		}
		rewardEntityList := make([]entity.RewardEntity, 0)
		for _, vl := range target.Reward {
			rewardEntity := new(entity.RewardEntity)
			rewardEntity.ItemTableId = vl[0]
			rewardEntity.Num = vl[1]
			rewardEntity.ExpireTimeId = 0
			rewardEntityList = append(rewardEntityList, *rewardEntity)
		}

		resParam := GetResParam(conf.SYSTEM_ID_TASK, conf.Reward)
		Backpack.BackpackAddItemListAndSave(msgBody.EntityID, rewardEntityList, *resParam)
	}
	log.Info("--OnPlayerAchievementLVClaimRewardRequest--end->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_AchievementLvClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Achievement) isAchievementLVid(achievementLVid uint32) bool {
	return achievementLVid > 0 && achievementLVid <= conf.MaxAchievementLV
}

// 领取列表
func (c *_Achievement) OnPlayerAchievementLVClaimRewardListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.AchievementLvClaimRewardListRequest{}
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

	msgResponse := &gmsg.AchievementLvClaimRewardListResponse{}
	msgResponse.RewardList = make([]*gmsg.AchievementLvReward, 0)
	for _, vl := range tEntityPlayer.AchievementLVRewardList {
		reward := new(gmsg.AchievementLvReward)
		stack.SimpleCopyProperties(reward, vl)
		msgResponse.RewardList = append(msgResponse.RewardList, reward)
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_AchievementLvClaimRewardListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 成就列表
func (c *_Achievement) OnPlayerAchievementListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.AchievementListRequest{}
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

	msgResponse := &gmsg.AchievementListResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.AchievementList = make([]*gmsg.Achievement, 0)
	for _, v := range tEntityPlayer.AchievementList {
		achievement := new(gmsg.Achievement)
		achievement.AchievementID = v.AchievementID
		achievement.TypeN = v.TypeN
		childList := make([]*gmsg.ChildAchievement, 0)
		for _, vl := range v.ChildList {
			child := new(gmsg.ChildAchievement)
			stack.SimpleCopyProperties(child, vl)
			child.AddTime = tools.FormatTimeStr(child.AddTime[0:10], "-")
			childList = append(childList, child)
		}
		achievement.ChildList = childList
		msgResponse.AchievementList = append(msgResponse.AchievementList, achievement)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_AchievementListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 获取子成就列表
func (c *_Achievement) OnPlayerAchievementChildListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.AchievementChildListRequest{}
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

	msgResponse := &gmsg.AchievementChildListResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.AchievementChildList = make([]*gmsg.ChildAchievement, 0)
	for _, v := range tEntityPlayer.GetChildAchievementList(msgBody.AchievementID) {
		achievement := new(gmsg.ChildAchievement)
		stack.SimpleCopyProperties(achievement, &v)
		achievement.AddTime = tools.FormatTimeStr(achievement.AddTime[0:10], "-")
		msgResponse.AchievementChildList = append(msgResponse.AchievementChildList, achievement)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_AchievementChildListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 获取游戏成就请求
func (c *_Achievement) OnGameAchievementListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.GameAchievementListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.GameAchievementListResponse{}
	msgResponse.AchievementList = make([]*gmsg.Achievement, 0)
	for _, v := range tEntityPlayer.AchievementList {
		if v.TypeN == conf.AchievementBattle {
			achievement := new(gmsg.Achievement)
			achievement.AchievementID = v.AchievementID
			achievement.TypeN = v.TypeN
			childList := make([]*gmsg.ChildAchievement, 0)
			for _, vl := range v.ChildList {
				child := new(gmsg.ChildAchievement)
				stack.SimpleCopyProperties(child, vl)
				child.AddTime = tools.FormatTimeStr(child.AddTime[0:10], "-")
				childList = append(childList, child)
			}
			achievement.ChildList = childList
			msgResponse.AchievementList = append(msgResponse.AchievementList, achievement)
		}
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_GameAchievementListResponse, msgResponse, []uint32{msgBody.EntityID})
}
