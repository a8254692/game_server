package vars

import (
	"BilliardServer/GameServer/initialize/consts"
	"gopkg.in/mgo.v2/bson"
)

var (
	SHOP_ITEM_TYPE_MAP = map[uint32]string{
		consts.ITEM_TYPE_UNKNOWN: "WZ",
		consts.ITEM_TYPE_GOLD:    "JB",
		consts.ITEM_TYPE_DIAMOND: "ZS",
		consts.ITEM_TYPE_VIP:     "VP",
		consts.ITEM_TYPE_CUE:     "QG",
		consts.ITEM_TYPE_DRESS:   "FS",
		consts.ITEM_TYPE_EFFECT:  "XG",
		consts.ITEM_TYPE_PROP:    "DJ",
	}
)

// 充值回调数据
type RechargeData struct {
	ObjID                  bson.ObjectId `bson:"_id"`                    //唯一ID
	OrderID                string        `bson:"OrderID"`                //sdk订单ID
	GMAccount              string        `bson:"GMAccount"`              //gm发奖帐号，如果不是gm发奖忽略
	OrderState             int           `bson:"OrderState"`             //订单状态,对应enOrderState枚举
	Account                string        `bson:"Account"`                //玩家帐号
	RecipientObjId         string        `bson:"RecipientObjId"`         //受赠者唯一id
	Success                string        `bson:"Success"`                //是否成功 1：成功，其他失败
	CustomOrderID          string        `bson:"CustomOrderID"`          //玩家订单ID ，后缀是_策划订单ID
	SendDate               string        `bson:"SendDate"`               //发送时间
	AddTime                string        `bson:"AddTime"`                //创建时间
	ProductId              string        `bson:"ProductId"`              //计费点ID（海外没有传金额，只传回计费点信息）
	MoneyType              int           `bson:"MoneyType"`              //金额类型，0为RMB,1为Dollar
	Money                  int           `bson:"Money"`                  //金额
	Dollar                 float64       `bson:"Dollar"`                 //美元金额
	Sign                   string        `bson:"Sign"`                   //签名信息
	PayType                string        `bson:"PayType"`                //充值类型
	ChannelId              int           `bson:"ChannelId"`              //渠道Id,对应Common.ChannelType枚举
	BundleID               string        `bson:"BundleID"`               //包名
	GameServerID           string        `bson:"GameServerID"`           //游戏服务器id
	RequestIp              string        `bson:"RequestIp"`              //请求iP
	DiamondBeforeRecharge  int           `bson:"DiamondBeforeRecharge"`  //充值前元宝
	Diamond                int           `bson:"Diamond"`                //充值获得元宝
	VipLevelBeforeRecharge int           `bson:"VipLevelBeforeRecharge"` //充值前vip等级
	VipLevelAfterRecharge  int           `bson:"VipLevelAfterRecharge"`  //充值后Vip等级
	VipExpBeforeRecharge   int           `bson:"VipExpBeforeRecharge"`   //充值前vip经验
	VipExpAfterRecharge    int           `bson:"VipExpAfterRecharge"`    //充值后Vip经验
	IsRead                 int           `bson:"IsRead"`                 //订单状态 OrderStatus
	Platform               int           `bson:"Platform"`
	ExtraParams            string        `bson:"ExtraParams"` //透传参数
}
