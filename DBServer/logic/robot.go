package logic

import (
	"BilliardServer/Common/entity"
	battleConf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"reflect"
	"time"
)

/***
 *@disc:
 *@author: lsj
 *@date: 2023/11/16
 */

type _Robot struct {
	DefaultNum int //每次拉取40个
	HighNum    int //每次拉取10个
	RobotList  map[uint32]uint8
}

var RobotMr _Robot

func (c *_Robot) Init() {
	c.DefaultNum, c.HighNum = 40, 10
	c.RobotList = make(map[uint32]uint8, 0)

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncRobotFromDb), reflect.ValueOf(c.GetGameSyncRobotRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_FirstLoadRobotRequest), reflect.ValueOf(c.FirstSyncRobotRequest))
}

func (c *_Robot) FirstSyncRobotRequest(msgEV *network.MsgBodyEvent) {
	buf := c.firstLoadEntityRobotPlayerArgs()
	if buf == nil {
		return
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_FirstLoadRobotResponse), buf, network.ServerType_Game)
}

func (c *_Robot) GetGameSyncRobotRequest(msgEV *network.MsgBodyEvent) {
	c.SyncEntityClubNoticeDB()
}

func (c *_Robot) SyncEntityClubNoticeDB() {
	buf := c.getEntityRobotPlayerArgs()
	if buf == nil {
		return
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncRobotToGame), buf, network.ServerType_Game)
}

func (c *_Robot) getEntityRobotPlayerArgs() []byte {
	tEntityRobotPlayer, _, _ := c.addIdleEntityRobot()
	buf, _ := stack.StructToBytes_Gob(tEntityRobotPlayer)
	if len(buf) < 1 {
		return nil
	}
	return buf
}

func (c *_Robot) FirstSyncEntityRobotToGame() {
	buf := c.firstLoadEntityRobotPlayerArgs()
	if buf == nil {
		return
	}
	time.AfterFunc(time.Second*5, func() {
		ConnectManager.SendMsgToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_FirstLoadRobotResponse), buf, network.ServerType_Game)
	})
}

func (c *_Robot) firstLoadEntityRobotPlayerArgs() []byte {
	c.RobotList = make(map[uint32]uint8, 0)
	tEntityRobotPlayer, _, _ := c.addIdleEntityRobot()

	buf, _ := stack.StructToBytes_Gob(tEntityRobotPlayer)
	if len(buf) < 1 {
		return nil
	}
	return buf
}

// 添加机器人
func (c *_Robot) addIdleEntityRobot() ([]entity.EntityPlayer, int, int) {
	tEntityRobotPlayer := make([]entity.EntityPlayer, 0)

	defaultNum, highNum := 0, 0
	for _, Value := range Entity.EmEntityPlayer.EntityMap {
		tEntityRobot := Value.(*entity.EntityPlayer)
		if !tEntityRobot.IsRobot || tEntityRobot.BehaviorStatus > battleConf.BEHAVIOR_STATUS_HALL {
			continue
		}
		if _, ok := c.RobotList[tEntityRobot.EntityID]; !ok && tEntityRobot.IsRobot {
			if defaultNum == c.DefaultNum && highNum == c.HighNum {
				break
			}
			if defaultNum < c.DefaultNum && tEntityRobot.NumGold < 50000 {
				tEntityRobotPlayer = append(tEntityRobotPlayer, *tEntityRobot)
				c.RobotList[tEntityRobot.EntityID] = 1
				defaultNum++
			}
			if highNum < c.HighNum && tEntityRobot.NumGold > 10000 {
				tEntityRobotPlayer = append(tEntityRobotPlayer, *tEntityRobot)
				c.RobotList[tEntityRobot.EntityID] = 1
				highNum++
			}
		}
	}

	return tEntityRobotPlayer, defaultNum, highNum
}

// 查询空闲的机器数量
func (c *_Robot) getIdleEntityRobot() (defaultNum, highNum int) {
	for _, Value := range Entity.EmEntityPlayer.EntityMap {
		tEntityRobot := Value.(*entity.EntityPlayer)
		if !tEntityRobot.IsRobot || tEntityRobot.BehaviorStatus > battleConf.BEHAVIOR_STATUS_HALL {
			continue
		}
		if _, ok := c.RobotList[tEntityRobot.EntityID]; !ok && tEntityRobot.IsRobot {
			if defaultNum == c.DefaultNum && highNum == c.HighNum {
				break
			}
			if defaultNum < c.DefaultNum && tEntityRobot.NumGold < 50000 {
				defaultNum++
			}
			if highNum < c.HighNum && tEntityRobot.NumGold > 10000 {
				highNum++
			}
		}
	}

	return defaultNum, highNum
}
