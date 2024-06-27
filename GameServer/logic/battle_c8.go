package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/xtimer"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// 球坐标信息
type C8BallCoordinateInfo struct {
	X float32 `json:"X"`
	Y float32 `json:"Y"`
	Z float32 `json:"Z"`
}

type C8BallInfo struct {
	BallId   uint32                ` json:"BallId"`
	Position *C8BallCoordinateInfo ` json:"Position"`
	Rotation *C8BallCoordinateInfo `json:"Rotation"`
}

type C8GoalsBallInfo struct {
	BallId          uint32 ` json:"BallId"`
	GoalsBallPocket uint32 `json:"GoalsBallPocket"` //进球球袋
}

// 创建一个对战信息
func NewBattleC8Info(roomID uint32) (b *BattleC8Info, err error) {
	b = new(BattleC8Info)
	b.roomID = roomID
	b.roomLevel = 0
	b.timerMg = xtimer.NewMpTimers() //创建定时器列表
	b.startTime = 0
	b.endTime = 0
	b.status = 0
	b.currenRoundEntityID = 0
	b.roundID = 0
	b.playerList = nil
	b.aLLBallsGoalsList = make([]C8GoalsBallInfo, 0)
	b.ballsSplittingList = make(map[uint32]uint32)
	b.entityViolation = make(map[uint32]uint32)

	b.tableFee = 0
	b.blind = 0
	b.incrBlindNum = 0
	b.incrBlindHis = make([]uint32, 0)

	b.currenRoundInfo = &RoundInfo{
		startTime:         time.Now().Unix(),
		endTime:           0,
		status:            0,
		roundID:           0,
		roundEntityID:     0,
		strength:          0,
		cueRotation:       C8BallCoordinateInfo{},
		firstBallCollider: 0,
		ballsGoalsList:    make([]C8GoalsBallInfo, 0),
		whiteBallPos:      C8BallCoordinateInfo{},
		tidBallsPosition:  make([]C8BallInfo, 0),
		endConfirmEntity:  make([]uint32, 0),
		isEnd:             false,
	}
	b.currenRoundHis = make([]*RoundInfo, 0)

	//对战创建之后就开启一个goroutine来处理对战上的各种消息。
	//go tb.Process(contexs.TODO())
	return
}

// 回合数据（一杆一回合）
type RoundInfo struct {
	startTime         int64                //回合开始时间
	endTime           int64                //回合结束时间
	status            uint8                //回合状态（从battle.proto的BilliardState中设置）
	roundID           uint32               //回合ID
	roundEntityID     uint32               //当前回合的EntityID
	strength          float32              //力度（以后要放到RoundInfo中）
	cueRotation       C8BallCoordinateInfo //球杆角度（以后要放到RoundInfo中）
	firstBallCollider uint32               //打到的第一个球
	ballsGoalsList    []C8GoalsBallInfo    //当前回合进球情况
	whiteBallPos      C8BallCoordinateInfo //白球设置的位置（当违规的时候）
	tidBallsPosition  []C8BallInfo         //当前回合桌面上球的位置信息
	isEnd             bool                 //当前回合是否结束
	vectorInfo        C8BallCoordinateInfo //加塞向量
	angle             uint32               //抬杆角度
	endConfirmEntity  []uint32             //回合结束收到消息的人
}

type BattleC8Info struct {
	timerMg   xtimer.MpTimersIntf //对战定时器
	startTime int64               //开局时间
	endTime   int64               //结束时间
	status    uint8               //状态 0.空闲 1.进行中 2.结算中  3.结算完成  4.已结束
	//ch  chan  						//通道：接收请求
	currenRoundEntityID uint32            //当前回合的EntityID
	roundID             uint32            //当前回合ID
	roomID              uint32            //房间ID
	roomLevel           uint32            //房间类型
	playerList          *BattlePlayerMgr  //房间人员列表
	aLLBallsGoalsList   []C8GoalsBallInfo //整局所有进球情况
	ballsSplittingList  map[uint32]uint32 //球的分边情况
	entityViolation     map[uint32]uint32 //连续违规次数

	currenRoundInfo *RoundInfo   //当前回合信息
	currenRoundHis  []*RoundInfo //历史回合信息

	tableFee       uint64   //台费
	blind          uint64   //盲注
	winExp         uint32   //胜利经验
	transporterExp uint32   //失败经验
	incrBlindNum   uint32   //成功加注次数
	incrBlindHis   []uint32 //成功加注记录（数组的值是申请加注的人）
}

// 回合切换时初始化对战信息
func (s *BattleC8Info) InitRound(roundID uint32, EntityID uint32) {
	s.currenRoundInfo = nil
	s.currenRoundInfo = &RoundInfo{
		startTime:         time.Now().Unix(),
		endTime:           0,
		status:            0,
		roundID:           roundID,
		roundEntityID:     EntityID,
		strength:          0,
		firstBallCollider: 0,
		ballsGoalsList:    make([]C8GoalsBallInfo, 0),
		tidBallsPosition:  make([]C8BallInfo, 0),
		endConfirmEntity:  make([]uint32, 0),
	}
	return
}

// 清理对战
func (s *BattleC8Info) Clear() {
	s.GameEnd()

	//TODO： 清理对局信息，记录对局日志

	s.playerList.Clear()
	s.playerList = nil
	return
}

// 设置当前的entityid
func (s *BattleC8Info) SetCurrenEntityID(entityID uint32) {
	s.currenRoundEntityID = entityID
	return
}

// 设置回合的entityid
func (s *BattleC8Info) SetRoundEntityID(entityID uint32) {
	s.currenRoundInfo.roundEntityID = entityID
	return
}

// 设置房间人员列表
func (s *BattleC8Info) SetEntityList(el map[uint32]entity.Entity) error {
	if len(el) <= 0 {
		return errors.New("-->logic--BattleC8Info--SetEntityList--el Is Empty")
	}

	playerMgr, err := NewBattlePlayerMgr()
	if err != nil {
		return errors.New(fmt.Sprintf("-->logic--BattleC8Info--SetEntityList--NewBattlePlayerMgr--Err:%s", err))
	}

	for _, v := range el {
		player := Entity.EmPlayer.GetEntityByID(v.GetEntityID())
		if player == nil {
			return errors.New(fmt.Sprintf("-->logic--BattleC8Info--SetEntityList--player == nil:%d", v))
		}
		entityPlayer := player.(*entity.EntityPlayer)
		if entityPlayer == nil || entityPlayer.EntityID <= 0 {
			return errors.New(fmt.Sprintf("-->logic--BattleC8Info--SetEntityList--entityPlayer == nil || entityPlayer.EntityID <= 0--%d", v))
		}

		playInfo, err := NewBattlePlayer(entityPlayer.EntityID, s.roomID, s.roomLevel, entityPlayer.IsRobot)
		if err != nil {
			return errors.New(fmt.Sprintf("-->logic--BattleC8Info--SetEntityList--NewBattlePlayer--Err:%s", err))
		}

		err = playerMgr.SetPlayer(playInfo)
		if err != nil {
			return errors.New(fmt.Sprintf("-->logic--BattleC8Info--SetEntityList--SetPlayer--Err:%s", err))
		}

		entityPlayer.SetBehaviorStatus(consts.BEHAVIOR_STATUS_BATTLE)
	}

	s.playerList = playerMgr

	return nil
}

// 设置进球情况
func (s *BattleC8Info) SetBallsGoalsList(C8GoalsBallInfo C8GoalsBallInfo) {
	if C8GoalsBallInfo.BallId != consts.BALLS_WHITE {
		s.aLLBallsGoalsList = append(s.aLLBallsGoalsList, C8GoalsBallInfo)
	}

	s.currenRoundInfo.ballsGoalsList = append(s.currenRoundInfo.ballsGoalsList, C8GoalsBallInfo)
	return
}

// 设置分边情况
func (s *BattleC8Info) SetBallsSplittingList(entityID uint32, splitting uint32) {
	s.ballsSplittingList[entityID] = splitting
	return
}

func (s *BattleC8Info) SetTidBallsPosition(tbs []C8BallInfo) {
	s.currenRoundInfo.tidBallsPosition = tbs
	return
}

// 设置房间类型
func (s *BattleC8Info) SetRoomLevel(lv uint32) {
	s.roomLevel = lv
	return
}

// 设置对战底分信息
func (s *BattleC8Info) SetBlind(blind uint64) {
	s.blind = blind
	return
}

// 设置胜利经验
func (s *BattleC8Info) SetWinExp(exp uint32) {
	s.winExp = exp
	return
}

// 设置失败经验
func (s *BattleC8Info) SetTransporterExp(exp uint32) {
	s.transporterExp = exp
	return
}

// 设置对战底分信息
func (s *BattleC8Info) SetTableFee(tableFee uint64) {
	s.tableFee = tableFee
	return
}

// 设置状态
func (s *BattleC8Info) SetTableStatus(status uint8) {
	s.status = status
	return
}

// 设置状态
func (s *BattleC8Info) SetRoundStatus(status uint8) {
	s.currenRoundInfo.status = status
	return
}

// 设置游戏开始时间
func (s *BattleC8Info) SetStartTime(tm int64) {
	s.startTime = tm
	return
}

// 设置对战的结束时间
func (s *BattleC8Info) SetEndTime(tm int64) {
	s.endTime = tm
	return
}

func (s *BattleC8Info) SetEntityOnlineInGame() {

}

func (s *BattleC8Info) SetEntityOnlineOutGame() {

}

// 设置加注次数
func (s *BattleC8Info) SetIncrBlindHis(entityID uint32) {
	s.incrBlindHis = append(s.incrBlindHis, entityID)
	return
}

// 向对战中发送消息
func (s *BattleC8Info) SetMsg(msg string) {
	//s.VInfo("消息进入channle，等待对战进行处理！ msg=%#v", msg)
	//s.ch <- msg
}

// 设置白球位置
func (s *BattleC8Info) SetCurrenCueRotation(cue C8BallCoordinateInfo) {
	s.currenRoundInfo.cueRotation = cue
	return
}

// 设置白球位置
func (s *BattleC8Info) SetCurrenStrength(strength float32) {
	s.currenRoundInfo.strength = strength
	return
}

// 设置白球位置
func (s *BattleC8Info) SetWhiteBallPos(ball C8BallCoordinateInfo) {
	s.currenRoundInfo.whiteBallPos = ball
	return
}

// 设置加塞向量
func (s *BattleC8Info) SetRoundVectorInfo(v C8BallCoordinateInfo) {
	s.currenRoundInfo.vectorInfo = v
	return
}

// 设置抬杆角度
func (s *BattleC8Info) SetRoundAngle(a uint32) {
	s.currenRoundInfo.angle = a
	return
}

// 设置回合历史列表
func (s *BattleC8Info) SetRoundHis() {
	s.currenRoundHis = append(s.currenRoundHis, s.currenRoundInfo)
	return
}

// 设置第一个进球
func (s *BattleC8Info) SetFirstBallCollider(ball uint32) {
	s.currenRoundInfo.firstBallCollider = ball
	return
}

// 设置对战的阶段值
func (s *BattleC8Info) IncrRound() uint32 {
	s.roundID++
	return s.roundID
}

// 设置加注次数
func (s *BattleC8Info) IncrBlindNum() {
	s.incrBlindNum++
	return
}

// 设置违规次数
func (s *BattleC8Info) IncrEntityViolationNum() uint32 {
	if s.currenRoundEntityID > 0 {
		s.entityViolation[s.currenRoundEntityID]++
	}
	return s.entityViolation[s.currenRoundEntityID]
}

// 重置违规次数
func (s *BattleC8Info) InitEntityViolationNum() {
	if s.currenRoundEntityID > 0 {
		s.entityViolation[s.currenRoundEntityID] = 0
	}
	return
}

func (s *BattleC8Info) SetRoundIsEnd() {
	s.currenRoundInfo.isEnd = true
	return
}

func (s *BattleC8Info) SetRoundEndConfirmId(e uint32) {
	s.currenRoundInfo.endConfirmEntity = append(s.currenRoundInfo.endConfirmEntity, e)
	return
}

// 获取对战状态
func (s *BattleC8Info) GetTableStatus() uint8 {
	return s.status
}

func (s *BattleC8Info) GetPlayerInfo(entityID uint32) *BattlePlayer {
	player, err := s.playerList.GetPlayerByID(entityID)
	if err != nil {
		return nil
	}

	return player
}

// 获取初始随机的球排序
func (s *BattleC8Info) GetInitBalls() []uint32 {
	arr := []uint32{2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15}

	//TODO:先屏蔽开局随机球逻辑
	//r := rand.New(rand.NewSource(time.Now().Unix()))
	//r.Shuffle(len(arr), func(i, j int) {
	//	arr[i], arr[j] = arr[j], arr[i]
	//})
	return arr
}

// 获取先手
func (s *BattleC8Info) GetStartCueEntityID() uint32 {
	var startCueEntityID uint32
	if s.currenRoundEntityID <= 0 {
		entityIDs := s.GetRoomALLEntityIDs()
		r := rand.New(rand.NewSource(time.Now().Unix()))
		randNum := r.Intn(2)
		if len(entityIDs) > randNum && entityIDs[randNum] > 0 {
			startCueEntityID = entityIDs[randNum]
			s.currenRoundEntityID = startCueEntityID
		}
	}

	return startCueEntityID
}

func (s *BattleC8Info) GetRoundId() uint32 {
	return s.roundID
}

func (s *BattleC8Info) GetRoundIsEnd() bool {
	return s.currenRoundInfo.isEnd
}

func (s *BattleC8Info) GetCurrenEntityID() uint32 {
	return s.currenRoundEntityID
}

func (s *BattleC8Info) GetAllBallsGoalsIdList() []uint32 {
	var list []uint32

	roundBallsGoalsList := s.aLLBallsGoalsList
	if len(roundBallsGoalsList) > 0 {
		for _, v := range roundBallsGoalsList {
			list = append(list, v.BallId)
		}
	}

	return list
}

func (s *BattleC8Info) GetEntityAllBallsGoalsList() []C8GoalsBallInfo {
	var list []C8GoalsBallInfo

	entityID := s.currenRoundEntityID
	if entityID <= 0 {
		return list
	}
	spl := s.ballsSplittingList[entityID]
	if spl <= 0 {
		return list
	}

	roundBallsGoalsList := s.aLLBallsGoalsList
	if len(roundBallsGoalsList) > 0 {
		for _, v := range roundBallsGoalsList {
			if spl == consts.BALLS_SPLITTING_SMALL && v.BallId >= 1 && v.BallId <= 7 {
				list = append(list, v)
			} else if spl == consts.BALLS_SPLITTING_BIG && v.BallId >= 9 && v.BallId <= 15 {
				list = append(list, v)
			}
		}
	}

	return list
}

func (s *BattleC8Info) GetCurrenRoundGoalsLastBall() C8GoalsBallInfo {
	var resp C8GoalsBallInfo
	goalsList := s.currenRoundInfo.ballsGoalsList
	if len(goalsList) > 0 {
		resp = goalsList[len(goalsList)-1]
	}

	return resp
}

// 获取自由球情况
func (s *BattleC8Info) GetCurrenRoundInfo() *RoundInfo {
	return s.currenRoundInfo
}

// 获取分边情况
func (s *BattleC8Info) GetBallsSplittingList() map[uint32]uint32 {
	return s.ballsSplittingList
}

// 获取当前回合进球列表
func (s *BattleC8Info) GetRoundBallsGoalsList() []C8GoalsBallInfo {
	return s.currenRoundInfo.ballsGoalsList
}

func (s *BattleC8Info) GetEntityViolationNum() uint32 {
	return s.entityViolation[s.currenRoundEntityID]
}

func (s *BattleC8Info) IsFirstRoundEndConfirmId() bool {
	return len(s.currenRoundInfo.endConfirmEntity) <= 0
}

func (s *BattleC8Info) CheckRoundEndConfirmId(e uint32) bool {
	var isIn bool

	if e <= 0 {
		return isIn
	}

	if len(s.currenRoundInfo.endConfirmEntity) > 0 {
		for _, v := range s.currenRoundInfo.endConfirmEntity {
			if v == e {
				isIn = true
			}
		}
	}

	return isIn
}

func (s *BattleC8Info) CheckRoundEndConfirmReady() bool {
	var isReady bool

	allIds := s.playerList.GetAllPlayerIds()
	if len(allIds) > 0 {
		checkMap := make(map[uint32]bool)
		for _, v := range allIds {
			if s.CheckRoundEndConfirmId(v) {
				checkMap[v] = true
			}
		}

		if len(checkMap) >= len(allIds) {
			isReady = true
		}
	}

	return isReady
}

func (s *BattleC8Info) CheckRoundDataStatistics(data map[uint32]uint32) {
	currenEntityID := s.GetCurrenEntityID()
	otherEntityID := s.GetRoomOtherEntityID(currenEntityID)

	currenPlayer := s.GetPlayerInfo(currenEntityID)
	if currenPlayer != nil {
		//处理连续回合数据
		crn := currenPlayer.IncrContinuousRoundNum()

		//处理进球数
		currenPlayer.IncrCumulativeGoalsNum(uint32(len(s.currenRoundInfo.ballsGoalsList)))

		currenPlayer.SetBattingStyleStatistic(data)

		//TODO: 一杆清台有bug,待测试
		//处理进球数一杆清台
		if crn > 7 {
			ballsGoalsList := s.GetEntityAllBallsGoalsList()
			if len(ballsGoalsList) >= 7 {
				//判断当前回合的列表是否进了黑球
				if s.IsInRoundBallsGoalsList(consts.BALLS_BLACK) {
					currenPlayer.SetIsOneCueClear()
				}
			}
		}
	}

	otherPlayer := s.GetPlayerInfo(otherEntityID)
	if otherPlayer != nil {
		//处理连续回合数据
		otherPlayer.ResetContinuousRoundNum()
	}

	//一杆进球数量
	currenPlayer.SetMaxOneCueGoal(uint32(len(s.GetEntityAllBallsGoalsList())))

	return
}

// 获取球是否在整局进球列表
func (s *BattleC8Info) IsInAllBallsGoalsList(ball uint32) bool {
	var isIn bool
	if ball < 0 {
		return isIn
	}

	if len(s.aLLBallsGoalsList) <= 0 {
		return isIn
	}

	for _, v := range s.aLLBallsGoalsList {
		if v.BallId == ball {
			isIn = true
		}
	}

	return isIn
}

// 获取球是否在当前回合进球列表
func (s *BattleC8Info) IsInRoundBallsGoalsList(ball uint32) bool {
	var isIn bool
	if ball < 0 {
		return isIn
	}

	if len(s.currenRoundInfo.ballsGoalsList) <= 0 {
		return isIn
	}

	for _, v := range s.currenRoundInfo.ballsGoalsList {
		if v.BallId == ball {
			isIn = true
		}
	}

	return isIn
}

func (s *BattleC8Info) IsSetTidBallsPosition() bool {
	var isSet bool
	if len(s.currenRoundInfo.tidBallsPosition) > 0 {
		isSet = true
	}
	return isSet
}

// 获取当前回合状态
func (s *BattleC8Info) GetRoundStatus() uint8 {
	return s.currenRoundInfo.status
}

// 获取第一个碰到的球是否违规
func (s *BattleC8Info) IsRoundFirstBallViolation() bool {
	var isLe bool

	first := s.currenRoundInfo.firstBallCollider
	if first <= 0 {
		isLe = true
		return isLe
	}

	if first == consts.BALLS_BLACK && len(s.GetEntityAllBallsGoalsList()) < 7 {
		isLe = true
		return isLe
	}

	entityID := s.currenRoundEntityID
	if entityID <= 0 {
		return isLe
	}
	spl := s.ballsSplittingList[entityID]
	if spl <= 0 {
		return isLe
	}

	if spl == consts.BALLS_SPLITTING_SMALL && first >= 9 && first <= 15 {
		isLe = true
	} else if spl == consts.BALLS_SPLITTING_BIG && first >= 1 && first <= 7 {
		isLe = true
	}

	return isLe
}

// 判断玩家的违规进球数
func (s *BattleC8Info) ViolationGoalsNum() uint8 {
	var isLeNum uint8
	entityID := s.currenRoundEntityID
	if entityID <= 0 {
		return isLeNum
	}

	spl := s.ballsSplittingList[entityID]
	if spl <= 0 {
		return isLeNum
	}

	roundBallsGoalsList := s.currenRoundInfo.ballsGoalsList
	if len(roundBallsGoalsList) > 0 {
		for _, v := range roundBallsGoalsList {
			if spl == consts.BALLS_SPLITTING_SMALL && v.BallId >= 9 && v.BallId <= 15 {
				isLeNum++
			} else if spl == consts.BALLS_SPLITTING_BIG && v.BallId >= 1 && v.BallId <= 7 {
				isLeNum++
			}
		}
	}

	return isLeNum
}

// 获取玩家合规进球数
func (s *BattleC8Info) ComplianceGoalsNum() uint8 {
	var isLeNum uint8
	entityID := s.currenRoundEntityID
	if entityID <= 0 {
		return isLeNum
	}

	spl := s.ballsSplittingList[entityID]
	if spl <= 0 {
		return isLeNum
	}

	roundBallsGoalsList := s.currenRoundInfo.ballsGoalsList
	if len(roundBallsGoalsList) > 0 {
		for _, v := range roundBallsGoalsList {
			if spl == consts.BALLS_SPLITTING_BIG && v.BallId >= 9 && v.BallId <= 15 {
				isLeNum++
			} else if spl == consts.BALLS_SPLITTING_SMALL && v.BallId >= 1 && v.BallId <= 7 {
				isLeNum++
			}
		}
	}

	return isLeNum
}

// 重新开始游戏
func (s *BattleC8Info) GameReStart() {

}

// 对战结束
func (s *BattleC8Info) GameEnd() {
	//设置游戏的结束时间
	s.SetEndTime(time.Now().Unix())
	return
}

func (s *BattleC8Info) GetRoomALLEntityIDs() []uint32 {
	entityIDs := make([]uint32, 0)
	pList, err := s.playerList.GetPlayerList()
	if err != nil {
		return entityIDs
	}

	if len(pList) <= 0 {
		return entityIDs
	}
	for k := range pList {
		entityIDs = append(entityIDs, k)
	}

	return entityIDs
}

func (s *BattleC8Info) GetRoomALLEntityIDsNoRobot() []uint32 {
	entityIDs := make([]uint32, 0)
	pList, err := s.playerList.GetPlayerList()
	if err != nil {
		return entityIDs
	}

	if len(pList) <= 0 {
		return entityIDs
	}

	for k, v := range pList {
		if v.IsRobot() {
			continue
		}
		entityIDs = append(entityIDs, k)
	}

	return entityIDs
}

func (s *BattleC8Info) IsRobot(entityID uint32) bool {
	player, err := s.playerList.GetPlayerByID(entityID)
	if err != nil {
		return false
	}

	return player.IsRobot()
}

// 获取房间内另外一个entityID
func (s *BattleC8Info) GetRoomOtherEntityID(entityID uint32) uint32 {
	var otherEntityID uint32
	if entityID <= 0 {
		return otherEntityID
	}

	entityIDs := s.GetRoomALLEntityIDs()
	if len(entityIDs) <= 0 {
		return otherEntityID
	}

	//获取切换后EntityID
	for _, v := range entityIDs {
		if v != entityID {
			otherEntityID = v
		}
	}

	return otherEntityID
}

// 获取胜利展示的金币数
func (s *BattleC8Info) GetShowWinSettleGold() uint64 {
	return (s.blind * 2) + (s.blind * uint64(s.incrBlindNum)) - s.tableFee
}

// 获取失败展示的金币数
func (s *BattleC8Info) GetShowTransporterSettleGold() uint64 {
	return s.blind + (s.blind * uint64(s.incrBlindNum))
}

// 结算对局
func (s *BattleC8Info) Settle(winner uint32, transporter uint32, settlementType uint32) error {
	//开始算钱
	incrGold := (s.blind - s.tableFee) + (s.blind * uint64(s.incrBlindNum))
	descGold := s.blind + (s.blind * uint64(s.incrBlindNum))

	//先结算胜利者的数据
	if winner > 0 {
		//设置人物经验相关
		winnerPlay := s.GetPlayerInfo(winner)

		winnerPlay.SetShowWinSettleGold(s.GetShowWinSettleGold())
		winnerPlay.SetShowTransporterSettleGold(s.GetShowTransporterSettleGold())

		winnerPlay.SetSettleType(settlementType)

		//设置胜负状态
		winnerPlay.SetWinner()

		//设置金币结算
		winnerPlay.SetSettleIncrGoldType(consts.OPERATE_TYPE_INCR)
		winnerPlay.SetSettleGold(incrGold)

		//设置经验结算
		winnerPlay.SetSettleExp(s.winExp)

		//设置巅峰赛经验结算
		winnerPlay.SetSettleIncrPeakRankType(consts.OPERATE_TYPE_INCR)
		winnerPlay.SetSettlePeakRankExp(consts.PEAK_RANK_SETTLEMENT_STAR_NUM)

		err := winnerPlay.SettlePlayer()
		if err != nil {
			log.Error("-->logic--BattleC8Info--Settle--winner--err:", err, winner)
			return err
		}
	}

	//结算输家的数据
	if transporter > 0 {
		transporterPlay := s.GetPlayerInfo(transporter)

		transporterPlay.SetShowWinSettleGold(s.GetShowWinSettleGold())
		transporterPlay.SetShowTransporterSettleGold(s.GetShowTransporterSettleGold())

		transporterPlay.SetSettleType(settlementType)
		//设置金币结算
		transporterPlay.SetSettleIncrGoldType(consts.OPERATE_TYPE_DECR)
		transporterPlay.SetSettleGold(descGold)

		//设置经验结算
		transporterPlay.SetSettleExp(s.transporterExp)

		//设置巅峰赛经验结算
		transporterPlay.SetSettleIncrPeakRankType(consts.OPERATE_TYPE_DECR)
		transporterPlay.SetSettlePeakRankExp(consts.PEAK_RANK_SETTLEMENT_STAR_NUM)

		err := transporterPlay.SettlePlayer()
		if err != nil {
			log.Error("-->logic--BattleC8Info--Settle--transporter--err:", err, transporter)
			return err
		}
	}

	return nil
}

// 清除定时器
func (s *BattleC8Info) DeleteAllTimer() error {
	err := s.timerMg.ClearAll()
	if err != nil {
		return err
	}
	return nil
}

//-----------------------------------------------------------------定时器相关-----------------------------------------------------------------

// 开启打球定时器
func (s *BattleC8Info) StartPlayBallCountdown() error {
	//针对整个对战的定时器
	duration := time.Duration(consts.BATTLE_PLAY_BALL_COUNTDOWN) * time.Second
	err := s.timerMg.StartOneTimer(consts.BATTLE_PLAY_BALL_COUNTDOWN_NAME, duration, s.playBallCountdownTimeout, 0)
	if err != nil {
		return err
	}
	return nil
}

// 删除打球定时器
func (s *BattleC8Info) DelPlayBallCountdown() error {
	//针对整个对战的定时器
	err := s.timerMg.CloseOneTimer(consts.BATTLE_PLAY_BALL_COUNTDOWN_NAME)
	if err != nil {
		return err
	}
	return nil
}

// 重置打球定时器
func (s *BattleC8Info) ResetPlayBallCountdown() error {
	//针对整个对战的定时器
	err := s.timerMg.CloseOneTimer(consts.BATTLE_PLAY_BALL_COUNTDOWN_NAME)
	if err != nil {
		return err
	}

	duration := time.Duration(consts.BATTLE_PLAY_BALL_COUNTDOWN) * time.Second
	err = s.timerMg.StartOneTimer(consts.BATTLE_PLAY_BALL_COUNTDOWN_NAME, duration, s.playBallCountdownTimeout, 0)
	if err != nil {
		return err
	}
	return nil
}

// 开启对战回合结束定时器
func (s *BattleC8Info) StartRoundEndCountdown() error {
	duration := time.Duration(consts.BATTLE_ROUND_END_COUNTDOWN) * time.Second
	err := s.timerMg.StartOneTimer(consts.BATTLE_ROUND_END_COUNTDOWN_NAME, duration, s.roundEndCountdownTimeout, 0)
	if err != nil {
		return err
	}
	return nil
}

// 删除对战回合结束定时器
func (s *BattleC8Info) DelRoundEndCountdown() error {
	err := s.timerMg.CloseOneTimer(consts.BATTLE_ROUND_END_COUNTDOWN_NAME)
	if err != nil {
		return err
	}
	return nil
}

func (s *BattleC8Info) playBallCountdownTimeout(i interface{}) {
	//开始可打球倒计时
	nowEntityID := s.currenRoundEntityID
	afterChangeEntityID := s.GetRoomOtherEntityID(s.currenRoundEntityID)

	//未击球违规
	if s.currenRoundInfo.firstBallCollider <= 0 && len(s.GetRoundBallsGoalsList()) <= 0 {
		violationNum := s.IncrEntityViolationNum()
		resp := &gmsg.BattleEntityViolationResponse{
			Code:         resp_code.CODE_SUCCESS,
			RoomID:       s.roomID,
			EntityID:     nowEntityID,
			IsViolation:  true,
			ViolationNum: violationNum,
		}

		log.Info("-->logic--BattleC8Info--playBallCountdownTimeout--BattleEntityViolationResponse:", resp)
		//广播同步消息
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_ViolationResponse, resp, []uint32{nowEntityID})
	}

	//判断违规次数
	if s.GetEntityViolationNum() >= consts.JUDG_TRANSPORT_NUM {
		s.startSettle(afterChangeEntityID, nowEntityID, consts.SETTLEMENT_TYPE_VIOLATION)
		return
	}

	/*********************************************开始切换回合信息****************************************/
	s.SetRoundHis()
	roundId := s.IncrRound()
	s.SetCurrenEntityID(afterChangeEntityID)
	s.InitRound(roundId, afterChangeEntityID)
	s.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Aiming))
	//倒计时删除重新开启一个
	err := s.ResetPlayBallCountdown()
	if err != nil {
		log.Waring("-->logic--_Battle--playBallCountdownTimeout--ResetPlayBallCountdown--err:", err)
		return
	}

	/*********************************************重新获取修改后的信息****************************************/

	roundInfo := s.GetCurrenRoundInfo()

	//开始桌面更新信息
	resp := &gmsg.BattleCountdownEndSync{}
	resp.RoundID = s.roundID
	resp.RoomID = s.roomID
	resp.CurrenRoundEntityID = afterChangeEntityID
	resp.RoundInfo = &gmsg.BattleRoundInfo{
		RoundID:             roundInfo.roundID,
		CurrenRoundEntityID: roundInfo.roundEntityID,
		State:               gmsg.BilliardState(roundInfo.status),
		WhiteDecimalVector3: &gmsg.BattleBallCoordinateInfo{
			X: roundInfo.whiteBallPos.X,
			Y: roundInfo.whiteBallPos.Y,
			Z: roundInfo.whiteBallPos.Z,
		},
		PotBalls: nil,
		Strength: roundInfo.strength,
		CueRotationInfo: &gmsg.BattleBallCoordinateInfo{
			X: roundInfo.cueRotation.X,
			Y: roundInfo.cueRotation.Y,
			Z: roundInfo.cueRotation.Z,
		},
	}
	resp.AllGoalsBalls = s.GetAllBallsGoalsIdList()

	targetEntityIDs := s.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--BattleC8Info--playBallCountdownTimeout--Resp:", resp, s.roomID, targetEntityIDs)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_CountdownEndSync, resp, targetEntityIDs)
	return
}

func (s *BattleC8Info) roundEndCountdownTimeout(i interface{}) {
	if s.CheckRoundEndConfirmReady() {
		return
	}

	var transporter uint32
	allIds := s.playerList.GetAllPlayerIds()
	if len(allIds) > 0 {
		for _, v := range allIds {
			if !s.CheckRoundEndConfirmId(v) {
				transporter = v
			}
		}
	}

	if transporter > 0 {
		winner := s.GetRoomOtherEntityID(transporter)
		//结算
		log.Info("-->logic--BattleC8Info--roundEndCountdownTimeout--Resp--win:", winner, "--transporter:", transporter)
		s.startSettle(winner, transporter, consts.SETTLEMENT_TYPE_OFFLINE)
		return
	}

	return
}

func (s *BattleC8Info) startSettle(winner uint32, transporter uint32, settlementType uint32) {
	//清除所有倒计时
	_ = s.DeleteAllTimer()

	//开始结算
	err := s.Settle(winner, transporter, settlementType)
	if err != nil {
		return
	}

	//设置对战状态
	s.SetTableStatus(consts.BATTLE_STATUS_SETTLE_OVER)

	_ = BattleC8Mgr.Clear(s.roomID)
	return
}
