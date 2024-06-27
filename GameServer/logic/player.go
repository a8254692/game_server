package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/tools"
	"gitee.com/go-package/carbon/v2"
	"reflect"
)

type _Player struct {
	PlayerBaseList map[uint32]*gmsg.PlayerBase
}

var Player _Player

func (s *_Player) Init() {
	s.PlayerBaseList = make(map[uint32]*gmsg.PlayerBase, 0)
	event.OnNet(gmsg.MsgTile_Player_ChangeIconRequest, reflect.ValueOf(s.OnChangeIconRequest))
	event.OnNet(gmsg.MsgTile_Player_ChangeNameRequest, reflect.ValueOf(s.OnChangeNameRequest))
	event.OnNet(gmsg.MsgTile_Player_ChangeSignRequest, reflect.ValueOf(s.OnChangePlayerSignRequest))
	event.OnNet(gmsg.MsgTile_Player_InfoRequest, reflect.ValueOf(s.OnPlayerInfoRequest))
	event.OnNet(gmsg.MsgTile_Player_InfoResponse, reflect.ValueOf(s.OnPlayerInfoResponse))
	event.OnNet(gmsg.MsgTile_Player_QueryEntityPlayerByIDRequest, reflect.ValueOf(s.OnQueryEntityPlayerByIDRequest))
	event.OnNet(gmsg.MsgTile_Player_QueryEntityPlayerByIDResponse, reflect.ValueOf(s.OnQueryEntityPlayerByIDResponse))
	event.OnNet(gmsg.MsgTile_Login_PlayerChangeSexRequest, reflect.ValueOf(s.OnChangePlayerSexRequest))
}

func (s *_Player) syncPlayerBase(list []*gmsg.PlayerBase) {
	for _, val := range list {
		s.PlayerBaseList[val.EntityID] = val
	}
}

// 修改角色头像，前端->游戏服->DB服
func (s *_Player) OnChangeIconRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ChangePlayerIconRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}

	s.ChangeIcon(msgBody.EntityID, *msgBody.PlayerIcon)

	msgResponse := &gmsg.ChangePlayerIconResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = 0
	msgResponse.PlayerIcon = msgBody.PlayerIcon
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_ChangeIconResponse, msgResponse, []uint32{msgBody.EntityID})
	return
}

// 修改角色头像
func (s *_Player) ChangeIcon(EntityID uint32, PlayerIcon uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.PlayerIcon = PlayerIcon
	tEntityPlayer.SyncEntity(1)

	msgBody := &gmsg.PlayerIconSync{}
	msgBody.PlayerIcon = PlayerIcon
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerIconSync, msgBody, []uint32{EntityID})
	return
}

// 修改角色球杆
func (s *_Player) ChangeCueTableID(EntityID uint32, CueTableID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.CueTableId = CueTableID
	tEntityPlayer.SyncEntity(1)

	msgBody := &gmsg.PlayerCueInfoSync{}
	msgBody.CueTableId = CueTableID
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerCueInfoSync, msgBody, []uint32{EntityID})
	return
}

// 修改角色服装
func (s *_Player) ChangePlayerDress(EntityID uint32, PlayerDress uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.PlayerDress = PlayerDress
	tEntityPlayer.SyncEntity(1)

	msgBody := &gmsg.PlayerDressSync{}
	msgBody.PlayerDress = PlayerDress
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerDressSync, msgBody, []uint32{EntityID})
	return
}

// 装扮同步
func (s *_Player) ChangeClothing(EntityID uint32, SubType, ClothingId uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if SubType == conf.Clothing_1 {
		tEntityPlayer.PlayerIcon = ClothingId
	} else if SubType == conf.Clothing_2 {
		tEntityPlayer.IconFrame = ClothingId
	} else if SubType == conf.Clothing_3 {
		tEntityPlayer.ClothingBubble = ClothingId
	} else if SubType == conf.Clothing_4 {
		tEntityPlayer.ClothingCountDown = ClothingId
	} else {
		log.Error("无效的id", "->ClothingId->", ClothingId, "->SubType->", SubType)
	}

	tEntityPlayer.SyncEntity(1)

	msgBody := &gmsg.PlayerClothingSync{}
	msgBody.PlayerIcon = tEntityPlayer.PlayerIcon
	msgBody.IconFrame = tEntityPlayer.IconFrame
	msgBody.ClothingBubble = tEntityPlayer.ClothingBubble
	msgBody.ClothingCountDown = tEntityPlayer.ClothingCountDown
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerClothingSync, msgBody, []uint32{EntityID})
	return
}

// 特效同步
func (s *_Player) ChangeEffect(EntityID uint32, SubType, EffectId uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if SubType == conf.Effect_1 {
		tEntityPlayer.TableCloth = EffectId
	} else if SubType == conf.Effect_2 {
		tEntityPlayer.BattingEffect = EffectId
	} else if SubType == conf.Effect_3 {
		tEntityPlayer.GoalInEffect = EffectId
	} else if SubType == conf.Effect_4 {
		tEntityPlayer.CueBall = EffectId
	} else {
		log.Error("无效的id", "->EffectId->", EffectId, "->SubType->", SubType)
	}

	tEntityPlayer.SyncEntity(1)

	msgBody := &gmsg.PlayerEffectSync{}
	msgBody.TableCloth = tEntityPlayer.TableCloth
	msgBody.BattingEffect = tEntityPlayer.BattingEffect
	msgBody.GoalInEffect = tEntityPlayer.GoalInEffect
	msgBody.CueBall = tEntityPlayer.CueBall
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerEffectSync, msgBody, []uint32{EntityID})
	return
}

// 同步角色称号
func (s *_Player) PlayerCollectIDSync(EntityID uint32, collectId uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.CollectId = collectId
	tEntityPlayer.SyncEntity(1)

	msgBody := &gmsg.PlayerCollectIDSync{}
	msgBody.CollectID = collectId
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerCollectIDSync, msgBody, []uint32{EntityID})
	return
}

// 同步成就和积分
func (s *_Player) PlayerAchievementLVAndScoreSync(EntityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgBody := &gmsg.PlayerAchievementLVAndScoreSync{}
	msgBody.AchievementLV = tEntityPlayer.AchievementLV
	msgBody.AchievementScore = tEntityPlayer.AchievementScore
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerAchievementLvAndScoreSync, msgBody, []uint32{EntityID})
	return
}

// 同步成就等级
func (s *_Player) PlayerAchievementLVSync(EntityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgBody := &gmsg.AchievementLvSync{}
	_, isCanClaim := Achievement.getAchievementLVRewardID(tEntityPlayer)
	msgBody.AchievementLvID = tEntityPlayer.AchievementLV
	msgBody.IsCanClaim = isCanClaim
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_AchievementLvSync, msgBody, []uint32{EntityID})
	return
}

// 修改角色名称，前端->游戏服->DB服
func (s *_Player) OnChangeNameRequest(msgEV *network.MsgBodyEvent) {
	log.Info("-->OnChangeNameRequest--------------game-------")
	msgBody := &gmsg.ChangePlayerNameRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityId)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse := &gmsg.ChangePlayerNameResponse{}
	if len(*msgBody.PlayerName) > 10 {
		msgResponse.EntityId = msgBody.EntityId
		msgResponse.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_ChangeNameResponse, msgResponse, []uint32{msgBody.EntityId})
		return
	}
	tEntityPlayer.PlayerName = *msgBody.PlayerName
	tEntityPlayer.SyncEntity(1)

	msgResponse.EntityId = msgBody.EntityId
	msgResponse.Code = 0
	msgResponse.PlayerName = msgBody.PlayerName
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_ChangeNameResponse, msgResponse, []uint32{msgBody.EntityId})
	return
}

// 修改性别 ，前端->游戏->db服
func (s *_Player) OnChangePlayerSexRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ChangePlayerSexRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse := &gmsg.ChangePlayerSexResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = 1
	num, index := Backpack.GetPlayerChangeSexRaw(tEntityPlayer)
	if num == uint32(0) {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerChangeSexResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	newSex := s.getPlayerNewSex(tEntityPlayer.Sex)

	//修改默认的服装
	newDressId := Table.DefaultDress[newSex]
	//更新消耗和道具
	err := Backpack.DeductPlayerChangeSexRes(tEntityPlayer, index, newSex, uint32(1))
	if err != nil {
		log.Error(err)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerChangeSexResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	tEntityPlayer.Sex = newSex
	//推送客户同步服装
	s.ChangePlayerDress(msgBody.EntityID, newDressId)
	tEntityPlayer.SyncEntity(1)

	msgResponse.Code = 0
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerChangeSexResponse, msgResponse, []uint32{msgBody.EntityID})
	return
}

// 修改签名 ，前端->游戏服->DB服
func (s *_Player) OnChangePlayerSignRequest(msgEV *network.MsgBodyEvent) {
	log.Info("-->OnChangePlayerSignRequest--------------begin-------")
	msgBody := &gmsg.ChangePlayerSignRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	msgResponse := &gmsg.ChangePlayerSignResponse{}

	if len(*msgBody.PlayerSign) > 20 {
		msgResponse.EntityID = msgBody.EntityID
		msgResponse.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_ChangeSignResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.PlayerSign = *msgBody.PlayerSign
	tEntityPlayer.SyncEntity(1)

	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = 0
	msgResponse.PlayerSign = msgBody.PlayerSign
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_ChangeSignResponse, msgResponse, []uint32{msgBody.EntityID})
	return
}

// 查询用户基本信息 前端->游戏->db服
func (s *_Player) OnQueryEntityPlayerByIDRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.QueryEntityPlayerByIDRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.QueryEntityPlayerByIDResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Player = make([]*gmsg.PlayerBase, 0)
	for _, val := range msgBody.QueryEntityID {
		if playerBase, ok := s.PlayerBaseList[val]; ok {
			msgResponse.Player = append(msgResponse.Player, playerBase)
		} else {
			ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Player_QueryEntityPlayerByIDRequest, msgBody, network.ServerType_DB)
			return
		}
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_QueryEntityPlayerByIDResponse, msgResponse, []uint32{msgResponse.EntityID})
	return
}

// 查询用户基本信息 DB服->游戏服－>前端
func (s *_Player) OnQueryEntityPlayerByIDResponse(msgEV *network.MsgBodyEvent) {
	msgResponse := &gmsg.QueryEntityPlayerByIDResponse{}
	if err := msgEV.Unmarshal(msgResponse); err != nil {
		return
	}

	if msgResponse.Code == 0 {
		s.syncPlayerBase(msgResponse.Player)
	}

	if msgResponse.EntityID > 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_QueryEntityPlayerByIDResponse, msgResponse, []uint32{msgResponse.EntityID})
	}
	return
}

// 查询用户个人资料
func (s *_Player) OnPlayerInfoRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PlayerInfoRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Player_InfoRequest, msgBody, network.ServerType_DB)
	return
}

// 查询用户个人资料
func (s *_Player) OnPlayerInfoResponse(msgEV *network.MsgBodyEvent) {
	msgResponse := &gmsg.PlayerInfoResponse{}
	if err := msgEV.Unmarshal(msgResponse); err != nil {
		return
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_InfoResponse, msgResponse, []uint32{msgResponse.EntityID})
	return
}

// 更新用户等级经验
func (s *_Player) UpdatePlayerLvExp(entityID uint32, incrExp uint32) {
	if entityID <= 0 || incrExp <= 0 {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	allExp := tEntityPlayer.NumExp + incrExp
	lv := Exp.Exp2Level(allExp)

	if lv > tEntityPlayer.PlayerLv {
		ConditionalMr.SyncConditionalPlayerLv(entityID, lv)
	}

	tEntityPlayer.PlayerLv = lv
	tEntityPlayer.NumExp = allExp
	tEntityPlayer.SyncEntity(0)

	msgBody := &gmsg.PlayerLvExpSync{}
	msgBody.PlayerLv = lv
	msgBody.NumExp = allExp

	log.Info("-->Logic--_Player--UpdatePlayerLvExp--resp--", msgBody)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerLvExpSync, msgBody, []uint32{entityID})
	return
}

// 更新用户Vip等级经验
func (s *_Player) UpdatePlayerVipLvExp(entityID uint32, incrExp uint32) {
	if entityID <= 0 || incrExp <= 0 {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	allExp := tEntityPlayer.VipExp + incrExp
	lv := VipExp.Exp2Level(allExp)
	if lv > tEntityPlayer.VipLv {
		ConditionalMr.SyncConditionalVipLv(entityID, lv)
	}

	tEntityPlayer.VipLv = lv
	tEntityPlayer.VipExp = allExp
	tEntityPlayer.SyncEntity(0)

	resp := &gmsg.PlayerVipLvExpSync{}
	resp.Lv = lv
	resp.Exp = allExp
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerVipLvExpSync, resp, []uint32{entityID})
	return
}

// 更新用户天梯等级经验
func (s *_Player) UpdatePlayerPeakRankLvExp(entityID uint32, incrExp int32) {
	if entityID <= 0 || incrExp <= 0 {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	allExp := int32(tEntityPlayer.PeakRankExp) + incrExp
	if allExp < 0 {
		allExp = 0
	}

	lv := PeakRankExp.Exp2Level(uint32(allExp))

	tEntityPlayer.PeakRankLv = lv
	tEntityPlayer.PeakRankExp = uint32(allExp)
	tEntityPlayer.SyncEntity(0)

	resp := &gmsg.PlayerPeakRankLvExpSync{}
	resp.Lv = lv
	resp.Exp = uint32(allExp)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerPeakRankLvExpSync, resp, []uint32{entityID})
}

// 同步用户属性道具
func (s *_Player) UpdatePlayerPropertyItem(entityID, tableID uint32, value int32, resParam entity.ResParam) {
	resp := &gmsg.PlayerPropertyItemSync{}
	resp.Item = new(gmsg.PropertyItem)
	res := s.UpPlayerPropertyItemFunc(entityID, tableID, value, resParam)
	resp.Item.TableID = tableID
	resp.Item.ItemValue = res.ItemValue

	// 以下三种数据不用再同步，否则变成覆盖关系了
	//todo 注意：UpdatePlayerLvExp，UpdatePlayerVipLvExp，UpdatePlayerPeakRankLvExp 已经同步，所以后续跳过处理，只记录产出和消耗
	//人气值不用同步
	if tableID == conf.LvExp || tableID == conf.VipLvExp || tableID == conf.PeakRankExp {
		return
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerPropertyItemSync, resp, []uint32{entityID})
}

// 同步用户多个属性道具
func (s *_Player) UpdatePlayerRepeatedPropertyItem(entityID uint32, items []entity.PropertyItem, resParam entity.ResParam) {
	resp := &gmsg.PlayerRepeatedPropertyItemSync{}
	resp.Items = make([]*gmsg.PropertyItem, 0)
	for _, val := range items {
		res := s.UpPlayerPropertyItemFunc(entityID, val.TableID, val.ItemValue, resParam)
		if res.ItemValue > 0 {
			resp.Items = append(resp.Items, res)
		}
	}

	if len(resp.Items) == 0 {
		return
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerRepeatedPropertyItemSync, resp, []uint32{entityID})
}

// 更新用户属性道具接口
func (s *_Player) UpPlayerPropertyItemFunc(entityID, tableID uint32, value int32, resParam entity.ResParam) *gmsg.PropertyItem {
	if entityID <= 0 {
		return nil
	}

	var updateFunc func(string, uint32, uint32, uint32, uint32, uint32, uint64, uint32, uint32, uint32)
	var incrType, afterModifyNum uint32

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	resp := new(gmsg.PropertyItem)
	resp.TableID = tableID
	if tableID == conf.Gold {
		tEntityPlayer.NumGold = tools.GetUint(int32(tEntityPlayer.NumGold) + value)
		resp.ItemValue = tEntityPlayer.NumGold
		afterModifyNum = tEntityPlayer.NumGold
	} else if tableID == conf.Diamond {
		tEntityPlayer.NumStone = tools.GetUint(int32(tEntityPlayer.NumStone) + value)
		resp.ItemValue = tEntityPlayer.NumStone
		afterModifyNum = tEntityPlayer.NumStone
	} else if tableID == conf.ClubGold {
		tEntityPlayer.ClubNumGold = tools.GetUint(int32(tEntityPlayer.ClubNumGold) + value)
		resp.ItemValue = tEntityPlayer.ClubNumGold
		afterModifyNum = tEntityPlayer.ClubNumGold
	} else if tableID == conf.Exchange {
		tEntityPlayer.ExchangeGold = tools.GetUint(int32(tEntityPlayer.ExchangeGold) + value)
		resp.ItemValue = tEntityPlayer.ExchangeGold
		afterModifyNum = tEntityPlayer.ExchangeGold
	} else if tableID == conf.ShopScore {
		tEntityPlayer.ShopScore = tools.GetUint(int32(tEntityPlayer.ShopScore) + value)
		resp.ItemValue = tEntityPlayer.ShopScore
		afterModifyNum = tEntityPlayer.ShopScore
	} else if tableID == conf.LvExp {
		s.UpdatePlayerLvExp(entityID, uint32(value))
		afterModifyNum = tEntityPlayer.NumExp
	} else if tableID == conf.VipLvExp {
		s.UpdatePlayerVipLvExp(entityID, uint32(value))
		afterModifyNum = tEntityPlayer.VipExp
	} else if tableID == conf.PeakRankExp {
		s.UpdatePlayerPeakRankLvExp(entityID, value)
		afterModifyNum = tEntityPlayer.PeakRankExp
	} else if tableID == conf.Popularity {
		tEntityPlayer.PopularityValue = tools.GetUint(int32(tEntityPlayer.PopularityValue) + value)
		resp.ItemValue = tEntityPlayer.PopularityValue
		afterModifyNum = tEntityPlayer.PopularityValue
	}

	tEntityPlayer.SyncEntity(1)

	if value > 0 {
		updateFunc, incrType = SendProductionResourceLogToDb, conf.RES_TYPE_INCR
		log.Info("-->entityID->", entityID, "-->value->", value, "-->tableID->", tableID, "--->resParam", resParam, "-->更新产出。")
	} else if value < 0 {
		updateFunc, incrType = SendConsumeResourceLogToDb, conf.RES_TYPE_DECR
		value = value * -1
		log.Info("-->entityID->", entityID, "-->value->", value, "-->tableID->", tableID, "--->resParam", resParam, "-->更新消耗。")
	}
	updateFunc(resParam.Uuid, entityID, conf.PropertyItem, 0, tableID, incrType, uint64(value), afterModifyNum, resParam.SysID, resParam.ActionID)
	return resp
}

func (s *_Player) UpdateAccAndPlayerStatus(entityID uint32, status uint32) {
	if entityID <= 0 || status <= 0 {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	//修改用户状态数据
	tEntityPlayer.SetPlayerState(status)
	tEntityPlayer.SyncEntity(1)

	tAcc := Entity.EmAcc.GetEntityByID(entityID)
	tEntityAcc := tAcc.(*entity.EntityAcc)
	//修改用户状态数据
	tEntityAcc.SetPlayerState(status)
	tEntityAcc.SyncEntity(1)

	switch status {
	case conf.USER_STATUS_PROHIBITION:
		//发禁言消息
		req := &gmsg.PlayerProhibitionRequest{
			EntityID: entityID,
		}
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerProhibitionRequest, req, []uint32{entityID})
	case conf.USER_STATUS_BAN_ACC, conf.USER_STATUS_BAN_IP:
		//踢下线
		s.KickOutPlayer(entityID)
	}

	return
}

// 将用户踢下线
func (s *_Player) KickOutPlayer(entityID uint32) {
	if entityID <= 0 {
		return
	}

	//先校验对战状态
	BattleC8Mgr.CheckEntityBattle(entityID)

	//先把用户踢下线
	req := &gmsg.PlayerKickOutRequest{
		EntityID: entityID,
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerKickOutRequest, req, []uint32{entityID})

	//移除用户内存数据
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity != nil {
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		tEntityPlayer.SyncEntity(1)

		Entity.EmPlayer.DelEntity(tEntity)
		tEntity = nil
	}

	tAcc := Entity.EmAcc.GetEntityByID(entityID)
	if tAcc != nil {
		tEntityAcc := tAcc.(*entity.EntityAcc)
		tEntityAcc.SyncEntity(1)

		Entity.EmAcc.DelEntity(tAcc)
		tAcc = nil
	}

	return
}

// 推送俱乐部id给前端
func (s *_Player) SyncClubToPlayer(entityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgBody := &gmsg.ClubToPlayerSync{}
	msgBody.ClubID = tEntityPlayer.ClubId
	msgBody.ClubName = tEntityPlayer.ClubName
	msgBody.ClubBadge = tEntityPlayer.ClubBadge
	msgBody.ClubRate = tEntityPlayer.ClubRate
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_ClubToPlayerSync, msgBody, []uint32{entityID})
	return
}

// 公用方法，获取用户在线时间或者离线时间
// todo 只有社交模块使用
func (s *_Player) GetGamePlayerOnline(EntityID uint32) uint32 {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	player, ok := s.PlayerBaseList[EntityID]
	if tEntity != nil {
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		if tEntityPlayer.BehaviorStatus >= conf.BEHAVIOR_STATUS_ROOM {
			return conf.PlayerIn
		}
		return conf.PlayerOnline
	}
	if ok {
		return uint32(carbon.Now().Timestamp()-tools.GetUnixFromStr(player.CurrentLoginTime)+int64(120)) / uint32(60)
	}
	return uint32(3600*7*24) / uint32(60)
}

// 公用方法，获取用户离线时间
func (s *_Player) GetGamePlayerCurrentLoginTime(EntityID uint32) string {
	player, ok := s.PlayerBaseList[EntityID]
	if ok {
		return player.CurrentLoginTime
	}
	return ""
}

// 获取角色自增id
func (s *_Player) GetMaxUuid(entityID uint32) uint32 {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return 0
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	return tEntityPlayer.GetMaxUuid()
}

// 获取角色是否在游戏中
func (s *_Player) GetBehaviorStatus(entityID uint32) uint32 {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return conf.PlayerOutline
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.BehaviorStatus >= conf.BEHAVIOR_STATUS_ROOM {
		return conf.PlayerIn
	}
	return conf.PlayerOnline
}

func (s *_Player) getPlayerNewSex(sex uint32) uint32 {
	if sex == conf.USER_MAN {
		return conf.USER_WOMEN
	}
	return conf.USER_MAN
}
