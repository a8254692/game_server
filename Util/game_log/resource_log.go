package game_log

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// 玩家资源日志
type ResourceLog struct {
	LogID          bson.ObjectId `bson:"_id"`            //唯一标识
	Uuid           string        `bson:"Uuid"`           //关联uuid
	EntityID       uint32        `bson:"EntityID"`       //账号
	Time           int64         `bson:"Time"`           //日志时间
	ResType        uint32        `bson:"ResType"`        //资源类型
	ResSubType     uint32        `bson:"ResSubType"`     //资源子类型
	ResID          uint32        `bson:"ResID"`          //资源ID
	IncrType       uint32        `bson:"IncrType"`       //增加/减少
	Count          uint64        `bson:"Count"`          //数量
	AfterModifyNum uint32        `bson:"AfterModifyNum"` //修改后数量
	SystemID       uint32        `bson:"SystemID"`       //所属系统
	ActionID       uint32        `bson:"ActionID"`       //行为(用户操作)
	DeviceID       string        `bson:"DeviceID"`       //设备Id
	ChannelID      uint32        `bson:"ChannelID"`      //渠道Id,对应Common.ChannelType枚举
	BundleID       string        `bson:"BundleID"`       //包名
}

// 记录设备拦截日志
type DeviceIDLog struct {
	LogID    bson.ObjectId `bson:"_id"`      //唯一标识
	EntityID uint32        `bson:"EntityID"` //账号
	Time     int64         `bson:"Time"`     //日志时间
	DeviceId string        `bson:"DeviceId"` //设备Id
	ServerId string        `bson:"ServerId"` //服务器ID
}

// 记录消耗日志
func SaveConsumeLog(uuid string, entityID uint32, resType uint32, resSubType uint32, resID uint32, incrType uint32, count uint64, afterModifyNum uint32, systemID uint32, actionID uint32) {
	resLog := new(ResourceLog)
	resLog.LogID = bson.NewObjectId()
	resLog.Time = time.Now().Unix()
	resLog.Uuid = uuid
	resLog.EntityID = entityID
	resLog.ResType = resType
	resLog.ResID = resID
	resLog.ResSubType = resSubType
	resLog.IncrType = incrType
	resLog.Count = count
	resLog.AfterModifyNum = afterModifyNum
	resLog.SystemID = systemID
	resLog.ActionID = actionID
	GGameLogManager.AddLog(GameLogType_ConsumeLog, resLog)
}

// 记录产出日志
func SaveProductionLog(uuid string, entityID uint32, resType uint32, resSubType uint32, resID uint32, incrType uint32, count uint64, afterModifyNum uint32, systemID uint32, actionID uint32) {
	resLog := new(ResourceLog)
	resLog.LogID = bson.NewObjectId()
	resLog.Time = time.Now().Unix()
	resLog.Uuid = uuid
	resLog.EntityID = entityID
	resLog.ResType = resType
	resLog.ResSubType = resSubType
	resLog.ResID = resID
	resLog.IncrType = incrType
	resLog.Count = count
	resLog.AfterModifyNum = afterModifyNum
	resLog.SystemID = systemID
	resLog.ActionID = actionID
	GGameLogManager.AddLog(GameLogType_Production, resLog)
}

// 记录设备拦截日志
func SaveDeviceIdLog(entityID uint32, deviceId string, serverId string) {
	resLog := new(DeviceIDLog)
	resLog.LogID = bson.NewObjectId()
	resLog.Time = time.Now().Unix()
	resLog.EntityID = entityID
	resLog.DeviceId = deviceId
	resLog.ServerId = serverId
	GGameLogManager.AddLog(GameLogType_DeviceIdLog, resLog)
}
