package entity

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/event"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// 主要的实体对象，包含多个实体部件
type UnitPlayerBase struct {
	EntityID   uint32 `bson:"EntityID"`   //EntityID
	PlayerID   uint32 `bson:"PlayerID"`   //角色ID
	PlayerName string `bson:"PlayerName"` //角色名称
	PlayerLv   uint32 `bson:"PlayerLv"`   //角色等级
	Sex        uint32 `bson:"Sex"`        //用户性别
	TimeCreate string `bson:"TimeCreate"` //创建时间
}

// 主要的实体对象，包含多个实体部件
type EntityAcc struct {
	CollectionName string        `bson:"-"`             //数据集名称
	FlagChange     bool          `bson:"-"`             //是否被修改
	FlagKick       bool          `bson:"-"`             //被T标记
	ObjID          bson.ObjectId `bson:"_id"`           //唯一ID
	EntityID       uint32        `bson:"EntityID"`      //帐号ID
	AccUnique      string        `bson:"AccUnique"`     //(多服的)唯一帐号
	PassWord       string        `bson:"PassWord"`      //密码
	IsIPhone       bool          `bson:"IsIPhone"`      //是否苹果设备
	Platform       uint32        `bson:"Platform"`      //平台
	LoginPlatform  uint32        `bson:"LoginPlatform"` //登录平台
	Channel        uint32        `bson:"Channel"`       //渠道
	DeviceId       string        `bson:"DeviceId"`      //设备Id
	Machine        string        `bson:"Machine"`       //机型
	RemoteAddr     string        `bson:"RemoteAddr"`    //远端ip
	PackageName    string        `bson:"PackageName"`   //当前客户端包名
	Language       uint32        `bson:"Language"`      //玩家的语言标记
	State          uint32        `bson:"State"`         //帐号当前状态  1禁言 2封禁
	LoginFlag      bool          `bson:"LoginFlag"`     //登录状态
	TimeCreate     string        `bson:"TimeCreate"`    //创建时间
	TimeUpdate     string        `bson:"TimeUpdate"`    //更新时间
	TimeExit       string        `bson:"TimeExit"`      //退出时间

	ListPlayer []UnitPlayerBase `bson:"ListPlayer"` //角色列表
}

// 初始化 第一次
func (this *EntityAcc) InitByFirst(collectionName string, tEntityID uint32) {
	this.CollectionName = collectionName
	this.State = 0
	this.FlagChange = false
	this.FlagKick = false
	this.ObjID = bson.NewObjectId()
	this.EntityID = tEntityID
	this.AccUnique = ""
	this.PassWord = ""
	this.TimeCreate = tools.GetTimeByTimeStamp(time.Now().Unix())
	this.TimeUpdate = this.TimeCreate
	this.TimeExit = this.TimeCreate
	this.IsIPhone = false
	this.DeviceId = ""
	this.Machine = ""
	this.RemoteAddr = ""
	this.PackageName = ""
	this.Language = 0
	this.LoginFlag = false

	this.ListPlayer = make([]UnitPlayerBase, 0)
}

// 获取ObjID
func (this *EntityAcc) GetObjID() string {
	return this.ObjID.String()
}

// 获取ObjID
func (this *EntityAcc) GetEntityID() uint32 {
	return this.EntityID
}

// 设置DBConnect
func (this *EntityAcc) SetDBConnect(collectionName string) {
	this.CollectionName = collectionName
}

// 初始化 by数据结构
func (this *EntityAcc) InitByData(playerData interface{}) {
	stack.SimpleCopyProperties(this, playerData)
}

// 初始化 by数据库
func (this *EntityAcc) InitFormDB(tEntityID uint32, tDBConnect *mongodb.DBConnect) (bool, error) {
	if tDBConnect == nil {
		return false, errors.New("tDBConnect == nil")
	}
	err := tDBConnect.GetData(this.CollectionName, "EntityID", tEntityID, this)
	if err != nil {
		return false, err
	}

	return true, err
}

// 插入数据库
func (this *EntityAcc) InsertEntity(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(this.CollectionName, this)
}

// 保存致数据库
func (this *EntityAcc) SaveEntity(tDBConnect *mongodb.DBConnect) {
	if tDBConnect == nil {
		return
	}
	tDBConnect.SaveData(this.CollectionName, "_id", this.ObjID, this)
}

// 清理实体
func (this *EntityAcc) ClearEntity() {
	this.CollectionName = ""
}

// 同步实体
// typeSave: 0定时同步 1根据环境默认 2立即同步
func (this *EntityAcc) SyncEntity(typeSave uint32) {
	evEntity := new(EntityEvent)
	evEntity.TypeSave = typeSave
	evEntity.TypeEntity = EntityTypeAcc
	evEntity.Entity = this
	event.Emit(UnitSyncentity, evEntity)
}

func (this *EntityAcc) SetPlayerState(status uint32) {
	this.State = status
	return
}
