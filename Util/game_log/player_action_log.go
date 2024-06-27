package game_log

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

// 行为日志
type PlayerActionLog struct {
	LogID       bson.ObjectId `bson:"_id"`         //唯一标识
	Time        int64         `bson:"Time"`        //消耗时间
	Account     string        `bson:"Account"`     //账号
	SystemID    int           `bson:"System"`      //所属系统
	ActionID    int           `bson:"Action"`      //行为
	PlayerLevel int           `bson:"PlayerLevel"` //玩家等级
	DeviceId    string        `bson:"DeviceId"`    //设备Id
	ChannelId   int           `bson:"ChannelId"`   //渠道Id,对应Common.ChannelType枚举
	BundleID    string        `bson:"BundleID"`    //包名
}

// 保存一条玩家行为日志
func CreateActionLog(account string, systemID int, actionID int, level int, guanqiaLevel int) {
	log := &PlayerActionLog{
		LogID:       bson.NewObjectId(),
		Time:        time.Now().Unix(),
		Account:     account,
		SystemID:    systemID,
		ActionID:    actionID,
		PlayerLevel: level,
	}

	GGameLogManager.AddLog(GameLogType_PlayerActionLog, log)
}
