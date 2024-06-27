package logic

import (
	"BilliardServer/DBServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/db/collection"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
	"time"
)

type _PointsShop struct{}

var PointsShop _PointsShop

func (s *_PointsShop) Init() {
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPointsShopToGameRequest), reflect.ValueOf(s.SyncPointsShopListToGame))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PointsShopRedeemedNumToDbSync), reflect.ValueOf(s.SyncPointsShopRedeemedNum))

}

func (s *_PointsShop) SyncPointsShopListToGame(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InPointsShopToDbRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	coll := new(collection.PointsMallData)
	coll.SetDBConnect(consts.COLLECTION_POINTSMALL)
	list := coll.GetDataOfQuery(DBConnect)

	respList := make([]*gmsg.InPointsShopInfo, 0)
	if len(list) > 0 {
		now := time.Now().Unix()
		for _, v := range list {
			if v.EndTime < now {
				continue
			}

			rewardList := make([]*gmsg.InRewardInfo, 0)
			if len(v.RewardList) <= 0 {
				continue
			}

			for _, rv := range v.RewardList {
				info := gmsg.InRewardInfo{
					ItemTableId:  rv.ItemTableId,
					Num:          rv.Num,
					ExpireTimeId: rv.ExpireTimeId,
				}
				rewardList = append(rewardList, &info)
			}

			respList = append(respList, &gmsg.InPointsShopInfo{
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
			})
		}
	}

	log.Info("-->logic--_PointsShop--SyncPointsShopListToGame--Resp:", respList)

	resp := &gmsg.InPointsShopList{
		List: respList,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPointsShopToGameResponse), resp, network.ServerType_Game)
	return
}

func (s *_PointsShop) SyncPointsShopRedeemedNum(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InPointsShopRedeemedNumToDbSync{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	if req.PointsMallId == "" {
		return
	}

	coll := new(collection.PointsMallData)
	coll.SetDBConnect(consts.COLLECTION_POINTSMALL)
	coll.InitFormDB(req.PointsMallId, DBConnect)

	coll.RedeemedNum = req.RedeemedNum
	_ = coll.Save(DBConnect)

	return
}
