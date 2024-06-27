package collection

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/stack"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type UserDataStatistics struct {
	CollectionName     string        `bson:"-"`                  // 数据集名称
	ObjID              bson.ObjectId `bson:"_id,omitempty"`      // 唯一ID
	EntityId           uint32        `bson:"EntityId"`           //用户id
	AccumulateGold     uint32        `bson:"AccumulateGold"`     //累计获取的金币
	AccumulateGoal     uint32        `bson:"AccumulateGoal"`     //累计进球
	OneCueClear        uint32        `bson:"OneCueClear"`        //一杆清台
	IncrBindNum        uint32        `bson:"IncrBindNum"`        //加注次数
	C8PlayNum          uint32        `bson:"C8PlayNum"`          //对局次数
	C8WinNum           uint32        `bson:"C8WinNum"`           //胜利次数
	C8EscapeNum        uint32        `bson:"C8EscapeNum"`        //逃跑次数
	C8ContinuousWin    uint32        `bson:"C8ContinuousWin"`    //连胜次数(输了清零)
	C8MaxContinuousWin uint32        `bson:"C8MaxContinuousWin"` //最大连胜次数
	C8DoubleGoalNum    uint32        `bson:"C8DoubleGoalNum"`    //二连杆次数
	C8ThreeGoalNum     uint32        `bson:"C8ThreeGoalNum"`     //三连杆次数
}

// 初始化 第一次
func (s *UserDataStatistics) InitByFirst(collectionName string, entityId uint32) {
	s.CollectionName = collectionName
	s.ObjID = bson.NewObjectId()
	s.EntityId = entityId
	s.AccumulateGold = 0
	s.AccumulateGoal = 0
	s.OneCueClear = 0
	s.IncrBindNum = 0
	s.C8PlayNum = 0
	s.C8WinNum = 0
	s.C8EscapeNum = 0
	s.C8ContinuousWin = 0
	s.C8MaxContinuousWin = 0
	s.C8DoubleGoalNum = 0
	s.C8ThreeGoalNum = 0
}

// 获取ObjID
func (s *UserDataStatistics) GetObjID() string {
	return s.ObjID.String()
}

// 设置DBConnect
func (s *UserDataStatistics) SetDBConnect(collectionName string) {
	s.CollectionName = collectionName
}

// 清理实体
func (s *UserDataStatistics) Clear() {
	s.CollectionName = ""
}

// 初始化 by数据结构
func (s *UserDataStatistics) InitByData(data interface{}) error {
	return stack.SimpleCopyProperties(s, data)
}

// 初始化 by数据库
func (s *UserDataStatistics) InitFormDB(entityId uint32, tDBConnect *mongodb.DBConnect) bool {
	if tDBConnect == nil {
		return false
	}
	if s.CollectionName == "" || entityId <= 0 {
		return false
	}

	err := tDBConnect.GetData(s.CollectionName, "EntityId", entityId, s)
	return err == nil
}

func (s *UserDataStatistics) GetDataByEntityId(entityId uint32, tDBConnect *mongodb.DBConnect) *UserDataStatistics {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" || entityId <= 0 {
		return nil
	}
	resp := &UserDataStatistics{}
	err := tDBConnect.GetData(s.CollectionName, "EntityId", entityId, resp)
	if err != nil {
		return nil
	}
	return resp
}

// 插入数据库
func (s *UserDataStatistics) Insert(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(s.CollectionName, s)
}

// 保存致数据库
func (s *UserDataStatistics) Save(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection.UserDataStatistics--Save--tDBConnect == nil")
	}
	return tDBConnect.SaveData(s.CollectionName, "_id", s.ObjID, s)
}
