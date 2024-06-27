package vars

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

// 包配置
type PackageConfig struct {
	ID   int    `bson:"ID" json:"ID"`     //ID
	Name string `bson:"Name" json:"Name"` //包配置名称/详情

	PackageNames []string `bson:"PackageNames" json:"PackageNames"` //包名

	ServerListIP   string `bson:"ServerListIP" json:"ServerListIP"`     //获取服务器列表的IP 如果为空，使用Game里配置的
	ServerListPort string `bson:"ServerListPort" json:"ServerListPort"` //获取服务器列表的端口 如果为空，使用Game里配置的

	WhiteServerListIP   string `bson:"WhiteServerListIP" json:"WhiteServerListIP"`     //获取白名单服务器列表的IP 如果为空，使用Game里配置的
	WhiteServerListPort string `bson:"WhiteServerListPort" json:"WhiteServerListPort"` //获取白名单服务器列表的端口 如果为空，使用Game里配置的

	UpdateFlag bool   `bson:"UpdateFlag" json:"UpdateFlag"` //更新标记
	UpdateURL  string `bson:"UpdateURL" json:"UpdateURL"`   //更新URL
	UploadURL  string `bson:"UploadURL" json:"UploadURL"`   //上传资源使用的URL
	UploadType int    `bson:"UploadType" json:"UploadType"` //上传资源的方式
	Version    string `bson:"Version" json:"Version"`       //原本是版本号 现在用来做热刷字段

	AlternateUpdateURL  string `bson:"AlternateUpdateURL" json:"AlternateUpdateURL"`   //备用更新URL
	AlternateUploadURL  string `bson:"AlternateUploadURL" json:"AlternateUploadURL"`   //备用更新URL
	AlternateUploadType int    `bson:"AlternateUploadType" json:"AlternateUploadType"` //备用上传资源的方式

	Flag bool `bson:"Flag" json:"Flag"` //标记

	WhiteIPList     []string `bson:"WhiteIPList" json:"WhiteIPList"`         //热更新白名单IP
	WhiteUpdateURL  string   `bson:"WhiteUpdateURL" json:"WhiteUpdateURL"`   //热更新的更新URL
	WhiteUploadURL  string   `bson:"WhiteUploadURL" json:"WhiteUploadURL"`   //上传热更新资源使用的URL
	WhiteUploadType int      `bson:"WhiteUploadType" json:"WhiteUploadType"` //上传资源的方式
	WhiteVersion    string   `bson:"WhiteVersion" json:"WhiteVersion"`       //原本是白名单版本号 现在用来做白名单热刷字段
	WhiteResVersion string   `bson:"WhiteResVersion" json:"WhiteResVersion"` //资源版本号，每次更新必须要换一个，否则填空

	ForceUpdateFlag    int    `bson:"ForceUpdateFlag" json:"ForceUpdateFlag"`       //强更标记 0 不强更 1 强更 2 需要更新但是可以不更新进去
	ForceUpdateReason  string `bson:"ForceUpdateReason" json:"ForceUpdateReason"`   //强更理由
	ForceUpdateURL     string `bson:"ForceUpdateURL" json:"ForceUpdateURL"`         //强更URL
	ForceUpdateVersion int    `bson:"ForceUpdateVersion" json:"ForceUpdateVersion"` //强更版本号

	ChannelTag int `bson:"ChannelTag" json:"ChannelTag"` //渠道标记，1魔亚（用来打开魔亚特殊的活动）

	ClientIP  string `bson:"ClientIP" json:"ClientIP"`   //客户端IP
	IsWhiteIP bool   `bson:"IsWhiteIP" json:"IsWhiteIP"` //是否是白名单

	Language int `bson:"Language" json:"Language"` //包的默认语言标记

	FunctionOpen []int `bson:"FunctionOpen" json:"FunctionOpen"` //功能开放标记

	/////////////////////////////以下字段长尾小七都不支持//////////////////////////////
	MultiLanguage bool `bson:"MultiLanguage" json:"MultiLanguage"` //该包是否多语言，这个标记是服务器使用的

	CDNUpdateURL            []string `bson:"CDNUpdateURL" json:"CDNUpdateURL"`                       //CDN下载地址代替 更新地址
	CDNUploadURL            []string `bson:"CDNUploadURL" json:"CDNUploadURL"`                       //CDN上传地址代替 更新地址
	CDNUploadType           []int    `bson:"CDNUploadType" json:"CDNUploadType"`                     //CDN上传方式
	CDNErrorCount           int      `bson:"CDNErrorCount" json:"CDNErrorCount"`                     //CDN下载最大错误次数
	FilesMD5                string   `bson:"FilesMD5" json:"FilesMD5"`                               //files文件的MD5，以防下载错误
	WebClientType           int      `bson:"WebClientType" json:"WebClientType"`                     //下载方式 1 Luaframework；2 振东写的下载
	ThreadCountPerProcessor int      `bson:"ThreadCountPerProcessor" json:"ThreadCountPerProcessor"` //下载的进程的核心数
	ResVersion              string   `bson:"ResVersion" json:"ResVersion"`                           //资源版本号，每次更新必须要换一个，否则填空

	//由于小七和长尾不支持这个字段，暂时不用了
	//改用Version和WhiteVersion作为热刷字段
	HeatFleshLua string `bson:"HeatFleshLua" json:"HeatFleshLua"` //热刷Lua脚本

	ShowLanuageOnStart bool `bson:"ShowLanuageOnStart" json:"ShowLanuageOnStart"` //启动就显示语言按钮

	Params []string `bson:"Params" json:"Params"` //备用字段

	//在海外英文包以后才有的
	ShowLanguage   []int    `bson:"ShowLanguage" json:"ShowLanguage"`     //后台发送的可以选择的语言
	ChooseLanguage bool     `bson:"ChooseLanguage" json:"ChooseLanguage"` //是否可以选择语言
	Hash           string   `bson:"Hash" json:"Hash"`                     //模式哈希
	LanguageFormat string   `bson:"LanguageFormat" json:"LanguageFormat"` //SDK语言对应游戏语言的字符串，格式如:"zh-Hans:1,en:2,zh-Hant:3,ko-KR:4"
	CantChangeLang bool     `bson:"CantChangeLang" json:"CantChangeLang"` //是否可以被修改语言，提审时这个值设置false就不会
	CenterPaths    []string `bson:"CenterPaths" json:"CenterPaths"`       //包中心服地址配置

	//服务器用的，白名单强更
	WhiteForceUpdateFlag    int    `bson:"WhiteForceUpdateFlag" json:"WhiteForceUpdateFlag"`       //强更标记 0 不强更 1 强更 2 需要更新但是可以不更新进去
	WhiteForceUpdateReason  string `bson:"WhiteForceUpdateReason" json:"WhiteForceUpdateReason"`   //强更理由
	WhiteForceUpdateURL     string `bson:"WhiteForceUpdateURL" json:"WhiteForceUpdateURL"`         //强更URL
	WhiteForceUpdateVersion int    `bson:"WhiteForceUpdateVersion" json:"WhiteForceUpdateVersion"` //强更版本号
}

type PackageSkip struct {
	ObjID       bson.ObjectId `bson:"_id"`
	PackageName string        `bson:"PackageName"` //包名
	StartIp     int64         `bson:"StartIp"`
	EndIp       int64         `bson:"EndIp"`
	ID          int           `bson:"ID"` //ID
}

// 包语言配置
type PackageLangConfig struct {
	Languange      []int    `bson:"Languange" json:"Languange"`
	StrParams      []string `bson:"StrParams" json:"StrParams"`
	IntParams      []int    `bson:"IntParams" json:"IntParams"`
	ChooseLanguage bool     `bson:"ChooseLanguage" json:"ChooseLanguage"` //是否可以选择语言
	Hash           string   `bson:"Hash" json:"Hash"`                     //模式哈希
}

// 包切换配置
type PackageSwitchData struct {
	SwitchTime time.Time `bson:"SwitchTime" json:"SwitchTime"` //切换时间
	Name       string    `bson:"Name" json:"Name"`             //名字
}

// 编辑器下的包配置
type PackageEditorConfig struct {
	ID   int    `bson:"ID" json:"ID"`     //ID
	Name string `bson:"Name" json:"Name"` //包配置名称/详情

	PackageNames []string `bson:"PackageNames" json:"PackageNames"` //包名

	ServerListIP   string `bson:"ServerListIP" json:"ServerListIP"`     //获取服务器列表的IP 如果为空，使用Game里配置的
	ServerListPort string `bson:"ServerListPort" json:"ServerListPort"` //获取服务器列表的端口 如果为空，使用Game里配置的

	WhiteServerListIP   string `bson:"WhiteServerListIP" json:"WhiteServerListIP"`     //获取白名单服务器列表的IP 如果为空，使用Game里配置的
	WhiteServerListPort string `bson:"WhiteServerListPort" json:"WhiteServerListPort"` //获取白名单服务器列表的端口 如果为空，使用Game里配置的

	UpdateFlag bool   `bson:"UpdateFlag" json:"UpdateFlag"` //更新标记
	UpdateURL  string `bson:"UpdateURL" json:"UpdateURL"`   //更新URL
	UploadURL  string `bson:"UploadURL" json:"UploadURL"`   //上传资源使用的URL
	UploadType int    `bson:"UploadType" json:"UploadType"` //上传资源的方式
	Version    string `bson:"Version" json:"Version"`       //更新版本名 已弃用

	AlternateUpdateURL  string `bson:"AlternateUpdateURL" json:"AlternateUpdateURL"`   //备用更新URL
	AlternateUploadURL  string `bson:"AlternateUploadURL" json:"AlternateUploadURL"`   //备用更新URL
	AlternateUploadType int    `bson:"AlternateUploadType" json:"AlternateUploadType"` //备用上传资源的方式

	Flag bool `bson:"Flag" json:"Flag"` //标记

	WhiteIPList     []string `bson:"WhiteIPList" json:"WhiteIPList"`         //热更新白名单IP
	WhiteUpdateURL  string   `bson:"WhiteUpdateURL" json:"WhiteUpdateURL"`   //热更新的更新URL
	WhiteUploadURL  string   `bson:"WhiteUploadURL" json:"WhiteUploadURL"`   //上传热更新资源使用的URL
	WhiteUploadType int      `bson:"WhiteUploadType" json:"WhiteUploadType"` //上传资源的方式
	WhiteVersion    string   `bson:"WhiteVersion" json:"WhiteVersion"`       //热更新版本 已弃用
	WhiteResVersion string   `bson:"WhiteResVersion" json:"WhiteResVersion"` //资源版本号，每次更新必须要换一个，否则填空

	ForceUpdateFlag    int    `bson:"ForceUpdateFlag" json:"ForceUpdateFlag"`       //强更标记 0 不强更 1 强更 2 需要更新但是可以不更新进去
	ForceUpdateReason  string `bson:"ForceUpdateReason" json:"ForceUpdateReason"`   //强更理由
	ForceUpdateURL     string `bson:"ForceUpdateURL" json:"ForceUpdateURL"`         //强更URL
	ForceUpdateVersion int    `bson:"ForceUpdateVersion" json:"ForceUpdateVersion"` //强更版本号

	ChannelTag int `bson:"ChannelTag" json:"ChannelTag"` //渠道标记，1魔亚（用来打开魔亚特殊的活动）

	ClientIP  string `bson:"ClientIP" json:"ClientIP"`   //客户端IP
	IsWhiteIP bool   `bson:"IsWhiteIP" json:"IsWhiteIP"` //是否是白名单

	Language int `bson:"Language" json:"Language"` //包的默认语言标记

	FunctionOpen []int `bson:"FunctionOpen" json:"FunctionOpen"` //功能开放标记

	/////////////////////////////以下字段长尾小七都不支持//////////////////////////////
	MultiLanguage bool `bson:"MultiLanguage" json:"MultiLanguage"` //该包是否多语言，这个标记是服务器使用的

	CDNUpdateURL            []string `bson:"CDNUpdateURL" json:"CDNUpdateURL"`                       //CDN下载地址代替 更新地址
	CDNUploadURL            []string `bson:"CDNUploadURL" json:"CDNUploadURL"`                       //CDN上传地址代替 更新地址
	CDNUploadType           []int    `bson:"CDNUploadType" json:"CDNUploadType"`                     //CDN上传方式
	CDNErrorCount           int      `bson:"CDNErrorCount" json:"CDNErrorCount"`                     //CDN下载最大错误次数
	FilesMD5                string   `bson:"FilesMD5" json:"FilesMD5"`                               //files文件的MD5，以防下载错误
	WebClientType           int      `bson:"WebClientType" json:"WebClientType"`                     //下载方式 1 Luaframework；2 振东写的下载
	ThreadCountPerProcessor int      `bson:"ThreadCountPerProcessor" json:"ThreadCountPerProcessor"` //下载的进程的核心数
	ResVersion              string   `bson:"ResVersion" json:"ResVersion"`                           //资源版本号，每次更新必须要换一个，否则填空

	HeatFleshLua string `bson:"HeatFleshLua" json:"HeatFleshLua"` //热刷Lua脚本

	ShowLanuageOnStart bool `bson:"ShowLanuageOnStart" json:"ShowLanuageOnStart"` //启动就显示语言按钮

	Params []string `bson:"Params" json:"Params"` //备用字段

	//在海外英文包以后才有的
	//后台发送的可以选择的语言包
	ShowLanguage   []int    `bson:"ShowLanguage" json:"ShowLanguage"`
	ChooseLanguage bool     `bson:"ChooseLanguage" json:"ChooseLanguage"` //是否可以选择语言
	Hash           string   `bson:"Hash" json:"Hash"`                     //模式哈希
	LanguageFormat string   `bson:"LanguageFormat" json:"LanguageFormat"` //SDK语言对应游戏语言的字符串，格式如:"zh-Hans:1,en:2,zh-Hant:3"
	CantChangeLang bool     `bson:"CantChangeLang" json:"CantChangeLang"` //是否可以被修改语言，提审时这个值设置false就不会
	CenterPaths    []string `bson:"CenterPaths" json:"CenterPaths"`       //包中心服地址配置

	/////////////////////////////以下字段用于兼容老版本的URL//////////////////////////////
	PreName string `bson:"PreName" json:"PreName"` //预更新名称

	PreUpdateURL  string `bson:"PreUpdateURL" json:"PreUpdateURL"`   //预更新的更新URL
	PreUploadURL  string `bson:"PreUploadURL" json:"PreUploadURL"`   //预更新上传资源使用的URL
	PreUploadType int    `bson:"PreUploadType" json:"PreUploadType"` //预更新上传资源的方式

	PreAltUpdateURL  string `bson:"PreAltUpdateURL" json:"PreAltUpdateURL"`   //预更新备用的更新URL
	PreAltUploadURL  string `bson:"PreAltUploadURL" json:"PreAltUploadURL"`   //预更新上传资源使用的URL
	PreAltUploadType int    `bson:"PreAltUploadType" json:"PreAltUploadType"` //预更新备用上传资源的方式

	PreCDNUpdateURL  []string `bson:"PreCDNUpdateURL" json:"PreCDNUpdateURL"`   //预更新CDN的更新URL
	PreCDNUploadURL  []string `bson:"PreCDNUploadURL" json:"PreCDNUploadURL"`   //预更新CDN上传地址代替 更新地址
	PreCDNUploadType []int    `bson:"PreCDNUploadType" json:"PreCDNUploadType"` //预更新CDN上传方式

	/////////////////////////////替换记录//////////////////////////////
	History []*PackageSwitchData //根据时间插入，只保留最近10次

	//服务器用的，白名单强更
	WhiteForceUpdateFlag    int    `bson:"WhiteForceUpdateFlag" json:"WhiteForceUpdateFlag"`       //强更标记 0 不强更 1 强更 2 需要更新但是可以不更新进去
	WhiteForceUpdateReason  string `bson:"WhiteForceUpdateReason" json:"WhiteForceUpdateReason"`   //强更理由
	WhiteForceUpdateURL     string `bson:"WhiteForceUpdateURL" json:"WhiteForceUpdateURL"`         //强更URL
	WhiteForceUpdateVersion int    `bson:"WhiteForceUpdateVersion" json:"WhiteForceUpdateVersion"` //强更版本号
}

// 包配置数组，给C#用
type PackageConfigArray struct {
	Array []*PackageConfig `bson:"Array"` //配置数组
}

// 包配置返回结果
type PackageResult struct {
	Result int    `bson:"Result"` //返回结果
	Error  string `bson:"Error"`
}

type DeviceManageAuth struct {
	DeviceID string `bson:"DeviceID"`
	IsPay    bool   `bson:"IsPay"`
}

type ServerListIP struct {
	Ip string `bson:"ip" json:"ip"`
}

// 服务器包配置
type ServerPackageConfig struct {
	ID            int    `bson:"ID" json:"ID"`                       //ID
	Name          string `bson:"Name" json:"Name"`                   //包配置名称/详情
	FileExtension string `bson:"FileExtension" json:"FileExtension"` //文件后缀名(打包文件自动加上后缀名)
	UploadURL     string `bson:"UploadURL" json:"UploadURL"`         //上传资源使用的URL
	FilePath      string `bson:"FilePath" json:"FilePath"`           //文件路径
}

// 服务器包配置
type ServerUpdateCMD struct {
	ID              int    `bson:"ID" json:"ID"`                           //ID
	Name            string `bson:"Name" json:"Name"`                       //包配置名称/详情
	ServerPackageID int    `bson:"ServerPackageID" json:"ServerPackageID"` //服务器包配置ID
	TargetGMTIP     string `bson:"TargetGMTIP" json:"TargetGMTIP"`         //服务器GMT的IP
	ServerIDList    string `bson:"ServerIDList" json:"ServerIDList"`       //服务器ID列表
}
