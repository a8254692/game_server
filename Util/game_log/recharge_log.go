package game_log

import (
	gmsg "BilliardServer/Proto/gmsg"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// 玩家充值
type RechargeLog struct {
	ID              bson.ObjectId        `bson:"_id"`             //唯一标识
	EntityID        uint32               `bson:"EntityID"`        //账号
	Time            int64                `bson:"Time"`            //日志时间
	Channel         uint32               `bson:"Channel"`         //充值渠道
	OrderId         uint32               `bson:"OrderId"`         //订单号
	BeforeRecharge  uint32               `bson:"BeforeRecharge"`  //充值前钻石
	AfterRecharge   uint32               `bson:"AfterRecharge"`   //充值后钻石
	OrderAmount     float32              `bson:"OrderAmount"`     //订单金额
	Discount        float32              `bson:"Discount"`        //折扣信息
	EventGifts      float32              `bson:"EventGifts"`      //活动赠送
	Deduction       float32              `bson:"Deduction"`       //扣款
	ActualReceipt   float32              `bson:"ActualReceipt"`   //实际到账
	CreateOrderTime int64                `bson:"CreateOrderTime"` //下单时间
	PayTime         int64                `bson:"PayTime"`         //支付时间
	TypeN           uint32               `bson:"TypeN"`           //支付类型 0普通充值，1首充，2活动充值
	RewardItem      []*gmsg.InRewardInfo `bson:"RewardItem"`      //物品列表
}

// 记录玩家充值日志
func SaveRechargeLog(entityID uint32, channel uint32, orderId uint32, beforeRecharge uint32, afterRecharge uint32, orderAmount float32, discount float32, eventGifts float32, deduction float32, actualReceipt float32, createOrderTime int64, payTime int64, rewardItems []*gmsg.InRewardInfo) {
	resLog := &RechargeLog{
		ID:              bson.NewObjectId(),
		EntityID:        entityID,
		Time:            time.Now().Unix(),
		Channel:         channel,
		OrderId:         orderId,
		BeforeRecharge:  beforeRecharge,
		AfterRecharge:   afterRecharge,
		OrderAmount:     orderAmount,
		Discount:        discount,
		EventGifts:      eventGifts,
		Deduction:       deduction,
		ActualReceipt:   actualReceipt,
		CreateOrderTime: createOrderTime,
		PayTime:         payTime,
		RewardItem:      rewardItems,
	}

	GGameLogManager.AddLog(GameLogType_RechargeLog, resLog)
	return
}
