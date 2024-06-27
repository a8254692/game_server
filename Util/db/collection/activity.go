package collection

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/stack"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type Activity struct {
	CollectionName string        `bson:"-"`             // 数据集名称
	ObjID          bson.ObjectId `bson:"_id"`           //唯一ID
	ActivityId     string        `bson:"ActivityId"`    //活动唯一ID
	TimeType       uint32        `bson:"TimeType"`      //时间类型
	StartTime      int64         `bson:"StartTime"`     //开始时间
	EndTime        int64         `bson:"EndTime"`       //结束时间
	AType          uint32        `bson:"AType"`         //活动类型
	SubType        uint32        `bson:"SubType"`       //活动主题类型
	ActivityName   string        `bson:"ActivityName"`  //活动名称
	CurrentState   uint32        `bson:"CurrentState"`  //活动状态
	PlatformLimit  uint32        `bson:"PlatformLimit"` //平台限制
	VipLimit       uint32        `bson:"VipLimit"`      //vip等级限制
	Config         string        `bson:"Config"`        //配置(json格式)
}

// 初始化 第一次
func (s *Activity) InitByFirst(collectionName string) {
	s.CollectionName = collectionName
	s.ObjID = bson.NewObjectId()
}

// 获取ObjID
func (s *Activity) GetObjID() string {
	return s.ObjID.String()
}

// 设置DBConnect
func (s *Activity) SetDBConnect(collectionName string) {
	s.CollectionName = collectionName
}

// 清理实体
func (s *Activity) Clear() {
	s.CollectionName = ""
}

// 初始化 by数据结构
func (s *Activity) InitByData(data interface{}) error {
	return stack.SimpleCopyProperties(s, data)
}

// 初始化 by数据库
func (s *Activity) InitFormDB(activityId string, tDBConnect *mongodb.DBConnect) bool {
	if tDBConnect == nil {
		return false
	}
	if s.CollectionName == "" || activityId == "" {
		return false
	}

	err := tDBConnect.GetData(s.CollectionName, "ActivityId", activityId, s)
	return err == nil
}

func (s *Activity) GetDataByActivityId(activityId string, tDBConnect *mongodb.DBConnect) *Activity {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" || activityId == "" {
		return nil
	}
	resp := &Activity{}
	err := tDBConnect.GetData(s.CollectionName, "ActivityId", activityId, resp)
	if err != nil {
		return nil
	}
	return resp
}

func (s *Activity) GetDataAfterTime(tDBConnect *mongodb.DBConnect) []Activity {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" {
		return nil
	}

	resp := &[]Activity{}
	query := bson.M{"IsRelease": true}
	err := tDBConnect.GetAll(s.CollectionName, query, nil, resp)
	if err != nil {
		return nil
	}
	return *resp
}

// 插入数据库
func (s *Activity) Insert(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(s.CollectionName, s)
}

// 保存致数据库
func (s *Activity) Save(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection--Activity--Save--tDBConnect == nil")
	}
	return tDBConnect.SaveData(s.CollectionName, "_id", s.ObjID, s)
}
