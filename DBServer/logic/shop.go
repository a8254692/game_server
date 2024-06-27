package logic

import (
	"BilliardServer/Common/resp_code"
	"BilliardServer/DBServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/db/collection"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
	"reflect"
)

type _Shop struct {
}

var Shop _Shop

func (s *_Shop) Init() {
	//注册逻辑业务事件
	//event.On("Msg_MultiNinjaPointWarEnemyTeam", reflect.ValueOf(TeamRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Shop_Create_Order_Request), reflect.ValueOf(s.OnShopCreateOrderRequest))
}

// 创建订单 DB服->游戏服
func (s *_Shop) OnShopCreateOrderRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ShopCreateOrderRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	order := new(collection.Order)
	order.InitByFirst(consts.COLLECTION_ORDER, req.OrderSn)
	err = order.InitByData(req)
	if err != nil {
		return
	}
	err = order.Insert(DBConnect)
	if err != nil {
		return
	}

	resp := &gmsg.ShopCreateOrderResponse{
		Code:    resp_code.CODE_SUCCESS,
		OrderSn: req.OrderSn,
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Shop_Create_Order_Response), resp, network.ServerType_Game)
}
