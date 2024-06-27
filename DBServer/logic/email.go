package logic

import (
	"BilliardServer/Common/entity"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
)

type _Email struct {
}

var Email _Email

func (s *_Email) Init() {
	//注册逻辑业务事件
	//event.On("Msg_MultiNinjaPointWarEnemyTeam", reflect.ValueOf(TeamRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Email_SendEmailToDbRequest), reflect.ValueOf(s.OnInEmailUpdateRequest))
}

// OnInEmailUpdateRequest 同步email消息 游戏服->DB服
func (s *_Email) OnInEmailUpdateRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InEmailUpdateRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}
	log.Info("-->OnInEmailUpdateRequest-->begin-->", req)
	if req.EntityID <= 0 {
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(req.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	info := new(entity.Email)
	info.EmailID = tEntityPlayer.GetMaxUuid()
	info.State = req.Email.State
	info.Date = req.Email.Date
	info.StateReward = req.Email.StateReward
	info.RewardList = make([]entity.RewardEntity, 0)
	for _, rewardInfo := range req.Email.RewardList {
		emailRewardEntity := new(entity.RewardEntity)
		emailRewardEntity.ItemTableId = rewardInfo.ItemTableId
		emailRewardEntity.Num = rewardInfo.Num
		emailRewardEntity.ExpireTimeId = rewardInfo.ExpireTimeId
		info.RewardList = append(info.RewardList, *emailRewardEntity)
	}
	info.Tittle = req.Email.Tittle
	info.Content = req.Email.Content

	tEntityPlayer.EmailList = append(tEntityPlayer.EmailList, *info)
	tEntityPlayer.SaveEntity(DBConnect)
	log.Info("-->OnInEmailUpdateRequest-->end")
	return
}
