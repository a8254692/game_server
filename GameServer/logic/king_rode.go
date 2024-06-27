package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"fmt"
	"reflect"
	"sort"
	"time"
)

/***
 *@disc:王者之路
 *@author: lsj
 *@date: 2024/1/12
 */

type _KingRode struct {
	AdvancedPrice uint32
	Reward        map[uint32]KingRodeRw
	ConditionalId uint32
}

type KingRodeRw struct {
	RewardId       uint32
	RewardElite    entity.RewardEntity
	RewardAdvanced entity.RewardEntity
	Count          uint32
}

var KingRodeMr _KingRode

func (c *_KingRode) Init() {
	c.setAdvancedPrice()
	c.initKConfig()
	event.OnNet(gmsg.MsgTile_Reward_KingRodeAdvancedUnlockRequest, reflect.ValueOf(c.KingRodeAdvancedUnlockRequest))
	event.OnNet(gmsg.MsgTile_Reward_ReceiveKingRodeActivityRewardRequest, reflect.ValueOf(c.ReceiveKingRodeActivityRewardRequest))
}

func (c *_KingRode) setAdvancedPrice() {
	config, ok := Table.GetConstMap()["7"]
	if !ok {
		log.Error("setAdvancedPrice is err!")
		return
	}
	c.AdvancedPrice = config.Paramater1
}

func (c *_KingRode) initKConfig() {
	c.Reward = make(map[uint32]KingRodeRw, 0)
	for _, val := range Table.KingCfg {
		kingRodeRw := new(KingRodeRw)
		kingRodeRw.RewardId = val.TableID
		if len(val.RewardID1) > 1 && len(val.RewardID2) > 1 {
			kingRodeRw.RewardElite = c.getRewardFrom(val.RewardID1)
			kingRodeRw.RewardAdvanced = c.getRewardFrom(val.RewardID2)
		}
		if len(val.Count) > 1 {
			kingRodeRw.Count = val.Count[1]
		}
		c.Reward[val.TableID] = *kingRodeRw
	}
}

func (c *_KingRode) getKConfig() (rewardElite []entity.KingRodeReward, rewardAdvanced []entity.KingRodeReward) {
	for _, val := range Table.KingCfg {
		kingRode := new(entity.KingRodeReward)
		kingRode.RewardId = val.TableID
		kingRode.StateReward = uint32(0)
		if len(val.Count) > 1 {
			c.ConditionalId = val.Count[0]
			kingRode.TargetProgress = val.Count[1]
		}
		rewardAdvanced = append(rewardAdvanced, *kingRode)
		kingRode.StateReward = uint32(1)
		rewardElite = append(rewardElite, *kingRode)
	}
	sort.Slice(rewardAdvanced, func(i, j int) bool {
		return rewardAdvanced[i].RewardId < rewardAdvanced[j].RewardId
	})
	sort.Slice(rewardElite, func(i, j int) bool {
		return rewardElite[i].RewardId < rewardElite[j].RewardId
	})

	return rewardElite, rewardAdvanced
}

func (c *_KingRode) getRewardFrom(data []uint32) (res entity.RewardEntity) {
	itemType, _ := tools.GetItemTypeByTableId(data[0])
	switch itemType {
	case consts.Item, consts.PropertyItem, consts.Clothing:
		res = entity.RewardEntity{ItemTableId: data[0], Num: data[1], ExpireTimeId: 0}
	case consts.Cue, consts.Dress, consts.Effect:
		res = entity.RewardEntity{ItemTableId: data[0], Num: uint32(1), ExpireTimeId: data[1]}
	default:
	}
	return
}

// 解锁进阶版
func (c *_KingRode) KingRodeAdvancedUnlockRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.KingRodeAdvancedUnlockRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--_Activity--KingRodeAdvancedUnlockRequest--GetEntityPlayerById--err--", err)
		return
	}
	msgResponse := &gmsg.KingRodeAdvancedUnlockResponse{
		Code: uint32(1),
	}

	if c.AdvancedPrice == uint32(0) {
		log.Error("--->KingRodeAdvancedUnlockRequest-->AdvancedPrice=0")
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_KingRodeAdvancedUnlockResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if len(tEntityPlayer.KingRodeActivityList) > 0 {
		for key, val := range tEntityPlayer.KingRodeActivityList {
			if val.ActivityId == msgBody.ActivityId {
				if val.IsUnlockAdvanced {
					break
				}
				for index, vs := range val.RewardAdvanced {
					value := vs
					if vs.TargetProgress <= val.CompleteProgress {
						value.StateReward = 2
					} else {
						value.StateReward = 1
					}
					val.RewardAdvanced[index] = value
				}
				tEntityPlayer.KingRodeActivityList[key].RewardAdvanced = val.RewardAdvanced
				tEntityPlayer.KingRodeActivityList[key].IsUnlockAdvanced = true
				msgResponse.Code = uint32(0)
				tEntityPlayer.SyncEntity(1)
				break
			}
		}
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_KingRodeAdvancedUnlockResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 批量领取奖励
func (c *_KingRode) ReceiveKingRodeActivityRewardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ReceiveKingRodeActivityRewardRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--_Activity--ReceiveKingRodeActivityRewardRequest--GetEntityPlayerById--err--", err)
		return
	}
	fmt.Println("-->ReceiveKingRodeActivityRewardRequest-->", msgBody)
	msgResponse := &gmsg.ReceiveKingRodeActivityRewardResponse{
		Code:       uint32(1),
		ActivityId: msgBody.ActivityId,
		RewardType: msgBody.RewardType,
		RewardId:   msgBody.RewardId,
	}
	rewardList := make([]entity.RewardEntity, 0)
	resParam := GetResParam(consts.SYSTEM_ID_ACTIVITY, consts.Reward)
	if len(tEntityPlayer.KingRodeActivityList) == 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceiveKingRodeActivityRewardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	now := time.Now().Unix()
	for _, v := range Activity.List {
		if v.ActivityId != msgBody.ActivityId {
			continue
		}

		if v.AType != consts.ActivityTplType_KingRode {
			continue
		}

		if now < v.StartTime || now > v.EndTime {
			continue
		}
		for index, val := range tEntityPlayer.KingRodeActivityList {
			if val.ActivityId != msgBody.ActivityId {
				continue
			}

			if msgBody.RewardType != 2 {
				for key1, ve := range val.RewardElite {
					if ve.TargetProgress > val.CompleteProgress {
						continue
					}
					if ve.StateReward != 2 {
						continue
					}

					if msgBody.RewardId == 0 {
						value := ve
						value.StateReward = 3
						value.AddTimestamp = now
						tEntityPlayer.KingRodeActivityList[index].RewardElite[key1] = value
						rewardList = append(rewardList, c.Reward[ve.RewardId].RewardElite)
					} else if msgBody.RewardId == ve.RewardId && (msgBody.RewardType == 3 || msgBody.RewardType == 1) {
						value := ve
						value.StateReward = 3
						value.AddTimestamp = now
						tEntityPlayer.KingRodeActivityList[index].RewardElite[key1] = value
						rewardList = append(rewardList, c.Reward[ve.RewardId].RewardElite)
						break
					}
				}
			}

			if msgBody.RewardType != 1 {
				for key2, va := range val.RewardAdvanced {
					if va.StateReward != 2 {
						continue
					}

					if msgBody.RewardId == 0 {
						value := va
						value.StateReward = 3
						value.AddTimestamp = now
						tEntityPlayer.KingRodeActivityList[index].RewardAdvanced[key2] = value
						rewardList = append(rewardList, c.Reward[va.RewardId].RewardAdvanced)
					} else if msgBody.RewardId == va.RewardId && msgBody.RewardType >= 2 {
						value := va
						value.StateReward = 3
						value.AddTimestamp = now
						tEntityPlayer.KingRodeActivityList[index].RewardAdvanced[key2] = value
						rewardList = append(rewardList, c.Reward[va.RewardId].RewardAdvanced)
						break
					}
				}

			}
		}
	}

	log.Info("--ReceiveKingRodeActivityRewardRequest-->res", msgResponse)

	if len(rewardList) > 0 {
		msgResponse.Code = uint32(0)
		log.Info("--ReceiveKingRodeActivityRewardRequest-->res", msgResponse)
		Backpack.BackpackAddItemListAndUpdateItemSync(msgBody.EntityID, rewardList, *resParam)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceiveKingRodeActivityRewardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_ReceiveKingRodeActivityRewardResponse, msgResponse, []uint32{msgBody.EntityID})
	return
}

// 初始化王者之路活动
func (c *_KingRode) AddKingRodeActivityList(activityId string, tEntityPlayer *entity.EntityPlayer) (res1, res2 []*gmsg.KingRodeReward) {
	kingRodeProgress := entity.KingRodeProgress{
		ActivityId:       activityId,
		ConditionalId:    c.ConditionalId,
		CompleteProgress: uint32(0),
	}

	kingRodeProgress.RewardElite, kingRodeProgress.RewardAdvanced = make([]entity.KingRodeReward, 0), make([]entity.KingRodeReward, 0)

	KConfig1, KConfig2 := c.getKConfig()
	kingRodeProgress.RewardElite = KConfig1
	kingRodeProgress.RewardAdvanced = KConfig2
	for _, val := range KConfig1 {
		resReward := new(gmsg.KingRodeReward)
		stack.SimpleCopyProperties(resReward, &val)
		res1 = append(res1, resReward)
	}
	for _, val := range KConfig2 {
		resReward := new(gmsg.KingRodeReward)
		stack.SimpleCopyProperties(resReward, &val)
		res2 = append(res2, resReward)
	}
	tEntityPlayer.KingRodeActivityList = append(tEntityPlayer.KingRodeActivityList, kingRodeProgress)
	tEntityPlayer.SyncEntity(1)
	return
}

// 初始化王者之路活动
func (c *_KingRode) resetKingRodeActivityList(EntityID uint32) {
	tEntityPlayer, err := GetEntityPlayerById(EntityID)
	if err != nil {
		return
	}
	tEntityPlayer.KingRodeActivityList = nil
	KConfig1, KConfig2 := c.getKConfig()
	for _, v := range Activity.List {
		if v.AType != consts.ActivityTplType_KingRode {
			continue
		}
		kingRodeProgress := entity.KingRodeProgress{
			ActivityId:       v.ActivityId,
			ConditionalId:    c.ConditionalId,
			CompleteProgress: uint32(0),
			RewardElite:      KConfig1,
			RewardAdvanced:   KConfig2,
		}
		tEntityPlayer.KingRodeActivityList = append(tEntityPlayer.KingRodeActivityList, kingRodeProgress)

	}
	tEntityPlayer.SyncEntity(1)

	return
}

func (c *_KingRode) addUpdateKingRodeActivityProgress(entityId uint32) {
	Activity.UpdateKingRodeActivityProgress(entityId)
}
