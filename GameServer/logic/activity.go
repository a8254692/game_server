package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/GameServer/initialize/vars"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"encoding/json"
	"github.com/google/uuid"
	"math/rand"
	"reflect"
	"time"
)

type _Activity struct {
	List []vars.ActivityData
}

var Activity _Activity

func (s *_Activity) Init() {
	s.List = make([]vars.ActivityData, 0)

	event.OnNet(gmsg.MsgTile_Reward_ReceiveBattleActivityRequest, reflect.ValueOf(s.OnReceiveBattleActivityRequest))
	event.OnNet(gmsg.MsgTile_Reward_ReceivePayActivityRequest, reflect.ValueOf(s.OnReceivePayActivityRequest))
	event.OnNet(gmsg.MsgTile_Reward_ReceiveTurntableActivityRequest, reflect.ValueOf(s.OnReceiveTurntableActivityRequest))
	event.OnNet(gmsg.MsgTile_Reward_ReceivePayLotteryActivityRequest, reflect.ValueOf(s.OnReceivePayLotteryActivityRequest))
	event.OnNet(gmsg.MsgTile_Reward_ReceivePayLotteryDrawNumRequest, reflect.ValueOf(s.OnReceivePayLotteryDrawNumRequest))
	event.OnNet(gmsg.MsgTile_Reward_GetActivityListRequest, reflect.ValueOf(s.GetActivityListRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncActivityListToGameResponse), reflect.ValueOf(s.SyncActivityListFromDb))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_ActivityOtherToGameSync), reflect.ValueOf(s.AdminChangeActivityList))

	time.AfterFunc(time.Millisecond*1000, s.SyncActivityListToDb)
}

func (s *_Activity) BattleSettleNotice(entityId uint32, isWinner bool) {
	s.UpdateTurntableActivityProgress(entityId)
	if isWinner {
		s.UpdateKingRodeActivityProgress(entityId)
	}
	return
}

func (s *_Activity) deleteListElem(aId string) {
	if len(s.List) <= 0 {
		return
	}

	for i := 0; i < len(s.List); i++ {
		if s.List[i].ActivityId == aId {
			s.List = append(s.List[:i], s.List[i+1:]...)
			i--
		}
	}

	return
}

func (s *_Activity) AdminChangeActivityList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InAdminActivityListSync{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--AdminChangeActivityList--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_Activity--AdminChangeActivityList--req", req)

	toReq := &gmsg.InActivityListToDbRequest{
		IsUpdate: true,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncActivityListToGameRequest), toReq, network.ServerType_DB)
	return
}

func (s *_Activity) SyncActivityListFromDb(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InActivityList{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--SyncActivityListFromDb--msgEV.Unmarshal(req) err:", err)
		return
	}

	if len(req.List) <= 0 && len(s.List) <= 0 {
		return
	}

	respList := make([]vars.ActivityData, 0)
	if len(req.List) <= 0 {
		if len(s.List) > 0 {
			s.List = respList
		}
		return
	}

	for _, v := range req.List {
		if v.AType <= 0 {
			continue
		}

		var configTurntable vars.ActivityConfigTurntable
		var configPayLottery vars.ActivityConfigPayLottery
		var configPay vars.ActivityConfigPay
		var configBattle vars.ActivityConfigBattle

		if v.AType == uint32(consts.ActivityTplType_Pay) {
			err = json.Unmarshal([]byte(v.Config), &configPay)
			if err != nil {
				continue
			}
		} else if v.AType == uint32(consts.ActivityTplType_Battle) {
			err = json.Unmarshal([]byte(v.Config), &configBattle)
			if err != nil {
				continue
			}
		} else if v.AType == uint32(consts.ActivityTplType_Turntable) {
			err = json.Unmarshal([]byte(v.Config), &configTurntable)
			if err != nil {
				continue
			}
		} else if v.AType == uint32(consts.ActivityTplType_PayLottery) {
			err = json.Unmarshal([]byte(v.Config), &configPayLottery)
			if err != nil {
				continue
			}
		}

		info := vars.ActivityData{
			ActivityId:       v.ActivityId,
			TimeType:         v.TimeType,
			StartTime:        v.StartTime,
			EndTime:          v.EndTime,
			AType:            consts.ActivityType(v.AType),
			SubType:          v.SubType,
			ActivityName:     v.ActivityName,
			PlatformLimit:    v.PlatformLimit,
			VipLimit:         v.VipLimit,
			ConfigTurntable:  configTurntable,
			ConfigPayLottery: configPayLottery,
			ConfigPay:        configPay,
			ConfigBattle:     configBattle,
		}

		respList = append(respList, info)
	}

	if len(respList) > 0 {
		s.List = respList
	}

	if req.IsUpdate {
		//开始初始化桌面信息
		syncResp := &gmsg.ActivityListUpdateNoticeSync{
			IsUpdate: true,
		}
		//广播初始化消息
		ConnectManager.SendMsgPbToGateBroadCastAll(gmsg.MsgTile_Reward_ActivityListUpdateNoticeSync, syncResp)
	}

	return
}

func (s *_Activity) SyncActivityListToDb() {
	//开始初始化桌面信息
	req := &gmsg.InActivityListToDbRequest{}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncActivityListToGameRequest), req, network.ServerType_DB)
	return
}

// 获取活动列表
func (s *_Activity) GetActivityListRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetActivityListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--GetActivityListRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityId <= 0 {
		return
	}

	//开始初始化桌面信息
	resp := &gmsg.GetActivityListResponse{
		List: s.GetLoginActivityListRequest(req.EntityId),
	}

	log.Info("-->logic--_Activity--GetActivityListRequest--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityId}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_GetActivityListResponse, resp, targetEntityIDs)
	return
}

func (s *_Activity) OnReceiveBattleActivityRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ReceiveBattleActivityRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceiveBattleActivityRequest--msgEV.Unmarshal(req) err:", err)
		return
	}
	entityID := req.EntityId
	activityId := req.ActivityId
	configSerial := req.ConfigSerial

	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 || activityId == "" || configSerial <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceiveBattleActivityRequest--GetEntityPlayerById--err--", err)
		return
	}

	resParam := GetResParam(consts.SYSTEM_ID_ACTIVITY, consts.Reward)
	now := time.Now().Unix()
	for _, v := range s.List {
		if v.ActivityId != activityId {
			continue
		}

		if v.AType != consts.ActivityTplType_Battle {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}

		if len(v.ConfigBattle.ConditionAndRewardList) <= 0 {
			continue
		}

		for _, cv := range v.ConfigBattle.ConditionAndRewardList {
			if cv.No != configSerial {
				continue
			}

			for pk, pv := range tEntityPlayer.ProgressActivityList {
				if pv.ActivityId != activityId {
					continue
				}
				if pv.ConfigSerial != configSerial && pv.TargetProgress == cv.ValueList {
					continue
				}
				if pv.CompleteProgress < pv.TargetProgress {
					continue
				}

				tEntityPlayer.ProgressActivityList[pk].StateReward = consts.RECEIVE_STATUS_YES
				RewardManager.AddReward(entityID, cv.RewardList, *resParam)
			}
		}
	}

	tEntityPlayer.SyncEntity(1)

	//开始初始化桌面信息
	resp := &gmsg.ReceiveBattleActivityResponse{
		Code:         0,
		ActivityId:   activityId,
		ConfigSerial: configSerial,
	}

	log.Info("-->logic--_Activity--OnReceiveBattleActivityRequest--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceiveBattleActivityResponse, resp, targetEntityIDs)
	return
}

func (s *_Activity) OnReceivePayActivityRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ReceivePayActivityRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceivePayActivityRequest--msgEV.Unmarshal(req) err:", err)
		return
	}
	entityID := req.EntityId
	activityId := req.ActivityId
	configSerial := req.ConfigSerial

	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 || activityId == "" || configSerial <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceivePayActivityRequest--GetEntityPlayerById--err--", err)
		return
	}

	resParam := GetResParam(consts.SYSTEM_ID_ACTIVITY, consts.Reward)
	now := time.Now().Unix()
	for _, v := range s.List {
		if v.ActivityId != activityId {
			continue
		}

		if v.AType != consts.ActivityTplType_Pay {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}

		if len(v.ConfigPay.ConditionAndRewardList) <= 0 {
			continue
		}

		for _, cv := range v.ConfigPay.ConditionAndRewardList {
			if cv.No != configSerial {
				continue
			}

			for pk, pv := range tEntityPlayer.ProgressActivityList {
				if pv.ActivityId != activityId {
					continue
				}
				if pv.ConfigSerial != configSerial && pv.TargetProgress == cv.ValueList {
					continue
				}
				if pv.CompleteProgress < pv.TargetProgress {
					continue
				}

				tEntityPlayer.ProgressActivityList[pk].StateReward = consts.RECEIVE_STATUS_YES
				RewardManager.AddReward(entityID, cv.RewardList, *resParam)
			}
		}
	}

	tEntityPlayer.SyncEntity(1)

	//开始初始化桌面信息
	resp := &gmsg.ReceivePayActivityResponse{
		Code:         0,
		ActivityId:   activityId,
		ConfigSerial: configSerial,
	}

	log.Info("-->logic--_Activity--OnReceivePayActivityRequest--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceivePayActivityResponse, resp, targetEntityIDs)
	return
}

func (s *_Activity) OnReceiveTurntableActivityRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ReceiveTurntableActivityRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceiveTurntableActivityRequest--msgEV.Unmarshal(req) err:", err)
		return
	}
	entityID := req.EntityId
	activityId := req.ActivityId

	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 || activityId == "" {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceiveTurntableActivityRequest--GetEntityPlayerById--err--", err)
		return
	}

	var showTotalUseNum uint32

	var lotteryNum uint32
	resParam := GetResParam(consts.SYSTEM_ID_ACTIVITY, consts.Reward)
	t := time.Now()
	now := t.Unix()
	addTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	for _, v := range s.List {
		if v.ActivityId != activityId {
			continue
		}

		if v.AType != consts.ActivityTplType_Turntable {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}

		if len(v.ConfigTurntable.TurntableRewardList) <= 0 {
			continue
		}

		//计算可转转盘次数
		totalNum := v.ConfigTurntable.FreeDrawNum
		if v.ConfigTurntable.DrawNumConfig <= 0 {
			totalNum = v.ConfigTurntable.TotalDrawNum
		} else {
			if tEntityPlayer.DayBattleNum.Timestamp == addTime {
				if v.ConfigTurntable.DrawNumConfig > 0 {
					if tEntityPlayer.DayBattleNum.Num > 0 {
						playNum := tEntityPlayer.DayBattleNum.Num / v.ConfigTurntable.DrawNumConfig
						totalNum += playNum
					}
				}
			}

			if totalNum > v.ConfigTurntable.TotalDrawNum {
				totalNum = v.ConfigTurntable.TotalDrawNum
			}
		}

		//计算已转转盘次数
		var useNum uint32
		if len(tEntityPlayer.DayReceiveStatusNumList) > 0 {
			for _, drv := range tEntityPlayer.DayReceiveStatusNumList {
				if drv.ActivityId == activityId && drv.Timestamp == addTime {
					useNum = drv.Num
				}
			}
		}

		showTotalUseNum = useNum

		if useNum >= totalNum {
			continue
		}

		//扣除次数
		var isIn bool
		var isDeduct bool
		if len(tEntityPlayer.DayReceiveStatusNumList) > 0 {
			for drrk, drrv := range tEntityPlayer.DayReceiveStatusNumList {
				if drrv.ActivityId == activityId {

					log.Waring("-->logic--_Activity--OnReceiveTurntableActivityRequest--test9--", activityId)

					if drrv.Timestamp == addTime {
						tEntityPlayer.DayReceiveStatusNumList[drrk].Num += 1
					} else {
						tEntityPlayer.DayReceiveStatusNumList[drrk] = entity.DayReceiveStatusNum{
							ActivityId: activityId,
							Num:        1,
							Timestamp:  addTime,
						}
					}

					isIn = true
					isDeduct = true
				}
			}
		}

		if !isIn {
			tEntityPlayer.DayReceiveStatusNumList = append(tEntityPlayer.DayReceiveStatusNumList, entity.DayReceiveStatusNum{
				ActivityId: activityId,
				Num:        1,
				Timestamp:  addTime,
			})
			isDeduct = true
		}

		if isDeduct {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			randNum := r.Intn(10000)

			var start, end, rewardKey int
			for tk, tv := range v.ConfigTurntable.TurntableRewardList {
				end += int(tv.Probability) * 100
				if start <= randNum && end > randNum {
					rewardKey = tk

					lotteryNum = tv.No
				}
				start = end
			}

			rewardInfo := v.ConfigTurntable.TurntableRewardList[rewardKey].Reward
			rewardList := make([]entity.RewardEntity, 0)
			rewardList = append(rewardList, rewardInfo)

			RewardManager.AddReward(entityID, rewardList, *resParam)
		}

		showTotalUseNum = useNum + 1
	}

	tEntityPlayer.SyncEntity(1)

	var code uint32
	if lotteryNum <= 0 {
		code = 1
	}

	//开始初始化桌面信息
	resp := &gmsg.ReceiveTurntableActivityResponse{
		Code:            code,
		ActivityId:      activityId,
		ConfigSerial:    lotteryNum,
		UseTotalDrawNum: showTotalUseNum,
	}

	log.Info("-->logic--_Activity--OnReceiveTurntableActivityRequest--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceiveTurntableActivityResponse, resp, targetEntityIDs)
	return
}

func (s *_Activity) OnReceivePayLotteryActivityRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ReceivePayLotteryActivityRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceivePayLotteryActivityRequest--msgEV.Unmarshal(req) err:", err)
		return
	}
	entityID := req.EntityId
	activityId := req.ActivityId

	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 || activityId == "" {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceivePayLotteryActivityRequest--GetEntityPlayerById--err--", err)
		return
	}

	var lotteryNum uint32
	resParam := GetResParam(consts.SYSTEM_ID_ACTIVITY, consts.Reward)
	t := time.Now()
	now := t.Unix()
	addTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	for _, v := range s.List {
		if v.ActivityId != activityId {
			continue
		}

		if v.AType != consts.ActivityTplType_PayLottery {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}

		if len(v.ConfigPayLottery.PayLotteryRewardList) <= 0 {
			continue
		}

		var canUseFree bool
		if v.ConfigPayLottery.DayFreeDrawNum > 0 {
			var isInFree bool
			for pk, pv := range tEntityPlayer.ReceivePayLotteryList {
				if pv.ActivityId == activityId {
					isInFree = true

					if pv.FreeCreateTime != addTime {
						tEntityPlayer.ReceivePayLotteryList[pk].FreeNum += 1
						tEntityPlayer.ReceivePayLotteryList[pk].FreeCreateTime = addTime

						canUseFree = true
					} else {
						if pv.FreeNum < v.ConfigPayLottery.DayFreeDrawNum {
							tEntityPlayer.ReceivePayLotteryList[pk].FreeNum += 1
							tEntityPlayer.ReceivePayLotteryList[pk].FreeCreateTime = addTime

							canUseFree = true
						}
					}
				}
			}

			if !isInFree {
				tEntityPlayer.ReceivePayLotteryList = append(tEntityPlayer.ReceivePayLotteryList, entity.PayLotteryStatus{
					ActivityId:         activityId,
					FreeNum:            1,
					FreeCreateTime:     addTime,
					TotalDrawNum:       0,
					DrawEndTime:        0,
					LuckNum:            0,
					LastGetLuckDrawNum: 0,
					LuckEndTime:        0,
					DrawNumStatus:      make(map[uint32]uint32),
				})

				canUseFree = true
			}
		}

		//没有免费次数使用则校验存在抽奖物品
		if !canUseFree {
			//TODO: 校验是否存在物品
			if v.ConfigPayLottery.ConsumeItemID <= 0 || v.ConfigPayLottery.ConsumeItemNum <= 0 {
				break
			}

		}

		//计算出有效期
		effectiveDrawTime := int64(-1)
		switch v.ConfigPayLottery.DrawResetType {
		case consts.ResetType_Day:
			effectiveDrawTime = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()).Unix()
		case consts.ResetType_Week:
			lastDay := tools.GetFirstDateOfWeek(t).AddDate(0, 0, 6)
			effectiveDrawTime = time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 23, 59, 59, 0, lastDay.Location()).Unix()
		}

		var isDrawIn bool
		for p1k, p1v := range tEntityPlayer.ReceivePayLotteryList {
			if p1v.ActivityId == activityId {
				isDrawIn = true

				if effectiveDrawTime == -1 {
					tEntityPlayer.ReceivePayLotteryList[p1k].TotalDrawNum += 1
				} else {
					if tEntityPlayer.ReceivePayLotteryList[p1k].DrawEndTime > effectiveDrawTime {
						tEntityPlayer.ReceivePayLotteryList[p1k].TotalDrawNum = 1
					} else {
						tEntityPlayer.ReceivePayLotteryList[p1k].TotalDrawNum += 1
					}
				}
				tEntityPlayer.ReceivePayLotteryList[p1k].DrawEndTime = effectiveDrawTime
			}
		}
		if !isDrawIn {
			tEntityPlayer.ReceivePayLotteryList = append(tEntityPlayer.ReceivePayLotteryList, entity.PayLotteryStatus{
				ActivityId:         activityId,
				FreeNum:            0,
				FreeCreateTime:     0,
				TotalDrawNum:       1,
				DrawEndTime:        effectiveDrawTime,
				LuckNum:            0,
				LastGetLuckDrawNum: 0,
				LuckEndTime:        0,
				DrawNumStatus:      make(map[uint32]uint32),
			})
		}

		//开始增加相关计数
		if v.ConfigPayLottery.IsOpenLucky {
			//计算出有效期
			effectiveLuckyTime := int64(-1)
			switch v.ConfigPayLottery.LuckyResetType {
			case consts.ResetType_Day:
				effectiveLuckyTime = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()).Unix()
			case consts.ResetType_Week:
				lastLuckyDay := tools.GetFirstDateOfWeek(t).AddDate(0, 0, 6)
				effectiveLuckyTime = time.Date(lastLuckyDay.Year(), lastLuckyDay.Month(), lastLuckyDay.Day(), 23, 59, 59, 0, lastLuckyDay.Location()).Unix()
			}

			var isLuckIn bool
			for p2k, p2v := range tEntityPlayer.ReceivePayLotteryList {
				if p2v.ActivityId == activityId {
					isLuckIn = true

					//计算需要增加幸运值
					var addLuckyNum uint32
					if p2v.TotalDrawNum > p2v.LastGetLuckDrawNum && v.ConfigPayLottery.TakeLuckyNum > 0 {
						mul := (p2v.TotalDrawNum - p2v.LastGetLuckDrawNum) / v.ConfigPayLottery.TakeLuckyNum
						if mul > 0 {
							addLuckyNum = v.ConfigPayLottery.LuckyNum * mul
						}
					}

					if addLuckyNum > 0 {
						if effectiveLuckyTime == -1 {
							tEntityPlayer.ReceivePayLotteryList[p2k].LuckNum += addLuckyNum
						} else {
							if tEntityPlayer.ReceivePayLotteryList[p2k].LuckEndTime > effectiveLuckyTime {
								tEntityPlayer.ReceivePayLotteryList[p2k].LuckNum = addLuckyNum
							} else {
								tEntityPlayer.ReceivePayLotteryList[p2k].LuckNum += addLuckyNum
							}
							tEntityPlayer.ReceivePayLotteryList[p2k].LuckEndTime = effectiveLuckyTime
						}
					}
				}
			}

			if !isLuckIn {
				tEntityPlayer.ReceivePayLotteryList = append(tEntityPlayer.ReceivePayLotteryList, entity.PayLotteryStatus{
					ActivityId:         activityId,
					FreeNum:            0,
					FreeCreateTime:     0,
					TotalDrawNum:       0,
					DrawEndTime:        0,
					LuckNum:            0,
					LastGetLuckDrawNum: 0,
					LuckEndTime:        0,
					DrawNumStatus:      make(map[uint32]uint32),
				})
			}
		}

		//开始计算是否保底
		var isGuarantee bool
		for p3k, p3v := range tEntityPlayer.ReceivePayLotteryList {
			if p3v.ActivityId == activityId {
				if p3v.LuckNum >= v.ConfigPayLottery.MaxLuckyNum {
					tEntityPlayer.ReceivePayLotteryList[p3k].LuckNum = 0

					isGuarantee = true
				}
			}
		}

		tEntityPlayer.SyncEntity(1)

		//开始随机物品
		var rewardKey int
		if isGuarantee {
			//保底直接发放保底物品
			for tk, tv := range v.ConfigPayLottery.PayLotteryRewardList {
				if tv.IsGuarantee {
					rewardKey = tk
				}
			}
		} else {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			randNum := r.Intn(10000)

			var start, end int
			for tk, tv := range v.ConfigPayLottery.PayLotteryRewardList {
				end += int(tv.Probability) * 100
				if start <= randNum && end > randNum {
					rewardKey = tk

					lotteryNum = tv.No
				}
				start = end
			}
		}

		rewardInfo := v.ConfigPayLottery.PayLotteryRewardList[rewardKey].Reward
		rewardList := make([]entity.RewardEntity, 0)
		rewardList = append(rewardList, rewardInfo)

		RewardManager.AddReward(entityID, rewardList, *resParam)
	}

	//开始初始化桌面信息
	resp := &gmsg.ReceivePayLotteryActivityResponse{
		Code:         0,
		ActivityId:   activityId,
		ConfigSerial: lotteryNum,
	}

	log.Info("-->logic--_Activity--OnReceivePayLotteryActivityRequest--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceivePayLotteryActivityResponse, resp, targetEntityIDs)
	return
}

func (s *_Activity) OnReceivePayLotteryDrawNumRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ReceivePayLotteryDrawNumRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceivePayLotteryDrawNumRequest--msgEV.Unmarshal(req) err:", err)
		return
	}
	entityID := req.EntityId
	activityId := req.ActivityId
	configSerial := req.ConfigSerial

	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 || activityId == "" || configSerial <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--OnReceivePayLotteryDrawNumRequest--GetEntityPlayerById--err--", err)
		return
	}

	resParam := GetResParam(consts.SYSTEM_ID_ACTIVITY, consts.Reward)
	t := time.Now()
	now := t.Unix()
	for _, v := range s.List {
		if v.ActivityId != activityId {
			continue
		}

		if v.AType != consts.ActivityTplType_PayLottery {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}

		if !v.ConfigPayLottery.IsOpenDrawNum {
			continue
		}

		if len(v.ConfigPayLottery.DrawNumRewardList) <= 0 {
			continue
		}

		for _, cv := range v.ConfigPayLottery.DrawNumRewardList {
			if cv.No != configSerial {
				continue
			}

			for p1k, p1v := range tEntityPlayer.ReceivePayLotteryList {
				if p1v.ActivityId != activityId {
					continue
				}

				if p1v.DrawNumStatus[configSerial] == consts.RECEIVE_STATUS_YES {
					continue
				}

				if p1v.TotalDrawNum < cv.ValueList {
					continue
				}

				tEntityPlayer.ReceivePayLotteryList[p1k].DrawNumStatus[configSerial] = consts.RECEIVE_STATUS_YES
				RewardManager.AddReward(entityID, cv.RewardList, *resParam)
			}
		}
	}

	tEntityPlayer.SyncEntity(1)

	//开始初始化桌面信息
	resp := &gmsg.ReceivePayLotteryDrawNumResponse{
		Code:         0,
		ActivityId:   activityId,
		ConfigSerial: configSerial,
	}

	log.Info("-->logic--_Activity--OnReceivePayLotteryDrawNumRequest--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceivePayLotteryDrawNumResponse, resp, targetEntityIDs)
	return
}

func (s *_Activity) GetLoginActivityListRequest(entityId uint32) []*gmsg.LoginActivityInfo {
	resp := make([]*gmsg.LoginActivityInfo, 0)

	if len(s.List) <= 0 {
		return resp
	}

	tEntityPlayer, err := GetEntityPlayerById(entityId)
	if err != nil {
		log.Waring("-->logic--_Activity--GetLoginActivityListRequest--GetEntityPlayerById--err--", err)
		return resp
	}

	t := time.Now()
	addTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	for _, v := range s.List {
		if t.Unix() > v.EndTime {
			continue
		}
		var configPay gmsg.ActivityConfigPay
		var configBattle gmsg.ActivityConfigBattle
		var configTurntable gmsg.ActivityConfigTurntable
		var configPayLottery gmsg.ActivityConfigPayLottery
		var KingRodeConfig gmsg.KingRodeProgressList

		switch v.AType {
		case consts.ActivityTplType_Pay:
			rList := make([]*gmsg.ConditionAndReward, 0)
			for _, pv := range v.ConfigPay.ConditionAndRewardList {
				rewardList := make([]*gmsg.RewardInfo, 0)
				for _, rv := range pv.RewardList {
					rewardList = append(rewardList, &gmsg.RewardInfo{
						ItemTableId:  rv.ItemTableId,
						Num:          rv.Num,
						ExpireTimeId: rv.ExpireTimeId,
					})
				}

				var progressNum uint32
				var status uint32
				if len(tEntityPlayer.ProgressActivityList) > 0 {
					for _, pav := range tEntityPlayer.ProgressActivityList {
						if v.ActivityId != pav.ActivityId && pav.ConfigSerial == pv.No {
							progressNum = pav.CompleteProgress
							status = pav.StateReward
							break
						}
					}
				}

				rList = append(rList, &gmsg.ConditionAndReward{
					No:          pv.No,
					ValueList:   pv.ValueList,
					RewardList:  rewardList,
					TabName:     pv.TabName,
					ProgressNum: progressNum,
					Status:      status,
				})
			}

			configPay = gmsg.ActivityConfigPay{
				PayType:                v.ConfigPay.PayType,
				ConditionAndRewardList: rList,
			}
		case consts.ActivityTplType_Battle:
			rList := make([]*gmsg.ConditionAndReward, 0)

			for _, pv := range v.ConfigBattle.ConditionAndRewardList {
				rewardList := make([]*gmsg.RewardInfo, 0)
				for _, rv := range pv.RewardList {
					rewardList = append(rewardList, &gmsg.RewardInfo{
						ItemTableId:  rv.ItemTableId,
						Num:          rv.Num,
						ExpireTimeId: rv.ExpireTimeId,
					})
				}

				var progressNum uint32
				var status uint32
				if len(tEntityPlayer.ProgressActivityList) > 0 {
					for _, pav := range tEntityPlayer.ProgressActivityList {
						if v.ActivityId != pav.ActivityId && pav.ConfigSerial == pv.No {
							progressNum = pav.CompleteProgress
							status = pav.StateReward
							break
						}
					}
				}

				rList = append(rList, &gmsg.ConditionAndReward{
					No:          pv.No,
					ValueList:   pv.ValueList,
					RewardList:  rewardList,
					TabName:     pv.TabName,
					ProgressNum: progressNum,
					Status:      status,
				})
			}

			configBattle = gmsg.ActivityConfigBattle{
				BattleType:             v.ConfigBattle.BattleType,
				OutcomeType:            v.ConfigBattle.OutcomeType,
				ConditionAndRewardList: rList,
			}
		case consts.ActivityTplType_Turntable:
			tList := make([]*gmsg.TurntableReward, 0)
			for _, tv := range v.ConfigTurntable.TurntableRewardList {
				rewardInfo := gmsg.RewardInfo{
					ItemTableId:  tv.Reward.ItemTableId,
					Num:          tv.Reward.Num,
					ExpireTimeId: tv.Reward.ExpireTimeId,
				}

				tList = append(tList, &gmsg.TurntableReward{
					No:          tv.No,
					Probability: tv.Probability,
					Reward:      &rewardInfo,
				})
			}

			//计算可转转盘次数
			totalNum := v.ConfigTurntable.FreeDrawNum
			if v.ConfigTurntable.DrawNumConfig <= 0 {
				totalNum = v.ConfigTurntable.TotalDrawNum
			} else {
				if tEntityPlayer.DayBattleNum.Timestamp == addTime {
					if v.ConfigTurntable.DrawNumConfig > 0 {
						if tEntityPlayer.DayBattleNum.Num > 0 {
							playNum := tEntityPlayer.DayBattleNum.Num / v.ConfigTurntable.DrawNumConfig
							totalNum += playNum
						}
					}
				}

				if totalNum > v.ConfigTurntable.TotalDrawNum {
					totalNum = v.ConfigTurntable.TotalDrawNum
				}
			}

			//计算已转转盘次数
			var useNum uint32
			if len(tEntityPlayer.DayReceiveStatusNumList) > 0 {
				for _, drv := range tEntityPlayer.DayReceiveStatusNumList {
					if drv.ActivityId == v.ActivityId && drv.Timestamp == addTime {
						useNum = drv.Num
					}
				}
			}

			configTurntable = gmsg.ActivityConfigTurntable{
				FreeDrawNum:         v.ConfigTurntable.FreeDrawNum,
				TotalDrawNum:        totalNum,
				DrawNumConfig:       v.ConfigTurntable.DrawNumConfig,
				UseTotalDrawNum:     useNum,
				TurntableRewardList: tList,
			}
		case consts.ActivityTplType_PayLottery:
			lrList := make([]*gmsg.PayLotteryReward, 0)
			drawList := make([]*gmsg.ConditionAndReward, 0)
			exList := make([]*gmsg.ConditionAndReward, 0)

			for _, lrv := range v.ConfigPayLottery.PayLotteryRewardList {
				rewardInfo := gmsg.RewardInfo{
					ItemTableId:  lrv.Reward.ItemTableId,
					Num:          lrv.Reward.Num,
					ExpireTimeId: lrv.Reward.ExpireTimeId,
				}
				lrList = append(lrList, &gmsg.PayLotteryReward{
					No:          lrv.No,
					Probability: lrv.Probability,
					Reward:      &rewardInfo,
					IsGuarantee: lrv.IsGuarantee,
				})
			}

			if v.ConfigPayLottery.IsOpenDrawNum {
				for _, dv := range v.ConfigPayLottery.DrawNumRewardList {
					rewardList := make([]*gmsg.RewardInfo, 0)
					for _, rv := range dv.RewardList {
						rewardList = append(rewardList, &gmsg.RewardInfo{
							ItemTableId:  rv.ItemTableId,
							Num:          rv.Num,
							ExpireTimeId: rv.ExpireTimeId,
						})
					}

					var progressNum uint32
					status := consts.RECEIVE_STATUS_NO
					for _, p1v := range tEntityPlayer.ReceivePayLotteryList {
						if p1v.ActivityId == v.ActivityId {

							progressNum = p1v.TotalDrawNum
							if p1v.TotalDrawNum > dv.ValueList {
								progressNum = dv.ValueList
							}

							if p1v.DrawNumStatus[dv.ValueList] == consts.RECEIVE_STATUS_YES {
								status = consts.RECEIVE_STATUS_YES
							}
						}
					}

					drawList = append(drawList, &gmsg.ConditionAndReward{
						No:          dv.No,
						ValueList:   dv.ValueList,
						RewardList:  rewardList,
						TabName:     dv.TabName,
						ProgressNum: progressNum,
						Status:      uint32(status),
					})
				}
			}

			if v.ConfigPayLottery.IsOpenExchange {
				for _, ev := range v.ConfigPayLottery.ExchangeRewardList {
					rewardList := make([]*gmsg.RewardInfo, 0)
					for _, e1v := range ev.RewardList {
						rewardList = append(rewardList, &gmsg.RewardInfo{
							ItemTableId:  e1v.ItemTableId,
							Num:          e1v.Num,
							ExpireTimeId: e1v.ExpireTimeId,
						})
					}

					drawList = append(drawList, &gmsg.ConditionAndReward{
						No:         ev.No,
						ValueList:  ev.ValueList,
						RewardList: rewardList,
						TabName:    ev.TabName,
					})
				}
			}

			configPayLottery = gmsg.ActivityConfigPayLottery{
				ConsumeItemID:         v.ConfigPayLottery.ConsumeItemID,
				DayFreeDrawNum:        v.ConfigPayLottery.DayFreeDrawNum,
				IsOpenLucky:           v.ConfigPayLottery.IsOpenLucky,
				TakeLuckyNum:          v.ConfigPayLottery.TakeLuckyNum,
				LuckyNum:              v.ConfigPayLottery.LuckyNum,
				LuckyResetType:        uint32(v.ConfigPayLottery.LuckyResetType),
				MaxLuckyNum:           v.ConfigPayLottery.MaxLuckyNum,
				IsOpenDrawNum:         v.ConfigPayLottery.IsOpenDrawNum,
				DrawResetType:         uint32(v.ConfigPayLottery.DrawResetType),
				DrawNumRewardList:     drawList,
				IsOpenExchange:        v.ConfigPayLottery.IsOpenExchange,
				ExchangeConsumeItemID: v.ConfigPayLottery.ExchangeConsumeItemID,
				ExchangeRewardList:    exList,
				PayLotteryRewardList:  lrList,
			}
		case consts.ActivityTplType_KingRode:
			completeProgress := uint32(0)
			rewardElite, rewardAdvanced := make([]*gmsg.KingRodeReward, 0), make([]*gmsg.KingRodeReward, 0)
			isHave := false
			if len(tEntityPlayer.KingRodeActivityList) > 0 {
				for _, val := range tEntityPlayer.KingRodeActivityList {
					if val.ActivityId == v.ActivityId {
						isHave = true
						completeProgress = val.CompleteProgress
						for _, vls := range val.RewardElite {
							reward := new(gmsg.KingRodeReward)
							stack.SimpleCopyProperties(reward, &vls)
							rewardElite = append(rewardElite, reward)
						}
						for _, vls := range val.RewardAdvanced {
							reward := new(gmsg.KingRodeReward)
							stack.SimpleCopyProperties(reward, &vls)
							rewardAdvanced = append(rewardAdvanced, reward)
						}
					}
				}
			}
			if !isHave {
				resRewardElite, resRewardAdvanced := KingRodeMr.AddKingRodeActivityList(v.ActivityId, tEntityPlayer)
				rewardElite, rewardAdvanced = resRewardElite, resRewardAdvanced
			}
			KingRodeConfig = gmsg.KingRodeProgressList{
				CompleteProgress: completeProgress,
				RewardElite:      rewardElite,
				RewardAdvanced:   rewardAdvanced,
			}
		}

		info := gmsg.LoginActivityInfo{
			ActivityId:           v.ActivityId,
			TimeType:             v.TimeType,
			StartTime:            v.StartTime,
			EndTime:              v.EndTime,
			AType:                uint32(v.AType),
			SubType:              v.SubType,
			ActivityName:         v.ActivityName,
			ConfigTurntable:      &configTurntable,
			ConfigPayLottery:     &configPayLottery,
			ConfigPay:            &configPay,
			ConfigBattle:         &configBattle,
			KingRodeProgressList: &KingRodeConfig,
		}
		resp = append(resp, &info)
	}

	log.Info("-->logic--_Activity--GetLoginActivityListRequest--Resp:", resp)

	return resp
}

// 更新对战活动的进度
func (s *_Activity) UpdateBattleActivityProgress(entityID uint32) {
	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--UpdateBattleActivityProgress--GetEntityPlayerById--err--", err)
		return
	}

	now := time.Now().Unix()
	for _, v := range s.List {
		if now < v.StartTime || now > v.EndTime {
			continue
		}
		if v.AType != consts.ActivityTplType_Battle {
			continue
		}
		if len(v.ConfigBattle.ConditionAndRewardList) <= 0 {
			continue
		}

		for _, cv := range v.ConfigBattle.ConditionAndRewardList {
			var isIn bool
			if len(tEntityPlayer.ProgressActivityList) > 0 {
				for pk, pv := range tEntityPlayer.ProgressActivityList {
					if v.ActivityId != pv.ActivityId && cv.ValueList == pv.TargetProgress {
						isIn = true

						tEntityPlayer.ProgressActivityList[pk].CompleteProgress += 1
					}
				}
			}

			if !isIn {
				uuID, _ := uuid.NewUUID()
				now := time.Now().Unix()
				tEntityPlayer.ProgressActivityList = append(tEntityPlayer.ProgressActivityList, entity.ProgressActivityStatus{
					Id:               uuID.String(),
					ActivityId:       v.ActivityId,
					ConfigSerial:     cv.No,
					TargetProgress:   cv.ValueList,
					CompleteProgress: 1,
					StateReward:      consts.RECEIVE_STATUS_NO,
					Timestamp:        now,
				})
			}
		}
	}

	tEntityPlayer.SyncEntity(1)

	respList := make([]*gmsg.ActivityInfo, 0)
	for _, pav := range tEntityPlayer.ProgressActivityList {
		if pav.StateReward == consts.RECEIVE_STATUS_YES {
			continue
		}

		var isIn bool
		for aik, aiv := range respList {
			if aiv.ActivityId == pav.ActivityId {
				isIn = true

				respList[aik].List = append(respList[aik].List, &gmsg.ActivityProgress{
					Id:               pav.Id,
					TargetProgress:   pav.TargetProgress,
					CompleteProgress: pav.CompleteProgress,
					StateReward:      pav.StateReward,
					ConfigSerial:     pav.ConfigSerial,
				})
			}
		}

		if !isIn {
			info := gmsg.ActivityProgress{
				Id:               pav.Id,
				TargetProgress:   pav.TargetProgress,
				CompleteProgress: pav.CompleteProgress,
				StateReward:      pav.StateReward,
				ConfigSerial:     pav.ConfigSerial,
			}
			progressList := make([]*gmsg.ActivityProgress, 0)
			progressList = append(progressList, &info)
			respList = append(respList, &gmsg.ActivityInfo{
				ActivityId: pav.ActivityId,
				List:       progressList,
			})
		}
	}
	//开始初始化桌面信息
	resp := &gmsg.UpdateBattleActivityProgressResponse{
		List: respList,
	}

	log.Info("-->logic--_Activity--UpdateBattleActivityProgress--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_UpdateBattleActivityProgressResponse, resp, targetEntityIDs)
	return
}

// 更新支付活动的进度
func (s *_Activity) UpdatePayActivityProgress(entityID uint32, num uint32) {
	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 || num <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--UpdatePayActivityProgress--GetEntityPlayerById--err--", err)
		return
	}

	now := time.Now().Unix()
	for _, v := range s.List {
		if now < v.StartTime || now > v.EndTime {
			continue
		}
		if v.AType != consts.ActivityTplType_Pay {
			continue
		}
		if len(v.ConfigPay.ConditionAndRewardList) <= 0 {
			continue
		}

		for _, cv := range v.ConfigPay.ConditionAndRewardList {
			var isIn bool
			if len(tEntityPlayer.ProgressActivityList) > 0 {
				for pk, pv := range tEntityPlayer.ProgressActivityList {
					if v.ActivityId != pv.ActivityId && cv.ValueList == pv.TargetProgress {
						isIn = true

						tEntityPlayer.ProgressActivityList[pk].CompleteProgress += num
					}
				}
			}

			if !isIn {
				uuID, _ := uuid.NewUUID()
				now := time.Now().Unix()
				tEntityPlayer.ProgressActivityList = append(tEntityPlayer.ProgressActivityList, entity.ProgressActivityStatus{
					Id:               uuID.String(),
					ActivityId:       v.ActivityId,
					TargetProgress:   cv.ValueList,
					CompleteProgress: num,
					StateReward:      consts.RECEIVE_STATUS_NO,
					Timestamp:        now,
				})
			}
		}
	}

	tEntityPlayer.SyncEntity(1)

	respList := make([]*gmsg.ActivityInfo, 0)
	for _, pav := range tEntityPlayer.ProgressActivityList {
		if pav.StateReward == consts.RECEIVE_STATUS_YES {
			continue
		}

		var isIn bool
		for aik, aiv := range respList {
			if aiv.ActivityId == pav.ActivityId {
				isIn = true

				respList[aik].List = append(respList[aik].List, &gmsg.ActivityProgress{
					Id:               pav.Id,
					TargetProgress:   pav.TargetProgress,
					CompleteProgress: pav.CompleteProgress,
					StateReward:      pav.StateReward,
					ConfigSerial:     pav.ConfigSerial,
				})
			}
		}

		if !isIn {
			info := gmsg.ActivityProgress{
				Id:               pav.Id,
				TargetProgress:   pav.TargetProgress,
				CompleteProgress: pav.CompleteProgress,
				StateReward:      pav.StateReward,
				ConfigSerial:     pav.ConfigSerial,
			}
			progressList := make([]*gmsg.ActivityProgress, 0)
			progressList = append(progressList, &info)
			respList = append(respList, &gmsg.ActivityInfo{
				ActivityId: pav.ActivityId,
				List:       progressList,
			})
		}
	}
	//开始初始化桌面信息
	resp := &gmsg.UpdatePayActivityProgressResponse{
		List: respList,
	}

	log.Info("-->logic--_Activity--UpdatePayActivityProgress--Resp:", resp)

	targetEntityIDs := []uint32{entityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_UpdatePayActivityProgressResponse, resp, targetEntityIDs)
	return
}

// 更新转盘活动的进度
func (s *_Activity) UpdateTurntableActivityProgress(entityID uint32) {
	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--UpdateTurntableActivityProgress--GetEntityPlayerById--err--", err)
		return
	}

	respList := make([]*gmsg.TurntableInfo, 0)
	var isChange bool
	t := time.Now()
	now := t.Unix()
	addTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	for _, v := range s.List {
		if v.AType != consts.ActivityTplType_Turntable {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}

		if v.ConfigTurntable.DrawNumConfig <= 0 {
			continue
		}

		oldPlayNum := tEntityPlayer.DayBattleNum.Num / v.ConfigTurntable.DrawNumConfig
		if tEntityPlayer.DayBattleNum.Timestamp == addTime {
			tEntityPlayer.DayBattleNum.Num += 1
		} else {
			tEntityPlayer.DayBattleNum = entity.DayBattleNum{
				Num:       1,
				Timestamp: addTime,
			}
		}

		//计算可转转盘次数
		totalNum := v.ConfigTurntable.FreeDrawNum
		if tEntityPlayer.DayBattleNum.Timestamp == addTime {
			if v.ConfigTurntable.DrawNumConfig > 0 {
				if tEntityPlayer.DayBattleNum.Num > 0 {
					playNum := tEntityPlayer.DayBattleNum.Num / v.ConfigTurntable.DrawNumConfig
					totalNum += playNum

					if playNum > oldPlayNum {
						isChange = true
					}
				}
			}
		}

		if totalNum > v.ConfigTurntable.TotalDrawNum {
			isChange = false
			continue
		}

		respList = append(respList, &gmsg.TurntableInfo{
			Id:           "",
			ActivityId:   v.ActivityId,
			TotalDrawNum: totalNum,
		})
	}

	tEntityPlayer.SyncEntity(1)

	if isChange {
		//开始初始化桌面信息
		resp := &gmsg.UpdateTurntableActivityProgressResponse{
			List: respList,
		}

		log.Info("-->logic--_Activity--UpdateTurntableActivityProgress--Resp:", resp)

		targetEntityIDs := []uint32{entityID}

		//广播初始化消息
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_UpdateTurntableActivityProgressResponse, resp, targetEntityIDs)
	}

	return
}

// 更新王者活动的进度
func (s *_Activity) UpdateKingRodeActivityProgress(entityID uint32) {
	if len(s.List) <= 0 {
		return
	}

	if entityID <= 0 {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err != nil {
		log.Waring("-->logic--_Activity--UpdateKingRodeActivityProgress--GetEntityPlayerById--err--", err)
		return
	}

	resp := &gmsg.KingRodeProgressSync{}

	t := time.Now()
	now := t.Unix()
	for _, v := range s.List {
		if v.AType != consts.ActivityTplType_KingRode {
			continue
		}
		if now < v.StartTime || now > v.EndTime {
			continue
		}

		for key, val := range tEntityPlayer.KingRodeActivityList {
			if val.ActivityId == v.ActivityId {
				tEntityPlayer.KingRodeActivityList[key].CompleteProgress += 1
				resp.CompleteProgress = tEntityPlayer.KingRodeActivityList[key].CompleteProgress
				for index, vs := range val.RewardElite {
					if vs.StateReward == 1 && vs.TargetProgress <= resp.CompleteProgress {
						value := &vs
						value.StateReward = 2
						value.AddTimestamp = time.Now().Unix()
						tEntityPlayer.KingRodeActivityList[key].RewardElite[index] = *value
					}
				}

				for index, vs := range val.RewardAdvanced {
					if vs.StateReward == 1 && vs.TargetProgress <= resp.CompleteProgress {
						value := &vs
						value.StateReward = 2
						value.AddTimestamp = time.Now().Unix()
						tEntityPlayer.KingRodeActivityList[key].RewardAdvanced[index] = *value
					}
				}

				break
			}
		}
	}

	targetEntityIDs := []uint32{entityID}
	if resp.CompleteProgress > uint32(0) {
		tEntityPlayer.SyncEntity(1)
		log.Info("-->logic--_Activity--UpdateKingRodeActivityProgress--Resp:", resp)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_UpdateKingRodeActivityProgressResponse, resp, targetEntityIDs)
	}
	return
}

func (s *_Activity) GetPeakRankSettle() *gmsg.PeakRankInfo {
	resp := gmsg.PeakRankInfo{
		Status:      uint32(gmsg.ReceiveStatus_Receive_Status_Yes),
		ID:          1,
		PeakRankLv:  24,
		PeakRankExp: 3,
	}

	return &resp
}

func (s *_Activity) checkPerActivityPeakRank(entityID uint32) entity.PeakRankHist {
	resp := entity.PeakRankHist{}
	if entityID <= 0 {
		return resp
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		log.Waring("-->logic--_LoginReward--checkPerPerActivityPeakRank--GetEntityByID--tEntity == nil")
		return resp
	}

	now := time.Now().Unix()
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	hisList := tEntityPlayer.PeakRankHist
	if len(hisList) > 0 {
		for _, v := range hisList {
			if v.Status == consts.RECEIVE_STATUS_YES && now < v.AwardTime {
				resp = v
			}
		}
	}

	return resp
}
