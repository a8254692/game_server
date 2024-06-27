package vars

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
)

// 活动条件和奖池
type ConditionAndReward struct {
	No         uint32                `json:"no"`          //编号
	ValueList  uint32                `json:"value_list"`  //条件值数组
	RewardList []entity.RewardEntity `json:"reward_list"` //奖品
	TabName    string                `json:"tab_name"`    //选项卡名字
}

// 付费抽奖奖池
type PayLotteryReward struct {
	No          uint32              `json:"no"`           //编号
	Probability uint32              `json:"probability"`  //概率
	Reward      entity.RewardEntity `json:"reward"`       //奖品
	IsGuarantee bool                `json:"is_guarantee"` //是否为保底物品（只能存在一个）
}

// 转盘奖池
type TurntableReward struct {
	No          uint32              `json:"no"`          //编号
	Probability uint32              `json:"probability"` //概率
	Reward      entity.RewardEntity `json:"reward"`      //奖品
}

// 对战活动配置
type ActivityConfigBattle struct {
	BattleType             uint32               `json:"battle_type"`
	OutcomeType            uint32               `json:"outcome_type"`
	ConditionAndRewardList []ConditionAndReward `json:"condition_and_reward_list"` //活动奖励和条件
}

// 支付活动配置
type ActivityConfigPay struct {
	PayType                uint32               `json:"pay_type"`
	ConditionAndRewardList []ConditionAndReward `json:"condition_and_reward_list"` //活动奖励和条件
}

// 付费抽奖活动配置（幸运值为大保底进度）
type ActivityConfigPayLottery struct {
	ConsumeItemID  uint32 `json:"consume_item_id"`   //消耗道具id消耗道具id
	ConsumeItemNum uint32 `json:"consume_item_num"`  //消耗道具id数量
	DayFreeDrawNum uint32 `json:"day_free_draw_num"` //每日免费抽奖次数

	IsOpenLucky    bool             `json:"is_open_lucky"`    //是否开启幸运值逻辑
	TakeLuckyNum   uint32           `json:"take_lucky_num"`   //获取幸运值配置-抽数"
	LuckyNum       uint32           `json:"lucky_num"`        //获取幸运值配置-获得幸运值"
	LuckyResetType consts.ResetType `json:"lucky_reset_type"` //幸运值重置类型
	MaxLuckyNum    uint32           `json:"max_lucky_num"`    //幸运值最大数（保底数）

	IsOpenDrawNum     bool                 `json:"is_open_draw_num"`     //是否开启抽奖次数
	DrawResetType     consts.ResetType     `json:"draw_reset_type"`      //抽奖次数重置类型
	DrawNumRewardList []ConditionAndReward `json:"draw_num_reward_list"` //抽奖次数奖励列表

	IsOpenExchange        bool                 `json:"is_open_exchange"`         //是否开启兑换列表
	ExchangeConsumeItemID uint32               `json:"exchange_consume_item_id"` //兑换消耗道具id
	ExchangeRewardList    []ConditionAndReward `json:"exchange_reward_list"`     //抽奖次数奖励列表

	PayLotteryRewardList []PayLotteryReward `json:"pay_lottery_reward_list"` //活动奖励和概率
}

// 转盘活动配置
type ActivityConfigTurntable struct {
	FreeDrawNum         uint32            `json:"free_draw_num"`         //免费抽奖次数
	TotalDrawNum        uint32            `json:"total_draw_num"`        //可使用总抽奖次数（含免费抽奖次数）
	DrawNumConfig       uint32            `json:"draw_num_config"`       //对局N次获取抽奖
	TurntableRewardList []TurntableReward `json:"turntable_reward_list"` //活动奖励和概率
}

// 后台活动数据结构，存在每个服务器数据库中
type ActivityData struct {
	ActivityId       string                   `json:"activity_id"`        //活动唯一ID
	TimeType         uint32                   `json:"time_type"`          //时间类型
	StartTime        int64                    `json:"start_time"`         //开始时间
	EndTime          int64                    `json:"end_time"`           //结束时间
	AType            consts.ActivityType      `json:"a_type"`             //活动类型
	SubType          uint32                   `json:"sub_type"`           //活动主题类型
	ActivityName     string                   `json:"activity_name"`      //活动名称
	PlatformLimit    uint32                   `json:"platform_limit"`     //平台限制
	VipLimit         uint32                   `json:"vip_limit"`          //vip等级限制
	IsRelease        bool                     `json:"is_release"`         //是否发布
	ConfigTurntable  ActivityConfigTurntable  `json:"config_turntable"`   //转盘活动配置
	ConfigPayLottery ActivityConfigPayLottery `json:"config_pay_lottery"` //付费抽奖活动配置
	ConfigPay        ActivityConfigPay        `json:"config_pay"`         //支付活动配置
	ConfigBattle     ActivityConfigBattle     `json:"config_battle"`      //对战活动配置
}

type LoginNotice struct {
	ObjId         string `json:"id"`                                   //唯一ID
	LoginNoticeId string `bson:"LoginNoticeId" json:"login_notice_id"` //登录公告唯一ID
	Title         string `bson:"Title" json:"title"`                   //标题
	Name          string `bson:"Name" json:"name"`                     //登录公告名称
	Context       string `bson:"Context" json:"context"`               //登录公告内容
	StartTime     int64  `bson:"StartTime" json:"start_time"`          //开始时间
	EndTime       int64  `bson:"EndTime" json:"end_time"`              //结束时间
	PlatformLimit uint32 `bson:"PlatformLimit" json:"platform_limit"`  //平台限制
	VipLimit      uint32 `bson:"VipLimit" json:"vip_limit"`            //vip等级限制
	IsRelease     bool   `bson:"IsRelease" json:"is_release"`          //是否暂停
}
