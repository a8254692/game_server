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
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

/*** 守护榜
 *@disc:
 *@author: lsj
 *@date: 2024/1/3
 */

type _Gifts struct {
	lock           sync.RWMutex
	MaxRankNum     int //统计榜上人数
	GiftsFixReward map[uint32][]uint32
	PopTotalRank   []*gmsg.InGiftInfo
	PopWeekRank    []*gmsg.InGiftInfo
}

var GiftsMr _Gifts

func (c *_Gifts) Init() {
	c.MaxRankNum = 50
	c.GiftsFixReward = make(map[uint32][]uint32, 0)
	c.PopTotalRank, c.PopWeekRank = make([]*gmsg.InGiftInfo, 0), make([]*gmsg.InGiftInfo, 0)
	c.initGiftsFixReward()
	time.AfterFunc(time.Millisecond*500, c.SendSyncMsgToDbForPopData)
	event.OnNet(gmsg.MsgTile_Player_ChangOpenGiftRequest, reflect.ValueOf(c.OnChangPlayerOpenGiftsRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_Player_GiveGiftResponse), reflect.ValueOf(c.OnPlayerGiveGiftDbResponse))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PopularitySync_Response), reflect.ValueOf(c.SetMsgToDbForPopularityData))
}

func (c *_Gifts) initGiftsFixReward() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, val := range Table.ItemCfg {
		if val.TypeN == conf.Item_7 && len(val.FixedReward) > 0 {
			c.GiftsFixReward[val.TableID] = val.FixedReward[0]
		}
	}
}

func (c *_Gifts) getGiftsPopularityValue(giftsId uint32) (uint32, uint32) {
	v, ok := c.GiftsFixReward[giftsId]
	if !ok {
		return uint32(0), uint32(0)
	}
	return v[0], v[1]
}

// 通知DB同步数据人气排行榜
func (c *_Gifts) SendSyncMsgToDbForPopData() {
	request := &gmsg.InPopularityRankRequest{}
	request.MaxRankNum = uint32(c.MaxRankNum)

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PopularitySync_Request), request, network.ServerType_DB)
}

// 同步人气排行榜数据
func (c *_Gifts) SetMsgToDbForPopularityData(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	var msgBody gmsg.InPopularityRankResponse

	stack.BytesToStruct_Gob(msgEV.MsgBody, &msgBody)

	if len(msgBody.TotalList) == 0 && len(msgBody.WeekList) == 0 {
		log.Info("人气榜为空。")
		return
	}

	c.PopTotalRank, c.PopWeekRank = nil, nil
	c.PopTotalRank = msgBody.TotalList
	c.PopWeekRank = msgBody.WeekList
	log.Info("-->同步人气榜成功。", c.PopTotalRank, "-->", c.PopWeekRank)
}

func (c *_Gifts) checkIsInRank(entityID uint32) (tn, wn int) {
	tn, wn = -1, -1
	for k, v := range c.PopTotalRank {
		if v.EntityID == entityID {
			tn = k
			break
		}
	}

	for k, v := range c.PopWeekRank {
		if v.EntityID == entityID {
			wn = k
			break
		}
	}
	return
}

// 插入排行榜
func (c *_Gifts) addPopRank(info *gmsg.InGiftInfo, afterValue, value, entityID uint32) {
	tn, wn := c.checkIsInRank(entityID)
	if tn < 0 {
		c.PopTotalRank = append(c.PopTotalRank, info)
	} else if tn >= 0 {
		data := c.PopTotalRank[tn]
		data.PopularityValue = afterValue
		data.AddTamp = time.Now().Unix()
		c.PopTotalRank[tn] = data
	}

	if len(c.PopTotalRank) > c.MaxRankNum {
		c.PopTotalRank = c.PopTotalRank[0:c.MaxRankNum]
	}
	sort.Slice(c.PopTotalRank, func(i, j int) bool {
		return c.PopTotalRank[i].PopularityValue > c.PopTotalRank[j].PopularityValue || (c.PopTotalRank[i].PopularityValue == c.PopTotalRank[j].PopularityValue && c.PopTotalRank[i].AddTamp < c.PopTotalRank[j].AddTamp)
	})
	//fmt.Println("c.PopTotalRank", c.PopTotalRank)

	if wn < 0 {
		c.PopWeekRank = append(c.PopWeekRank, info)
	} else if wn >= 0 {
		data := c.PopWeekRank[wn]
		data.PopularityValue += value
		data.AddTamp = time.Now().Unix()
		c.PopWeekRank[wn] = data
	}

	if len(c.PopWeekRank) > c.MaxRankNum {
		c.PopWeekRank = c.PopWeekRank[0:c.MaxRankNum]
	}
	sort.Slice(c.PopWeekRank, func(i, j int) bool {
		return c.PopWeekRank[i].PopularityValue > c.PopWeekRank[j].PopularityValue || (c.PopWeekRank[i].PopularityValue == c.PopWeekRank[j].PopularityValue && c.PopWeekRank[i].AddTamp < c.PopWeekRank[j].AddTamp)
	})
	//fmt.Println("c.PopWeekRank", c.PopWeekRank)
}

// 礼物开关切换
func (c *_Gifts) OnChangPlayerOpenGiftsRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ChangPlayerOpenGiftsRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.OpenGifts {
		tEntityPlayer.OpenGifts = false
	} else {
		tEntityPlayer.OpenGifts = true
	}
	tEntityPlayer.SyncEntity(1)
	msgResponse := &gmsg.ChangPlayerOpenGiftsResponse{}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_ChangOPenGiftResponse, msgResponse, []uint32{msgBody.EntityID})
	return
}

// 赠送礼物 前端->游戏->db服
func (c *_Gifts) GiveGiftRequest(useItemReq *gmsg.UseItemRequest, giftId uint32, resParam entity.ResParam) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, value := c.getGiftsPopularityValue(giftId)
	if value == 0 {
		return errors.New(fmt.Sprintf("giftId is err,%d", giftId))
	}

	nowMin := tools.GetTimeMinFormat()
	toPlayer, _ := GetEntityPlayerById(useItemReq.ToEntityID)
	if toPlayer == nil {
		//离线用户，去db处理
		inReq := &gmsg.InGiveGiftRequest{}
		inReq.ToEntityID = useItemReq.ToEntityID
		inReq.Number = useItemReq.Number
		inReq.ItemID = useItemReq.ItemID
		inReq.PopularityValue = useItemReq.Number * value
		inReq.EntityID = useItemReq.EntityID
		inReq.NowMin = nowMin
		inReq.GiftsId = giftId
		inReq.Uuid = resParam.Uuid
		inReq.SysID = resParam.SysID
		inReq.ActionID = resParam.ActionID
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_Player_GiveGiftRequest), inReq, network.ServerType_DB)
		return nil
	}
	//在线用户，直接赠送
	gerr := c.giveGiftsFunc(useItemReq.EntityID, useItemReq.ToEntityID, giftId, useItemReq.Number, resParam, nowMin)
	if gerr != nil {
		return errors.New("giveGifts is err")
	}
	c.receivingFunc(toPlayer, useItemReq.EntityID, giftId, useItemReq.Number, resParam, nowMin)

	inGiftInfo := new(gmsg.InGiftInfo)
	stack.SimpleCopyProperties(inGiftInfo, toPlayer)
	inGiftInfo.AddTamp = time.Now().Unix()
	go c.addPopRank(inGiftInfo, toPlayer.PopularityValue, useItemReq.Number*value, useItemReq.ToEntityID)

	return nil
}

// 赠送礼物 前端->游戏->db服
func (c *_Gifts) OnPlayerGiveGiftDbResponse(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()

	msgBody := &gmsg.InGiveGiftResponse{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	if msgBody.Code > 0 {
		log.Error("-->OnPlayerGiveGiftDbResponse-->err", "-->entityid-->", msgBody.EntityID, "-->toEntityID->", msgBody.ToEntityID)
		return
	}

	resParam := &entity.ResParam{Uuid: msgBody.Uuid, SysID: msgBody.SysID, ActionID: msgBody.ActionID}
	gerr := c.giveGiftsFunc(msgBody.EntityID, msgBody.ToEntityID, msgBody.GiftsId, msgBody.Number, *resParam, msgBody.NowMin)
	if gerr != nil {
		return
	}

	inGiftInfo := new(gmsg.InGiftInfo)
	stack.SimpleCopyProperties(inGiftInfo, msgBody)
	inGiftInfo.EntityID = msgBody.ToEntityID
	inGiftInfo.AddTamp = time.Now().Unix()
	go c.addPopRank(inGiftInfo, msgBody.AfterPopularityValue, msgBody.PopularityValue, msgBody.ToEntityID)
	return
}

// 赠送礼物
func (c *_Gifts) giveGiftsFunc(entityID, toEntityID, giftId, num uint32, resParam entity.ResParam, nowMin string) error {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return errors.New("EntityID is err")
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	_, value := c.getGiftsPopularityValue(giftId)
	if value == 0 {
		return errors.New(fmt.Sprintf("giftId is err,%d", giftId))
	}

	resGiveGifts, indexg := tEntityPlayer.GetGiftsList(toEntityID)
	if resGiveGifts == nil {
		give := &entity.GiveGift{
			EntityID:        toEntityID,
			LastAddTime:     nowMin,
			GiveNum:         uint32(1),
			PopularityValue: value * num,
		}
		give.IdLog = make([]entity.GiftLog, 0)
		give.IdLog = append(give.IdLog, entity.GiftLog{GiftID: giftId, Number: num})
		tEntityPlayer.GiftsList = append(tEntityPlayer.GiftsList, *give)
	} else {
		resGiveGifts.LastAddTime = nowMin
		resGiveGifts.PopularityValue += value * num
		resGiveGifts.GiveNum += uint32(1)
		isAdd := true
		for k, v := range resGiveGifts.IdLog {
			if v.GiftID == giftId {
				v.Number += num
				resGiveGifts.IdLog[k] = v
				isAdd = false
				break
			}
		}
		if isAdd {
			resGiveGifts.IdLog = append(resGiveGifts.IdLog, entity.GiftLog{GiftID: giftId, Number: num})
		}

		tEntityPlayer.GiftsList[indexg] = *resGiveGifts
	}

	tEntityPlayer.SyncEntity(1)
	return nil
}

// 接收礼物
func (c *_Gifts) receivingFunc(toPlayer *entity.EntityPlayer, entityID, giftId, num uint32, resParam entity.ResParam, nowMin string) {
	tableID, value := c.getGiftsPopularityValue(giftId)
	receivingGifts, indexr := toPlayer.GetReceivingGifts(entityID)
	if receivingGifts == nil {
		receive := &entity.RecGift{
			EntityID:        entityID,
			PopularityValue: value * num,
			GiveNum:         uint32(1),
			LastAddTime:     nowMin,
		}

		receive.Log = make([]entity.RecGiftLog, 0)
		receive.Log = append(receive.Log,
			entity.RecGiftLog{
				AddTime:         nowMin,
				GiveNum:         uint32(1),
				PopularityValue: value * num})

		receive.Log[0].IdLog = make([]entity.GiftLog, 0)
		receive.Log[0].IdLog = append(receive.Log[0].IdLog,
			entity.GiftLog{GiftID: giftId, Number: num})

		toPlayer.ReceivingGifts = append(toPlayer.ReceivingGifts, *receive)
	} else {
		receivingGifts.PopularityValue += value * num
		receivingGifts.GiveNum += uint32(1)
		receivingGifts.LastAddTime = nowMin

		//查询最后一个数据
		log := receivingGifts.Log[len(receivingGifts.Log)-1]
		//存在数据
		if log.AddTime == nowMin {
			isAdd := true
			log.PopularityValue += value * num
			log.GiveNum += uint32(1)
			for j, d := range log.IdLog {
				if d.GiftID == giftId {
					d.Number += num
					log.IdLog[j] = d
					isAdd = false
					break
				}
			}
			if isAdd {
				log.IdLog = append(log.IdLog, entity.GiftLog{GiftID: giftId, Number: num})
			}
			receivingGifts.Log[len(receivingGifts.Log)-1] = log
		} else {
			//不存在数据
			receivingGifts.Log = append(receivingGifts.Log,
				entity.RecGiftLog{
					AddTime:         nowMin,
					GiveNum:         uint32(1),
					PopularityValue: value * num,
					IdLog:           []entity.GiftLog{{GiftID: giftId, Number: num}},
				})
		}

		toPlayer.ReceivingGifts[indexr] = *receivingGifts
	}

	Player.UpdatePlayerPropertyItem(toPlayer.EntityID, tableID, int32(value*num), resParam)
}
