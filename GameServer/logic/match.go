package logic

import (
	"BilliardServer/Common/entity"
	battleConf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"

	//"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"errors"
	"fmt"
	"gitee.com/go-package/carbon/v2"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

/***
 *@disc: 匹配队列系统
 *@author: lsj
 *@date: 2023/8/21
 */

type _BattleMatch struct {
	MatchSecond      uint32 // 匹配秒数
	EnterRobotSecond uint32 //5秒匹配机器人
	IsDisable        bool   // 否禁用
	MaxEntity        int
	RoomManger       map[uint32]*UnitRoom // 房间

	DefaultChanMr *MatchChanManger
	InitialChanMr *MatchChanManger
	MiddleChanMr  *MatchChanManger
	HighChanMr    *MatchChanManger
}

type MatchChanManger struct {
	ChanManagerName string //管理器名称
	Count           uint32 //人数
	Chan            chan MatchEntityID
	MapEntity       map[uint32]MatchEntityID
}

type MatchEntityID struct {
	EntityID   uint32 // 请求的id
	IsRobot    bool   //true代表机器人，false代表玩家
	EnterStamp int64
	EnterTime  time.Time
	MatchCount uint32 // 匹配次数
	Level      uint32 //匹配房间
}

var MatchManager _BattleMatch

var mutex sync.Mutex

func (c *_BattleMatch) Init() {
	c.MaxEntity = battleConf.Num
	c.IsDisable = false
	c.MatchSecond = battleConf.MatchTimes
	c.EnterRobotSecond = battleConf.EnterRobotSecond
	c.RoomManger = make(map[uint32]*UnitRoom, 0)
	c.DefaultChanMr = new(MatchChanManger)
	c.DefaultChanMr.ChanInit("DefaultChanMr")
	c.InitialChanMr = new(MatchChanManger)
	c.InitialChanMr.ChanInit("InitialChanMr")
	c.MiddleChanMr = new(MatchChanManger)
	c.MiddleChanMr.ChanInit("MiddleChanMr")
	c.HighChanMr = new(MatchChanManger)
	c.HighChanMr.ChanInit("HighChanMr")

	go MatchManager.Start()
	event.OnNet(gmsg.MsgTile_Hall_EightMatchRequest, reflect.ValueOf(c.EightMatchRequest))
	event.OnNet(gmsg.MsgTile_Hall_EightMatchCancelRequest, reflect.ValueOf(c.MatchCancelRequest))
	event.OnNet(gmsg.MsgTile_Hall_ExitRoomRequest, reflect.ValueOf(c.ExitRoomRequest))
	event.OnNet(gmsg.MsgTile_Hall_UseItemFromRoomIDRequest, reflect.ValueOf(c.UseItemFromRoomIDRequest))
	event.OnNet(gmsg.MsgTile_Hall_RoomListRequest, reflect.ValueOf(c.OnHallRoomRequest))
	event.OnNet(gmsg.MsgTile_Hall_EightReplayRequest, reflect.ValueOf(c.OnEightReplayRequest))
	event.OnNet(gmsg.MsgTile_Hall_EightReplayConfirmResponse, reflect.ValueOf(c.OnEightReplayConfirmResponse))
	timer.AddTimer(c, "ClearRoomTimer", 60000, true)
}

func (this *MatchChanManger) ChanInit(name string) {
	this.ChanManagerName, this.Count, this.MapEntity, this.Chan = name, 0, make(map[uint32]MatchEntityID, 0), make(chan MatchEntityID, 1000)
	log.Info("-->管道初始化完成 ", name)
}

func (c *_BattleMatch) EightMatchRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.HallEightMatchRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		log.Error("err请求：", err)
		return
	}

	msgMatchResponse := &gmsg.HallEightMatchResponse{}
	msgMatchResponse.Code = uint32(2)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchResponse, msgMatchResponse, []uint32{msgBody.EntityID})
		return
	}

	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	//快速匹配
	if msgBody.IsFastMatch {
		errs, newLevel := c.getPlayerLevel(tEntityPlayer.NumGold)
		if errs != nil {
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchResponse, msgMatchResponse, []uint32{msgBody.EntityID})
			return
		}
		c.RequestMatch(msgBody.EntityID, newLevel)
		return
	}

	errs, code, resLevel := c.checkPlayerGold(tEntityPlayer.NumGold, msgBody.Level)
	if msgBody.Level < 0 || errs != nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchResponse, msgMatchResponse, []uint32{msgBody.EntityID})
		return
	}

	if code > 0 {
		msgMatchResponse.Code = code
		msgMatchResponse.ResLevel = resLevel
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchResponse, msgMatchResponse, []uint32{msgBody.EntityID})
		return
	}

	c.RequestMatch(msgBody.EntityID, msgBody.Level)
}

func (c *_BattleMatch) RequestMatch(EntityID, level uint32) {
	msgMatchResponse := &gmsg.HallEightMatchResponse{}
	msgMatchResponse.Code = uint32(1)
	ch, chName, _ := c.getMatchChan(level)
	if ch == nil {
		log.Error("-->MatchRequest-->无效的管道。")
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchResponse, msgMatchResponse, []uint32{EntityID})
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	log.Info("匹配前的room_id:", tEntityPlayer.RoomId)

	// 进入队列，通知前端请求成功
	EnterMatch(EntityID, level, false, ch, chName)

	msgMatchResponse.Code = uint32(0)
	msgMatchResponse.ResLevel = level
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchResponse, msgMatchResponse, []uint32{EntityID})
}

func (c *_BattleMatch) getMatchChan(level uint32) (ch chan MatchEntityID, chName string, mapEntity map[uint32]MatchEntityID) {
	switch level {
	case battleConf.DefaultRoom:
		ch = c.DefaultChanMr.Chan
		chName = c.DefaultChanMr.ChanManagerName
		mapEntity = c.DefaultChanMr.MapEntity
	case battleConf.InitialRoom:
		ch = c.InitialChanMr.Chan
		chName = c.InitialChanMr.ChanManagerName
		mapEntity = c.InitialChanMr.MapEntity
	case battleConf.MiddleRoom:
		ch = c.MiddleChanMr.Chan
		chName = c.MiddleChanMr.ChanManagerName
		mapEntity = c.MiddleChanMr.MapEntity
	case battleConf.HighRoom:
		ch = c.HighChanMr.Chan
		chName = c.HighChanMr.ChanManagerName
		mapEntity = c.HighChanMr.MapEntity
	default:
		ch = nil
	}
	return
}

func EnterMatch(EntityID, Level uint32, isRobot bool, ch chan MatchEntityID, chName string) {
	select {
	case ch <- MatchEntityID{EntityID, isRobot, time.Now().UnixMilli(), time.Now(), 0, Level}:
		log.Info("->ch->", chName, "-->进队列成功-->level:", Level, "-->EntityID-->", EntityID)
	default:
		log.Error("队列溢出...")
	}
}

func (c *_BattleMatch) Start() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		c.enterLoop()
		go c.SendToMatch(c.DefaultChanMr.MapEntity, c.DefaultChanMr.ChanManagerName)
		go c.SendToMatch(c.InitialChanMr.MapEntity, c.InitialChanMr.ChanManagerName)
		go c.SendToMatch(c.MiddleChanMr.MapEntity, c.MiddleChanMr.ChanManagerName)
		go c.SendToMatch(c.HighChanMr.MapEntity, c.HighChanMr.ChanManagerName)
	}
}

func (c *_BattleMatch) enterLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("enterLoop", err)
			stack.PrintCallStack()
		}
	}()

out:
	for {
		select {
		case data := <-c.DefaultChanMr.Chan:
			if last, ok := c.DefaultChanMr.MapEntity[data.EntityID]; ok && time.Since(last.EnterTime) < time.Second*3 {
				return
			}
			c.DefaultChanMr.MapEntity[data.EntityID] = data
		case data := <-c.InitialChanMr.Chan:
			if last, ok := c.InitialChanMr.MapEntity[data.EntityID]; ok && time.Since(last.EnterTime) < time.Second*3 {
				return
			}
			c.InitialChanMr.MapEntity[data.EntityID] = data
		case data := <-c.MiddleChanMr.Chan:
			if last, ok := c.MiddleChanMr.MapEntity[data.EntityID]; ok && time.Since(last.EnterTime) < time.Second*3 {
				return
			}
			c.MiddleChanMr.MapEntity[data.EntityID] = data
		case data := <-c.HighChanMr.Chan:
			if last, ok := c.HighChanMr.MapEntity[data.EntityID]; ok && time.Since(last.EnterTime) < time.Second*3 {
				return
			}
			c.HighChanMr.MapEntity[data.EntityID] = data
		default:
			break out
		}
	}
}

// 匹配规则，奇数的，随机去掉一位，等下一次匹配；偶数直接匹配
func (c *_BattleMatch) SendToMatch(mapEntity map[uint32]MatchEntityID, chanManagerName string) {
	entityLen := len(mapEntity)
	if entityLen == 0 {
		return
	}

	if entityLen == 1 {
		c.toNextMatch(mapEntity)
		return
	}

	log.Info("-->SendToMatch,数据条数：", len(mapEntity))

	// 判断个数是奇数还是偶数
	remainder := entityLen % battleConf.Num
	var entityMakeLen int
	if remainder == 0 {
		entityMakeLen = entityLen
	} else {
		entityMakeLen = entityLen - 1
	}
	c.formRemainderMatch(mapEntity, remainder, entityMakeLen, chanManagerName)

}

func (c *_BattleMatch) toNextMatch(MapEntity map[uint32]MatchEntityID) {
	for k, v := range MapEntity {
		one := new(MatchEntityID)
		stack.SimpleCopyProperties(one, &v)
		one.MatchCount = v.MatchCount + 1
		MapEntity[v.EntityID] = *one

		if !one.IsRobot && one.MatchCount >= c.EnterRobotSecond {
			go c.inviteRobot(v.Level)
		}

		if one.MatchCount > c.MatchSecond {
			// 超过15秒，没有数据匹配
			if !one.IsRobot {
				c.MatchTimeOutResponse(k)
			}
			c.DelMatch(one.IsRobot, k, MapEntity)
		}
	}
}

// 删除匹配
func (c *_BattleMatch) DelMatch(isRobot bool, entityID uint32, MapEntity map[uint32]MatchEntityID) {
	if isRobot {
		tEntity := Entity.EmPlayer.GetEntityByID(entityID)
		if tEntity == nil {
			return
		}
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		tEntityPlayer.SetBehaviorStatus(battleConf.BEHAVIOR_STATUS_NONE)
		tEntityPlayer.SyncEntity(1)
	}
	delete(MapEntity, entityID)
}

// 邀请机器人
func (c *_BattleMatch) inviteRobot(level uint32) {
	robotID := uint32(0)
	isHigh := c.isRobotType(level)
	if isHigh {
		robotID = RobotMr.MatchHighRobot()
	} else {
		robotID = RobotMr.MatchDefaultRobot()
	}

	if robotID == uint32(0) {
		log.Info("-------->没有空闲的机器人-->")
		return
	}

	channel, chName, _ := c.getMatchChan(level)
	EnterMatch(robotID, level, true, channel, chName)
	log.Info("-->robotID-->", robotID, "机器人入场成功。")
	return
}

func (c *_BattleMatch) isRobotType(level uint32) bool {
	return level >= battleConf.HighRoom
}

type MatchEntityIDSlice []MatchEntityID

func (c *_BattleMatch) formRemainderMatch(MapEntity map[uint32]MatchEntityID, remainder, entityMakeLen int, chName string) {

	entitySlice := make(MatchEntityIDSlice, 0)

	if remainder == 0 {
		for key, value := range MapEntity {
			entitySlice = append(entitySlice, value)
			delete(MapEntity, key)
		}
	} else {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(entityMakeLen)
		index := 0
		for key, value := range MapEntity {
			if randIndex != index {
				entitySlice = append(entitySlice, value)
				delete(MapEntity, key)
			}
			index++
		}
		c.toNextMatch(MapEntity)
	}

	matchResult := c.MatchGroupFunc(entitySlice, battleConf.Num)

	log.Info(fmt.Sprintf("-->chName:%s-->匹配%d对成功.", chName, len(matchResult)))
	//创建房间,并通知网关
	go c.MatchSuccessResponse(matchResult)
}

func (c *_BattleMatch) MatchGroupFunc(matchArr []MatchEntityID, num int) [][]MatchEntityID {
	max := len(matchArr)

	if max <= num {
		return [][]MatchEntityID{matchArr}
	}

	var quantity int
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}

	var res = make([][]MatchEntityID, 0)

	var start, end, i int
	for i = 1; i <= quantity; i++ {
		end = i * num
		if i != quantity {
			res = append(res, matchArr[start:end])
		} else {
			res = append(res, matchArr[start:])
		}
		start = i * num
	}
	return res
}

// 关闭匹配
func (c *_BattleMatch) Stop() {
	//close(c.DefaultChannel)
}

func (c *_BattleMatch) IsToDisable() {
	c.IsDisable = true
}

// 匹配超时返回
func (c *_BattleMatch) MatchTimeOutResponse(EntityID uint32) {
	MatchTimeOutResponse := &gmsg.HallMatchTimeOutResponse{}
	MatchTimeOutResponse.Code = *proto.Uint32(1)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchTimeOutResponse, MatchTimeOutResponse, []uint32{EntityID})
}

func (c *_BattleMatch) getMatchResponsePlayer(entityID, roomID uint32) *gmsg.MatchResponsePlayer {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return nil
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgMatchResponsePlayer := &gmsg.MatchResponsePlayer{}

	msgMatchResponsePlayer.Player = new(gmsg.PlayerData)
	msgMatchResponsePlayer.EntityID = entityID
	msgMatchResponsePlayer.Player.IsAi = *proto.Uint32(0)
	if tEntityPlayer.IsRobot {
		msgMatchResponsePlayer.Player.IsAi = *proto.Uint32(1)
	}
	msgMatchResponsePlayer.Player.PlayerIcon = tEntityPlayer.PlayerIcon
	msgMatchResponsePlayer.Player.IconFrame = tEntityPlayer.IconFrame
	msgMatchResponsePlayer.Player.PlayerName = tEntityPlayer.PlayerName
	msgMatchResponsePlayer.Player.PlayerLv = tEntityPlayer.PlayerLv
	msgMatchResponsePlayer.Player.PeakRankLv = tEntityPlayer.PeakRankLv

	msgMatchResponsePlayer.PlayerItem = new(gmsg.PlayerItemInfo)
	msgMatchResponsePlayer.PlayerItem.EntityID = entityID
	msgMatchResponsePlayer.PlayerItem.CueTableID = tEntityPlayer.CueTableId
	msgMatchResponsePlayer.PlayerItem.BattingEffect = tEntityPlayer.BattingEffect
	msgMatchResponsePlayer.PlayerItem.GoalInEffect = tEntityPlayer.GoalInEffect
	msgMatchResponsePlayer.PlayerItem.CueBall = tEntityPlayer.CueBall
	msgMatchResponsePlayer.PlayerItem.TableCloth = tEntityPlayer.TableCloth
	tEntityPlayer.RoomId = roomID
	tEntityPlayer.SyncEntity(1)
	return msgMatchResponsePlayer
}

// 匹配配对成功返回
func (c *_BattleMatch) MatchSuccessResponse(matchRes [][]MatchEntityID) {
	for _, resBroadCast := range matchRes {
		roomID := c.InitCreateRoomData(resBroadCast[0].Level)
		msgMatchResponse := &gmsg.HallMatchSuccessResponse{}
		msgMatchResponse.Code = *proto.Uint32(0)
		msgMatchResponse.RoomID = roomID

		for _, v := range resBroadCast {
			msgMatchResponsePlayer := c.getMatchResponsePlayer(v.EntityID, roomID)
			if msgMatchResponsePlayer == nil {
				continue
			}
			msgMatchResponse.MatchResponsePlayer = append(msgMatchResponse.MatchResponsePlayer, msgMatchResponsePlayer)
		}
		for _, broadCast := range resBroadCast {
			tEntity := Entity.EmPlayer.GetEntityByID(broadCast.EntityID)
			tEntityPlayer := tEntity.(*entity.EntityPlayer)
			c.MatchEnterRoomData(tEntity, roomID)
			if !broadCast.IsRobot {
				ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightMatchSuccessResponse, msgMatchResponse, []uint32{broadCast.EntityID})
			} else {
				RobotMr.DeductRobot(RobotMr.IsRobotHighLevel(tEntityPlayer.NumGold), broadCast.EntityID)
			}
		}

		//通知对战服务开始对战
		BattleC8Mgr.OnMatchSuccessResponse(roomID)
	}
}

// 创建房间
func (c *_BattleMatch) InitCreateRoomData(level uint32) uint32 {
	mutex.Lock()
	defer mutex.Unlock()
	levelCfg := Table.GetEightBallRoomCfgLevel(level)
	RoomID := c.getRoomRandomID(100000, 999999)
	this := new(UnitRoom)
	this.RoomID = RoomID
	this.CountEntity = 0
	this.Level = level
	this.Blind = uint64(levelCfg.MinCoin)
	this.TableFee = uint64(levelCfg.TableFee)
	this.WinExp = levelCfg.WinExp
	this.TransporterExp = levelCfg.TransporterExp
	this.MaxEntity = c.MaxEntity
	this.CreateTime = carbon.Now().ToDateTimeString()
	this.UpdateTime = this.CreateTime
	this.MapEntity = make(map[uint32]entity.Entity)
	this.PlayNum += 1
	c.RoomManger[RoomID] = this
	log.Info("-------------->room id create -->:", RoomID)
	return RoomID
}

func (c *_BattleMatch) MatchEnterRoomData(Entity entity.Entity, RoomID uint32) {
	room := c.GetRoomID(RoomID)
	if room == nil {
		return
	}
	if room.YesInRoom(Entity.GetEntityID()) {
		return
	}
	room.UpdateTime = carbon.Now().ToDateTimeString()
	room.MapEntity[Entity.GetEntityID()] = Entity
	room.CountEntity = len(room.MapEntity)
	c.incrPlayerNum(room.Level)
	log.Info(Entity.GetEntityID(), "----------->进入房间-->", room)
}

func (c *_BattleMatch) incrPlayerNum(level uint32) {
	mutex.Lock()
	defer mutex.Unlock()
	switch level {
	case battleConf.DefaultRoom:
		c.DefaultChanMr.Count += 1
	case battleConf.InitialRoom:
		c.InitialChanMr.Count += 1
	case battleConf.MiddleRoom:
		c.MiddleChanMr.Count += 1
	case battleConf.HighRoom:
		c.HighChanMr.Count += 1
	default:
		log.Error("level is err")
	}
}

func (c *_BattleMatch) deductPlayerNum(level uint32) {
	mutex.Lock()
	defer mutex.Unlock()
	switch level {
	case battleConf.DefaultRoom:
		c.DefaultChanMr.Count -= 1
	case battleConf.InitialRoom:
		c.InitialChanMr.Count -= 1
	case battleConf.MiddleRoom:
		c.MiddleChanMr.Count -= 1
	case battleConf.HighRoom:
		c.HighChanMr.Count -= 1
	default:
	}
}

// 退出房间通知房间的人
func (c *_BattleMatch) ExitRoomRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.HallExitRoomRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	log.Info("-->ExitRoomRequest->begin->", msgBody)
	ExitRoomResponse := &gmsg.HallExitRoomResponse{}
	ExitRoomResponse.Code = *proto.Uint32(0)
	room := c.GetRoomID(msgBody.RoomID)
	if room == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ExitRoomResponse, ExitRoomResponse, []uint32{msgBody.EntityID})
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	c.initPlayerBehavior(battleConf.BEHAVIOR_STATUS_NONE, msgBody.RoomID, tEntityPlayer)
	if len(room.MapEntity) == 0 {
		c.DelRoomFromID(msgBody.RoomID)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ExitRoomResponse, ExitRoomResponse, []uint32{msgBody.EntityID})
}

func (c *_BattleMatch) GetRoomID(RoomID uint32) *UnitRoom {
	if len(c.RoomManger) == 0 {
		return nil
	}
	if v, ok := c.RoomManger[RoomID]; ok {
		return v
	}
	return nil
}

func (c *_BattleMatch) IsExistRoomID(RoomID uint32) bool {
	r := c.RoomManger[RoomID]
	if r == nil {
		return false
	}
	return true
}

func (c *_BattleMatch) MatchCancelRequest(msgEV *network.MsgBodyEvent) {
	mutex.Lock()
	defer mutex.Unlock()
	msgBody := &gmsg.HallEightMatchCancelRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	_, _, mapEntity := c.getMatchChan(msgBody.Level)
	delete(mapEntity, msgBody.EntityID)
}

func (c *_BattleMatch) DelRoomFromID(roomID uint32) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(c.RoomManger, roomID)
	log.Info("-->DelRoomFromID:", roomID)
}

func (c *_BattleMatch) UseItemFromRoomIDRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.UseItemFromRoomIDRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("-->UseItemFromRoomIDRequest->begin->", msgBody)
	msgResponse := &gmsg.UseItemFromRoomIDResponse{}
	msgResponse.Code = uint32(1)
	if msgBody.RoomID == 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UseItemFromRoomIDResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	room := c.GetRoomID(msgBody.RoomID)
	if room == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UseItemFromRoomIDResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	useReq := &gmsg.UseItemRequest{
		EntityID: msgBody.RoomID,
		ItemID:   msgBody.ItemID,
	}
	item, code, _ := Backpack.UpdatePlayInfoByItemType(tEntityPlayer, useReq)
	msgResponse.Code = code
	if item == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UseItemFromRoomIDResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	//角色数据同步
	if item.ItemType == battleConf.Cue {
		Player.ChangeCueTableID(msgBody.EntityID, item.TableID)
	} else if item.ItemType == battleConf.Effect {
		Player.ChangeEffect(msgBody.EntityID, item.SubType, item.TableID)
	}

	msgResponse.PlayerItem = make([]*gmsg.PlayerItemInfo, 0)
	for _, vEntity := range room.MapEntity {
		val := vEntity.(*entity.EntityPlayer)
		playItemInfo := new(gmsg.PlayerItemInfo)
		playItemInfo.EntityID = val.GetEntityID()
		playItemInfo.CueTableID = val.CueTableId
		playItemInfo.BattingEffect = val.BattingEffect
		playItemInfo.GoalInEffect = val.GoalInEffect
		playItemInfo.CueBall = val.CueBall
		msgResponse.PlayerItem = append(msgResponse.PlayerItem, playItemInfo)
	}
	log.Info("-->UseItemFromRoomIDRequest->end->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UseItemFromRoomIDResponse, msgResponse, room.GetAllEntityID())
}

// 房间列表请求
func (c *_BattleMatch) OnHallRoomRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.HallRoomListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.HallRoomListResponse{}
	msgResponse.GameType = msgBody.GameType
	msgResponse.RoomList = make([]*gmsg.RoomInfo, 0)

	if msgBody.GameType == battleConf.GameDefault {
		for i := battleConf.DefaultRoom; i <= battleConf.HighRoom; i++ {
			roomInfo := new(gmsg.RoomInfo)
			levelCfg := Table.GetEightBallRoomCfgLevel(uint32(i))
			if levelCfg == nil {
				continue
			}
			roomInfo.RoomTableID = levelCfg.TableID
			roomInfo.PlayerNum = c.getPlayerNum(uint32(i))
			msgResponse.RoomList = append(msgResponse.RoomList, roomInfo)
		}
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_RoomListResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_BattleMatch) getPlayerNum(level uint32) (num uint32) {
	switch level {
	case battleConf.DefaultRoom:
		num = uint32(c.DefaultChanMr.Count)
	case battleConf.InitialRoom:
		num = uint32(c.InitialChanMr.Count)
	case battleConf.MiddleRoom:
		num = uint32(c.MiddleChanMr.Count)
	case battleConf.HighRoom:
		num = uint32(c.HighChanMr.Count)
	default:
		num = 0
	}

	return
}

// 重赛请求//0成功并开局，1房间不存在,2对手离开房间，3金币不足，4推荐去高级场，5对手金币不足
func (c *_BattleMatch) OnEightReplayRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EightReplayRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.EightReplayResponse{}
	msgResponse.Code = uint32(1)
	room := c.GetRoomID(msgBody.RoomID)
	if room == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	//有空位表示有人离开房间
	if room.YesForFree() {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	room.lock.Lock()
	defer room.lock.Unlock()
	if room.ReplayConfirm > uint32(0) {
		log.Error("不可重复重赛", room.RoomID, msgBody.EntityID)
		return
	}
	room.ReplayConfirm += uint32(1)
	log.Info("room", room)

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	errs, code, _ := c.checkPlayerGold(tEntityPlayer.NumGold, room.Level)
	if errs != nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if code > 1 {
		msgResponse.Code = code + uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	other := Entity.EmPlayer.GetEntityByID(room.GetOtherPlayerId(msgBody.EntityID))
	if other == nil {
		return
	}
	otherPlayer := other.(*entity.EntityPlayer)
	err, rescode, _ := c.checkPlayerMinGold(otherPlayer.NumGold, room.Level)
	if err != nil || rescode > 0 {
		msgResponse.Code = uint32(5)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	c.OnEightReplayConfirmRequest(room.GetOtherPlayerId(msgBody.EntityID), room.RoomID)

	time.AfterFunc(time.Second*5, func() {
		c.EightReplayConfirmTimeOutResponse(msgBody.EntityID, room)
	})
}

// 对手确认同步
func (c *_BattleMatch) OnEightReplayConfirmRequest(toEntityID, roomId uint32) {
	resBody := &gmsg.EightReplayConfirmRequest{EntityID: toEntityID, RoomID: roomId}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayConfirmRequest, resBody, []uint32{toEntityID})
}

// 对手确认返回
func (c *_BattleMatch) OnEightReplayConfirmResponse(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EightReplayConfirmResponse{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	room := c.GetRoomID(msgBody.RoomID)
	if room == nil {
		log.Error("房间不存在", msgBody.RoomID)
		return
	}

	room.ResetReplayConfirm()
	msgResponse := &gmsg.EightReplayResponse{}
	if !msgBody.IsAgree {
		msgResponse.Code = uint32(6)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{room.GetOtherPlayerId(msgBody.EntityID)})
		return
	}

	msgResponse.Code = uint32(0)
	msgResponse.RoomID = msgBody.RoomID
	msgResponse.MatchResponsePlayer = make([]*gmsg.MatchResponsePlayer, 0)
	for _, val := range room.GetAllEntityID() {
		player := c.getMatchResponsePlayer(val, msgBody.RoomID)
		msgResponse.MatchResponsePlayer = append(msgResponse.MatchResponsePlayer, player)
	}
	//通知对战服务开始对战
	BattleC8Mgr.OnMatchSuccessResponse(msgBody.RoomID)
	room.AddRoomPlayNum()

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, room.GetAllEntityID())
}

func (c *_BattleMatch) EightReplayConfirmTimeOutResponse(entityID uint32, room *UnitRoom) {
	room.lock.Lock()
	defer room.lock.Unlock()
	if room.ReplayConfirm == 0 {
		return
	}

	room.ReplayConfirm = 0
	msgResponse := &gmsg.EightReplayResponse{}
	msgResponse.Code = uint32(6)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_EightReplayResponse, msgResponse, []uint32{entityID})
	return
}

func (c *_BattleMatch) checkPlayerMinGold(gold, level uint32) (err error, code, resLevel uint32) {
	levelCfg := Table.GetEightBallRoomCfgLevel(level)
	if levelCfg == nil {
		return errors.New("无配置。"), 1, level
	}

	if gold < levelCfg.MinCoin {
		return nil, 1, level
	}

	return nil, 0, level
}

func (c *_BattleMatch) checkPlayerGold(gold, level uint32) (err error, code, resLevel uint32) {
	levelCfg := Table.GetEightBallRoomCfgLevel(level)
	if levelCfg == nil {
		return errors.New("无配置。"), 0, level
	}

	if gold < levelCfg.MinCoin {
		return nil, 2, level
	}

	if gold > levelCfg.MaxCoin && level < battleConf.HighRoom {
		errs, newLevel := c.getPlayerLevel(gold)
		if errs != nil {
			return errs, 0, level
		}
		code = 1
		if newLevel > level {
			code = 3
		}
		return nil, code, newLevel
	}

	return nil, 0, level
}

func (c *_BattleMatch) getPlayerLevel(gold uint32) (error, uint32) {
	for _, v := range Table.EightBallRoomSlice {
		if v.Level == battleConf.HighRoom && gold > v.MinCoin {
			return nil, v.Level
		}
		if gold < v.MinCoin || gold > v.MaxCoin {
			continue
		}
		return nil, v.Level
	}

	return errors.New("is err"), 0
}

func (c *_BattleMatch) SetRobotBehaviorStatus(status uint8, entityID uint32, roomID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	c.initPlayerBehavior(status, roomID, tEntityPlayer)
}

func (c *_BattleMatch) initPlayerBehavior(status uint8, roomID uint32, tEntityPlayer *entity.EntityPlayer) {
	if tEntityPlayer.RoomId == roomID {
		tEntityPlayer.SetBehaviorStatus(status)
		tEntityPlayer.RoomId = 0
		tEntityPlayer.SyncEntity(1)
	}
	room := c.GetRoomID(roomID)
	if room == nil {
		return
	}
	room.ExitRoom(tEntityPlayer.EntityID)
	if tEntityPlayer.IsRobot {
		isHigh := RobotMr.IsRobotHighLevel(tEntityPlayer.NumGold)
		RobotMr.IncrRobot(isHigh, tEntityPlayer.EntityID)
	}
	c.deductPlayerNum(room.Level)
}

func (c *_BattleMatch) ClearRoomTimer() {
	if len(c.RoomManger) == 0 {
		return
	}
	now := time.Now().Unix()

	for rid, val := range c.RoomManger {
		bt, _ := BattleC8Mgr.getBattleByRoomID(rid)
		t := now - tools.GetTimeByString(val.CreateTime).Unix()
		if bt == nil || bt.status >= battleConf.BATTLE_STATUS_SETTLE_OVER || t > 40*60 {
			c.clearRoom(val)
		} else if t > 20*60 {
			log.Info("未处理的房间ID：", val)
		}
	}
}

func (c *_BattleMatch) clearRoom(room *UnitRoom) {
	for entityID, _ := range room.MapEntity {
		c.SetRobotBehaviorStatus(battleConf.BEHAVIOR_STATUS_NONE, entityID, room.RoomID)
	}
	if len(room.MapEntity) == 0 {
		c.DelRoomFromID(room.RoomID)
		log.Info("清理房间成功：", room.RoomID)
	}
}

func (m MatchEntityIDSlice) Len() int {
	return len(m)
}

func (m MatchEntityIDSlice) Less(i, j int) bool {
	return m[i].EnterStamp < m[j].EnterStamp
}

func (m MatchEntityIDSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// 生成不重复的房间ID
func (this *_BattleMatch) getRoomRandomID(start int, end int) uint32 {
	//范围检查
	if end < start {
		return 0
	}
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//生成随机数
	num := uint32(r.Intn((end - start)) + start)
	if this.IsExistRoomID(num) {
		num = this.getRoomRandomID(100000, 999999)
	}
	return num
}

//func (c *_BattleMatch) testMatch(lv uint32) {
//	r := rand.New(rand.NewSource(time.Now().UnixNano()))
//	num := r.Intn(1000) + 100
//	fmt.Println(num, "-", lv)
//	ch, chName, _ := c.getMatchChan(lv)
//	for i := 0; i < num; i++ {
//		EnterMatch(uint32(i), lv, false, ch, chName)
//	}
//}
