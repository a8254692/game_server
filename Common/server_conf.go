package Common

import (
	"time"
)

// 语言类型
const (
	ModeLocal = "local"
	ModeDev   = "dev"
	ModeProd  = "prod"

	LocalConfPath  = "../Common/conf/"
	LocalTablePath = "../Common/table/"
	DevConfPath    = "conf/"
	ProdConfPath   = "conf/"
	CfgFileName    = "server.ini"

	LANGUAGE_SCN = 0 //中文简体
	LANGUAGE_TCN = 1 //中文繁体
	LANGUAGE_EN  = 2 //英文
)

// 平台枚举
const (
	PLATFORM_TEST    = iota + 1 //测试平台
	PLATFORM_ANDROID            //首发安卓平台
	PLATFORM_IOS                //首发ios平台
)

var PLATFORM_STR_MAP = map[int]string{
	PLATFORM_TEST:    "test",
	PLATFORM_ANDROID: "android",
	PLATFORM_IOS:     "ios",
}

// 登录平台枚举
const (
	LOGIN_PLATFORM_GOOGLE   = iota + 1 //测试平台
	LOGIN_PLATFORM_FACEBOOK            //首发安卓平台
)

var LOGIN_PLATFORM_STR_MAP = map[int]string{
	LOGIN_PLATFORM_GOOGLE:   "google",
	LOGIN_PLATFORM_FACEBOOK: "facebook",
}

// 渠道枚举
const ( //ID
	CHANNEL_TYPE_NULLSDK = iota + 1 //白包
	CHANNEL_TYPE_JY                 //91
)

var CHANNEL_TYPE_STR_MAP = map[int]string{
	CHANNEL_TYPE_NULLSDK: "test",
	CHANNEL_TYPE_JY:      "aaa",
}

// 支付渠道枚举
const ( //ID
	PAY_CHANNEL_GOOGLE = iota + 1 //白包
	PAY_CHANNEL_JY                //91
)

var PAY_CHANNEL_TYPE_STR_MAP = map[int]string{
	PAY_CHANNEL_GOOGLE: "test",
	PAY_CHANNEL_JY:     "android",
}

// 服务器配置
type ServerConfig struct {
	ServerID         string    `bson:"ServerID"`         //通过ID获取数据
	TcpAddr          string    `bson:"TcpAddr"`          //Tcp开放地址
	MaxConnectNum    int       `bson:"MaxConnectNum"`    //最大连接数量
	DBType           string    `bson:"DBType"`           //使用的数据库类型
	ProtoType        string    `bson:"ProtoType"`        //使用的协议类型
	NetBuffLen       int       `bson:"NetBuffLen"`       //数据发送缓冲
	StartTime        time.Time `bson:"StartTime"`        //开服时间
	DBName           string    `bson:"DBName"`           //数据库名称
	WebAddr          string    `bson:"WebAddr"`          //公开的Web端口
	IP               string    `bson:"IP"`               //服务器IP地址
	Platform         int       `bson:"Platform"`         //Platform 0 内部，1 安卓，2 苹果
	InnerIP          string    `bson:"InnerIP"`          //内网地址 192.168.1.250
	DBIP             string    `bson:"DBIP"`             //数据库地址
	ServerName       string    `bson:"ServerName"`       //服务器名称
	GroupID          int       `bson:"GroupID"`          //服务器组ID
	ServerType       int       `bson:"ServerType"`       //服务器类型，为1的则为普通玩家，为2为白名单(后台关掉白名单要用到)，自动开服检测会修改这个值
	PushID           string    `bson:"PushID"`           //推送ID
	PushTag          string    `bson:"PushTag"`          //推送tag，玩家按服务器做区分
	DirName          string    `bson:"DirName"`          //服务器文件夹名字（一键更新需要）
	ReplSetDBs       []string  `bson:"ReplSetDBs"`       //mongo副本集	弃用，不再使用
	LogDBAddr        string    `bson:"LogDBAddr"`        //日志库地址（包括端口）
	Language         int       `bson:"Language"`         //翻译表的语言类型 0，中文，1，英文
	OpenTime         time.Time `bson:"OpenTime"`         //对外开放时间
	HubGroupID       int       `bson:"HubGroupID"`       //HubSvr分组ID
	HubGFClose       bool      `bson:"HubGFClose"`       //HubSvr组织战关闭标记
	CombineServerIDs []string  `bson:"CombineServerIDs"` //所合并的服务器ID，如果没合服，这个数组为空
	DirectID         string    `bson:"DirectID"`         //合服后所指向的服务器，用来显示是几服的玩家
	IsCombine        bool      `bson:"IsCombine"`        //已被合服的标记
	CombineServerId  string    `bson:"CombineServerId"`  //合服后服务器id
	//开启配置
	CenterWebAddr string `bson:"CenterWebAddr"` //中心服内网地址
	CSvrTcpAddr   string `bson:"CSvrTcpAddr"`   //跨服战服务器内网地址

	// 2020-9-7 限时活动合服显示
	MergeEndTime time.Time `bson:"MergeEndTime"` //合服结束时间

	CSActivityDBName string `bson:"CSActivityDBName"` //跨服活动数据库名称
}

// 服务器列表信息
type ServerInfo struct {
	Id              string `bson:"Id" json:"Id"`                           //服务器id
	Title           string `bson:"Title" json:"Title"`                     //服务器名字
	Ip              string `bson:"Ip" json:"Ip"`                           //服务器ip
	Port            string `bson:"Port" json:"Port"`                       //服务器端口
	State           int    `bson:"State" json:"State"`                     //服务器状态
	LastState       int    `bson:"LastState" json:"LastState"`             //更新前服务器状态
	ServerType      int    `bson:"ServerType" json:"ServerType"`           //服务器类型，为1的则为普通玩家，为2为白名单
	StateTips       string `bson:"StateTips" json:"StateTips"`             //服务器状态提示
	PushTag         string `bson:"PushTag" json:"PushTag"`                 //推送tag，玩家按服务器做区分
	CombineServerId string `bson:"CombineServerId" json:"CombineServerId"` //合服后服务器
}
