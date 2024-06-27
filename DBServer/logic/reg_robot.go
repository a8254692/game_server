package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"github.com/beego/beego/v2/core/logs"
	"reflect"
)

/***
 *@disc:
 *@author: lsj
 *@date: 2023/11/17
 */

type _RegRobot struct {
	AccRobotEntityIdNow int //当前自增长ID
	AccRobotCount       int //当前帐号总数
}

var RegRobotMr _RegRobot

func (c *_RegRobot) Init() {
	c.AccRobotEntityIdNow = 200000000
	c.AccRobotCount = 0
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_RegRobot), reflect.ValueOf(c.OnRegRobot))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_ResetRobot), reflect.ValueOf(c.OnResetRobot))
	timer.AddTimer(c, "OnMongoDBInItComplete", 200, false)
}

type LoginBack struct {
	Code     int
	Err      string
	Token    string
	EntityID uint32
	GateAdr  string
}

func (c *_RegRobot) OnMongoDBInItComplete() {
	tEntityMainType := entity.UnitAcc
	entityCount, err := DBConnect.GetDataCountTotal(tEntityMainType)
	if err != nil {
		logs.Warning("-->RegRobot----OnMongoDBInItComplete----Err:", err.Error())
		return
	}

	c.AccRobotCount = entityCount
	c.AccRobotEntityIdNow = c.AccRobotEntityIdNow + entityCount
	logs.Info("-->RegRobot Init Complete, AccEntityIdNow:", c.AccRobotEntityIdNow)
}

func (c *_RegRobot) regAccForName(username string) (uint32, string) {
	password := "123456"
	tEntityAcc := new(entity.EntityAcc)
	errExist := DBConnect.GetData(entity.UnitAcc, "AccUnique", username, tEntityAcc)
	if errExist == nil {
		return 0, ""
	}
	tEntityID32 := int32(c.AccRobotEntityIdNow)
	tEntityID := tools.GetEntityID(&tEntityID32)
	c.AccRobotEntityIdNow = int(tEntityID)
	tEntityAcc.InitByFirst(entity.UnitAcc, uint32(c.AccRobotEntityIdNow))
	tEntityAcc.AccUnique = username
	tEntityAcc.PassWord = password
	errInsert := tEntityAcc.InsertEntity(DBConnect)
	if errInsert != nil {
		logs.Warning("-->LoginRegister RegisterAcc Error:", errInsert)
	}

	return tEntityAcc.EntityID, username
}

func (c *_RegRobot) regAcc() (uint32, string) {
	username := Table.GetPlayerRandName()
	password := "123456"
	tEntityAcc := new(entity.EntityAcc)
	errExist := DBConnect.GetData(entity.UnitAcc, "AccUnique", username, tEntityAcc)
	if errExist == nil {
		return 0, ""
	}
	tEntityID32 := int32(c.AccRobotEntityIdNow)
	tEntityID := tools.GetEntityID(&tEntityID32)
	c.AccRobotEntityIdNow = int(tEntityID)
	tEntityAcc.InitByFirst(entity.UnitAcc, uint32(c.AccRobotEntityIdNow))
	tEntityAcc.AccUnique = username
	tEntityAcc.PassWord = password
	errInsert := tEntityAcc.InsertEntity(DBConnect)
	if errInsert != nil {
		logs.Warning("-->LoginRegister RegisterAcc Error:", errInsert)
	}

	return tEntityAcc.EntityID, username
}

func (c *_RegRobot) player(entityID uint32, u string, gold uint32) {
	tEntityAcc := new(entity.EntityAcc)
	tEntityAcc.SetDBConnect(entity.UnitAcc)
	ok, err := tEntityAcc.InitFormDB(entityID, DBConnect)

	if err != nil {
		log.Error("-->logic._Entity--OnSyncEntityFormGame--tEntityAcc.InitFormDB--err:", err, entityID)
		return
	}
	if !ok {
		return
	}
	tEntityAcc.CollectionName = entity.UnitAcc

	Entity.EmEntityAcc.AddEntity(tEntityAcc) //添加进实体管理器

	tEntityPlayer := new(entity.EntityPlayer)
	tEntityPlayer.InitByFirst(entity.UnitPlayer, entityID)
	tEntityPlayer.IsRobot = true
	tEntityPlayer.PlayerName = u
	tEntityPlayer.Sex = 1
	if gold > 0 {
		tEntityPlayer.NumGold = gold
	}

	tEntityPlayer.InsertEntity(DBConnect)

	tUnitPlayerBase := new(entity.UnitPlayerBase)
	stack.SimpleCopyProperties(tUnitPlayerBase, tEntityPlayer)
	tEntityAcc.ListPlayer = append(tEntityAcc.ListPlayer, *tUnitPlayerBase)
	tEntityAcc.SaveEntity(DBConnect)
	Entity.EmEntityPlayer.AddEntity(tEntityPlayer)
	InitPlayerMr.InitPlayerData(entityID)
}

func (c *_RegRobot) OnRegRobot(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InRegRobotRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		return
	}

	defaultNum, highNum := RobotMr.getIdleEntityRobot()
	if defaultNum > 0 && highNum > 0 {
		RobotMr.SyncEntityClubNoticeDB()
		return
	}

	if req.Param < uint32(0) {
		return
	}
	//普通机器人
	c.regRobot(req.Param, 10000)
	//高级机器人
	c.regRobot(req.High, 50000)

	RobotMr.SyncEntityClubNoticeDB()
}

func (c *_RegRobot) regRobot(num, gold uint32) {
	for i := uint32(1); i <= num; i++ {
		enid, username := c.regAcc()
		if enid == 0 {
			continue
		}
		c.player(enid, username, gold)
	}
}

func (c *_RegRobot) OnResetRobot(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InRegRobotRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		return
	}

	for _, Value := range Entity.EmEntityPlayer.EntityMap {
		tEntityRobot := Value.(*entity.EntityPlayer)
		if tEntityRobot.IsRobot {
			tEntityRobot.BehaviorStatus = conf.BEHAVIOR_STATUS_NONE
			tEntityRobot.RoomId = 0
			tEntityRobot.SaveEntity(DBConnect)
		}
	}

	RobotMr.SyncEntityClubNoticeDB()
}
