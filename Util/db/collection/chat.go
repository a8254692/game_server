package collection

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/stack"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type PrivateChatEntity struct {
	EntityID   uint32
	PlayerName string
	Sex        uint32
	PlayerIcon uint32
	IconFrame  uint32
	VipLv      uint32
}

type Chat struct {
	CollectionName  string              `bson:"-"`               // 数据集名称
	ObjID           bson.ObjectId       `bson:"_id,omitempty"`   // 唯一ID
	EntityId        uint32              `bson:"EntityId"`        //用户id
	FriendsChatList []PrivateChatEntity `bson:"FriendsChatList"` //私聊列表
}

// 初始化 第一次
func (s *Chat) InitByFirst(collectionName string, entityId uint32) {
	s.CollectionName = collectionName
	s.ObjID = bson.NewObjectId()
	s.EntityId = entityId
	s.FriendsChatList = make([]PrivateChatEntity, 0)
}

// 获取ObjID
func (s *Chat) GetObjID() string {
	return s.ObjID.String()
}

// 设置DBConnect
func (s *Chat) SetDBConnect(collectionName string) {
	s.CollectionName = collectionName
}

// 清理实体
func (s *Chat) Clear() {
	s.CollectionName = ""
}

// 初始化 by数据结构
func (s *Chat) InitByData(data interface{}) error {
	return stack.SimpleCopyProperties(s, data)
}

// 初始化 by数据库
func (s *Chat) InitFormDB(entityId uint32, tDBConnect *mongodb.DBConnect) bool {
	if tDBConnect == nil {
		return false
	}
	if s.CollectionName == "" || entityId <= 0 {
		return false
	}

	err := tDBConnect.GetData(s.CollectionName, "EntityId", entityId, s)
	return err == nil
}

func (s *Chat) GetDataByEntityId(entityId uint32, tDBConnect *mongodb.DBConnect) *Chat {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" || entityId <= 0 {
		return nil
	}
	resp := &Chat{}
	err := tDBConnect.GetData(s.CollectionName, "EntityId", entityId, resp)
	if err != nil {
		return nil
	}
	return resp
}

// 插入数据库
func (s *Chat) Insert(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(s.CollectionName, s)
}

// 保存致数据库
func (s *Chat) Save(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection.Chat--Save--tDBConnect == nil")
	}
	return tDBConnect.SaveData(s.CollectionName, "_id", s.ObjID, s)
}

// 保存致数据库
func (s *Chat) RemoveAllData(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection.Chat--RemoveAllData--tDBConnect == nil")
	}

	tDBConnect.RemoveAllData(s.CollectionName)

	return nil
}
