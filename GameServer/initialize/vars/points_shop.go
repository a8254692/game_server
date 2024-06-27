package vars

import "BilliardServer/Common/entity"

type PointsShopData struct {
	PointsMallId         string                `bson:"PointsMallId" json:"points_mall_id"`                 //积分商城唯一ID
	Name                 string                `bson:"Name" json:"name"`                                   //名称
	StartTime            int64                 `bson:"StartTime" json:"start_time"`                        //开始时间
	EndTime              int64                 `bson:"EndTime" json:"end_time"`                            //结束时间
	RewardList           []entity.RewardEntity `bson:"RewardList" json:"reward_list"`                      //奖品
	Resources            string                `bson:"Resources" json:"resources"`                         //客户端资源地址
	LimitNum             uint32                `bson:"LimitNum" json:"limit_num"`                          //兑换总数限制
	ExchangeAmount       uint32                `bson:"ExchangeAmount" json:"exchange_amount"`              //兑换金额
	ExchangeCurrencyType uint32                `bson:"ExchangeCurrencyType" json:"exchange_currency_type"` //兑换代币类型
	ExchangeMaxNum       uint32                `bson:"ExchangeMaxNum" json:"exchange_max_num"`             //每人最大兑换数量
	RedeemedNum          uint32                `bson:"RedeemedNum" json:"redeemed_num"`                    //已兑换数量
}
