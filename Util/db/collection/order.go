package collection

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/stack"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type Order struct {
	CollectionName string        `bson:"-"`             // 数据集名称
	ObjID          bson.ObjectId `bson:"_id,omitempty"` // 唯一ID
	OrderSn        string        `bson:"OrderSn"`       // 订单编号
	EntityId       uint32        `bson:"EntityId"`      // 用户ID
	FinalPrice     uint32        `bson:"FinalPrice"`    // 订单最终金额
	CouponPrice    uint32        `bson:"CouponPrice"`   // 促销抵扣的金额
	OriginalPrice  uint32        `bson:"OriginalPrice"` // 原始的金额
	ItemId         string        `bson:"ItemId"`        // 商品id
	ItemName       string        `bson:"ItemName"`      // 商品名称
	ItemNum        uint32        `bson:"ItemNum"`       // 购买数量
	PayType        uint32        `bson:"PayType"`       // 支付方式：0->未支付 1->支付宝 2->微信
	SourceType     uint32        `bson:"SourceType"`    // 订单来源：0->app订单 1->其他订单
	AddVipExp      uint32        `bson:"AddVipExp"`     // 可以获得的VIP值
	Note           string        `bson:"Note"`          // 订单备注
	PaymentTime    string        `bson:"PaymentTime"`   // 支付时间
	TimeCreate     string        `bson:"TimeCreate"`    // 提交时间
	TimeUpdate     string        `bson:"TimeUpdate"`    // 修改时间
	TimeDelete     string        `bson:"TimeDelete"`    // 删除时间
	DeleteStatus   uint32        `bson:"DeleteStatus"`  // 删除状态：0->未删除 1->已删除
}

// 初始化 第一次
func (s *Order) InitByFirst(collectionName string, orderSn string) {
	s.CollectionName = collectionName
	s.ObjID = bson.NewObjectId()
	s.OrderSn = orderSn
	s.EntityId = 0
	s.FinalPrice = 0
	s.CouponPrice = 0
	s.OriginalPrice = 0
	s.ItemId = ""
	s.ItemName = ""
	s.ItemNum = 0
	s.PayType = 0
	s.SourceType = 0
	s.AddVipExp = 0
	s.Note = ""
	s.PaymentTime = ""
	s.TimeCreate = ""
	s.TimeUpdate = ""
	s.TimeDelete = ""
	s.DeleteStatus = 0
}

// 获取ObjID
func (s *Order) GetObjID() string {
	return s.ObjID.String()
}

// 设置DBConnect
func (s *Order) SetDBConnect(collectionName string) {
	s.CollectionName = collectionName
}

// 清理实体
func (s *Order) Clear() {
	s.CollectionName = ""
}

// 初始化 by数据结构
func (s *Order) InitByData(data interface{}) error {
	return stack.SimpleCopyProperties(s, data)
}

// 初始化 by数据库
func (s *Order) InitFormDB(orderSn string, tDBConnect *mongodb.DBConnect) bool {
	if tDBConnect == nil {
		return false
	}
	if s.CollectionName == "" || orderSn == "" {
		return false
	}

	err := tDBConnect.GetData(s.CollectionName, "OrderSn", orderSn, s)
	return err == nil
}

// 插入数据库
func (s *Order) Insert(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(s.CollectionName, s)
}

// 保存致数据库
func (s *Order) Save(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return errors.New("-->collection.Order--Save--tDBConnect == nil")
	}
	return tDBConnect.SaveData(s.CollectionName, "_id", s.ObjID, s)
}
