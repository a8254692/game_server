package entity

import "BilliardServer/Util/db/mongodb"

type Entity interface {
	GetObjID() string
	GetEntityID() uint32
	SetDBConnect(collectionName string)
	//初始化 第一次
	InitByFirst(collectionName string, tEntityID uint32)
	//初始化 by数据结构
	InitByData(eData interface{})
	//初始化 by数据库
	InitFormDB(tEntityID uint32, tDBConnect *mongodb.DBConnect) (bool, error)
	//插入数据库
	InsertEntity(tDBConnect *mongodb.DBConnect) error
	//保存致数据库
	SaveEntity(tDBConnect *mongodb.DBConnect)
	//清理实体
	ClearEntity()
	//同步实体
	SyncEntity(typeSave uint32)
}

// 实体同步事件
type EntityEvent struct {
	TypeSave   uint32 //保存类型 0定时存，1马上存
	TypeEntity uint32 //实体类型 0-acc,1-player,3-----
	Entity     Entity //实体接口
}

// 部件名称，做为循环始化mongo数据集使用
type Unit struct {
	main   string
	acc    string
	player string
	bag    string
	hose   string
	club   string
}

// 部件名称常量,做为获得指定数据集使用，与Unit完全对应
const (
	UnitSyncentity string = "SyncEntity"
	UnitMain       string = "entity"
	UnitAcc        string = "acc"
	UnitPlayer     string = "player"
	UnitBag        string = "bag"
	UnitHose       string = "hose"
	UnitClub       string = "club"
)
const (
	EntityTypeMain uint32 = iota
	EntityTypeAcc
	EntityTypePlayer
	EntityTypeBag
	EntityTypeHose
	EntityTypeClub
)
