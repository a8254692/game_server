package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/tools"
	"reflect"
	"time"
)

var AdminOperation _AdminOperation

type _AdminOperation struct {
}

func (s *_AdminOperation) Init() {
	//注册逻辑业务事件
	event.OnNet(gmsg.MsgTile_Sys_GmMsgRequest, reflect.ValueOf(s.OnGetGmMsgRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Edit_User_Status_Request), reflect.ValueOf(s.OnAdminEditUserStatusRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Edit_User_Attr_Request), reflect.ValueOf(s.OnAdminEditUserAttrRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SendMarqueeMsgSync), reflect.ValueOf(s.OnAdminSendMarqueeMsg))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SendEmailRequest), reflect.ValueOf(s.OnAdminSendEmail))
}

func (s *_AdminOperation) OnGetGmMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetGmMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_AdminOperation--OnGetGmMsgRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}

	s.editUserAttr(req.EntityID, req.GType, req.Param)

	//targetEntityIDs := []uint32{req.EntityID}
	//resp := &gmsg.GetGmMsgResponse{}
	//resp.Code = resp_code.CODE_SUCCESS
	//
	//log.Info("-->logic--_AdminOperation--OnGetGmMsgRequest--Resp:", resp)
	//
	////广播同步消息
	//ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_LoginPerActivityListResponse, resp, targetEntityIDs)
	return
}

func (s *_AdminOperation) OnAdminEditUserStatusRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InEditUserStatusRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_AdminOperation--OnEditUserStatusRequestRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}

	log.Info("-->logic--_AdminOperation--OnEditUserStatusRequest--req--", req)

	if req.EntityID <= 0 || req.OType <= 0 {
		log.Waring("-->logic--_AdminOperation--OnEditUserStatusRequestRequest--req.OType <= 0")
		return
	}

	//TODO：增加登录校验用户状态逻辑

	switch req.OType {
	case consts.USER_STATUS_KICK_OUT:
		Player.KickOutPlayer(req.EntityID)
	case consts.USER_STATUS_PROHIBITION, consts.USER_STATUS_BAN_ACC, consts.USER_STATUS_BAN_IP:
		Player.UpdateAccAndPlayerStatus(req.EntityID, req.OType)
	}

	return
}

func (s *_AdminOperation) OnAdminEditUserAttrRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InEditUserAttrRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_AdminOperation--OnEditUserAttrRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}

	log.Info("-->logic--_AdminOperation--OnEditUserAttrRequest--req--", req)

	s.editUserAttr(req.EntityID, req.OType, req.Param)

	return
}

func (s *_AdminOperation) editUserAttr(entityID uint32, gType uint32, param uint32) {
	if entityID <= 0 || gType <= 0 || param <= 0 {
		return
	}

	resParam := GetResParam(consts.SYSTEM_ID_BAG, consts.GM)
	switch gType {
	case 1:
		//加经验
		Player.UpdatePlayerLvExp(entityID, param)
	case 2:
		//加金币
		Player.UpdatePlayerPropertyItem(entityID, consts.Gold, int32(param), *resParam)
	case 3:
		//加钻石
		Player.UpdatePlayerPropertyItem(entityID, consts.Diamond, int32(param), *resParam)
	case 4:
		//加物品
		commReward := make([]entity.RewardEntity, 0)
		commReward = append(commReward, entity.RewardEntity{
			ItemTableId:  param,
			Num:          1,
			ExpireTimeId: 0,
		})
		_ = RewardManager.AddReward(entityID, commReward, *resParam)
		//增加机器人
	case 5:
		reg := &gmsg.InRegRobotRequest{}
		reg.Param = param
		reg.High = param
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_RegRobot), reg, network.ServerType_DB)
		//机器人初始化
	case 6:
		req := &gmsg.InResetRobotRequest{}
		req.TimeStamp = uint32(time.Now().Unix())
		//RobotMr.ResetInitRobot()
		//ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_ResetRobot), req, network.ServerType_DB)
	case 7:
		Task.ResetTaskTest(entityID)
	case 80:
		//运行评级函数
		ClubManager.statisticsClubFunc()
	case 81:
		//注册俱乐部
		ClubManager.OnBatchReqClubRequest(param)
	case 82:
		ClubManager.OnAddAllClubData()
	case 90:
		KingRodeMr.resetKingRodeActivityList(entityID)
	case 91:
		KingRodeMr.addUpdateKingRodeActivityProgress(entityID)
	default:
		//更新成就和称号
		if gType > 100 && gType < 10000 {
			AchievementMr.AddAchievementFromConditionID(entityID, gType, param)
			CollectMr.AddCollectFromConditionID(entityID, gType, param)
		}
		//更新俱乐部
		if gType > 10000 && gType < 10000000 {
			//添加俱乐部参数
			ClubManager.OnAddClubData(gType, param)
			return
		}
		//加物品
		if gType > 10000000 {
			//加物品
			commReward := make([]entity.RewardEntity, 0)
			commReward = append(commReward, entity.RewardEntity{
				ItemTableId:  gType,
				Num:          param,
				ExpireTimeId: 0,
			})
			_ = RewardManager.AddReward(entityID, commReward, *resParam)
		}
	}

	return
}

func (s *_AdminOperation) OnAdminSendMarqueeMsg(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InMarqueeMsgSync{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_AdminOperation--OnSendMarqueeMsgSync--msgEV.Unmarshal(req)--err:", err)
		return
	}

	log.Info("-->logic--_AdminOperation--OnSendMarqueeMsgSync--req--", req)

	if req.Context == "" {
		log.Waring("-->logic--_AdminOperation--OnSendMarqueeMsgSync--req.Context == nil--", err)
		return
	}

	SendMarqueeMsgSync(req.MarqueeType, req.Context)

	return
}

func (s *_AdminOperation) OnAdminSendEmail(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InAddEmailRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_AdminOperation--OnAdminSendEmail--msgEV.Unmarshal(req)--err:", err)
		return
	}

	log.Info("-->logic--_AdminOperation--OnAdminSendEmail--req--", req)

	if req.EntityID <= 0 {
		log.Waring("-->logic--_AdminOperation--OnSendMarqueeMsgSync--req.Context == nil--", err)
		return
	}

	rList := make([]*gmsg.RewardInfo, 0)
	if len(req.Email.RewardList) > 0 {
		for _, v := range req.Email.RewardList {
			rList = append(rList, &gmsg.RewardInfo{
				ItemTableId:  v.ItemTableId,
				Num:          v.Num,
				ExpireTimeId: v.ExpireTimeId,
			})
		}
	}

	email := &gmsg.Email{
		EmailID:    Player.GetMaxUuid(req.EntityID),
		Date:       tools.GetTimeByTimeStamp(time.Now().Unix()),
		Tittle:     req.Email.Tittle,
		Content:    req.Email.Content,
		RewardList: rList,
	}
	Email.AddEmail(req.EntityID, email)

	return
}
