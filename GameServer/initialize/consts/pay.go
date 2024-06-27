package consts

type OrderState int

const (
	OrderState_Normal               OrderState = iota //正常订单，但是可能写入对应数据库失败
	OrderState_2IntMoneyFail                          //转换整型金额失败
	OrderState_FailOrder                              //失败订单
	OrderState_SignCheckFail                          //签名校验失败
	OrderState_GameServerError                        //游戏服获取失败
	OrderState_Write2GameServerFail                   //写入游戏服失败
	OrderState_RepeatOrder                            //重复订单（订单ID一样）
	OrderState_Error                                  //其他错误
)

// 订单类型
type OrderType int

const (
	OrderType_Normal           OrderType = iota //普通订单
	OrderType_MonCard                           //月卡
	OrderType_SuperGift                         //超值礼包
	OrderType_GuildMonCard                      //公公月卡
	OrderType_ThirdPart                         //第三方充值
	OrderType_ThirdPartMonCard                  //第三方充值月卡
	OrderType_LevelGift                         //等级礼包
	OrderType_GuideBook                         //新手手册
	OrderType_TrafficPermit                     //通行证
	OrderType_Activity                          //活动充值
)

// 充值金额类型
type MoneyType int

const (
	MoneyType_RMB MoneyType = iota
	MoneyType_Dollar
	MoneyType_THB
)

// 订单状态
type OrderStatus int

const (
	OrderStatus_GotGoods    = 0 //已发货
	OrderStatus_NotGetGoods = 1 //未发货
	OrderStatus_NotPay      = 2 //订单创建未支付
)
