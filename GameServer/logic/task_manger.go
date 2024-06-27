package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/stack"
)

/***
 *@disc: 任务管理
 *@author: lsj
 *@date: 2023/10/10
 */

type _TaskMr struct {
}

var TaskManger _TaskMr

// 更新任务
func (c *_TaskMr) UpdateTaskFromConditionID(tEntityPlayer *entity.EntityPlayer, conditional, progress uint32) {
	taskInfo := tEntityPlayer.UpdateTaskFromConditionID(conditional, progress)
	if conditional == conf.ClubTaskID {
		tEntityPlayer.UpdateClubTags()
	}
	if taskInfo != nil {
		taskInfoSync := &gmsg.TaskListSync{}
		taskInfoSync.EntityID = tEntityPlayer.EntityID
		taskInfoSync.TaskInfo = new(gmsg.TaskInfo)
		stack.SimpleCopyProperties(taskInfoSync.TaskInfo, taskInfo)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_TaskListSync, taskInfoSync, []uint32{taskInfoSync.EntityID})
	}
}

func (c *_TaskMr) UpdateBattleTask(result uint32) []conf.ConditionData {
	cond := make([]conf.ConditionData, 0)
	if result == conf.RESULT_VICTORY {
		cond = append(cond, conf.ConditionData{conf.FirstWinX, 1, false})
	}
	cond = append(cond, conf.ConditionData{conf.RealPlayerBattle, 1, false})
	return cond
}
