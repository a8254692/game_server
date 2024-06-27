package logic

import (
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
	"sync"
)

var DataStatisticsMgr _DataStatisticsMgr

type _DataStatisticsMgr struct {
	userStatisticsList map[uint32]*DataStatistics

	lock sync.RWMutex
}

func (s *_DataStatisticsMgr) Init() {
	s.userStatisticsList = make(map[uint32]*DataStatistics)

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_User_Statistics_Sync_Response), reflect.ValueOf(s.OnUserLoginSyncData))
}

//func (s *_DataStatisticsMgr) InitUserData(id uint32) *DataStatistics {
//	if id <= 0 {
//		return nil
//	}
//	data := s.GetUserStatisticsByID(id)
//	if data == nil {
//		data = NewDataStatistics(id)
//		s.addUserStatistics(id, data)
//	}
//	return data
//}

func (s *_DataStatisticsMgr) GetUserStatisticsByID(id uint32) *DataStatistics {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if id <= 0 {
		return nil
	}
	return s.userStatisticsList[id]
}

func (s *_DataStatisticsMgr) DelUserStatistics(id uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if id <= 0 {
		return
	}
	delete(s.userStatisticsList, id)
}

func (s *_DataStatisticsMgr) addUserStatistics(id uint32, data *DataStatistics) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if id <= 0 {
		return
	}
	s.userStatisticsList[id] = data
}

func (s *_DataStatisticsMgr) OnUserLoginSyncData(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InUserDataStatisticsData{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_DataStatisticsMgr--OnUserLoginSyncData--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityId <= 0 {
		log.Waring("-->logic--_DataStatisticsMgr--OnUserLoginSyncData--req.EntityId <= 0")
		return
	}

	data := &DataStatistics{
		entityId:           req.EntityId,
		accumulateGold:     req.AccumulateGold,
		accumulateGoal:     req.AccumulateGoal,
		oneCueClear:        req.OneCueClear,
		incrBindNum:        req.IncrBindNum,
		c8PlayNum:          req.C8PlayNum,
		c8WinNum:           req.C8WinNum,
		c8EscapeNum:        req.C8EscapeNum,
		c8ContinuousWin:    req.C8ContinuousWin,
		c8MaxContinuousWin: req.C8MaxContinuousWin,
		c8DoubleGoalNum:    req.C8DoubleGoalNum,
		c8ThreeGoalNum:     req.C8ThreeGoalNum,
	}
	s.addUserStatistics(req.EntityId, data)

	return
}
