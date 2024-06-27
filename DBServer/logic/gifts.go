package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/game_log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"reflect"
	"sort"
	"sync"
)

/***
 *@disc:
 *@author: lsj
 *@date: 2024/1/3
 */

type Gifts struct {
	lock sync.RWMutex
}

var GiftsMr Gifts

func (c *Gifts) Init() {
	//c.PopularitySync()
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_Player_GiveGiftRequest), reflect.ValueOf(c.OnPlayerGiveGiftRequestFromGame))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PopularitySync_Request), reflect.ValueOf(c.GetGameSyncPopRequest))
}

func (c *Gifts) GetGameSyncPopRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.InPopularityRankRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	buf := c.getPopularityArgs(msgBody.MaxRankNum)
	if buf == nil {
		return
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PopularitySync_Response), buf, network.ServerType_Game)
}

//func (c *Gifts) PopularitySync() {
//	buf := c.getPopularityArgs(50)
//	if buf == nil {
//		return
//	}
//	time.AfterFunc(time.Second*5, func() {
//		ConnectManager.SendMsgToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PopularitySync_Response), buf, network.ServerType_Game)
//	})
//}

func (c *Gifts) getPopularityArgs(maxRankNum uint32) []byte {
	resBody := &gmsg.InPopularityRankResponse{}
	resBody.TotalList, resBody.WeekList = make([]*gmsg.InGiftInfo, 0), make([]*gmsg.InGiftInfo, 0)

	beginTime := tools.GetThisWeekFirstDate()
	for _, val := range Entity.EmEntityPlayer.EntityMap {
		tEntityPlayer := val.(*entity.EntityPlayer)
		if tEntityPlayer.IsRobot || tEntityPlayer.PopularityValue == 0 {
			continue
		}
		value := uint32(0)
		giftInfoTotal := new(gmsg.InGiftInfo)
		stack.SimpleCopyProperties(giftInfoTotal, tEntityPlayer)
		resBody.TotalList = append(resBody.TotalList, giftInfoTotal)
		for _, vl := range tEntityPlayer.ReceivingGifts {
			if tools.GetUnixFromStr(vl.LastAddTime) >= beginTime {
				value += vl.PopularityValue
			}
		}
		if value > 0 {
			giftInfoWeek := new(gmsg.InGiftInfo)
			stack.SimpleCopyProperties(giftInfoWeek, tEntityPlayer)
			giftInfoWeek.PopularityValue = value
			resBody.WeekList = append(resBody.WeekList, giftInfoWeek)
		}
	}
	if len(resBody.TotalList) > int(maxRankNum) {
		resBody.TotalList = resBody.TotalList[0:int(maxRankNum)]
	}

	if len(resBody.WeekList) > int(maxRankNum) {
		resBody.WeekList = resBody.WeekList[0:int(maxRankNum)]
	}
	sort.Slice(resBody.TotalList, func(i, j int) bool {
		return resBody.TotalList[i].PopularityValue > resBody.TotalList[j].PopularityValue
	})
	sort.Slice(resBody.WeekList, func(i, j int) bool {
		return resBody.WeekList[i].PopularityValue > resBody.WeekList[j].PopularityValue
	})
	buf, _ := stack.StructToBytes_Gob(resBody)
	if len(buf) < 1 {
		return nil
	}
	return buf
}

func (c *Gifts) OnPlayerGiveGiftRequestFromGame(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()

	msgBody := &gmsg.InGiveGiftRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.InGiveGiftResponse{}
	msgResponse.Code = uint32(1)
	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.ToEntityID)
	if tEntity == nil {
		//接收者不存在
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_Player_GiveGiftResponse), msgResponse, network.ServerType_Game)
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse.AfterPopularityValue = tEntityPlayer.PopularityValue + msgBody.PopularityValue
	c.receivingDbFunc(tEntityPlayer, msgBody)

	stack.SimpleCopyProperties(msgResponse, msgBody)
	stack.SimpleCopyProperties(msgResponse, tEntityPlayer)
	msgResponse.PopularityValue = msgBody.PopularityValue
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = uint32(0)

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_Player_GiveGiftResponse), msgResponse, network.ServerType_Game)
	return
}

// 接收礼物
func (c *Gifts) receivingDbFunc(toPlayer *entity.EntityPlayer, req *gmsg.InGiveGiftRequest) {
	receivingGifts, indexr := toPlayer.GetReceivingGifts(req.EntityID)
	if receivingGifts == nil {
		receive := &entity.RecGift{
			EntityID:        req.EntityID,
			PopularityValue: req.PopularityValue,
			GiveNum:         uint32(1),
			LastAddTime:     req.NowMin,
		}

		receive.Log = make([]entity.RecGiftLog, 0)
		receive.Log = append(receive.Log,
			entity.RecGiftLog{
				AddTime:         req.NowMin,
				GiveNum:         uint32(1),
				PopularityValue: req.PopularityValue})

		receive.Log[0].IdLog = make([]entity.GiftLog, 0)
		receive.Log[0].IdLog = append(receive.Log[0].IdLog,
			entity.GiftLog{GiftID: req.GiftsId, Number: req.Number})

		toPlayer.ReceivingGifts = append(toPlayer.ReceivingGifts, *receive)
	} else {
		receivingGifts.PopularityValue += req.PopularityValue
		receivingGifts.GiveNum += uint32(1)
		receivingGifts.LastAddTime = req.NowMin

		//查询最后一个数据
		log := receivingGifts.Log[len(receivingGifts.Log)-1]
		//存在数据
		if log.AddTime == req.NowMin {
			isAdd := true
			log.PopularityValue += req.PopularityValue
			log.GiveNum += uint32(1)
			for j, d := range log.IdLog {
				if d.GiftID == req.GiftsId {
					d.Number += req.Number
					log.IdLog[j] = d
					isAdd = false
					break
				}
			}
			if isAdd {
				log.IdLog = append(log.IdLog, entity.GiftLog{GiftID: req.GiftsId, Number: req.Number})
			}
			receivingGifts.Log[len(receivingGifts.Log)-1] = log
		} else {
			//不存在数据
			receivingGifts.Log = append(receivingGifts.Log,
				entity.RecGiftLog{
					AddTime:         req.NowMin,
					GiveNum:         uint32(1),
					PopularityValue: req.PopularityValue,
					IdLog:           []entity.GiftLog{{GiftID: req.GiftsId, Number: req.Number}},
				})
		}

		toPlayer.ReceivingGifts[indexr] = *receivingGifts
	}

	toPlayer.PopularityValue += req.PopularityValue
	toPlayer.FlagChang()
	game_log.SaveProductionLog(req.Uuid, req.ToEntityID, conf.PropertyItem, uint32(0), conf.Popularity, conf.RES_TYPE_INCR, uint64(req.Number), toPlayer.PopularityValue, req.SysID, req.ActionID)
}
