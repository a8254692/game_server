package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"errors"
	"reflect"
)

type _Email struct {
	EmailMax uint32 //邮件最大值
}

var Email _Email

func (s *_Email) Init() {
	s.EmailMax = 100

	//event.OnNet(gmsg.MsgTile_Hall_EmailRequest, reflect.ValueOf(s.OnEmailRequest))
	//event.OnNet(gmsg.MsgTile_Hall_EmailAddRequest, reflect.ValueOf(s.OnEmailAddRequest))
	event.OnNet(gmsg.MsgTile_Hall_EmailDelRequest, reflect.ValueOf(s.OnEmailDelRequest))
	//event.OnNet(gmsg.MsgTile_Hall_EmailUpdateRequest, reflect.ValueOf(s.OnEmailUpdateRequest))
	event.OnNet(gmsg.MsgTile_Hall_EmailReadRequest, reflect.ValueOf(s.OnEmailReadRequest))
	event.OnNet(gmsg.MsgTile_Hall_EmailGetRewardRequest, reflect.ValueOf(s.OnEmailGetRewardRequest))
}

// 获取全部邮件
func (s *_Email) getAllEmailByPlayer(tEntityPlayer *entity.EntityPlayer) []*gmsg.Email {
	if tEntityPlayer == nil {
		return nil
	}

	resp := make([]*gmsg.Email, 0)
	for _, email := range tEntityPlayer.EmailList {
		rList := make([]*gmsg.RewardInfo, 0)

		var isRewardEmail bool
		if len(email.RewardList) > 0 {
			isRewardEmail = true

			for _, rv := range email.RewardList {
				rInfo := &gmsg.RewardInfo{
					ItemTableId: rv.ItemTableId,
					Num:         rv.Num,
				}
				rList = append(rList, rInfo)
			}
		}

		e := &gmsg.Email{
			EmailID:       email.EmailID,
			State:         email.State,
			Date:          email.Date,
			StateReward:   email.StateReward,
			RewardList:    rList,
			Tittle:        email.Tittle,
			Content:       email.Content,
			IsRewardEmail: isRewardEmail,
		}

		resp = append(resp, e)
	}

	return resp
}

// 获取邮件数据请求
//func (s *_Email) OnEmailRequest(msgEV *network.MsgBodyEvent) {
//	msgBody := &gmsg.EmailRequest{}
//	err := msgEV.Unmarshal(msgBody)
//	if err != nil {
//		return
//	}
//	s.getAllEmail(msgBody.EntityID)
//	return
//}
//
//// 获取全部邮件
//func (s *_Email) getAllEmail(EntityID uint32) {
//	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
//	tEntityPlayer := tEntity.(*entity.EntityPlayer)
//	resp := &gmsg.EmailResponse{}
//	resp.EntityID = EntityID
//	resp.Code = 0
//
//	for _, email := range tEntityPlayer.EmailList {
//
//		rList := make([]*gmsg.RewardInfo, 0)
//		if len(email.RewardList) > 0 {
//			for _, rv := range email.RewardList {
//				rInfo := &gmsg.RewardInfo{
//					ItemTableId: rv.ItemTableId,
//					Num:         rv.Num,
//				}
//				rList = append(rList, rInfo)
//			}
//		}
//
//		e := new(gmsg.Email)
//		e.EmailID = email.EmailID
//		e.State = email.State
//		e.Date = email.Date
//		e.StateReward = email.StateReward
//		e.RewardList = rList
//		e.Tittle = proto.String(email.Tittle)
//		e.Content = proto.String(email.Content)
//
//		resp.EmailList = append(resp.EmailList, e)
//	}
//
//	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailResponse, resp, []uint32{EntityID})
//	return
//}

// 增加邮件请求
//func (s *_Email) OnEmailAddRequest(msgEV *network.MsgBodyEvent) {
//	msgBody := &gmsg.EmailAddRequest{}
//	err := msgEV.Unmarshal(msgBody)
//	if err != nil {
//		return
//	}
//	s.AddEmail(msgBody.EntityID, msgBody.Email)
//	return
//}

// 删除邮件请求
func (s *_Email) OnEmailDelRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EmailDelRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	s.delEmail(msgBody.EntityID, msgBody.EmailID)
	return
}

// 读邮件请求
func (s *_Email) OnEmailReadRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EmailReadRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	s.readEmail(msgBody.EntityID, msgBody.EmailID)
	return
}

// 领取邮件奖励请求
func (s *_Email) OnEmailGetRewardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EmailGetRewardRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}

	resp := &gmsg.EmailGetRewardResponse{}
	emailData, err := s.getEmailReward(msgBody.EntityID, msgBody.EmailID)
	if err != nil {
		resp.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailGetRewardResponse, resp, []uint32{msgBody.EntityID})
		return
	}

	resp.Code = 2
	commReward := make([]entity.RewardEntity, 0)
	var successList []*gmsg.RewardInfo

	resParam := GetResParam(consts.SYSTEM_ID_EMAIL, consts.Reward)
	if len(emailData.RewardList) > 0 {
		resp.Code = 1
		for _, rewardEntity := range emailData.RewardList {
			reward := new(entity.RewardEntity)
			reward.ItemTableId = rewardEntity.ItemTableId
			reward.Num = rewardEntity.Num
			reward.ExpireTimeId = rewardEntity.ExpireTimeId
			commReward = append(commReward, *reward)
			rewardInfo := new(gmsg.RewardInfo)
			stack.SimpleCopyProperties(rewardInfo, reward)
			successList = append(successList, rewardInfo)
		}

		if len(commReward) > 0 {
			RewardManager.AddReward(msgBody.EntityID, commReward, *resParam)
		}
	}

	resp.EmailID = emailData.EmailID
	resp.StateReward = true
	resp.RewardList = successList
	resp.Code = 0
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailGetRewardResponse, resp, []uint32{msgBody.EntityID})
	return
}

// 增加邮件
func (s *_Email) AddEmail(EntityID uint32, newEmail *gmsg.Email) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		//不在线直接发消息给db暂存邮件
		s.sendEmailToDb(EntityID, newEmail)
		return
	}

	info := new(entity.Email)
	info.EmailID = newEmail.EmailID
	info.State = newEmail.State
	info.Date = newEmail.Date
	info.StateReward = newEmail.StateReward
	info.RewardList = make([]entity.RewardEntity, 0)
	for _, rewardInfo := range newEmail.RewardList {
		emailRewardEntity := new(entity.RewardEntity)
		emailRewardEntity.ItemTableId = rewardInfo.ItemTableId
		emailRewardEntity.Num = rewardInfo.Num
		emailRewardEntity.ExpireTimeId = rewardInfo.ExpireTimeId
		info.RewardList = append(info.RewardList, *emailRewardEntity)
	}
	info.Tittle = newEmail.Tittle
	info.Content = newEmail.Content

	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.EmailList = append(tEntityPlayer.EmailList, *info)
	tEntityPlayer.SyncEntity(1)

	resp := &gmsg.NewEmailSync{}
	resp.EntityID = EntityID
	resp.Code = 0
	resp.Email = newEmail

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_NewEmailSync, resp, []uint32{EntityID})
	return
}

// 发送邮件至db服持久化数据
func (s *_Email) sendEmailToDb(EntityID uint32, newEmail *gmsg.Email) {
	if newEmail == nil {
		return
	}

	var rewardList []*gmsg.InRewardInfo
	if len(newEmail.RewardList) > 0 {
		for _, v := range newEmail.RewardList {
			info := &gmsg.InRewardInfo{
				ItemTableId:  v.ItemTableId,
				Num:          v.Num,
				ExpireTimeId: v.ExpireTimeId,
			}
			rewardList = append(rewardList, info)
		}
	}

	email := &gmsg.InEmail{
		EmailID:     newEmail.EmailID,
		Date:        newEmail.Date,
		State:       newEmail.State,
		StateReward: newEmail.StateReward,
		RewardList:  rewardList,
		Tittle:      newEmail.Tittle,
		Content:     newEmail.Content,
	}
	request := &gmsg.InEmailUpdateRequest{
		EntityID: EntityID,
		EmailID:  newEmail.EmailID,
		Email:    email,
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Email_SendEmailToDbRequest), request, network.ServerType_DB)
	return
}

// 删除邮件
func (s *_Email) delEmail(EntityID uint32, EmailID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	resp := &gmsg.EmailDelResponse{}
	resp.EntityID = EntityID

	for index, emailData := range tEntityPlayer.EmailList {
		if emailData.EmailID == EmailID {
			tEntityPlayer.EmailList = append(tEntityPlayer.EmailList[:index], tEntityPlayer.EmailList[(index+1):]...)

			resp.Code = 0
			email := new(gmsg.Email)
			email.EmailID = EmailID
			email.State = emailData.State
			email.StateReward = emailData.StateReward
			email.Date = emailData.Date
			email.RewardList = make([]*gmsg.RewardInfo, 0)
			email.Tittle = "testTitle"
			email.Content = "testContent"
			resp.Email = email

			tEntityPlayer.SyncEntity(1)

			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailDelResponse, resp, []uint32{EntityID})
			break
		}
	}

	if len(tEntityPlayer.EmailList) == 0 {
		resp.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailDelResponse, resp, []uint32{EntityID})
		return
	}

	return
}

// 读取邮件
func (s *_Email) readEmail(EntityID uint32, EmailID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	resp := &gmsg.EmailReadResponse{}

	if len(tEntityPlayer.EmailList) >= int(s.EmailMax) {
		resp.Code = 2
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailReadResponse, resp, []uint32{EntityID})
		return
	}

	emailData, index, err := s.getEmailByID(*tEntityPlayer, EmailID)
	if err != nil {
		resp.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailReadResponse, resp, []uint32{EntityID})
		return
	}

	entityEmail := emailData
	entityEmail.State = true
	tEntityPlayer.EmailList[index] = *entityEmail
	tEntityPlayer.SyncEntity(1)

	resp.EmailID = EmailID
	resp.Code = 0
	resp.State = true
	resp.EmailID = EmailID

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EmailReadResponse, resp, []uint32{EntityID})
	return
}

// 根据邮件ID获取邮件数据
func (s *_Email) getEmailByID(tEntityPlayer entity.EntityPlayer, EmailID uint32) (*entity.Email, int, error) {
	for index, emailData := range tEntityPlayer.EmailList {
		if emailData.EmailID == EmailID {
			return &tEntityPlayer.EmailList[index], index, nil
		}
	}
	return nil, -1, errors.New("-->logic--_Email--getEmailReward err")
}

// 领取邮件奖励
func (s *_Email) getEmailReward(EntityID uint32, EmailID uint32) (*entity.Email, error) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if len(tEntityPlayer.EmailList) >= int(s.EmailMax) {
		return nil, errors.New("-->logic--_Email--getEmailReward err")
	}

	emailData, index, err := s.getEmailByID(*tEntityPlayer, EmailID)

	if err != nil {
		return nil, errors.New("-->logic--_Email--getEmailReward err")

	}
	entityEmail := emailData
	entityEmail.StateReward = true
	tEntityPlayer.EmailList[index] = *entityEmail
	tEntityPlayer.SyncEntity(1)
	return emailData, nil
}
