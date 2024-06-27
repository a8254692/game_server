package game_log

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

// 在线曲线
type OnlineCurve struct {
	ObjID          bson.ObjectId `bson:"_id"`            //唯一ID
	Time           int64         `bson:"Time"`           //时间
	HighOlineCount uint32        `bson:"HighOlineCount"` //最高在线
}

// 在线记录(每天)
type OnlineData struct {
	ObjID               bson.ObjectId `bson:"_id"`                 //唯一ID
	Time                int64         `bson:"Time"`                //时间
	MaxLoginTimes       uint32        `bson:"MaxLoginTimes"`       //最大登录次数（某个玩家）
	AvgLoginTimes       uint32        `bson:"AvgLoginTimes"`       //平均登录次数（登录次数/登录人数）
	MaxOnlineTimeLength uint32        `bson:"MaxOnlineTimeLength"` //最大在线时长
	AvgOnlineTimeLength uint32        `bson:"AvgOnlineTimeLength"` //平均在线时长
	MaxOnlineTime       int64         `bson:"MaxOnlineTime"`       //最高在线时间点
	ACU                 uint32        `bson:"ACU"`                 //平均同时在线玩家人数
	PCU                 uint32        `bson:"PCU"`                 //最高同时在线玩家人数
}

// 保存最高在线曲线数据
func CreateHighOlineLog(num uint32) {
	data := &OnlineCurve{
		ObjID:          bson.NewObjectId(),
		Time:           time.Now().Unix(),
		HighOlineCount: num,
	}
	GGameLogManager.AddLog(GameLogType_OnlineCurveLog, data)
}

// 保存在线数据
func CreateOnlineLog(maxLogin uint32, avgLogin uint32, maxOlineLen uint32, avgOlineLen uint32, maxOline int64, acu uint32, pcu uint32) {
	log := &OnlineData{
		ObjID:               bson.NewObjectId(),
		Time:                time.Now().Unix(),
		MaxLoginTimes:       maxLogin,
		AvgLoginTimes:       avgLogin,
		MaxOnlineTimeLength: maxOlineLen,
		AvgOnlineTimeLength: avgOlineLen,
		MaxOnlineTime:       maxOline,
		ACU:                 acu,
		PCU:                 pcu,
	}
	GGameLogManager.AddLog(GameLogType_OnlineRecordLog, log)
}
