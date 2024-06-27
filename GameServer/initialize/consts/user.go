package consts

const (
	LOGIN_GAME_MSG  = "欢迎VIP用户<color=#ffff00ff><b>%s</b></color>闪亮登场！"
	CREATE_GAME_MSG = "VIP用户<color=#ffff00ff><b>%s</b></color>注册成功！"

	USER_MAN   = 1
	USER_WOMEN = 0
)

const (
	PopWeekRank uint32 = iota + 1 // 人气排行榜
	PopRank                       //总榜
)

const (
	DayGiveGoldTimes = 20  //每天赠送次数
	GoldAmount       = 100 //每次赠送金额

	MSGMAXNUM = 40
	MaxBoxNum = 4

	USER_STATUS_KICK_OUT    = 1 //踢人
	USER_STATUS_PROHIBITION = 2 //禁言
	USER_STATUS_BAN_ACC     = 3 //封账号
	USER_STATUS_BAN_IP      = 4 //封IP
)

// 0是在线，1为游戏中，2离线
const (
	PlayerOnline uint32 = iota
	PlayerIn
	PlayerOutline
)

// 客户端运行点类型
const (
	ClientRunningPointType_StartUp        = "0" //启动
	ClientRunningPointType_AfterHotUpdate = "1" //热更
	ClientRunningPointType_Login          = "2" //登录
	ClientRunningPointType_Guide          = "3" //引导
	ClientRunningPointType_Level          = "4" //等级
	ClientRunningPointType_Download       = "5" //下载
)

// 客户端运行记录数据集
const (
	ClientRunningPointCollection          = "ClientRunningRecord"
	ClientRunningPointName_StartUp        = "StartUpRecord"
	ClientRunningPointName_AfterHotUpdate = "AfterHotUpdateRecord"
	ClientRunningPointName_Login          = "LoginRecord"
	ClientRunningPointName_Guide          = "GuideRecord"
	ClientRunningPointName_Level          = "LevelRecord"
)

//// 客户端运行打点数据
//type ClientRunningRecord struct {
//	ObjID       bson.ObjectId `bson:"_id"`         //唯一ID
//	Date        time.Time     `bson:"Date"`        //时间
//	DeviceId    string        `bson:"DeviceId"`    //设备唯一id
//	AccoutId    string        `bson:"AccoutId"`    //账号id
//	Platform    string        `bson:"Platform"`    //渠道
//	RecordPoint string        `bson:"RecordPoint"` //记录点，对应ClientRecordPointType枚举
//	PointParam  string        `bson:"PointParam"`  //记录点参数
//	BundleID    string        `bson:"BundleID"`    //包名
//	Reason      string        `bson:"Reason"`      //原因，有些错误打点会有原因
//	DeviceName  string        `bson:"DeviceName"`  //设备名称，方便查询
//	IP          string        `bson:"IP"`          //IP地址
//	Unique      string        `bson:"Unique"`      //一次操作的标记
//}

var (
	DefaultCueKey    uint32 = 1 //默认球杆
	DefaultIconFrame uint32 = 50200001
)

// 0正常领取，1视频补领
const (
	LoginStateReward_0 = iota
	LoginStateReward_1
)
