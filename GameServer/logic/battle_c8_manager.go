package logic

import (
	"BilliardServer/Common/resp_code"
	"BilliardServer/Util/log"
	"errors"
	"reflect"
	"sync"
	"time"

	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
)

// 所有对战的统一管理器
type _BattleC8Mgr struct {
	battleList map[uint32]*BattleC8Info
	lock       sync.RWMutex
}

var BattleC8Mgr _BattleC8Mgr

func (s *_BattleC8Mgr) Init() {
	//创建对战
	s.battleList = make(map[uint32]*BattleC8Info)

	//注册逻辑业务事件
	event.OnNet(gmsg.MsgTile_Battle_StartRequest, reflect.ValueOf(s.OnBattleStartRequest))
	event.OnNet(gmsg.MsgTile_Battle_AimingRequest, reflect.ValueOf(s.OnBattleAimingRequest))
	event.OnNet(gmsg.MsgTile_Battle_EnergyStorageRequest, reflect.ValueOf(s.OnBattleEnergyStorageRequest))
	event.OnNet(gmsg.MsgTile_Battle_StrokeBallRequest, reflect.ValueOf(s.OnBattleStrokeBallRequest))
	event.OnNet(gmsg.MsgTile_Battle_ObstructRequest, reflect.ValueOf(s.OnBattleObstructRequest))
	event.OnNet(gmsg.MsgTile_Battle_ObstructCueRequest, reflect.ValueOf(s.OnBattleObstructCueRequest))
	event.OnNet(gmsg.MsgTile_Battle_RoundEndRequest, reflect.ValueOf(s.OnRoundEndRequest))
	event.OnNet(gmsg.MsgTile_Battle_BallGoalRequest, reflect.ValueOf(s.OnBallGoalRequest))
	event.OnNet(gmsg.MsgTile_Battle_SurrenderRequest, reflect.ValueOf(s.OnSurrenderRequest))
	event.OnNet(gmsg.MsgTile_Battle_ApplyIncrBindRequest, reflect.ValueOf(s.OnApplyIncrBindRequest))
	event.OnNet(gmsg.MsgTile_Battle_FeedbackIncrBindRequest, reflect.ValueOf(s.OnFeedbackIncrBindRequest))
	event.OnNet(gmsg.MsgTile_Battle_FirstBallColliderRequest, reflect.ValueOf(s.OnBattleFirstBallColliderRequest))
	event.OnNet(gmsg.MsgTile_Battle_SetWhiteBallLocationIngRequest, reflect.ValueOf(s.OnBattleSetWhiteBallLocationIngRequest))
	event.OnNet(gmsg.MsgTile_Battle_SetWhiteBallLocationEndRequest, reflect.ValueOf(s.OnBattleSetWhiteBallLocationEndRequest))
	event.OnNet(gmsg.MsgTile_Battle_UserChartMsgRequest, reflect.ValueOf(s.OnUserChartMsgRequest))
}

func (s *_BattleC8Mgr) CheckEntityBattle(entityId uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if entityId <= 0 || len(s.battleList) <= 0 {
		return
	}

	for k, v := range s.battleList {
		if len(v.playerList.GetAllPlayerIds()) <= 0 {
			continue
		}

		for _, rv := range v.playerList.GetAllPlayerIds() {
			if entityId == rv {
				if v.GetTableStatus() < consts.BATTLE_STATUS_SETTLE_ING {
					otherEntityID := v.GetRoomOtherEntityID(entityId)
					s.startSettlement(k, otherEntityID, entityId, consts.SETTLEMENT_TYPE_SURRENDER)
				}
			}
		}
	}

	return
}

// 对战加入到对战管理器中
func (s *_BattleC8Mgr) SetBattleToMgr(bt *BattleC8Info) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.battleList[bt.roomID] = bt
	return nil
}

func (s *_BattleC8Mgr) Clear(roomID uint32) error {
	if roomID <= 0 {
		log.Waring("-->logic--_BattleC8Mgr--Clear--RoomID is empty")
		return nil
	}

	bt, err := s.getBattleByRoomID(roomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--Clear--getBattleByRoomID err:", err)
		return errors.New("-->logic--_BattleC8Mgr--Clear--Err")
	}

	//设置对战状态
	bt.SetTableStatus(consts.BATTLE_STATUS_END)

	bt.Clear()

	s.battleList[roomID] = nil
	delete(s.battleList, roomID)
	bt = nil

	return nil
}

func (s *_BattleC8Mgr) getBattleByRoomID(roomID uint32) (*BattleC8Info, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if roomID <= 0 {
		return nil, errors.New("-->_BattleC8Mgr--GetBattleByRoomID--RoomID Is Empty")
	}

	bt, ok := s.battleList[roomID]
	if !ok {
		return nil, errors.New("-->_BattleC8Mgr--GetBattleByRoomID--BattleInfo Not  Found!")
	}
	return bt, nil
}

// OnBattleStartRequest 开始对战(如果是房间对战则调用此方法，需要校验两个人的确认状态)
func (s *_BattleC8Mgr) OnBattleStartRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleStartRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleStartRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBattleStartRequest--Req:", req)

	if req.RoomID <= 0 {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleStartRequest--RoomID is empty")
		return
	}

	bt, err := s.initBattleInfo(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleStartRequest--GetBattleInfo err:", err)
		return
	}

	//开始初始化桌面信息
	resp := &gmsg.BattleBilliardInitInfoSync{}
	resp.RoomID = req.RoomID
	resp.StartCueEntityID = bt.GetCurrenEntityID()
	resp.InitBalls = bt.GetInitBalls()
	resp.ShowSettleGoldNum = uint32(bt.GetShowWinSettleGold())

	log.Info("-->logic--_BattleC8Mgr--OnBattleStartRequest--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_BilliardInitInfoSync, resp, targetEntityIDs)
	return
}

// OnMatchSuccessResponse 开始对战（匹配成功只会被请求一次）
func (s *_BattleC8Mgr) OnMatchSuccessResponse(roomId uint32) {
	if roomId <= 0 {
		log.Waring("-->logic--_BattleC8Mgr--OnMatchSuccessResponse--RoomID is empty")
		return
	}

	bt, err := s.initBattleInfo(roomId)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnMatchSuccessResponse--GetBattleInfo err:", err)
		return
	}

	//开始初始化桌面信息
	resp := &gmsg.BattleBilliardInitInfoSync{}
	resp.RoomID = roomId
	resp.StartCueEntityID = bt.GetCurrenEntityID()
	resp.InitBalls = bt.GetInitBalls()
	resp.ShowSettleGoldNum = uint32(bt.GetShowWinSettleGold())
	resp.RoundInfo = &gmsg.BattleRoundInfo{
		RoundID:             bt.GetRoundId(),
		CurrenRoundEntityID: bt.GetCurrenEntityID(),
		State:               gmsg.BilliardState_Enum_Aiming,
	}

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnMatchSuccessResponse--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_BilliardInitInfoSync, resp, targetEntityIDs)
	return
}

// 获取对战数据
func (s *_BattleC8Mgr) initBattleInfo(roomID uint32) (*BattleC8Info, error) {
	var bt *BattleC8Info
	var err error

	if s.battleList[roomID] == nil {
		//开始初始化房间
		//创建对战
		bt, err = NewBattleC8Info(roomID)
		if err != nil {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--NewBattleInfo err:", err)
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--NewBattleInfo err")
		}

		bt.SetStartTime(time.Now().Unix())

		roomInfo := MatchManager.GetRoomID(roomID)
		if roomInfo == nil || len(roomInfo.MapEntity) <= 0 {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.MapEntity is nil")
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.MapEntity is nil")
		}

		bt.SetRoomLevel(roomInfo.Level)

		if len(roomInfo.MapEntity) <= 0 {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.MapEntity is empty")
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.MapEntity is empty")
		}

		err = bt.SetEntityList(roomInfo.MapEntity)
		if err != nil {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--SetEntityList err:", err)
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--SetEntityList err")
		}

		if roomInfo.Blind <= 0 || roomInfo.TableFee <= 0 {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.Blind  || TableFee <= 0")
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.Blind  || TableFee <= 0")
		}
		bt.SetBlind(roomInfo.Blind)
		bt.SetTableFee(roomInfo.TableFee)

		if roomInfo.Blind <= 0 || roomInfo.TableFee <= 0 {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.WinExp || TransporterExp <= 0")
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--roomInfo.WinExp || TransporterExp <= 0")
		}
		bt.SetWinExp(roomInfo.WinExp)
		bt.SetTransporterExp(roomInfo.TransporterExp)

		err = bt.StartPlayBallCountdown()
		if err != nil {
			log.Waring("-->logic--_BattleC8Mgr--initBattleInfo--StartPlayBallCountdown err:", err)
			return nil, errors.New("-->logic--_BattleC8Mgr--initBattleInfo--StartPlayBallCountdown err")
		}

		//随机一个先手
		_ = bt.GetStartCueEntityID()
		bt.SetTableStatus(consts.BATTLE_STATUS_BATTLE_ING)

		err = s.SetBattleToMgr(bt)
	} else {
		bt = s.battleList[roomID]
	}
	return bt, err
}

// OnBattleAimingRequest 瞄准
func (s *_BattleC8Mgr) OnBattleAimingRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleAimingRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleAimingRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBattleAimingRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleAimingRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.RoundID != bt.GetRoundId() || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleAimingRequest--req.EntityID != bt.GetCurrenEntityID()")
		return

	}
	bt.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Aiming))

	battleInfoCueRotation := C8BallCoordinateInfo{
		X: req.CueRotationInfo.X,
		Y: req.CueRotationInfo.Y,
		Z: req.CueRotationInfo.Z,
	}
	bt.SetCurrenCueRotation(battleInfoCueRotation)

	//开始桌面更新信息
	cueRotation := gmsg.BattleBallCoordinateInfo{
		X: req.CueRotationInfo.X,
		Y: req.CueRotationInfo.Y,
		Z: req.CueRotationInfo.Z,
	}
	resp := &gmsg.BattleBilliardUpdateAimingSync{}
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()
	resp.CueRotationInfo = &cueRotation
	resp.CurrenRoundEntityID = bt.GetCurrenEntityID()

	log.Info("-->logic--_BattleC8Mgr--OnBattleAimingRequest--Resp:", resp)

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_BilliardUpdateAimingSync, resp, targetEntityIDs)
	return
}

// OnBattleEnergyStorageRequest 蓄力
func (s *_BattleC8Mgr) OnBattleEnergyStorageRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleEnergyStorageRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleEnergyStorageRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}
	log.Info("-->logic--_BattleC8Mgr--OnBattleEnergyStorageRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleEnergyStorageRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.RoundID != bt.GetRoundId() || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleEnergyStorageRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	bt.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Accumulating))
	bt.SetCurrenStrength(req.Strength)

	//开始桌面更新信息
	resp := &gmsg.BattleBilliardUpdateStrengthSync{}
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()
	resp.Strength = req.Strength
	resp.CurrenRoundEntityID = bt.GetCurrenEntityID()

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnBattleEnergyStorageRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_BilliardUpdateStrengthSync, resp, targetEntityIDs)
	return
}

// OnBattleStrokeBallRequest 击球
func (s *_BattleC8Mgr) OnBattleStrokeBallRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleStrokeBallRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleStrokeBallRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBattleStrokeBallRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleStrokeBallRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.RoundID != bt.GetRoundId() || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleStrokeBallRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	//先清除定时器
	_ = bt.DelPlayBallCountdown()

	bt.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Scrolling))

	//返回击球信息
	resp := &gmsg.BattleStrokeBallResponse{}
	resp.Code = 0
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()

	log.Info("-->logic--_BattleC8Mgr--OnBattleStrokeBallRequest--Resp:", resp)

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_StrokeBallResponse, resp, targetEntityIDs)

	return
}

// OnBattleObstructRequest 加塞
func (s *_BattleC8Mgr) OnBattleObstructRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleObstructRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleObstructRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}
	log.Info("-->logic--_BattleC8Mgr--OnBattleObstructRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleObstructRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.RoundID != bt.GetRoundId() || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleObstructRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	vector := C8BallCoordinateInfo{
		X: req.VectorInfo.X,
		Y: req.VectorInfo.Y,
		Z: req.VectorInfo.Z,
	}
	bt.SetRoundVectorInfo(vector)

	//开始桌面更新信息
	resp := &gmsg.BattleObstructSync{}
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()
	resp.VectorInfo = &gmsg.BattleBallCoordinateInfo{
		X: req.VectorInfo.X,
		Y: req.VectorInfo.Y,
		Z: req.VectorInfo.Z,
	}

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnBattleObstructRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_ObstructSync, resp, targetEntityIDs)
	return
}

// OnBattleObstructCueRequest 抬杆
func (s *_BattleC8Mgr) OnBattleObstructCueRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleObstructCueRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleObstructCueRequest--msgEV.Unmarshal(req)--err:", err)
		return
	}
	log.Info("-->logic--_BattleC8Mgr--OnBattleObstructCueRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleObstructCueRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.RoundID != bt.GetRoundId() || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleObstructCueRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	bt.SetRoundAngle(req.Angle)

	//开始桌面更新信息
	resp := &gmsg.BattleObstructCueSync{}
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()
	resp.Angle = req.Angle

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnBattleObstructCueRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_ObstructCueSync, resp, targetEntityIDs)
	return
}

// OnBallGoalRequest 进球
func (s *_BattleC8Mgr) OnBallGoalRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleBallGoalRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBallGoalRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBallGoalRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBallGoalRequest--s.getBattleByRoomID--err:", err)
		return
	}

	//获取是否在本局的进球列表
	isInRoundGoals := bt.IsInRoundBallsGoalsList(req.BallGoalsID)
	if isInRoundGoals {
		log.Info("-->logic--_BattleC8Mgr--OnBallGoalRequest--bt.IsInRoundBallsGoalsList")
		return
	}

	//获取是否在所有的进球列表
	isInAllGoals := bt.IsInAllBallsGoalsList(req.BallGoalsID)
	if isInAllGoals {
		log.Info("-->logic--_BattleC8Mgr--OnBallGoalRequest--bt.IsInAllBallsGoalsList")
		return
	}

	//更新进球列表
	goalsInfo := C8GoalsBallInfo{
		BallId:          req.BallGoalsID,
		GoalsBallPocket: req.GoalsBallPocket,
	}
	bt.SetBallsGoalsList(goalsInfo)

	//开始桌面更新信息
	resp := &gmsg.BattleBallGoalResponse{}
	resp.Code = 0
	resp.RoomID = req.RoomID

	log.Info("-->logic--_BattleC8Mgr--OnBallGoalRequest--Resp:", resp)

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()
	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_BallGoalResponse, resp, targetEntityIDs)
	return
}

// OnRoundEndRequest 回合结束
func (s *_BattleC8Mgr) OnRoundEndRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleRoundEndRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnRoundEndRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnRoundEndRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnRoundEndRequest--getBattleByRoomID--err:", err)
		return
	}

	if bt.GetRoundIsEnd() {
		log.Info("-->logic--_BattleC8Mgr--OnRoundEndRequest--bt.GetRoundIsEnd()")
		return
	}

	roundId := bt.GetRoundId()
	if req.RoundID != roundId {
		log.Info("-->logic--_BattleC8Mgr--OnRoundEndRequest--req.RoundID != bt.GetRoundId()")
		return
	}

	if bt.GetRoundStatus() != uint8(gmsg.BilliardState_Enum_Scrolling) {
		log.Waring("-->logic--_BattleC8Mgr--OnRoundEndRequest--bt.GetRoundStatus() != uint8(gmsg.BilliardState_Enum_Scrolling)--", req.EntityID)
		return
	}

	//判断是否是第一个到达的请求
	if bt.IsFirstRoundEndConfirmId() {
		//1.第一个到达的回合结束请求,记录不处理
		bt.SetRoundEndConfirmId(req.EntityID)
		//2.开启第二个请求的倒计时
		_ = bt.StartRoundEndCountdown()
		return
	}

	//先判断是否已记录
	if bt.CheckRoundEndConfirmId(req.EntityID) {
		return
	}
	bt.SetRoundEndConfirmId(req.EntityID)
	//确认是否收到全部回合结束请求
	if !bt.CheckRoundEndConfirmReady() {
		return
	}
	//删除第二个请求的倒计时
	_ = bt.DelRoundEndCountdown()

	//开始处理时再设置处理状态
	bt.SetRoundIsEnd()
	bt.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Stop))

	//先初始化所有球的信息
	bt.SetTidBallsPosition(nil)
	//更新球的位置信息
	ballsList := make([]C8BallInfo, 0)
	for _, v := range req.BallsInfo {
		bbpInfo := C8BallCoordinateInfo{
			X: v.Position.X,
			Y: v.Position.Y,
			Z: v.Position.Z,
		}

		//如果是白球需要更新当前回合白球的位置信息
		if v.BallId == consts.BALLS_WHITE {
			bt.SetWhiteBallPos(bbpInfo)
		}

		bbrInfo := C8BallCoordinateInfo{
			X: v.Rotation.X,
			Y: v.Rotation.Y,
			Z: v.Rotation.Z,
		}
		pInfo := C8BallInfo{
			BallId:   v.BallId,
			Position: &bbpInfo,
			Rotation: &bbrInfo,
		}
		ballsList = append(ballsList, pInfo)
	}
	bt.SetTidBallsPosition(ballsList)

	currenEntityID := bt.GetCurrenEntityID()
	otherEntityID := bt.GetRoomOtherEntityID(currenEntityID)
	if currenEntityID <= 0 {
		log.Waring("-->logic--_BattleC8Mgr--OnRoundEndRequest--currenEntityID <= 0--")
		return
	}

	if otherEntityID <= 0 {
		log.Waring("-->logic--_BattleC8Mgr--OnRoundEndRequest--otherEntityID <= 0--")
		return
	}

	//处理统记的数据
	bt.CheckRoundDataStatistics(req.BattingStyleStatistic)

	//获取全部进球列表
	ballsGoalsList := bt.GetEntityAllBallsGoalsList()
	//是否进了白球
	whiteBallIsIn := bt.IsInRoundBallsGoalsList(consts.BALLS_WHITE)
	//碰到第一个球是否是违规
	firstBallViolation := bt.IsRoundFirstBallViolation()

	//是否违规
	//仅有对方目标球入袋，亦不犯规，换由对方击球 打到第一个球是自己的
	if whiteBallIsIn || firstBallViolation || (req.IsNotContactFrame && len(ballsGoalsList) <= 0) {
		//修改违规次数
		violationNum := bt.IncrEntityViolationNum()
		s.sendViolationMsg(currenEntityID, req.RoomID, violationNum)
	} else {
		//重置违规次数
		bt.InitEntityViolationNum()
	}

	if bt.GetEntityViolationNum() >= consts.JUDG_TRANSPORT_NUM {
		s.startSettlement(req.RoomID, otherEntityID, currenEntityID, consts.SETTLEMENT_TYPE_VIOLATION)
		return
	}

	//判断当前回合的列表是否进了黑球
	blackBallIsIn := bt.IsInRoundBallsGoalsList(consts.BALLS_BLACK)
	if len(ballsGoalsList) >= 7 && blackBallIsIn {
		if whiteBallIsIn {
			s.startSettlement(req.RoomID, otherEntityID, currenEntityID, consts.SETTLEMENT_TYPE_NORMAL)
		} else {
			s.startSettlement(req.RoomID, currenEntityID, otherEntityID, consts.SETTLEMENT_TYPE_NORMAL)
		}
		return
	}

	//黑球直接判负
	if len(ballsGoalsList) < 7 && blackBallIsIn {
		s.startSettlement(req.RoomID, otherEntityID, currenEntityID, consts.SETTLEMENT_TYPE_NORMAL)
		return
	}

	roundGoalsList := bt.GetRoundBallsGoalsList()
	splittingList := bt.GetBallsSplittingList()

	//先判断分边
	if len(splittingList) <= 0 && roundId > 0 && len(roundGoalsList) > 0 {
		//最后一个进球
		goalsLastBall := bt.GetCurrenRoundGoalsLastBall()
		if goalsLastBall.BallId >= 1 && goalsLastBall.BallId <= 7 {
			bt.SetBallsSplittingList(currenEntityID, consts.BALLS_SPLITTING_SMALL)
			bt.SetBallsSplittingList(otherEntityID, consts.BALLS_SPLITTING_BIG)
		} else {
			bt.SetBallsSplittingList(otherEntityID, consts.BALLS_SPLITTING_SMALL)
			bt.SetBallsSplittingList(currenEntityID, consts.BALLS_SPLITTING_BIG)
		}
	}

	//违规进球的数量
	violationNum := bt.ViolationGoalsNum()
	complianceNum := bt.ComplianceGoalsNum()

	//是否换人
	nextEntityId := currenEntityID
	if len(roundGoalsList) <= 0 || whiteBallIsIn || firstBallViolation || (!firstBallViolation && complianceNum <= 0 && violationNum > 0) {
		nextEntityId = otherEntityID
	}

	/*********************************************开始切换回合信息****************************************/
	bt.SetRoundHis()
	incrRoundId := bt.IncrRound()
	bt.SetCurrenEntityID(nextEntityId)
	bt.InitRound(incrRoundId, nextEntityId)
	//先设置个默认状态
	bt.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Aiming))
	if whiteBallIsIn || firstBallViolation {
		//设置放置白球状态
		bt.SetRoundStatus(uint8(gmsg.BilliardState_Enum_Spoting))
	}
	//开启新倒计时
	_ = bt.StartPlayBallCountdown()

	/*********************************************重新获取修改后的信息****************************************/
	//开始处理返回的数据
	var respSplittingInfo []*gmsg.SplittingInfo
	splittingListNext := bt.GetBallsSplittingList()
	if len(splittingListNext) > 0 {
		for k, v := range splittingListNext {
			sInfo := &gmsg.SplittingInfo{
				EntityID:  k,
				Splitting: v,
			}
			respSplittingInfo = append(respSplittingInfo, sInfo)
		}
	}

	roundInfo := bt.GetCurrenRoundInfo()

	resp := &gmsg.BattleRoundEndResponse{}
	resp.Code = 0
	resp.CurrenRoundEntityID = bt.GetCurrenEntityID()
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()
	resp.RoundInfo = &gmsg.BattleRoundInfo{
		RoundID:             roundInfo.roundID,
		CurrenRoundEntityID: roundInfo.roundEntityID,
		State:               gmsg.BilliardState(roundInfo.status),
	}
	resp.SplittingInfo = respSplittingInfo
	resp.AllGoalsBalls = bt.GetAllBallsGoalsIdList()
	resp.BallsInfo = req.BallsInfo

	log.Info("-->logic--_BattleC8Mgr--OnRoundEndRequest--Resp:", resp)

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_RoundEndResponse, resp, targetEntityIDs)
	return
}

func (s *_BattleC8Mgr) sendViolationMsg(entityId uint32, roomId uint32, violationNum uint32) {
	if entityId <= 0 {
		return
	}

	resp := &gmsg.BattleEntityViolationResponse{
		Code:         resp_code.CODE_SUCCESS,
		RoomID:       roomId,
		EntityID:     entityId,
		IsViolation:  true,
		ViolationNum: violationNum,
	}

	log.Info("-->logic--_BattleC8Mgr--sendViolationMsg--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_ViolationResponse, resp, []uint32{entityId})
	return
}

func (s *_BattleC8Mgr) startSettlement(roomID uint32, winner uint32, transporter uint32, settlementType uint32) {
	bt, err := s.getBattleByRoomID(roomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--startSettlement--getBattleByRoomID--err:", err)
		return
	}

	//设置对战状态
	bt.SetTableStatus(consts.BATTLE_STATUS_SETTLE_ING)

	//清除所有倒计时
	_ = bt.DeleteAllTimer()

	//开始结算
	err = bt.Settle(winner, transporter, settlementType)
	if err != nil {
		return
	}

	//设置对战状态
	bt.SetTableStatus(consts.BATTLE_STATUS_SETTLE_OVER)

	_ = s.Clear(roomID)
	return
}

// OnSurrenderRequest 投降
func (s *_BattleC8Mgr) OnSurrenderRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleSurrenderRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnSurrenderRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnSurrenderRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnSurrenderRequest--s.getBattleByRoomID--err:", err)
		return
	}

	//投降方输家不需要推送消息
	otherEntityID := bt.GetRoomOtherEntityID(req.EntityID)
	s.startSettlement(req.RoomID, otherEntityID, req.EntityID, consts.SETTLEMENT_TYPE_SURRENDER)

	//开始桌面更新信息
	resp := &gmsg.BattleSurrenderResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.RoomID = req.RoomID
	resp.EntityID = req.EntityID

	log.Info("-->logic--_BattleC8Mgr--OnSurrenderRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_SurrenderResponse, resp, []uint32{otherEntityID})
	return
}

// OnApplyIncrBindRequest 申请加注
func (s *_BattleC8Mgr) OnApplyIncrBindRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ApplyIncrBindRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnApplyIncrBindRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnApplyIncrBindRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnApplyIncrBindRequest--getBattleByRoomID--err:", err)
		return
	}

	roomOtherEntityID := bt.GetRoomOtherEntityID(req.EntityID)
	if bt.IsRobot(roomOtherEntityID) {
		return
	}

	//开始桌面更新信息
	resp := &gmsg.ApplyIncrBindResponse{}
	resp.RoomID = req.RoomID
	resp.ApplyEntityID = req.EntityID

	log.Info("-->logic--_BattleC8Mgr--OnApplyIncrBindRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_ApplyIncrBindResponse, resp, []uint32{roomOtherEntityID})
	return
}

// OnFeedbackIncrBindRequest 反馈加注
func (s *_BattleC8Mgr) OnFeedbackIncrBindRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.FeedbackIncrBindRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnFeedbackIncrBindRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnFeedbackIncrBindRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnFeedbackIncrBindRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.IsAgree {
		roomOtherEntityID := bt.GetRoomOtherEntityID(req.EntityID)

		bt.IncrBlindNum()
		bt.SetIncrBlindHis(roomOtherEntityID)
	}

	//开始桌面更新信息
	resp := &gmsg.FeedbackIncrBindResponse{}
	resp.RoomID = req.RoomID
	resp.FeedbackEntityID = req.EntityID
	resp.IsAgree = req.IsAgree
	resp.ShowSettleGoldNum = uint32(bt.GetShowWinSettleGold())

	log.Info("-->logic--_BattleC8Mgr--OnFeedbackIncrBindRequest--Resp:", resp)

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_FeedbackIncrBindResponse, resp, targetEntityIDs)
	return
}

// OnBattleFirstBallColliderRequest 碰到第一个球请求
func (s *_BattleC8Mgr) OnBattleFirstBallColliderRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleFirstBallColliderRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleFirstBallColliderRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBattleFirstBallColliderRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleFirstBallColliderRequest--getBattleByRoomID--err:", err)
		return
	}

	if req.RoundID != bt.GetRoundId() || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleFirstBallColliderRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	bt.SetFirstBallCollider(req.BallID)

	//开始桌面更新信息
	resp := &gmsg.BattleFirstBallColliderResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.RoomID = req.RoomID
	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_FirstBallColliderResponse, resp, []uint32{req.EntityID})
	return
}

// OnBattleSetWhiteBallLocationIngRequest 白球放置中
func (s *_BattleC8Mgr) OnBattleSetWhiteBallLocationIngRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.SetWhiteBallLocationIngRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationIngRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationIngRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationIngRequest--getBattleByRoomID--err:", err)
		return
	}

	roundId := bt.GetRoundId()
	if req.RoundID != roundId || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationIngRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	//if bt.GetRoundStatus() != uint8(gmsg.BilliardState_Enum_Spoting) {
	//	log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationIngRequest--bt.GetRoundStatus() != uint8(gmsg.BilliardState_Enum_Spoting)")
	//	return
	//}

	//开始桌面更新信息
	resp := &gmsg.SetWhiteBallLocationIngSync{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.RoomID = req.RoomID
	resp.RoundID = roundId
	resp.WhiteBallsInfo = &gmsg.BattleBallInfo{
		BallId: req.WhiteBallsInfo.BallId,
		Position: &gmsg.BattleBallCoordinateInfo{
			X: req.WhiteBallsInfo.Position.X,
			Y: req.WhiteBallsInfo.Position.Y,
			Z: req.WhiteBallsInfo.Position.Z,
		},
		Rotation: &gmsg.BattleBallCoordinateInfo{
			X: req.WhiteBallsInfo.Rotation.X,
			Y: req.WhiteBallsInfo.Rotation.Y,
			Z: req.WhiteBallsInfo.Rotation.Z,
		},
		GoalsBallPocket: 0,
	}

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationIngRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_SetWhiteBallLocationIngSync, resp, targetEntityIDs)
	return
}

// OnBattleSetWhiteBallLocationEndRequest 白球放置结束
func (s *_BattleC8Mgr) OnBattleSetWhiteBallLocationEndRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.SetWhiteBallLocationEndRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationEndRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationEndRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationEndRequest--getBattleByRoomID--err:", err)
		return
	}

	roundId := bt.GetRoundId()
	if req.RoundID != roundId || req.EntityID != bt.GetCurrenEntityID() {
		log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationEndRequest--req.EntityID != bt.GetCurrenEntityID()")
		return
	}

	//if bt.GetRoundStatus() != uint8(gmsg.BilliardState_Enum_Spoting) {
	//	log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationEndRequest--bt.GetRoundStatus() != uint8(gmsg.BilliardState_Enum_Spoting)")
	//	return
	//}

	//开始桌面更新信息
	resp := &gmsg.SetWhiteBallLocationEndSync{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.RoomID = req.RoomID
	resp.RoundID = roundId
	resp.WhiteBallsInfo = &gmsg.BattleBallInfo{
		BallId: req.WhiteBallsInfo.BallId,
		Position: &gmsg.BattleBallCoordinateInfo{
			X: req.WhiteBallsInfo.Position.X,
			Y: req.WhiteBallsInfo.Position.Y,
			Z: req.WhiteBallsInfo.Position.Z,
		},
		Rotation: &gmsg.BattleBallCoordinateInfo{
			X: req.WhiteBallsInfo.Rotation.X,
			Y: req.WhiteBallsInfo.Rotation.Y,
			Z: req.WhiteBallsInfo.Rotation.Z,
		},
		GoalsBallPocket: 0,
	}

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnBattleSetWhiteBallLocationEndRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_SetWhiteBallLocationEndSync, resp, targetEntityIDs)
	return
}

// OnUserChartMsgRequest 对局聊天
func (s *_BattleC8Mgr) OnUserChartMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.BattleUserChartMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnUserChartMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_BattleC8Mgr--OnUserChartMsgRequest--Req:", req)

	bt, err := s.getBattleByRoomID(req.RoomID)
	if err != nil {
		log.Waring("-->logic--_BattleC8Mgr--OnUserChartMsgRequest--getBattleByRoomID--err:", err)
		return
	}

	//开始桌面更新信息
	resp := &gmsg.BattleUserChartMsgSync{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.RoomID = req.RoomID
	resp.RoundID = bt.GetRoundId()
	resp.EntityID = req.EntityID
	resp.MType = req.MType
	resp.Context = req.Context

	targetEntityIDs := bt.GetRoomALLEntityIDsNoRobot()

	log.Info("-->logic--_BattleC8Mgr--OnUserChartMsgRequest--Resp:", resp)

	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Battle_UserChartMsgSync, resp, targetEntityIDs)
	return
}
