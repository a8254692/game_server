package entity

import "gopkg.in/mgo.v2/bson"

// 数据实体 物品
type Item struct {
	ItemID     uint32  `bson:"ItemID"`  //ItemID
	TableID    uint32  `bson:"TableID"` //TableID
	ItemNum    uint32  `bson:"ItemNum"` //总数
	ItemType   uint32  `bson:"ItemType"`
	SubType    uint32  `bson:"SubType"`    // 小类别
	EndTime    uint32  `bson:"EndTime"`    //0表示永久
	ItemStatus uint32  `bson:"ItemStatus"` // 使用状态，0为未使用，1为使用中
	CueInfo    CueInfo `bson:"CueInfo"`    //球杆属性
}

// 球杆属性 物品
type CueInfo struct {
	Quality uint32 `bson:"Quality"` //阶级
	Star    uint32 `bson:"Star"`    //星级
}

// 背包球杆集合
type CueQualityS struct {
	QualityS   int //s品质数量
	QualitySs  int //ss品质数量
	QualitySss int //sss品质数量
}

// 数据实体 任务
type Task struct {
	TaskId           uint32 `bson:"TaskId"`           //TaskId
	ConditionId      uint32 `bson:"ConditionId"`      //条件id
	State            uint32 `bson:"State"`            //任务状态
	StateReward      uint32 `bson:"StateReward"`      //奖励状态
	CompleteProgress uint32 `bson:"CompleteProgress"` //完成的进度
	TaskProgress     uint32 `bson:"TaskProgress"`     //任务配置的进度
	Timestamp        int64  `bson:"Timestamp"`        //时间戳
}

// 进度活动的进度及领奖
type ProgressActivityStatus struct {
	Id               string `bson:"Id"`               //唯一id
	ActivityId       string `bson:"ActivityId"`       //活动id
	ConfigSerial     uint32 `bson:"ConfigSerial"`     //配置奖励序号
	TargetProgress   uint32 `bson:"TargetProgress"`   //进度目标
	CompleteProgress uint32 `bson:"CompleteProgress"` //完成的进度
	StateReward      uint32 `bson:"StateReward"`      //奖励状态
	Timestamp        int64  `bson:"Timestamp"`        //时间戳
}

// 转盘活动日对局数统记
type DayBattleNum struct {
	Num       uint32 `bson:"Num"`       //奖励状态
	Timestamp int64  `bson:"Timestamp"` //时间戳
}

// 转盘活动日领取次数
type DayReceiveStatusNum struct {
	ActivityId string `bson:"ActivityId"` //活动id
	Num        uint32 `bson:"Num"`        //奖励领取次数
	Timestamp  int64  `bson:"Timestamp"`  //时间戳
}

// 付费抽奖记录
type PayLotteryStatus struct {
	ActivityId         string            `bson:"ActivityId"`     //活动id
	FreeNum            uint32            `bson:"FreeNum"`        //免费抽奖次数
	FreeCreateTime     int64             `bson:"FreeCreateTime"` //免费抽奖时间戳
	TotalDrawNum       uint32            `bson:"TotalDrawNum"`   //总抽奖次数
	DrawEndTime        int64             `bson:"LuckEndTime"`    //抽奖次数过期时间戳
	DrawNumStatus      map[uint32]uint32 `bson:"DrawNumStatus"`  //抽奖次数领奖状态
	LuckNum            uint32            `bson:"LuckNum"`        //幸运值
	LastGetLuckDrawNum uint32            `bson:"LuckNum"`        //最后获取幸运值的抽数
	LuckEndTime        int64             `bson:"LuckEndTime"`    //幸运值过期时间戳
}

// 王者之路进度表
type KingRodeProgress struct {
	ActivityId       string           `bson:"ActivityId"`       //活动id
	ConditionalId    uint32           `bson:"ConditionalId"`    //条件id
	CompleteProgress uint32           `bson:"CompleteProgress"` //完成次数
	RewardElite      []KingRodeReward `bson:"RewardElite"`      //精英版
	RewardAdvanced   []KingRodeReward `bson:"RewardAdvanced"`   //进阶版
	IsUnlockAdvanced bool             `bson:"IsUnlockAdvanced"` //是否解锁进阶，默认未解锁
}

type KingRodeReward struct {
	RewardId       uint32 `bson:"RewardId"`
	StateReward    uint32 `bson:"StateReward"`    //0未解锁，1解锁不能领取，2可领取，3已领取
	TargetProgress uint32 `bson:"TargetProgress"` //进度目标
	AddTimestamp   int64  `bson:"AddTimestamp"`
}

// 任务重置日期
type TaskResetDate struct {
	DayDate  string `bson:"DayDate"`
	WeekDate string `bson:"WeekDate"`
}

// 赠送礼物记录
type GiveGift struct {
	EntityID        uint32    `bson:"EntityID"`        //接收人id
	LastAddTime     string    `bson:"LastAddTime"`     //最近赠送时间
	GiveNum         uint32    `bson:"GiveNum"`         //赠送次数
	PopularityValue uint32    `bson:"PopularityValue"` //人气值
	IdLog           []GiftLog `bson:"IdLog"`           //礼物列表记录
}

type GiftLog struct {
	GiftID uint32 `bson:"GiftID"`
	Number uint32 `bson:"Number"`
}

// 接收送礼物记录
type RecGift struct {
	EntityID        uint32       `bson:"EntityID"`        //赠送人id
	LastAddTime     string       `bson:"LastAddTime"`     //最近赠送时间
	GiveNum         uint32       `bson:"GiveNum"`         //赠送次数
	PopularityValue uint32       `bson:"PopularityValue"` //人气值
	Log             []RecGiftLog `bson:"Log"`             //礼物记录
}

type RecGiftLog struct {
	PopularityValue uint32    `bson:"PopularityValue"` //人气值
	AddTime         string    `bson:"AddTime"`         //赠送时间
	GiveNum         uint32    `bson:"GiveNum"`         //赠送次数
	IdLog           []GiftLog `bson:"IdLog"`           //礼物列表记录
}

// 数据实体 红点提示
type RedTips struct {
	RedType uint32 `bson:"RedType"` //RedType
	Name    string `bson:"Name"`    //名称
	State   uint32 `bson:"State"`   //State
}

type RewardEntity struct {
	ItemTableId  uint32 `bson:"ItemTableId" json:"item_table_id"`
	Num          uint32 `bson:"Num" json:"num"`
	ExpireTimeId uint32 `bson:"ExpireTimeId" json:"expire_time_id"`
}

// 数据实体 邮件
type Email struct {
	EmailID     uint32         `bson:"EmailID"`     //EmailID
	State       bool           `bson:"State"`       //邮件状态-是否已读
	StateReward bool           `bson:"StateReward"` //奖励状态-是否领取
	Date        string         `bson:"Date"`        //日期
	RewardList  []RewardEntity `bson:"RewardList"`  //物品列表
	Tittle      string         `bson:"Tittle"`      //标题
	Content     string         `bson:"Content"`     //内容
}

// 赛季历史记录
type PeakRankHist struct {
	ID          uint32 `bson:"ID"`          //赛季ID
	PeakRankLv  uint32 `bson:"PeakRankLv"`  //赛季等级
	PeakRankExp uint32 `bson:"PeakRankExp"` //赛季星数
	Status      uint32 `bson:"Status"`      //赛季结算领奖状态
	AwardTime   int64  `bson:"AwardTime"`   //赛季结算可领奖截止时间
}

// 关注的好友
type Friend struct {
	EntityID    uint32 `bson:"EntityID"`
	AddTime     uint64 `bson:"AddTime"`
	GiveGoldSec int64  `bson:"GiveGoldSec"` //赠送时间
	Gold        uint32 `bson:"Gold"`        //金额
}

type FansAttribute struct {
	FansUnixSec int64  `bson:"FansUnixSec"`
	List        []Fans `bson:"List"`
}

// 粉丝
type Fans struct {
	EntityID uint32
	AddTime  uint64
}

// 赠送金币
type GiveGold struct {
	EntityID    uint32 //被赠送人id
	Gold        uint32 //金额
	GiveGoldSec int64  //时间戳
}

// 签到记录
type SignInReward struct {
	ID                 uint64   // 签到活动的唯一ID
	SignLog            []uint64 //签到记录bitmap
	LastSignInUnixSec  int64    // 上次签到秒(时间戳)
	FirstSignInUnixSec int64    // 首次签到秒(时间戳)
}

// 数据实体 物品
type ShopItem struct {
	ItemID  uint32 `bson:"ItemID"`  //商店ItemID
	TableID uint32 `bson:"TableID"` //商店配置表ID
	BuyNum  uint32 `bson:"ItemNum"` //购买数量
	BuyTime string `bson:"BuyTime"` //购买时间
}

// 进度领取表
type ProgressList struct {
	DateStamp          int64            `bson:"DateStamp"` //日期时间戳
	ProgressRewardList []ProgressReward `bson:"ProgressRewardList"`
}

// 领取奖励
type ProgressReward struct {
	ProgressID  uint32 `bson:"ProgressID"`  // 进度表id
	StateReward uint32 `bson:"StateReward"` // 0未领取，1领取
	RewardTime  string `bson:"RewardTime"`  //领取时间
}

// 任务进度表集合
type TaskProgress struct {
	DayProgressValue       uint32
	WeekProgressValue      uint32
	DayProgressRewardList  []ProgressReward
	WeekProgressRewardList []ProgressReward
}

// 称号单元
type Collect struct {
	CollectID        uint32 `bson:"CollectID"`        // 称号id
	ConditionID      uint32 `bson:"ConditionID"`      //条件id
	State            uint32 `bson:"State"`            //0未完成，1可激活，2已激活
	CompleteProgress uint32 `bson:"CompleteProgress"` //完成的进度
	TaskProgress     uint32 `bson:"TaskProgress"`     //任务配置的进度
	AddTime          string `bson:"AddTime"`
	Apply            uint32 `bson:"Apply"` // 0未使用，1使用中
}

// 成就奖励
type AchievementLVReward struct {
	AchievementLvID uint32 `bson:"AchievementLvID"` //奖励等级id
	StateReward     uint32 `bson:"StateReward"`     // 是否领取，0未领取，1领取
	Score           uint32 `bson:"Score"`           //积分
	AddTime         string `bson:"AddTime"`         //领取时间
}

// 成就表
type Achievement struct {
	AchievementID uint32             `bson:"AchievementID"`
	ChildList     []ChildAchievement `bson:"ChildList"`
	TypeN         uint32             `bson:"TypeN"`
}

// 成就元素表
type ChildAchievement struct {
	ChildID          uint32 `bson:"ChildID"`          // 成就元素id
	ConditionID      uint32 `bson:"ConditionID"`      // 条件id
	State            uint32 `bson:"State"`            //0未完成，1完成
	CompleteProgress uint32 `bson:"CompleteProgress"` //完成的进度
	TaskProgress     uint32 `bson:"TaskProgress"`     //任务配置的进度
	AddTime          string `bson:"AddTime"`          // 添加时间
}

// 每日签到记录
type DailySignInElement struct {
	ObjID              bson.ObjectId `bson:"ID"` //唯一ID
	MonthKey           int           `bson:"MonthKey"`
	SignLog            []uint64      `bson:"SignLog"`            //签到记录bitmap（连续签到7天会重置）
	SummarySignLog     []uint64      `bson:"SummarySignLog"`     //总签到记录bitmap（不会重置）
	LastSignInUnixSec  int64         `bson:"LastSignInUnixSec"`  // 上次签到秒(时间戳)
	FirstSignInUnixSec int64         `bson:"FirstSignInUnixSec"` // 首次签到秒(时间戳)
	SignType           uint32        `bson:"SignType"`           //0普通，1广告
}

// 赠送金币记录
type GiveGoldDate struct {
	LastGiveGoldUnixSec int64         `bson:"LastGiveGoldUnixSec"` // 上次赠送秒(时间戳)
	GiveElementList     []GiveElement `bson:"GiveElementList"`
}

type GiveElement struct {
	ObjID      bson.ObjectId `bson:"ID"` //唯一ID
	MonthKey   int           `bson:"MonthKey"`
	ElementLog []uint64      `bson:"SignLog"` //赠送记录
}

// 俱乐部相关
type ClubAttribute struct {
	ClubReFreshUnix        int64                `bson:"ClubReFreshUnix"`        //俱乐部刷新时间（用来判断是否发生了重置）
	DailySignInUnixSec     int64                `bson:"DailySignInUnixSec"`     //打卡时间
	ClubActiveValue        uint32               `bson:"ClubActiveValue"`        //俱乐部任务活跃值
	ExitClubUnixSec        int64                `bson:"ExitClubUnixSec"`        //退出俱乐部时间，不可以重置 ，只有退出俱乐部才写这个字段
	ClubProgressRewardList []ClubProgressReward `bson:"ClubProgressRewardList"` //活跃值领取表
	ClubTaskProgressList   []ClubTaskProgress   `bson:"ClubTaskProgressList"`   //我的活跃领取表
	ClubTaskList           []ClubWeekTask       `bson:"ClubTaskList"`           //俱乐部任务表
}

type ClubProgressReward struct {
	ProgressID  uint32 `bson:"ProgressID"`
	Progress    uint32 `bson:"Progress"`
	StateReward uint32 `bson:"StateReward"`
	AddTime     string `bson:"AddTime"` // 添加时间
}

type ClubTaskProgress struct {
	ProgressID  uint32 `bson:"ProgressID"`
	Progress    uint32 `bson:"Progress"`
	StateReward uint32 `bson:"StateReward"`
	AddTime     string `bson:"AddTime"` // 添加时间
}

type ClubWeekTask struct {
	TaskID            uint32          `bson:"TaskID"`           // 任务id
	ConditionID       uint32          `bson:"ConditionID"`      //条件id
	State             uint32          `bson:"State"`            //0未完成，1完成
	CompleteProgress  uint32          `bson:"CompleteProgress"` //完成的进度
	TaskProgress      uint32          `bson:"TaskProgress"`     //任务配置的进度
	AddTime           string          `bson:"AddTime"`
	ClubDailyTaskList []ClubDailyTask `bson:"ClubDailyTaskList"` //天任务
}

type ClubDailyTask struct {
	State            uint32 `bson:"State"`            //0未完成，1完成
	CompleteProgress uint32 `bson:"CompleteProgress"` //完成的进度
	TaskProgress     uint32 `bson:"TaskProgress"`     //任务配置的进度
	AddTime          string `bson:"AddTime"`
}

type Box struct {
	ObjID           bson.ObjectId `bson:"ID"`              //唯一ID
	BoxNum          uint32        `bson:"BoxNum"`          //宝箱位置
	BoxID           uint32        `bson:"BoxID"`           //宝箱id
	GameType        uint32        `bson:"GameType"`        // 游戏类型，0：8球，1：血流，2：斯诺克
	RoomType        uint32        `bson:"RoomType"`        // 房间类型，0：新手，1初级，2中级，3高级，4巅峰
	AddTime         string        `bson:"AddTime"`         //加入时间
	UnlockTimeStamp int64         `bson:"UnlockTimeStamp"` //超过当前时间，可以领取；
	ReduceTime      uint32        `bson:"ReduceTime"`      //扣减秒数
}

type ElemBook struct {
	Key        uint32 `bson:"Key"`
	CueID      uint32 `bson:"CueID"`
	State      uint32 `bson:"State"` //0未完成，1完成
	AddTime    string `bson:"AddTime"`
	CueQuality uint32 `bson:"CueQuality"` //品质
}

// 更新资源参数配置
type ResParam struct {
	Uuid     string
	SysID    uint32
	ActionID uint32
}

type PropertyItem struct {
	TableID   uint32
	ItemValue int32
}

// 免费商店参数
type FreeShopRefresh struct {
	RefreshAdTimes   uint32 `bson:"RefreshAdTimes"`   //剩余刷新次数
	LastRefreshStamp int64  `bson:"LastRefreshStamp"` //最后一次刷新时间
}

type LoginElem struct {
	LastLoginTime int64         `bson:"LastLoginTime"` //今日最早登录时间
	RewardList    []LoginReward `bson:"RewardList"`    //奖励领取列表
}

// 登录奖励字段
type LoginReward struct {
	TimeKey  uint32 `bson:"TimeKey"`
	IsReward bool   `bson:"IsReward"` //false未领取，true已领取
}

// 积分商城商品购买记录
type PointsShopBuy struct {
	PointsMallId string `bson:"PointsMallId"` //积分商城ID
	Num          uint32 `bson:"Num"`          //累计购买数量
}

// 充值
type Recharge struct {
	TableID uint32 `bson:"TableID"` //配置表key
	IsBuy   bool   `bson:"IsBuy"`   //false未购买，true已购买
}
