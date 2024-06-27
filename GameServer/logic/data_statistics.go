package logic

import (
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/network"
)

type DataStatistics struct {
	entityId       uint32
	accumulateGold uint32 //累计获取的金币
	accumulateGoal uint32 //累计进球
	oneCueClear    uint32 //一杆清台
	incrBindNum    uint32 //加注次数

	c8PlayNum              uint32 //对局次数
	c8WinNum               uint32 //胜利次数
	c8EscapeNum            uint32 //逃跑次数
	c8ContinuousWin        uint32 //连胜次数(输了清零)
	c8MaxContinuousWin     uint32 //最大连胜次数
	c8DoubleGoalNum        uint32 //二连杆次数
	c8ThreeGoalNum         uint32 //三连杆次数
	c8MaxOneCueGoal        uint32 //最大一杆进球数
	c8MaxContinuousGoalNum uint32 //最大连杆次数
}

func (s *DataStatistics) SaveDataToDb() {
	req := &gmsg.InUserDataStatisticsData{
		EntityId:           s.entityId,
		AccumulateGold:     s.accumulateGold,
		AccumulateGoal:     s.accumulateGoal,
		OneCueClear:        s.oneCueClear,
		IncrBindNum:        s.incrBindNum,
		C8PlayNum:          s.c8PlayNum,
		C8WinNum:           s.c8WinNum,
		C8EscapeNum:        s.c8EscapeNum,
		C8ContinuousWin:    s.c8ContinuousWin,
		C8MaxContinuousWin: s.c8MaxContinuousWin,
		C8DoubleGoalNum:    s.c8DoubleGoalNum,
		C8ThreeGoalNum:     s.c8ThreeGoalNum,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_User_Statistics_Db_Save_Request), req, network.ServerType_DB)
	return
}

func (s *DataStatistics) C8SettleOver(n uint32) {
	if n <= 0 {
		return
	}
	s.accumulateGold += n
	return
}

func (s *DataStatistics) IncrAccumulateGold(n uint32) {
	if n > 0 {
		s.accumulateGold += n
	}
	return
}

func (s *DataStatistics) IncrAccumulateGoal(n uint32) {
	if n > 0 {
		s.accumulateGoal += n
	}
	return
}

func (s *DataStatistics) IncrOneCueClear(n uint32) {
	if n > 0 {
		s.oneCueClear += n
	}
	return
}

func (s *DataStatistics) IncrIncrBindNum(n uint32) {
	if n > 0 {
		s.incrBindNum += n
	}
	return
}

func (s *DataStatistics) IncrC8PlayNum(n uint32) {
	if n > 0 {
		s.c8PlayNum += n
	}
	return
}

func (s *DataStatistics) IncrC8WinNum(n uint32) {
	if n > 0 {
		s.c8WinNum += n
	}
	return
}

func (s *DataStatistics) IncrC8EscapeNum(n uint32) {
	if n > 0 {
		s.c8EscapeNum += n
	}
	return
}

func (s *DataStatistics) SetC8MaxOneCueGoal(n uint32) {
	if n > 0 && n > s.c8MaxOneCueGoal {
		s.c8MaxOneCueGoal = n
	}
	return
}

func (s *DataStatistics) SetC8MaxContinuousGoalNum(n uint32) {
	if n > 0 && n > s.c8MaxContinuousGoalNum {
		s.c8MaxContinuousGoalNum = n
	}
	return
}

func (s *DataStatistics) ResetC8ContinuousWin() {
	s.c8ContinuousWin = 0
	return
}

func (s *DataStatistics) IncrC8ContinuousWin(n uint32) {
	if n > 0 {
		s.c8ContinuousWin += n
		if s.c8MaxContinuousWin < s.c8ContinuousWin {
			s.c8MaxContinuousWin = s.c8ContinuousWin
		}
	}
	return
}

func (s *DataStatistics) IncrC8DoubleGoalNum(n uint32) {
	if n > 0 {
		s.c8DoubleGoalNum += n
	}
	return
}

func (s *DataStatistics) IncrC8ThreeGoalNum(n uint32) {
	if n > 0 {
		s.c8ThreeGoalNum += n
	}
	return
}

// 升级消息
func (s *DataStatistics) IsChangePlayerLv() {
	return
}

func (s *DataStatistics) GetAccumulateGold() uint32 {
	return s.accumulateGold
}

func (s *DataStatistics) GetAccumulateGoal() uint32 {
	return s.accumulateGoal
}
func (s *DataStatistics) GetOneCueClear() uint32 {
	return s.oneCueClear
}
func (s *DataStatistics) GetIncrBindNum() uint32 {
	return s.incrBindNum
}
func (s *DataStatistics) GetC8PlayNum() uint32 {
	return s.c8PlayNum
}

func (s *DataStatistics) GetC8WinNum() uint32 {
	return s.c8WinNum
}

func (s *DataStatistics) GetC8EscapeNum() uint32 {
	return s.c8EscapeNum
}

func (s *DataStatistics) GetC8ContinuousWin() uint32 {
	return s.c8ContinuousWin
}

func (s *DataStatistics) GetC8MaxContinuousWin() uint32 {
	return s.c8MaxContinuousWin
}

func (s *DataStatistics) GetC8DoubleGoalNum() uint32 {
	return s.c8DoubleGoalNum
}

func (s *DataStatistics) GetC8ThreeGoalNum() uint32 {
	return s.c8ThreeGoalNum
}

func (s *DataStatistics) GetC8MaxOneCueGoal() uint32 {
	return s.c8MaxOneCueGoal
}

func (s *DataStatistics) GetC8MaxContinuousGoalNum() uint32 {
	return s.c8MaxContinuousGoalNum
}
