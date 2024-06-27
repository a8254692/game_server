package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/table"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
	"sort"
	"time"
)

/***
 *@disc:充值
 *@author: lsj
 *@date: 2024/1/20
 */

type _Recharge struct {
	FirstRechargeData map[uint32]FirstRecharge
	FirstRechargeCfg  map[uint32]*table.FirstRechargeCfg
}

type FirstRecharge struct {
	FirstRechargeRewardInfo []*gmsg.InRewardInfo
	FirstRechargeReward     []entity.RewardEntity
}

var RechargeMr _Recharge

func (c *_Recharge) Init() {
	c.initFirstRechargeReward()
	event.OnNet(gmsg.MsgTile_Hall_FirstRechargeRequest, reflect.ValueOf(c.OnFirstRechargeRequest))
}

func (c *_Recharge) initFirstRechargeReward() {
	c.FirstRechargeData, c.FirstRechargeCfg = make(map[uint32]FirstRecharge, 0), make(map[uint32]*table.FirstRechargeCfg, 0)
	for _, val := range Table.FirstRechargeCfg {
		inRewardInfoList, rewardList := make([]*gmsg.InRewardInfo, 0), make([]entity.RewardEntity, 0)
		for _, v := range val.RewardList {
			if len(v) < 1 {
				continue
			}
			inRewardInfo, reward := new(gmsg.InRewardInfo), new(entity.RewardEntity)
			inRewardInfo.ItemTableId = v[0]
			inRewardInfo.Num = v[1]
			inRewardInfoList = append(inRewardInfoList, inRewardInfo)
			reward.ItemTableId = v[0]
			reward.Num = v[1]
			rewardList = append(rewardList, *reward)
		}
		firstRecharge := new(FirstRecharge)
		firstRecharge.FirstRechargeRewardInfo = inRewardInfoList
		firstRecharge.FirstRechargeReward = rewardList
		c.FirstRechargeData[val.TableID] = *firstRecharge
		c.FirstRechargeCfg[val.TableID] = val
	}
	log.Info("c.FirstRechargeData", c.FirstRechargeData)
}

// 角色初始化首充列表
func (c *_Recharge) playerFirstRechargeInit(tEntityPlayer *entity.EntityPlayer) {
	if len(tEntityPlayer.FirstRecharge) > 0 {
		return
	}

	for _, val := range c.FirstRechargeCfg {
		tEntityPlayer.FirstRecharge = append(tEntityPlayer.FirstRecharge, entity.Recharge{TableID: val.TableID, IsBuy: false})
	}
	sort.Slice(tEntityPlayer.FirstRecharge, func(i, j int) bool {
		return tEntityPlayer.FirstRecharge[i].TableID < tEntityPlayer.FirstRecharge[j].TableID
	})
}

func (c *_Recharge) getPlayerIsHaveRecharge(entityID uint32) bool {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return false
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	for _, val := range tEntityPlayer.FirstRecharge {
		if !val.IsBuy {
			return false
		}
	}
	return true
}

// 首充购买
func (c *_Recharge) OnFirstRechargeRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.FirstRechargeRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	resResponse := &gmsg.FirstRechargeResponse{
		Code: uint32(1),
	}

	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--OnFirstRechargeRequest--GetEntityPlayerById--err--", err)
		return
	}

	cfg, ok := c.FirstRechargeCfg[msgBody.TableID]
	if !ok || c.isHaveFirstRecharge(msgBody.TableID, tEntityPlayer) {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_FirstRechargeResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	resParam := GetResParam(consts.Shop, consts.Buy)
	err, _ = Backpack.BackpackAddItemListAndSave(msgBody.EntityID, c.FirstRechargeData[msgBody.TableID].FirstRechargeReward, *resParam)
	if err != nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_FirstRechargeResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	c.updateFirstRecharge(msgBody.TableID, tEntityPlayer)
	SendRechargeLogToDb(msgBody.EntityID, consts.Shop, uint32(time.Now().Unix()), tEntityPlayer.NumStone, tEntityPlayer.NumStone, float32(cfg.Price),
		0, 0, 0, 0, time.Now().Unix(), time.Now().Unix(), consts.FirstRecharge, c.FirstRechargeData[msgBody.TableID].FirstRechargeRewardInfo)

	resResponse.Code = uint32(0)
	tEntityPlayer.SyncEntity(1)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_FirstRechargeResponse, resResponse, []uint32{msgBody.EntityID})
}

func (c *_Recharge) isHaveFirstRecharge(tableID uint32, tEntityPlayer *entity.EntityPlayer) bool {
	for _, val := range tEntityPlayer.FirstRecharge {
		if tableID == val.TableID && val.IsBuy == true {
			return true
		}
	}
	return false
}

func (c *_Recharge) updateFirstRecharge(tableID uint32, tEntityPlayer *entity.EntityPlayer) {
	for key, val := range tEntityPlayer.FirstRecharge {
		if tableID == val.TableID && val.IsBuy == false {
			v := val
			v.IsBuy = true
			tEntityPlayer.FirstRecharge[key] = v
			break
		}
	}
}
