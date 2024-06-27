package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"reflect"
	"sort"
	"time"
)

/***
 *@disc: 特殊商店
 *@author: lsj
 *@date: 2023/11/27
 */

type _SpecialShop struct {
	List []*gmsg.SpecialShopInfo
}

var SpecialShopMr _SpecialShop

func (c *_SpecialShop) Init() {
	c.List = make([]*gmsg.SpecialShopInfo, 0)
	c.SetSpecialShopList()
	event.OnNet(gmsg.MsgTile_Hall_SpecialShopListRequest, reflect.ValueOf(c.OnSpecialShopListRequest))
	event.OnNet(gmsg.MsgTile_Hall_BuySpecialShopRequest, reflect.ValueOf(c.OnBuySpecialShopRequest))
}

func (c *_SpecialShop) SetSpecialShopList() {
	for _, val := range Table.GetAllSpecialShopCfg() {
		shopInfo := new(gmsg.SpecialShopInfo)
		stack.SimpleCopyProperties(shopInfo, val)
		shopInfo.Key = val.TableID
		if len(val.Item) > 0 {
			shopInfo.TableId = val.Item[0]
			shopInfo.Num = val.Item[1]
		}
		if len(val.GiftNum) > 0 {
			shopInfo.GiftNum += val.GiftNum[1]
		}
		c.List = append(c.List, shopInfo)
	}
	sort.Slice(c.List, func(i, j int) bool {
		return c.List[i].Key < c.List[j].Key
	})
}

// 特殊商店列表
func (c *_SpecialShop) OnSpecialShopListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SpecialShopListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.SpecialShopListResponse{}
	msgResponse.ShopList = make([]*gmsg.SpecialShopInfo, 0)
	msgResponse.ShopList = c.List
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_SpecialShopListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 购买特殊商品
func (c *_SpecialShop) OnBuySpecialShopRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.BuySpecialShopRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse := &gmsg.BuySpecialShopResponse{}
	msgResponse.Code = uint32(1)
	specialCfg := Table.GetSpecialShopCfg(msgBody.Key)
	if specialCfg == nil || len(specialCfg.Item) == 0 {
		log.Error("--->special shop cfg is err")
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BuySpecialShopResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if !c.checkPlayerGold(tEntityPlayer, specialCfg.PayType, specialCfg.Price) {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BuySpecialShopResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	msgResponse.Code = uint32(0)
	rewardEntity := new(entity.RewardEntity)

	rewardEntity.ItemTableId = specialCfg.Item[0]
	rewardEntity.Num = specialCfg.Item[1]
	if len(specialCfg.GiftNum) > 0 {
		rewardEntity.Num += specialCfg.GiftNum[1]
	}
	rewardEntity.ExpireTimeId = 0
	resParam := GetResParam(consts.SYSTEM_ID_SPECIAL_SHOP, consts.Buy)
	BeforeRecharge := tEntityPlayer.NumStone
	//发放物品
	addItemErr, _ := Backpack.BackpackAddItemListAndSave(msgBody.EntityID, []entity.RewardEntity{*rewardEntity}, *resParam)
	if addItemErr != nil {
		log.Error("-->addItemErr:", addItemErr)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BuySpecialShopResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if specialCfg.PayType == uint32(0) {
		inRewardEntity := new(gmsg.InRewardInfo)
		stack.SimpleCopyProperties(inRewardEntity, rewardEntity)
		//充值钻石
		ClubManager.UpdateConsumeTask(msgBody.EntityID, rewardEntity.Num, tEntityPlayer.GetClubID(), *resParam)
		Player.UpdatePlayerPropertyItem(tEntityPlayer.EntityID, consts.VipLvExp, int32(rewardEntity.Num), *resParam)
		SendRechargeLogToDb(msgBody.EntityID, consts.SpecialShop, uint32(time.Now().Unix()), BeforeRecharge, tEntityPlayer.NumStone, float32(specialCfg.Price),
			0, 0, 0, 0, time.Now().Unix(), time.Now().Unix(), consts.DefaultRecharge, []*gmsg.InRewardInfo{inRewardEntity})
		ConditionalMr.SyncConditionalRecharge(msgBody.EntityID, specialCfg.Price)
	} else if specialCfg.PayType == consts.Diamond {
		Player.UpdatePlayerPropertyItem(tEntityPlayer.EntityID, consts.Diamond, int32(-specialCfg.Price), *resParam)
		//消耗钻石
		ClubManager.UpdateConsumeTask(msgBody.EntityID, specialCfg.Price, tEntityPlayer.GetClubID(), *resParam)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BuySpecialShopResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_SpecialShop) checkPlayerGold(tEntityPlayer *entity.EntityPlayer, payType, payGold uint32) (isDeduct bool) {
	if payType == consts.Diamond {
		isDeduct = tEntityPlayer.NumStone >= payGold
	} else if payType == uint32(0) {
		//todo 支付以后接
		isDeduct = true
	}

	return
}
