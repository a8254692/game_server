package logic

import (
	"BilliardServer/DBServer/initialize/consts"
	"BilliardServer/Util/db/collection"
	"BilliardServer/Util/log"
	"reflect"

	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
)

type _Statistics struct {
}

// 统记相关
var Statistics _Statistics

func (s *_Statistics) Init() {
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_User_Statistics_Db_Save_Request), reflect.ValueOf(s.OnUserDataStatisticsSaveRequest))
}

func (s *_Statistics) SyncUserDataStatisticsToGame(entityId uint32) {
	if entityId <= 0 {
		return
	}

	statistics := new(collection.UserDataStatistics)
	statistics.SetDBConnect(consts.COLLECTION_STATISTICS)
	statistics.InitFormDB(entityId, DBConnect)

	resp := &gmsg.InUserDataStatisticsData{
		EntityId:           entityId,
		AccumulateGold:     statistics.AccumulateGold,
		AccumulateGoal:     statistics.AccumulateGoal,
		OneCueClear:        statistics.OneCueClear,
		IncrBindNum:        statistics.IncrBindNum,
		C8PlayNum:          statistics.C8PlayNum,
		C8WinNum:           statistics.C8WinNum,
		C8EscapeNum:        statistics.C8EscapeNum,
		C8ContinuousWin:    statistics.C8ContinuousWin,
		C8MaxContinuousWin: statistics.C8MaxContinuousWin,
		C8DoubleGoalNum:    statistics.C8DoubleGoalNum,
		C8ThreeGoalNum:     statistics.C8ThreeGoalNum,
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_User_Statistics_Sync_Response), resp, network.ServerType_Game)
	return
}

// 同步用户统记数据(游戏服->DB服)
func (s *_Statistics) OnUserDataStatisticsSaveRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InUserDataStatisticsData{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}
	if req.EntityId <= 0 {
		log.Waring("-->logic--_Statistics--OnSyncDataStatisticsRequest--req.EntityId <= 0")
		return
	}

	statistics := new(collection.UserDataStatistics)
	statistics.InitByFirst(consts.COLLECTION_STATISTICS, req.EntityId)
	err = statistics.InitByData(req)
	if err != nil {
		log.Waring("-->logic--_Statistics--OnSyncDataStatisticsRequest--InitByData--err:", err.Error())
		return
	}
	dbInfo := statistics.GetDataByEntityId(req.EntityId, DBConnect)
	if dbInfo != nil && dbInfo.EntityId > 0 && dbInfo.ObjID != "" {
		statistics.ObjID = dbInfo.ObjID
	}
	err = statistics.Save(DBConnect)
	if err != nil {
		log.Waring("-->logic--_Statistics--OnSyncDataStatisticsRequest--dbInsert--err:", err.Error())
		return
	}

	return
}
