package collection

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/stack"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type PointsMallData struct {
	CollectionName       string                `bson:"-"`                                                  // 数据集名称
	ObjID                bson.ObjectId         `bson:"_id"`                                                //唯一ID
	PointsMallId         string                `bson:"PointsMallId" json:"points_mall_id"`                 //积分商城唯一ID
	ConfigId             string                `bson:"ConfigId" json:"config_id"`                          //配置对应唯一ID
	Name                 string                `bson:"Name" json:"name"`                                   //名称
	StartTime            int64                 `bson:"StartTime" json:"start_time"`                        //开始时间
	EndTime              int64                 `bson:"EndTime" json:"end_time"`                            //结束时间
	RewardList           []entity.RewardEntity `bson:"RewardList" json:"reward_list"`                      //奖品
	Resources            string                `bson:"Resources" json:"resources"`                         //客户端资源地址
	LimitNum             uint32                `bson:"LimitNum" json:"limit_num"`                          //兑换总数限制
	ExchangeAmount       uint32                `bson:"ExchangeAmount" json:"exchange_amount"`              //兑换金额
	ExchangeCurrencyType uint32                `bson:"ExchangeCurrencyType" json:"exchange_currency_type"` //兑换代币类型
	ExchangeMaxNum       uint32                `bson:"ExchangeMaxNum" json:"exchange_max_num"`             //每人最大兑换数量
	IsRelease            bool                  `bson:"IsRelease" json:"is_release"`                        //是否发布
	RedeemedNum          uint32                `bson:"RedeemedNum" json:"redeemed_num"`                    //已兑换数量
}

// 初始化 第一次
func (s *PointsMallData) InitByFirst(collectionName string) {
	s.CollectionName = collectionName
	s.ObjID = bson.NewObjectId()
}

// 获取ObjID
func (s *PointsMallData) GetObjID() string {
	return s.ObjID.String()
}

// 设置DBConnect
func (s *PointsMallData) SetDBConnect(collectionName string) {
	s.CollectionName = collectionName
}

// 清理实体
func (s *PointsMallData) Clear() {
	s.CollectionName = ""
}

// 初始化 by数据结构
func (s *PointsMallData) InitByData(data interface{}) error {
	return stack.SimpleCopyProperties(s, data)
}

// 初始化 by数据库
func (s *PointsMallData) InitFormDB(pointsMallDataId string, tDBConnect *mongodb.DBConnect) bool {
	if tDBConnect == nil {
		return false
	}
	if s.CollectionName == "" || pointsMallDataId == "" {
		return false
	}

	err := tDBConnect.GetData(s.CollectionName, "PointsMallDataId", pointsMallDataId, s)
	return err == nil
}

func (s *PointsMallData) GetDataByPointsMallDataId(pointsMallDataId string, tDBConnect *mongodb.DBConnect) *PointsMallData {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" || pointsMallDataId == "" {
		return nil
	}
	resp := &PointsMallData{}
	err := tDBConnect.GetData(s.CollectionName, "PointsMallDataId", pointsMallDataId, resp)
	if err != nil {
		return nil
	}
	return resp
}

func (s *PointsMallData) GetDataOfQuery(tDBConnect *mongodb.DBConnect) []PointsMallData {
	if tDBConnect == nil {
		return nil
	}
	if s.CollectionName == "" {
		return nil
	}

	resp := &[]PointsMallData{}
	query := bson.M{"IsRelease": true}
	err := tDBConnect.GetAll(s.CollectionName, query, nil, resp)
	if err != nil {
		return nil
	}
	return *resp
}

// 插入数据库
func (s *PointsMallData) Insert(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(s.CollectionName, s)
}

// 保存致数据库
func (s *PointsMallData) Save(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection--PointsMallData--Save--tDBConnect == nil")
	}
	return tDBConnect.SaveData(s.CollectionName, "_id", s.ObjID, s)
}
