package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"reflect"
	"strconv"
	"time"
)

var ShopMgr _ShopMgr

type _ShopMgr struct {
	shop *Shop
}

func (s *_ShopMgr) Init() {
	s.shop = NewShop()

	//注册逻辑业务事件
	event.OnNet(gmsg.MsgTile_Shop_GetShopListRequest, reflect.ValueOf(s.OnGetShopListRequest))
	event.OnNet(gmsg.MsgTile_Shop_BuyItemRequest, reflect.ValueOf(s.OnShopBuyItemRequest))
}

// OnGetShopListRequest 获取商城商品列表
func (s *_ShopMgr) OnGetShopListRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetShopListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ShopMgr--OnGetShopListRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	list, err := s.shop.GetShopList()
	if err != nil {
		log.Waring("-->logic--_ShopMgr--OnGetShopListRequest--s.shop.GetShopList err:", err)
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	//开始初始化桌面信息
	resp := &gmsg.GetShopListResponse{}
	resp.Code = resp_code.CODE_SUCCESS

	if len(list) > 0 {
		for _, v := range list {

			switch v.TypeN {
			case consts.Cue:
				isHave := Backpack.CheckIsHaveCueByTableID(tEntityPlayer, v.ItemID)
				if isHave {
					continue
				}
			case consts.Dress, consts.Effect, consts.Clothing:
				isHave := Backpack.CheckIsHaveItemByTableID(tEntityPlayer, v.ItemID)
				if isHave {
					continue
				}
			}

			info := &gmsg.ShopInfo{
				ID:            v.TableID,
				ShowStartTime: v.ShowStartTime,
				ShowEndTime:   v.ShowEndTime,
				Discount:      v.Discount,
			}
			resp.List = append(resp.List, info)
		}
	}

	log.Info("-->logic--_ShopMgr--OnGetShopListRequest--Resp--", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_GetShopListResponse, resp, []uint32{req.EntityID})
	return
}

// 道具商城购买物品，前端->游戏服->DB服
func (s *_ShopMgr) OnShopBuyItemRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ShopBuyItemRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	//购买返回消息
	resp := &gmsg.ShopBuyItemResponse{}

	//校验商品是否存在
	if req.ShopItemID <= 0 {
		return
	}
	shopItemInfo, err := s.shop.GetItemByID(strconv.Itoa(int(req.ShopItemID)))
	if err != nil {
		resp.Code = 3
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	now := time.Now().Unix()
	//是否在展示时间内
	if shopItemInfo.ShowStartTime > 0 && shopItemInfo.ShowEndTime > 0 {
		if shopItemInfo.ShowStartTime <= now || shopItemInfo.ShowEndTime >= now {
			resp.Code = 4
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
			return
		}
	}

	//检查背包是否存在永久的物品
	if shopItemInfo.TypeN == consts.Cue || shopItemInfo.TypeN == consts.Dress || shopItemInfo.TypeN == consts.Effect || shopItemInfo.TypeN == consts.Clothing {
		tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
		if tEntity == nil {
			resp.Code = 5
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
			return
		}

		tEntityPlayer := tEntity.(*entity.EntityPlayer)

		switch shopItemInfo.TypeN {
		case consts.Cue:
			isHave := Backpack.CheckIsHaveCueByTableID(tEntityPlayer, shopItemInfo.ItemID)
			if isHave {
				resp.Code = 5
				ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
				return
			}
		case consts.Dress, consts.Effect, consts.Clothing:
			isHave := Backpack.CheckIsHaveItemByTableID(tEntityPlayer, shopItemInfo.ItemID)
			if isHave {
				resp.Code = 5
				ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
				return
			}
		}
	}

	rewardEntity := new(entity.RewardEntity)
	rewardEntity.ItemTableId = shopItemInfo.ItemID
	rewardEntity.Num = req.Num
	rewardEntity.ExpireTimeId = req.ExpireTimeId

	payGold := rewardEntity.Num * shopItemInfo.Price

	//TODO：折扣以后加上
	//扣钱(必须先扣费)
	if !s.checkPlayerGold(req.EntityID, shopItemInfo.TokenType, payGold) {
		resp.Code = 2
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	//TODO：折扣以后加上
	resParam := GetResParam(consts.SYSTEM_ID_SHOP, consts.Buy)
	//发放物品
	addItemErr, _, _ := Backpack.BackpackAddOneItemAndSave(req.EntityID, *rewardEntity, *resParam)
	if addItemErr != nil {
		log.Waring("-->logic--_ShopMgr--OnShopBuyItemRequest--BackpackAddOneItemAndSave--err", addItemErr)
		resp.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	Player.UpdatePlayerPropertyItem(req.EntityID, shopItemInfo.TokenType, int32(-payGold), *resParam)

	//TODO:如果允许创建订单失败丢失日志则直接发一条消息即可，反之
	//创建订单
	//ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Login_EnterGameRequest, msgBody, network.ServerType_DB)

	resp.Code = 0
	resp.ShopItemID = req.ShopItemID
	msgRewardEntity := new(gmsg.RewardInfo)
	stack.SimpleCopyProperties(msgRewardEntity, rewardEntity)
	resp.RewardInfo = msgRewardEntity
	ConditionalMr.SyncConditional(req.EntityID, []consts.ConditionData{{consts.BuyCommodityTimes, 1, false}})
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
}

func (s *_ShopMgr) checkPlayerGold(EntityID, tokenType, payGold uint32) (isDeduct bool) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if tokenType == consts.Gold {
		isDeduct = tEntityPlayer.NumGold >= payGold
	} else if tokenType == consts.Diamond {
		isDeduct = tEntityPlayer.NumStone >= payGold
	} else if tokenType == consts.ClubGold {
		isDeduct = tEntityPlayer.ClubNumGold >= payGold
	} else if tokenType == consts.Exchange {
		isDeduct = tEntityPlayer.ExchangeGold >= payGold
	} else if tokenType == consts.ShopScore {
		isDeduct = tEntityPlayer.ShopScore >= payGold
	}

	return
}
