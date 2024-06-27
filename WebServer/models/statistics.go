package models

import (
	"BilliardServer/WebServer/utils"
	"github.com/beego/beego/v2/core/logs"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// 玩家创建日志
type UserCreateLog struct {
	LogID         bson.ObjectId `bson:"_id"`                       //唯一标识
	Time          int64         `bson:"Time"`                      //创建时间
	Account       string        `bson:"Account"`                   //账号
	EntityID      uint32        `bson:"EntityID" json:"entity_id"` //帐号ID
	IsIPhone      bool          `bson:"IsIPhone"`                  //是否苹果设备
	Platform      uint32        `bson:"Platform"`                  //平台
	LoginPlatform uint32        `bson:"LoginPlatform"`             //登录平台
	Channel       uint32        `bson:"Channel"`                   //渠道
	DeviceId      string        `bson:"DeviceId"`                  //设备Id
	Machine       string        `bson:"Machine"`                   //机型
	RemoteAddr    string        `bson:"RemoteAddr"`                //远端ip
	PackageName   string        `bson:"PackageName"`               //当前客户端包名
	Language      uint32        `bson:"Language"`                  //玩家的语言标记
}

// 玩家登录日志
type UserLoginLog struct {
	LogID         bson.ObjectId `bson:"_id"`                       //唯一标识
	Time          int64         `bson:"Time"`                      //登录时间
	Account       string        `bson:"Account"`                   //账号
	EntityID      uint32        `bson:"EntityID" json:"entity_id"` //帐号ID
	IsIPhone      bool          `bson:"IsIPhone"`                  //是否苹果设备
	Platform      uint32        `bson:"Platform"`                  //平台
	LoginPlatform uint32        `bson:"LoginPlatform"`             //登录平台
	Channel       uint32        `bson:"Channel"`                   //渠道
	DeviceId      string        `bson:"DeviceId"`                  //设备Id
	Machine       string        `bson:"Machine"`                   //机型
	RemoteAddr    string        `bson:"RemoteAddr"`                //远端ip
	PackageName   string        `bson:"PackageName"`               //当前客户端包名
	Language      uint32        `bson:"Language"`                  //玩家的语言标记
	IsNew         bool          `bson:"IsNew"`                     //是否为新用户
}

func AddCreateAccountLog(log *UserCreateLog) {
	if log == nil {
		return
	}

	log.LogID = bson.NewObjectId()
	err := utils.LogDB.InsertData("user_create_log", log)
	if err != nil {
		logs.Warning("-->models--AddCreateAccountLog--Error:", err)
		return
	}

	return
}

func AddLoginLog(log *UserLoginLog) {
	if log == nil {
		return
	}

	log.LogID = bson.NewObjectId()
	err := utils.LogDB.InsertData("user_login_log", log)
	if err != nil {
		logs.Warning("-->models--AddLoginLog--Error:", err)
		return
	}

	return
}

// 创建一个玩家登出日志
func CreateLogoutLog(account string, logoutTime time.Time, onLineTime int, deviceId string, channelId int, bundleId string) {
	return
}
