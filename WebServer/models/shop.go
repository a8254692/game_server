package models

import (
	"BilliardServer/WebServer/utils"
	"github.com/beego/beego/v2/core/logs"
	"gopkg.in/mgo.v2/bson"
)

type ShopOrder struct {
	ID       bson.ObjectId `bson:"_id"`      //唯一标识
	OrderId  string        `bson:"OrderId"`  //订单id
	EntityID uint32        `bson:"EntityID"` //帐号ID
	ItemId   uint32        `bson:"ItemId"`   //物品id
	Price    float64       `bson:"price"`    //价格
}

func CreateShopOrder(o *ShopOrder) {
	if o == nil {
		return
	}

	o.ID = bson.NewObjectId()
	err := utils.LogDB.InsertData("order", o)
	if err != nil {
		logs.Warning("-->models--AddCreateAccountLog--Error:", err)
		return
	}

	return
}
