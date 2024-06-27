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
	"gitee.com/go-package/carbon/v2"
	"reflect"
	"sync"
	"time"
)

/***
 *@disc:机器人管理
 *@author: lsj
 *@date: 2023/11/16
 */

type _Robot struct {
	IsSync         bool
	DefaultCount   uint32
	DefaultCollect map[uint32]uint8
	DefaultFree    uint32
	HighCount      uint32
	HighCollect    map[uint32]uint8
	HighFree       uint32
}

const RobotGold uint32 = 10000

const (
	IdleState uint8 = iota
	BusyState
)

var RobotMr _Robot

var RobotMutex sync.Mutex

func (c *_Robot) Init() {
	c.ResetInitRobot()
	c.IsSync = false
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncRobotToGame), reflect.ValueOf(c.SyncEntityRobotFromDB))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_FirstLoadRobotResponse), reflect.ValueOf(c.FirstLoadRobotEntityRobotFromDB))
	timer.AddTimer(c, "GetRobotSum", 5000, true)

	time.AfterFunc(time.Millisecond*500, c.FirstSyncRobotFromDb)
}

// 初始化机器人
func (c *_Robot) ResetInitRobot() {
	c.DefaultCount, c.HighCount, c.DefaultFree, c.HighFree, c.DefaultCollect, c.HighCollect = 0, 0, 0, 0, make(map[uint32]uint8, 0), make(map[uint32]uint8, 0)
}

// game启动，并延迟通知DB同步robot数据,全量同步
func (c *_Robot) FirstSyncRobotFromDb() {
	resBody := &gmsg.SyncEntityRobotNoticeDB{}
	resBody.TimeStamp = uint32(carbon.Now().Timestamp())

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_FirstLoadRobotRequest), resBody, network.ServerType_DB)
}

// db服同步robot到游戏服，增量拉取
func (c *_Robot) SyncEntityRobotFromDB(msgEV *network.MsgBodyEvent) {
	RobotMutex.Lock()
	defer RobotMutex.Unlock()
	argLen, defaultNum, highNum := c.syncRobotFromMsgBody(msgEV.MsgBody)
	c.setIsSyncRobot(argLen)
	log.Info("本次同步机器人数量：", argLen, "，普通机器：", defaultNum, "，高级机器人：", highNum)
}

// db服同步robot到游戏服，增量拉取
func (c *_Robot) FirstLoadRobotEntityRobotFromDB(msgEV *network.MsgBodyEvent) {
	RobotMutex.Lock()
	defer RobotMutex.Unlock()
	c.ResetInitRobot()
	argLen, _, _ := c.syncRobotFromMsgBody(msgEV.MsgBody)
	log.Info("机器人初始化成功：", argLen, "，普通机器总数：", c.DefaultCount, "，高级机器人总数：", c.HighCount)
	c.setIsSyncRobot(argLen)

	if argLen == 0 {
		reg := &gmsg.InRegRobotRequest{}
		reg.Param = 40
		reg.High = 10
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_RegRobot), reg, network.ServerType_DB)
	}
}

func (c *_Robot) syncRobotFromMsgBody(msgBody []byte) (int, uint32, uint32) {
	var tEntityRobotArgs []entity.EntityPlayer

	stack.BytesToStruct_Gob(msgBody, &tEntityRobotArgs)

	defaultNum, highNum := uint32(0), uint32(0)
	for _, tEntityRobot := range tEntityRobotArgs {
		robot := tEntityRobot
		if robot.NumGold < 50000 {
			c.DefaultCollect[robot.EntityID] = IdleState
			defaultNum++
		} else {
			c.HighCollect[robot.EntityID] = IdleState
			highNum++
		}
		Entity.EmPlayer.AddEntity(&robot)
	}

	c.DefaultCount = c.DefaultCount + defaultNum
	c.DefaultFree = c.DefaultFree + defaultNum
	c.HighCount = c.HighCount + highNum
	c.HighFree = c.HighFree + highNum
	return len(tEntityRobotArgs), defaultNum, highNum
}

func (c *_Robot) setIsSyncRobot(len int) {
	if len == 0 {
		return
	}
	c.IsSync = true
}

func (c *_Robot) GetRobotSum() {
	if !c.IsSync {
		return
	}
	log.Info("->空闲->普通机器人：", c.DefaultFree, " || 高级机器人：", c.HighFree)
	if c.DefaultFree < 30 || c.HighFree < 10 {
		reg := &gmsg.InRegRobotRequest{}
		if c.DefaultFree < 30 {
			reg.Param = 20
		}
		if c.HighFree < 10 {
			reg.High = 10
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_RegRobot), reg, network.ServerType_DB)
	}
}

func (c *_Robot) IncrRobot(isHigh bool, entityID uint32) {
	RobotMutex.Lock()
	defer RobotMutex.Unlock()
	if isHigh {
		c.HighFree += 1
		c.HighCollect[entityID] = IdleState
	} else {
		c.DefaultFree += 1
		c.DefaultCollect[entityID] = IdleState
	}
}

func (c *_Robot) DeductRobot(isHigh bool, entityID uint32) {
	RobotMutex.Lock()
	defer RobotMutex.Unlock()
	if isHigh {
		c.HighFree -= 1
		delete(c.HighCollect, entityID)
	} else {
		c.DefaultFree -= 1
		delete(c.DefaultCollect, entityID)
	}
}

func (c *_Robot) IsRobotHighLevel(gold uint32) bool {
	return gold > RobotGold
}

func (c *_Robot) MatchHighRobot() uint32 {
	RobotMutex.Lock()
	defer RobotMutex.Unlock()
	for val, _ := range c.HighCollect {
		tEntity := Entity.EmPlayer.GetEntityByID(val)
		if tEntity == nil {
			continue
		}
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		if tEntityPlayer.BehaviorStatus <= battleConf.BEHAVIOR_STATUS_HALL {
			tEntityPlayer.SetBehaviorStatus(battleConf.BEHAVIOR_STATUS_MATCH)
			tEntityPlayer.SyncEntity(0)
			return val
		}
	}
	return uint32(0)
}

func (c *_Robot) MatchDefaultRobot() uint32 {
	RobotMutex.Lock()
	defer RobotMutex.Unlock()
	for val, _ := range c.DefaultCollect {
		tEntity := Entity.EmPlayer.GetEntityByID(val)
		if tEntity == nil {
			continue
		}
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		if tEntityPlayer.BehaviorStatus <= battleConf.BEHAVIOR_STATUS_HALL {
			tEntityPlayer.SetBehaviorStatus(battleConf.BEHAVIOR_STATUS_MATCH)
			tEntityPlayer.SyncEntity(0)
			return val
		}
	}
	return uint32(0)
}
