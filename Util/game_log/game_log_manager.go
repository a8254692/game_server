package game_log

import (
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/log"
	"BilliardServer/Util/timer"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	SAVE_LOG_INTERVAL = 300
	MAX_LOG_COUNT     = 500
)

// 日志类型枚举
type GameLogType int

const (
	GameLogType_UserCreateLog    GameLogType = iota //玩家创建日志
	GameLogType_UserLoginLog                        //玩家登录日志
	GameLogType_UserLogoutLog                       //玩家登出日志z
	GameLogType_PlayerActionLog                     //玩家行为日志
	GameLogType_ConsumeLog                          //消耗日志
	GameLogType_Production                          //产出日志
	GameLogType_OnlineRecordLog                     //在线数据
	GameLogType_OnlineCurveLog                      //最高在线曲线数据
	GameLogType_ChatLog                             //聊天记录
	GameLogType_DeviceIdLog                         //设备ID记录
	GameLogType_FirstRechargeLog                    //玩家首充日志
	GameLogType_RechargeLog                         //玩家充值日志
)

// 日志数据集名称
var GameLogTableName = []string{
	"user_create_log",
	"user_login_log",
	"user_logout_log",
	"player_action_log",
	"consume_log",
	"production_log",
	"online_record_log",
	"online_curve_log",
	"chat_log",
	"device_id_log",
	"first_recharge_log",
	"recharge_log",
}

type GameLogList struct {
	LogType GameLogType     //类型
	List    []interface{}   //记录的日志
	GLMgr   *GameLogManager //日志管理器
	LstTime time.Time       //上次保存的时间
}

// 初始化
func (this *GameLogList) Init(logType GameLogType, mgr *GameLogManager) {
	this.LogType = logType
	this.List = make([]interface{}, 0)
	this.GLMgr = mgr
}

// 增加一个日志
func (this *GameLogList) Add(logs ...interface{}) {
	this.List = append(this.List, logs...)

	if !this.IsNeedSave() {
		return
	}

	this.Save(false)
}

// 是否需要保存
func (this *GameLogList) IsNeedSave() bool {
	if this.LogType == GameLogType_ChatLog {
		return true
	}
	if this.LogType == GameLogType_DeviceIdLog {
		return true
	}

	if len(this.List) >= 300 {
		return true
	}

	return false
}

// 保存
func (this *GameLogList) Save(isTimer bool) {
	if len(this.List) <= 0 {
		return
	}

	logTypeInt := int(this.LogType)
	if len(GameLogTableName) <= logTypeInt {
		log.Error("日志数据集名称错误:", logTypeInt)
		return
	}
	now := time.Now()
	if !isTimer && now.UnixNano()-this.LstTime.UnixNano() <= 1e8 {
		log.Error("连续保存时间间隔过短，日志类型：", this.LogType, "，当前日志长度：", len(this.List))
		return
	}

	dbCollectionName := GameLogTableName[logTypeInt]

	//保存
	if this.LogType == GameLogType_ChatLog {
		//为了方便查找，聊天就一个ID就好
		this.GLMgr.LogDB.InsertDataAsync(dbCollectionName, dbCollectionName, nil, nil, this.List...)
	} else if this.LogType == GameLogType_DeviceIdLog {
		//记录设备ID拦截信息
		this.GLMgr.LogDB.InsertDataAsync(dbCollectionName, dbCollectionName, nil, nil, this.List...)
	} else {
		this.GLMgr.LogDB.InsertDataAsync(this.GLMgr.DBName, dbCollectionName, nil, nil, this.List...)
	}

	this.List = make([]interface{}, 0)

	this.LstTime = time.Now()
}

type GameLogManager struct {
	LogDB       *mongodb.DialContext         //日志库
	DBName      string                       //日志库名(跟游戏库名字一样)
	CacheLogMap map[GameLogType]*GameLogList //需要保存的日志
}

var GGameLogManager GameLogManager

// 初始化日志管理器
func (this *GameLogManager) Init(logDB *mongodb.DialContext, dbName string) {
	this.LogDB = logDB
	this.DBName = dbName
	this.CacheLogMap = make(map[GameLogType]*GameLogList)

	timer.AddTimer(this, "GLSave", 30*1000, true)
}

// 日更的时候定期清理日志
// 避免清理日志的时间一样，导致数据库有压力，需要通过服务器ID进行
func (this *GameLogManager) CleanCache() {
	//活动的日志数量不多，可能有用，先不清除

	//清理时间：2年
	overdueTime := time.Now().AddDate(0, 3, 0)

	//清除消耗日志
	this.LogDB.RemoveDataByTimeAsync(this.DBName, "ConsumeLog", "Time", "$lt", overdueTime, nil, nil)

	//清除登录日志
	this.LogDB.RemoveDataByTimeAsync(this.DBName, "LoginLog", "Time", "$lt", overdueTime, nil, nil)

	//清除产出日志
	this.LogDB.RemoveDataByTimeAsync(this.DBName, "Production", "Time", "$lt", overdueTime, nil, nil)

	//清除玩家行为日志
	this.LogDB.RemoveDataByTimeAsync(this.DBName, "PlayerActionLog", "Time", "$lt", overdueTime, nil, nil)
}

// 增加日志
func (this *GameLogManager) AddLog(logType GameLogType, logs ...interface{}) {

	gll := this.CacheLogMap[logType]
	if gll == nil {
		gll = new(GameLogList)
		gll.Init(logType, this)
		this.CacheLogMap[logType] = gll
	}

	gll.Add(logs...)
}

// 定时保存
func (this *GameLogManager) GLSave() {
	for _, gameLogList := range this.CacheLogMap {
		gameLogList.Save(true)
	}
}

// 获取聊天日志
func (this *GameLogManager) GetChatLog(searchValue interface{}, count int, i interface{}, sortFieldName ...string) {
	dbCollectionName := GameLogTableName[GameLogType_ChatLog]
	if err := this.LogDB.GetLimitDataAndSort(dbCollectionName, dbCollectionName, count, searchValue, i, sortFieldName...); err != nil {
		log.Error("加载聊天日志失败!err=", err.Error())
	}

	return
}

// 删除聊天日志
func (this *GameLogManager) DelChatLog(query bson.M) {
	dbCollectionName := GameLogTableName[GameLogType_ChatLog]
	if err := this.LogDB.RemoveAllByQuery(dbCollectionName, dbCollectionName, query); err != nil {
		log.Error("删除聊天日志失败!err=", err.Error())
	}
}
