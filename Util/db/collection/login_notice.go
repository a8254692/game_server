package collection

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/stack"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type LoginNotice struct {
	CollectionName string        `bson:"-"`                                    // 数据集名称
	ObjID          bson.ObjectId `bson:"_id"`                                  //唯一ID
	LoginNoticeId  string        `bson:"LoginNoticeId" json:"login_notice_id"` //登录公告唯一ID
	Title          string        `bson:"Title" json:"title"`                   //标题
	Name           string        `bson:"Name" json:"name"`                     //登录公告名称
	Context        string        `bson:"Context" json:"context"`               //登录公告内容
	StartTime      int64         `bson:"StartTime" json:"start_time"`          //开始时间
	EndTime        int64         `bson:"EndTime" json:"end_time"`              //结束时间
	PlatformLimit  uint32        `bson:"PlatformLimit" json:"platform_limit"`  //平台限制
	VipLimit       uint32        `bson:"VipLimit" json:"vip_limit"`            //vip等级限制
	IsRelease      bool          `bson:"IsRelease" json:"is_release"`          //是否暂停
}

// 初始化 第一次
func (s *LoginNotice) InitByFirst(collectionName string) {
	s.CollectionName = collectionName
	s.ObjID = bson.NewObjectId()
}

// 获取ObjID
func (s *LoginNotice) GetObjID() string {
	return s.ObjID.String()
}

// 设置DBConnect
func (s *LoginNotice) SetDBConnect(collectionName string) {
	s.CollectionName = collectionName
}

// 清理实体
func (s *LoginNotice) Clear() {
	s.CollectionName = ""
}

// 初始化 by数据结构
func (s *LoginNotice) InitByData(data interface{}) error {
	return stack.SimpleCopyProperties(s, data)
}

// 初始化 by数据库
func (s *LoginNotice) InitFormDB(loginNoticeId string, tDBConnect *mongodb.DBConnect) bool {
	if tDBConnect == nil {
		return false
	}
	if s.CollectionName == "" || loginNoticeId == "" {
		return false
	}

	err := tDBConnect.GetData(s.CollectionName, "LoginNoticeId", loginNoticeId, s)
	return err == nil
}

func (s *LoginNotice) GetDataByLoginNoticeId(loginNoticeId string, tDBConnect *mongodb.DBConnect) *LoginNotice {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" || loginNoticeId == "" {
		return nil
	}
	resp := &LoginNotice{}
	err := tDBConnect.GetData(s.CollectionName, "LoginNoticeId", loginNoticeId, resp)
	if err != nil {
		return nil
	}
	return resp
}

func (s *LoginNotice) GetDataOfQuery(tDBConnect *mongodb.DBConnect) []LoginNotice {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" {
		return nil
	}

	resp := &[]LoginNotice{}
	query := bson.M{"IsRelease": true}
	err := tDBConnect.GetAll(s.CollectionName, query, nil, resp)
	if err != nil {
		return nil
	}
	return *resp
}

// 插入数据库
func (s *LoginNotice) Insert(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(s.CollectionName, s)
}

// 保存致数据库
func (s *LoginNotice) Save(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection--LoginNotice--Save--tDBConnect == nil")
	}
	return tDBConnect.SaveData(s.CollectionName, "_id", s.ObjID, s)
}
