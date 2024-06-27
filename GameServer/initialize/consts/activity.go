package consts

// 活动模版类型
type ActivityType uint32

const (
	ActivityTplType_Pay        ActivityType = iota + 1 //付费充值活动
	ActivityTplType_Battle                             //对局活动
	ActivityTplType_Turntable                          //转盘活动
	ActivityTplType_PayLottery                         //付费抽奖活动
	ActivityTplType_KingRode                           //王者之路（唯一）
	ActivityTplType_Navigation                         //航海探险（唯一）
	ActivityTplType_Festival                           //节日主题活动（唯一）
)

// 活动重置类型
type ResetType uint32

const (
	ResetType_None ResetType = iota + 1 //不重置
	ResetType_Day                       //每天
	ResetType_Week                      //每周
)

const (
	TimeType uint32 = iota //默认
	TimeType_Month
	TimeType_Week
)
