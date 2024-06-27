package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"errors"
)

// 创建一个对战用户对象
func NewBattlePlayer(entityID uint32, roomID uint32, roomLevel uint32, isRobot bool) (p *BattlePlayer, err error) {
	p = new(BattlePlayer)
	p.roomID = roomID
	p.roomLevel = roomLevel
	p.seatID = 0
	p.entityID = entityID
	p.isRobot = isRobot
	p.isWinner = false
	p.currentExpLevel = 0
	p.currentExp = 0
	p.settleIncrExp = 0
	p.settleIncrLv = 0
	p.peakRankLv = 1
	p.peakRankExp = 0
	p.settleIncrPeakRankType = 0
	p.settleIncrPeakRankExp = 0
	p.settleIncrPeakRankLv = 1
	p.gold = 0
	p.settleIncrGoldType = 0
	p.settleIncrGold = 0
	p.winAdditionalGold = 0
	p.vipAddMoneyRate = 0
	p.isCapping = false
	p.isOpenExpAddition = false
	p.isOpenGoldAddition = false
	p.isOpenGoldReserve = false
	//p.expAwards		 		  	= 0
	//	p.props            			= 0
	//	p.interactiveIds				= 0

	//统计相关数据
	p.doubleGoalTimes = 0
	p.threeGoalTimes = 0
	p.continuousRoundNum = 0
	p.maxContinuousRoundNum = 0
	p.cumulativeGoalsNum = 0
	p.isOneCueClear = false
	p.battingStyleStatistic = make(map[uint32]uint32)

	//连接和状态信息
	p.status = 0
	p.isOffLine = false

	return
}

type BattlePlayer struct {
	roomID                 uint32 //房间ID
	roomLevel              uint32 //房间类型
	seatID                 uint8  //玩家在对战上的座位号
	entityID               uint32
	isRobot                bool    //是否为机器人
	isWinner               bool    //是否是胜利者
	settleType             uint32  //结算类型  正常结算/投降结算
	currentExpLevel        uint32  //当前等级
	currentExp             uint32  //当前经验值
	settleIncrExp          uint32  //结算变动的经验值
	settleIncrLv           uint32  //结算变动的经验等级
	peakRankLv             uint32  //当前赛季等级
	peakRankExp            uint32  //当前赛季经验
	settleIncrPeakRankLv   uint32  //结算变动的赛季等级
	settleIncrPeakRankType uint32  //结算变动的赛季经验增减类型 1增加 2减少
	settleIncrPeakRankExp  uint32  //结算变动的赛季经验
	gold                   uint32  //当前金币数额
	settleIncrGoldType     uint32  //结算变动的金币增减类型 1增加 2减少
	settleIncrGold         uint64  //结算变动的金币数
	winAdditionalGold      uint64  //玩家赢牌时额外获取的奖励（使用"金币加成卡"...额外获得的奖励）
	vipAddMoneyRate        float32 //vip赢的时候，额外获取的金币奖励比例。
	isCapping              bool    //玩家是否赢到的封顶值
	isOpenExpAddition      bool    //玩家是否开启了经验加成
	isOpenGoldAddition     bool    //玩家是否开启了金币加成
	isOpenGoldReserve      bool    //玩家是否开启了金币保留
	//expAwards		 *protoGame.LevelUpgradeInfo //经验等级-玩家在本局结束之后升级奖励信息
	//props            []*protoProps.PropsInfo //玩家持有的道具信息
	//interactiveIds	[]protoGame.InteractiveInfo 				//玩家所拥有的互动表情id以及数量信息

	showWinSettleGold         uint64 //胜利展示的金币数
	showTransporterSettleGold uint64 //失败展示的金币数

	//统计相关数据
	doubleGoalTimes       uint32            //二连的次数
	threeGoalTimes        uint32            //三连的次数
	continuousRoundNum    uint32            //连续回合数(会清零)
	maxContinuousRoundNum uint32            //最大连续回合数(判断一杆清台)
	cumulativeGoalsNum    uint32            //累计进球数量
	maxOneCueGoal         uint32            //最大一杆进球数量
	isOneCueClear         bool              //是否是一杆清台
	battingStyleStatistic map[uint32]uint32 //击球风格统计

	//连接和状态信息
	status    uint8 //用户状态用
	isOffLine bool  //是否掉线的标记
}

// 同步用户信息至对战用户信息
//func (s *BattlePlayer) SyncEntityToBattlePlayer(e entity.Entity) error {
//	entityPlayer := e.(*entity.EntityPlayer)
//	if entityPlayer == nil {
//		return errors.New("-->BattlePlayer--SyncEntityToBattlePlayer--IsNil")
//	}
//
//	s.entityID = entityPlayer.EntityID
//	s.isRobot = entityPlayer.IsRobot
//	return nil
//}

// 清理玩家对战数据并写记录
func (s *BattlePlayer) Clear() {
	if s.isRobot {
		return
	}
	//TODO： 清理对战玩家个人对局信息，记录个人对局日志

	return
}

// 设置胜利状态
func (s *BattlePlayer) SetWinner() {
	s.isWinner = true
	return
}

// 设置结算类型
func (s *BattlePlayer) SetSettleType(t uint32) {
	s.settleType = t
	return
}

// 设置经验值
func (s *BattlePlayer) SetSettleExp(expNum uint32) {
	if expNum > 0 {
		s.settleIncrExp += expNum
	}
	return
}

// 设置巅峰赛经验值
func (s *BattlePlayer) SetSettlePeakRankExp(expNum uint32) {
	if expNum > 0 {
		s.settleIncrPeakRankExp += expNum
	}
	return
}

// 设置玩家的座位号
func (s *BattlePlayer) SetSeatId(seatId uint8) error {
	s.seatID = seatId
	return nil
}

// 设置结算的钱
func (s *BattlePlayer) SetSettleGold(num uint64) {
	if num > 0 {
		s.settleIncrGold += num
	}
	return
}

// 设置赛季经验结算增减类型
func (s *BattlePlayer) SetSettleIncrPeakRankType(settleType uint32) {
	s.settleIncrPeakRankType = settleType
	return
}

// 设置结算的钱增减类型
func (s *BattlePlayer) SetSettleIncrGoldType(settleType uint32) {
	s.settleIncrGoldType = settleType
	return
}

func (s *BattlePlayer) SetOffline() {
	s.isOffLine = true
	return
}

func (s *BattlePlayer) SetIsOneCueClear() {
	s.isOneCueClear = true
	return
}

func (s *BattlePlayer) ResetContinuousRoundNum() {
	s.continuousRoundNum = 0
	return
}

func (s *BattlePlayer) IncrContinuousRoundNum() uint32 {
	s.continuousRoundNum += 1
	//记录最大连续数
	if s.maxContinuousRoundNum < s.continuousRoundNum {
		s.maxContinuousRoundNum = s.continuousRoundNum
	}

	//处理二连杆的次数
	if s.continuousRoundNum == 3 {
		s.doubleGoalTimes += 1
	}
	//处理三连杆的次数
	if s.continuousRoundNum == 4 {
		s.threeGoalTimes += 1
	}

	return s.continuousRoundNum
}

func (s *BattlePlayer) SetMaxOneCueGoal(n uint32) {
	if n > 0 && n > s.maxOneCueGoal {
		s.maxOneCueGoal = n
	}
	return
}

func (s *BattlePlayer) SetBattingStyleStatistic(data map[uint32]uint32) {
	if len(data) <= 0 {
		return
	}

	for k, v := range data {
		s.battingStyleStatistic[k] += v
	}

	return
}

func (s *BattlePlayer) IncrCumulativeGoalsNum(n uint32) {
	if n > 0 {
		s.cumulativeGoalsNum += n
	}
	return
}

// 胜利展示的金币数
func (s *BattlePlayer) SetShowWinSettleGold(num uint64) {
	s.showWinSettleGold = num
	return
}

// 失败展示的金币数
func (s *BattlePlayer) SetShowTransporterSettleGold(num uint64) {
	s.showTransporterSettleGold = num
	return
}

// 是否是机器人
func (s *BattlePlayer) IsRobot() bool {
	return s.isRobot
}

func (s *BattlePlayer) GetSettleIncrPeakRankLv() uint32 {
	return s.settleIncrPeakRankLv
}

func (s *BattlePlayer) GetSettleIncrPeakRankType() uint32 {
	return s.settleIncrPeakRankType
}

// 结算展示的金币数
func (s *BattlePlayer) GetShowSettleGold() uint64 {
	var resp uint64

	if s.isWinner {
		resp = s.showWinSettleGold
	} else {
		resp = s.showTransporterSettleGold
	}

	return resp
}

// 结算玩家的数据
func (s *BattlePlayer) SettlePlayer() error {
	//先判断是否为机器人
	if !s.isRobot {
		tEntity := Entity.EmPlayer.GetEntityByID(s.entityID)
		if tEntity == nil {
			return errors.New("-->logic--BattlePlayer--SettlePlayer--tEntity == nil")
		}

		//结算时候先赋值(写日志使用)
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		s.gold = tEntityPlayer.NumGold
		s.currentExpLevel = tEntityPlayer.PlayerLv
		s.currentExp = tEntityPlayer.NumExp
		s.peakRankLv = tEntityPlayer.PeakRankLv
		s.peakRankExp = tEntityPlayer.PeakRankExp

		taskResult := consts.RESULT_TRANSPORT
		if s.isWinner {
			taskResult = consts.RESULT_VICTORY
		}
		taskParam := UpdateClubData{
			GameType:       0,
			RoomType:       s.roomLevel,
			Result:         uint32(taskResult),
			SettlementType: s.settleType,
			EntityID:       s.entityID,
		}

		buySource := GetResParam(consts.SYSTEM_ID_C8_BATTLE, consts.Reward)
		PropertyItems := make([]entity.PropertyItem, 0)
		if s.settleIncrExp > 0 {
			PropertyItems = append(PropertyItems, entity.PropertyItem{TableID: consts.LvExp, ItemValue: int32(s.settleIncrExp)})
		}

		if s.settleIncrGold > 0 {
			switch s.settleIncrGoldType {
			case consts.OPERATE_TYPE_INCR:
				PropertyItems = append(PropertyItems, entity.PropertyItem{TableID: consts.Gold, ItemValue: int32(s.settleIncrGold)})

				taskParam.Gold = uint32(s.settleIncrGold)
			case consts.OPERATE_TYPE_DECR:
				PropertyItems = append(PropertyItems, entity.PropertyItem{TableID: consts.Gold, ItemValue: int32(-s.settleIncrGold)})

				taskParam.Gold = 0
			}
		}

		var rLv uint32
		var rExp uint32
		var rExpShow uint32
		if s.settleIncrPeakRankExp > 0 {
			switch s.settleIncrPeakRankType {
			case consts.OPERATE_TYPE_INCR:
				rExp = s.peakRankExp + s.settleIncrPeakRankExp
				PropertyItems = append(PropertyItems, entity.PropertyItem{TableID: consts.PeakRankExp, ItemValue: int32(s.settleIncrPeakRankExp)})
			case consts.OPERATE_TYPE_DECR:
				if s.peakRankExp >= s.settleIncrPeakRankExp {
					rExp = s.peakRankExp - s.settleIncrPeakRankExp
				} else {
					rExp = 0
				}
				PropertyItems = append(PropertyItems, entity.PropertyItem{TableID: consts.PeakRankExp, ItemValue: int32(-s.settleIncrPeakRankExp)})
			}

			rLv = PeakRankExp.Exp2Level(rExp)
			//获取排位展示星数
			rExpShow = PeakRankExp.GetLvShowExp(rLv, rExp)
		}

		if len(PropertyItems) > 0 {
			//更新属性道具
			Player.UpdatePlayerRepeatedPropertyItem(s.entityID, PropertyItems, *buySource)
		}

		tEntityPlayer.SetBehaviorStatus(consts.BEHAVIOR_STATUS_ROOM)
		tEntityPlayer.SyncEntity(1)

		battleResult := consts.RESULT_TRANSPORT
		goldAddType := consts.OPERATE_TYPE_DECR
		if s.isWinner {
			battleResult = consts.RESULT_VICTORY
			goldAddType = consts.OPERATE_TYPE_INCR
		}

		resp := &gmsg.BattleSettlementSync{
			BattleResult:       uint32(battleResult),
			RoomID:             s.roomID,
			EntityID:           s.entityID,
			Exp:                s.settleIncrExp,
			Gold:               uint32(s.GetShowSettleGold()),
			GoldAddType:        gmsg.OperateType(goldAddType),
			SettlementType:     s.settleType,
			StarAddType:        gmsg.OperateType(s.GetSettleIncrPeakRankType()),
			ChangePeakRankStar: s.settleIncrPeakRankExp,
			PeakRankLv:         rLv,
			PeakRankStar:       rExpShow,
		}
		log.Info("-->logic--_BattleC8Mgr--SettlePlayer--结算消息--Resp:", resp)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_SettlementSync, resp, []uint32{s.entityID})

		ClubManager.UpdateClubFromGameResult(taskParam, *buySource)

		//开始处理统记数据
		statistics := DataStatisticsMgr.GetUserStatisticsByID(s.entityID)
		if statistics != nil {
			statistics.IncrC8PlayNum(1)
			//计算累计获胜次数
			if s.isWinner {
				statistics.IncrC8WinNum(1)
				statistics.IncrC8ContinuousWin(1)
			} else {
				statistics.ResetC8ContinuousWin()
				//输的人才算逃跑
				if s.settleType == consts.SETTLEMENT_TYPE_SURRENDER {
					statistics.IncrC8EscapeNum(1)
				}
			}

			//累计获得金币
			if s.settleIncrGold > 0 {
				if s.settleIncrGoldType == consts.OPERATE_TYPE_INCR {
					statistics.IncrAccumulateGold(uint32(s.settleIncrGold))
				}
			}
			//累计二连杆次数
			if s.doubleGoalTimes > 0 {
				statistics.IncrC8DoubleGoalNum(s.doubleGoalTimes)
			}
			//累计三连杆次数
			if s.threeGoalTimes > 0 {
				statistics.IncrC8ThreeGoalNum(s.threeGoalTimes)
			}
			//累计进球次数
			if s.cumulativeGoalsNum > 0 {
				statistics.IncrAccumulateGoal(s.cumulativeGoalsNum)
			}
			//一杆清台
			if s.isOneCueClear {
				statistics.IncrOneCueClear(1)
			}
			//最大连杆次数
			if s.maxContinuousRoundNum > 0 {
				statistics.SetC8MaxContinuousGoalNum(s.maxContinuousRoundNum)
			}
			//最大一杆进球数
			if s.maxOneCueGoal > 0 {
				statistics.SetC8MaxOneCueGoal(s.maxOneCueGoal)
			}
			statistics.SaveDataToDb()

			//更新成就
			ConditionalMr.SyncConditionalStatics(statistics, uint32(taskResult))

			Activity.BattleSettleNotice(s.entityID, s.isWinner)

			//TODO: 增加调用条件更新
			if len(s.battingStyleStatistic) > 0 {

			}
		}
	} else {
		//修改机器人行为状态
		MatchManager.SetRobotBehaviorStatus(consts.BEHAVIOR_STATUS_HALL, s.entityID, s.roomID)
	}

	return nil
}
