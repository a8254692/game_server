package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/GameServer/initialize/vars"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
	"time"
)

type _PointsShop struct {
	List []vars.PointsShopData
}

var PointsShop _PointsShop

func (s *_PointsShop) Init() {
	s.List = make([]vars.PointsShopData, 0)

	event.OnNet(gmsg.MsgTile_Shop_PointsShopBuyItemRequest, reflect.ValueOf(s.OnPointsShopBuyItemRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPointsShopToGameResponse), reflect.ValueOf(s.SyncPointsShopListFromDb))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PointsShopOtherToGameSync), reflect.ValueOf(s.AdminChangePointsShopList))

	time.AfterFunc(time.Millisecond*1000, s.SyncPointsShopListToDb)
}

func (s *_PointsShop) SyncPointsShopListFromDb(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InPointsShopList{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_PointsShop--SyncPointsShopListFromDb--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_PointsShop--SyncPointsShopListFromDb--req--", req)

	respList := make([]vars.PointsShopData, 0)
	if len(req.List) > 0 {
		for _, v := range req.List {

			rewardList := make([]entity.RewardEntity, 0)
			if len(v.RewardList) > 0 {
				for _, rv := range v.RewardList {
					info := entity.RewardEntity{
						ItemTableId:  rv.ItemTableId,
						Num:          rv.Num,
						ExpireTimeId: rv.ExpireTimeId,
					}
					rewardList = append(rewardList, info)
				}
			}

			info := vars.PointsShopData{
				PointsMallId:         v.PointsMallId,
				Name:                 v.Name,
				StartTime:            v.StartTime,
				EndTime:              v.EndTime,
				RewardList:           rewardList,
				Resources:            v.Resources,
				LimitNum:             v.LimitNum,
				ExchangeAmount:       v.ExchangeAmount,
				ExchangeCurrencyType: v.ExchangeCurrencyType,
				ExchangeMaxNum:       v.ExchangeMaxNum,
				RedeemedNum:          v.RedeemedNum,
			}

			respList = append(respList, info)
		}
	}

	s.List = respList
	return
}

func (s *_PointsShop) AdminChangePointsShopList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InAdminPointsShopListSync{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_PointsShop--AdminChangePointsShopList--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_PointsShop--AdminChangePointsShopList--req", req)

	toReq := &gmsg.InPointsShopToDbRequest{}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPointsShopToGameRequest), toReq, network.ServerType_DB)
	return
}

// 道具商城购买物品，前端->游戏服->DB服
func (s *_PointsShop) OnPointsShopBuyItemRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.PointsShopBuyItemRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	log.Info("-->logic--_PointsShop--OnPointsShopBuyItemRequest--req", req)

	//校验参数
	if req.EntityID <= 0 || req.Num <= 0 || req.PointsMallId == "" || len(s.List) <= 0 {
		return
	}

	//购买返回消息
	resp := &gmsg.PointsShopBuyItemResponse{}

	var itemInfo vars.PointsShopData
	for _, v := range s.List {
		if v.PointsMallId == req.PointsMallId {
			itemInfo = v
		}
	}

	if itemInfo.PointsMallId == "" {
		resp.Code = 3
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	now := time.Now().Unix()
	//是否在展示时间内
	if now < itemInfo.StartTime && now > itemInfo.EndTime {
		resp.Code = 4
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	//检查背包是否存在永久的物品
	//if shopItemInfo.TypeN == consts.Cue || shopItemInfo.TypeN == consts.Dress || shopItemInfo.TypeN == consts.Effect || shopItemInfo.TypeN == consts.Clothing {
	//	tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
	//	if tEntity == nil {
	//		resp.Code = 5
	//		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
	//		return
	//	}
	//
	//	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	//
	//	switch shopItemInfo.TypeN {
	//	case consts.Cue:
	//		isHave := Backpack.CheckIsHaveCueByTableID(tEntityPlayer, shopItemInfo.ItemID)
	//		if isHave {
	//			resp.Code = 5
	//			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
	//			return
	//		}
	//	case consts.Dress, consts.Effect, consts.Clothing:
	//		isHave := Backpack.CheckIsHaveItemByTableID(tEntityPlayer, shopItemInfo.ItemID)
	//		if isHave {
	//			resp.Code = 5
	//			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_BuyItemResponse, resp, []uint32{req.EntityID})
	//			return
	//		}
	//	}
	//}

	if len(itemInfo.RewardList) <= 0 {
		resp.Code = 3
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	if itemInfo.LimitNum > 0 {
		if itemInfo.LimitNum < itemInfo.RedeemedNum+req.Num {
			resp.Code = 5
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
			return
		}
	}

	if itemInfo.ExchangeMaxNum > 0 {
		tEntityPlayer, err := GetEntityPlayerById(req.EntityID)
		if err != nil {
			log.Waring("-->logic--_PointsShop--OnPointsShopBuyItemRequest--GetEntityPlayerById--err--", err)
			return
		}
		var isBuyLimit bool
		if len(tEntityPlayer.PointsShopBuyList) > 0 {
			for _, pv := range tEntityPlayer.PointsShopBuyList {
				if pv.PointsMallId == req.PointsMallId && pv.Num+req.Num > itemInfo.ExchangeMaxNum {
					isBuyLimit = true
				}
			}
		}
		if isBuyLimit {
			resp.Code = 5
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
			return
		}
	}

	//TODO：折扣以后加上
	//扣钱(必须先扣费)
	if !s.checkPlayerGold(req.EntityID, itemInfo.ExchangeCurrencyType, itemInfo.ExchangeAmount*req.Num) {
		resp.Code = 2
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	//增加总数限制
	var totalRedeemedNum uint32
	if itemInfo.LimitNum > 0 {
		for lk, lv := range s.List {
			if lv.PointsMallId == req.PointsMallId {
				s.List[lk].RedeemedNum += req.Num

				totalRedeemedNum = s.List[lk].RedeemedNum
			}
		}
	}

	var userRedeemedNum uint32
	//增加个人购买数量限制
	if itemInfo.ExchangeMaxNum > 0 {
		tEntityPlayer, err := GetEntityPlayerById(req.EntityID)
		if err != nil {
			log.Waring("-->logic--_PointsShop--OnPointsShopBuyItemRequest--GetEntityPlayerById--err--", err)
			return
		}
		var isIn bool
		if len(tEntityPlayer.PointsShopBuyList) > 0 {
			for p1k, p1v := range tEntityPlayer.PointsShopBuyList {
				if p1v.PointsMallId == req.PointsMallId {
					isIn = true

					tEntityPlayer.PointsShopBuyList[p1k].Num += req.Num

					userRedeemedNum = tEntityPlayer.PointsShopBuyList[p1k].Num
				}
			}
		}
		if !isIn {
			tEntityPlayer.PointsShopBuyList = append(tEntityPlayer.PointsShopBuyList, entity.PointsShopBuy{
				PointsMallId: req.PointsMallId,
				Num:          req.Num,
			})

			userRedeemedNum = req.Num
		}

		tEntityPlayer.SyncEntity(1)
	}

	resParam := GetResParam(consts.SYSTEM_ID_SHOP, consts.Buy)
	rewardList := make([]entity.RewardEntity, 0)
	for _, rv := range itemInfo.RewardList {
		rewardList = append(rewardList, entity.RewardEntity{
			ItemTableId:  rv.ItemTableId,
			Num:          rv.Num * req.Num,
			ExpireTimeId: rv.ExpireTimeId,
		})
	}

	//发放物品
	rList := RewardManager.AddReward(req.EntityID, rewardList, *resParam)
	if len(rList) <= 0 {
		log.Waring("-->logic--_PointsShop--OnPointsShopBuyItemRequest--RewardManager.AddReward--err")
		resp.Code = 1
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
		return
	}

	Player.UpdatePlayerPropertyItem(req.EntityID, itemInfo.ExchangeCurrencyType, int32(-itemInfo.ExchangeAmount*req.Num), *resParam)

	//TODO:如果允许创建订单失败丢失日志则直接发一条消息即可，反之
	//创建订单
	//ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Login_EnterGameRequest, msgBody, network.ServerType_DB)

	if itemInfo.LimitNum > 0 {
		broadCastResp := &gmsg.PointsShopBuyItemBroadCast{
			TotalRedeemedNum: totalRedeemedNum,
		}
		ConnectManager.SendMsgPbToGateBroadCastAll(gmsg.MsgTile_Shop_PointsShopBuyItemBroadCast, broadCastResp)
	}

	resp.Code = 0
	resp.PointsMallId = req.PointsMallId
	resp.UserRedeemedNum = userRedeemedNum
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Shop_PointsShopBuyItemResponse, resp, []uint32{req.EntityID})
}

func (s *_PointsShop) SyncPointsShopRedeemedNumToDb(pointsMallId string, num uint32) {
	//开始初始化桌面信息
	req := &gmsg.InPointsShopRedeemedNumToDbSync{
		PointsMallId: pointsMallId,
		RedeemedNum:  num,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PointsShopRedeemedNumToDbSync), req, network.ServerType_DB)
	return
}

func (s *_PointsShop) SyncPointsShopListToDb() {
	//开始初始化桌面信息
	req := &gmsg.InPointsShopToDbRequest{}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPointsShopToGameRequest), req, network.ServerType_DB)
	return
}

func (s *_PointsShop) GetPointsShopListRequest(entityId uint32) []*gmsg.PointsShopInfo {
	resp := make([]*gmsg.PointsShopInfo, 0)

	if len(s.List) <= 0 {
		return resp
	}

	tEntityPlayer, err := GetEntityPlayerById(entityId)
	if err != nil {
		log.Waring("-->logic--_PointsShop--GetPointsShopListRequest--GetEntityPlayerById--err--", err)
		return resp
	}

	now := time.Now().Unix()
	for _, v := range s.List {
		if v.EndTime < now {
			continue
		}

		rewardList := make([]*gmsg.RewardInfo, 0)
		if len(v.RewardList) > 0 {
			for _, rv := range v.RewardList {
				info := gmsg.RewardInfo{
					ItemTableId:  rv.ItemTableId,
					Num:          rv.Num,
					ExpireTimeId: rv.ExpireTimeId,
				}
				rewardList = append(rewardList, &info)
			}
		}

		var userRedeemedNum uint32
		if v.ExchangeMaxNum > 0 {
			if len(tEntityPlayer.PointsShopBuyList) > 0 {
				for _, p1v := range tEntityPlayer.PointsShopBuyList {
					if p1v.PointsMallId == v.PointsMallId {
						userRedeemedNum = p1v.Num
					}
				}
			}
		}

		var totalRedeemedNum uint32
		if v.LimitNum > 0 {
			totalRedeemedNum = v.RedeemedNum
		}

		info := gmsg.PointsShopInfo{
			PointsMallId:         v.PointsMallId,
			Name:                 v.Name,
			StartTime:            v.StartTime,
			EndTime:              v.EndTime,
			RewardList:           rewardList,
			Resources:            v.Resources,
			LimitNum:             v.LimitNum,
			ExchangeAmount:       v.ExchangeAmount,
			ExchangeCurrencyType: v.ExchangeCurrencyType,
			ExchangeMaxNum:       v.ExchangeMaxNum,
			TotalRedeemedNum:     totalRedeemedNum,
			UserRedeemedNum:      userRedeemedNum,
		}
		resp = append(resp, &info)
	}

	log.Info("-->logic--_PointsShop--GetPointsShopListRequest--Resp:", resp)

	return resp
}

func (s *_PointsShop) checkPlayerGold(EntityID, tokenType, payGold uint32) (isDeduct bool) {
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
