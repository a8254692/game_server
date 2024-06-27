package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/DBServer/initialize/consts"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/db/collection"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"fmt"
	"reflect"
	"sort"
)

/***
 *@disc:
 *@author: lsj
 *@date: 2023/9/11
 */

type _Player struct {
}

var Player _Player

func (c *_Player) Init() {
	event.OnNet(gmsg.MsgTile_Player_InfoRequest, reflect.ValueOf(c.OnPlayerInfoRequestFromGame))
}

func (c *_Player) getPlayerCueList(tEntityPlayer *entity.EntityPlayer) (int, uint32, []*gmsg.CueData) {
	list, cueCharmScore := make([]*gmsg.CueData, 0), uint32(0)
	for _, value := range tEntityPlayer.BagList {
		if value.ItemType == conf.Cue {
			cueData := new(gmsg.CueData)
			cueData.TableID = value.TableID
			cueData.CharmScore = Table.GetCueCharmScore(value.TableID)
			cueCharmScore += cueData.CharmScore
			list = append(list, cueData)
		}
	}
	return len(list), cueCharmScore, list
}

// 游戏生涯请求
func (c *_Player) getPlayerStatistics(entityID uint32) *gmsg.PlayStatisticsData {
	tEntity := Entity.EmEntityPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return nil
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	statistics := new(collection.UserDataStatistics)
	statistics.SetDBConnect(consts.COLLECTION_STATISTICS)
	yes := statistics.InitFormDB(tEntityPlayer.EntityID, DBConnect)
	if !yes {
		log.Info("UserDataStatistics entityid is err，用户没有对局数据", entityID)
	}

	msgResponse := &gmsg.PlayStatisticsData{}
	msgResponse.PeakRankLv = tEntityPlayer.PeakRankLv
	msgResponse.AchievementLV = tEntityPlayer.AchievementLV
	msgResponse.ProfitGold = uint64(statistics.AccumulateGold)
	msgResponse.ReceivingGifts = 0
	msgResponse.AccumulateGoal = statistics.AccumulateGoal
	msgResponse.OneCueClear = statistics.OneCueClear
	msgResponse.C8MaxContinuousWin = statistics.C8MaxContinuousWin

	msgResponse.List = make([]*gmsg.GameStatistics, 0)
	c8GameStatistics := new(gmsg.GameStatistics)
	c8GameStatistics.GameType = 0
	c8GameStatistics.PlayNum = statistics.C8PlayNum
	c8GameStatistics.C8MaxContinuousWin = statistics.C8MaxContinuousWin
	c8GameStatistics.OneCueClear = statistics.OneCueClear
	c8GameStatistics.EscapePer = 0
	c8GameStatistics.WinPer = 0
	if statistics.C8PlayNum > 0 {
		c8GameStatistics.EscapePer = float32(tools.FloatRound(float64(statistics.C8EscapeNum)/float64(statistics.C8PlayNum)*100, 2))
		c8GameStatistics.WinPer = float32(tools.FloatRound(float64(statistics.C8WinNum)/float64(statistics.C8PlayNum)*100, 2))
	}

	msgResponse.List = append(msgResponse.List, c8GameStatistics)
	return msgResponse
}

// 赠送的礼物记录
func (c *_Player) getGiftsList(data []entity.GiveGift) (res []*gmsg.GiftData) {
	for _, val := range data {
		gift := new(gmsg.GiftData)
		tEntitys := Entity.EmEntityPlayer.GetEntityByID(val.EntityID)
		if tEntitys == nil {
			continue
		}
		tEntityPlayers := tEntitys.(*entity.EntityPlayer)
		stack.SimpleCopyProperties(gift, tEntityPlayers)
		gift.LastAddTime = val.LastAddTime[0 : len(val.LastAddTime)-3]
		gift.GiveNum = val.GiveNum
		gift.PopularityValue = val.PopularityValue
		gift.GiftsList = make([]*gmsg.GiftIds, 0)
		for _, va := range val.IdLog {
			giftlids := new(gmsg.GiftIds)
			stack.SimpleCopyProperties(giftlids, &va)
			gift.GiftsList = append(gift.GiftsList, giftlids)
		}
		sort.Slice(gift.GiftsList, func(i, j int) bool {
			return gift.GiftsList[i].GiftID < gift.GiftsList[j].GiftID
		})
		res = append(res, gift)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].PopularityValue > res[j].PopularityValue
	})

	return
}

// 守护榜
func (c *_Player) getReceivingGiftsRank(data []entity.RecGift) (res []*gmsg.GiftData) {
	for _, vl := range data {
		gift := new(gmsg.GiftData)
		tEntitys := Entity.EmEntityPlayer.GetEntityByID(vl.EntityID)
		if tEntitys == nil {
			continue
		}
		tEntityPlayers := tEntitys.(*entity.EntityPlayer)
		stack.SimpleCopyProperties(gift, tEntityPlayers)
		gift.LastAddTime = vl.LastAddTime[0 : len(vl.LastAddTime)-3]
		gift.GiveNum = vl.GiveNum
		gift.GiftsList = make([]*gmsg.GiftIds, 0)
		gift.PopularityValue = vl.PopularityValue
		giftMap := make(map[uint32]uint32, 0)
		for _, va := range vl.Log {
			for _, v := range va.IdLog {
				if vs, ok := giftMap[v.GiftID]; ok {
					giftMap[v.GiftID] = vs + v.Number
				} else {
					giftMap[v.GiftID] = v.Number
				}
			}
		}
		for key, value := range giftMap {
			giftlids := new(gmsg.GiftIds)
			giftlids.GiftID = key
			giftlids.Number = value
			gift.GiftsList = append(gift.GiftsList, giftlids)
		}
		sort.Slice(gift.GiftsList, func(i, j int) bool {
			return gift.GiftsList[i].GiftID < gift.GiftsList[j].GiftID
		})
		res = append(res, gift)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].PopularityValue > res[j].PopularityValue
	})

	return
}

// 接收礼物记录
func (c *_Player) getReceivingGiftList(data []entity.RecGift) (res []*gmsg.GiftData) {
	for _, vl := range data {
		for _, va := range vl.Log {
			gift := new(gmsg.GiftData)
			tEntitys := Entity.EmEntityPlayer.GetEntityByID(vl.EntityID)
			if tEntitys == nil {
				continue
			}
			tEntityPlayers := tEntitys.(*entity.EntityPlayer)
			stack.SimpleCopyProperties(gift, tEntityPlayers)
			gift.GiftsList = make([]*gmsg.GiftIds, 0)
			gift.LastAddTime = va.AddTime[0 : len(va.AddTime)-3]
			gift.GiveNum = va.GiveNum
			gift.PopularityValue = va.PopularityValue
			for _, v := range va.IdLog {
				giftlids := new(gmsg.GiftIds)
				stack.SimpleCopyProperties(giftlids, &v)
				gift.GiftsList = append(gift.GiftsList, giftlids)
			}
			sort.Slice(gift.GiftsList, func(i, j int) bool {
				return gift.GiftsList[i].GiftID < gift.GiftsList[j].GiftID
			})
			res = append(res, gift)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return tools.GetUnixFromStr(fmt.Sprintf("%s:00", res[i].LastAddTime)) > tools.GetUnixFromStr(fmt.Sprintf("%s:00", res[j].LastAddTime))
	})

	return
}

func (c *_Player) OnPlayerInfoRequestFromGame(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PlayerInfoRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	var tEntityPlayer *entity.EntityPlayer

	msgResponse := &gmsg.PlayerInfoResponse{}
	msgResponse.IsFriend = false
	msgResponse.IsOpen = msgBody.IsOpen
	msgResponse.PlayStatisticsData = new(gmsg.PlayStatisticsData)

	entityId := msgBody.EntityID
	if *msgBody.QEntityID > uint32(0) {
		entityId = *msgBody.QEntityID
		tEntity := Entity.EmEntityPlayer.GetEntityByID(entityId)
		if tEntity == nil {
			log.Error("-->OnPlayerInfoRequestFromGame-->QEntityID is err", entityId)
			return
		}
		tEntityPlayer = tEntity.(*entity.EntityPlayer)
		msgResponse.IsFriend = tEntityPlayer.IsInFansList(msgBody.EntityID)
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(entityId)
	if tEntity == nil {
		log.Error("-->OnPlayerInfoRequestFromGame-->is nil", entityId)
		return
	}
	tEntityPlayer = tEntity.(*entity.EntityPlayer)

	cueNum, charmScore, cueList := c.getPlayerCueList(tEntityPlayer)
	msgResponse.CueList = make([]*gmsg.CueData, 0)
	msgResponse.CueList = cueList
	msgResponse.CueNum = uint32(cueNum)
	msgResponse.CueCharmScore = charmScore

	msgResponse.MainPlayer = &gmsg.EntityPlayer{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.QEntityID = msgBody.QEntityID

	err := stack.StructCopySame_Json(msgResponse.MainPlayer, tEntityPlayer)
	if err != nil {
		log.Error(err)
		return
	}
	msgResponse.PlayStatisticsData = c.getPlayerStatistics(entityId)

	msgResponse.OpenGifts = tEntityPlayer.OpenGifts
	if *msgBody.QEntityID == 0 || tEntityPlayer.OpenGifts && *msgBody.QEntityID > 0 {
		msgResponse.TAGiftsList = c.getGiftsList(tEntityPlayer.GiftsList)
	}
	msgResponse.ReceivingGiftsList = c.getReceivingGiftList(tEntityPlayer.ReceivingGifts)
	msgResponse.ReceivingGiftsRank = c.getReceivingGiftsRank(tEntityPlayer.ReceivingGifts)

	//fmt.Println("msgResponse.ReceivingGiftsList", msgResponse.ReceivingGiftsList)
	//fmt.Println("msgResponse.ReceivingGiftsRank", msgResponse.ReceivingGiftsRank)

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Player_InfoResponse, msgResponse, network.ServerType_Game)
}
