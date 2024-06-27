package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/tools"
	"reflect"
	"sort"
	"time"
)

/***
 *@disc:
 *@author: lsj
 *@date: 2023/10/10
 */

type _TaskDB struct {
	TaskDefaultList        []entity.Task
	TaskLessClub           []entity.Task
	DayProgressRewardList  []entity.ProgressReward
	WeekProgressRewardList []entity.ProgressReward
}

var TaskDBManger _TaskDB

func (c *_TaskDB) Init() {
	c.TaskDefaultList = make([]entity.Task, 0)
	c.TaskLessClub = make([]entity.Task, 0)
	c.DayProgressRewardList = make([]entity.ProgressReward, 0)
	c.WeekProgressRewardList = make([]entity.ProgressReward, 0)
	c.setDayTaskList()
	c.initProgressRewardList()
	c.tick()
	time.AfterFunc(time.Second*5, c.DbServerStartUpRequest)
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncConditionalToDB), reflect.ValueOf(c.OnSyncConditionalToDB))
}

// 设置每日任务
func (c *_TaskDB) setDayTaskList() {
	if len(Table.GetTaskList()) == 0 {
		log.Error("GetTaskList异常。")
		return
	}
	c.TaskDefaultList = nil
	c.TaskLessClub = nil
	list1, list2 := make([]entity.Task, 0), make([]entity.Task, 0)
	for _, vl := range Table.GetTaskList() {
		task := new(entity.Task)
		task.TaskId = vl.TaskID
		task.State = 0
		task.StateReward = 0
		task.CompleteProgress = 0
		task.TaskProgress = vl.Condition
		task.Timestamp = c.getTodayBeginTime()
		task.ConditionId = vl.ConditionID
		// 131是创建俱乐部或者加入俱乐部
		if vl.ConditionID != conf.ClubTaskID {
			list2 = append(list2, *task)
		}
		list1 = append(list1, *task)
	}
	c.TaskDefaultList = list1
	c.TaskLessClub = list2

	return
}

// 初始化领取表
func (c *_TaskDB) initProgressRewardList() {
	for _, vl := range Table.GetTaskProgressData() {
		progressReward := new(entity.ProgressReward)
		progressReward.StateReward = 0
		progressReward.ProgressID = vl.TableID
		if vl.TableID > conf.DayProgress*100 && vl.TableID < conf.WeekProgress*100 {
			c.DayProgressRewardList = append(c.DayProgressRewardList, *progressReward)
		} else {
			c.WeekProgressRewardList = append(c.WeekProgressRewardList, *progressReward)
		}
	}

	sort.Slice(c.DayProgressRewardList, func(i, j int) bool {
		return c.DayProgressRewardList[i].ProgressID < c.DayProgressRewardList[j].ProgressID
	})

	sort.Slice(c.WeekProgressRewardList, func(i, j int) bool {
		return c.WeekProgressRewardList[i].ProgressID < c.WeekProgressRewardList[j].ProgressID
	})
}

func (c *_TaskDB) setPlayerDailyTaskList(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.DayProgressValue = 0
	tEntityPlayer.TaskList = nil
	if tEntityPlayer.ClubTags {
		tEntityPlayer.TaskList = c.TaskLessClub
	} else {
		tEntityPlayer.TaskList = c.TaskDefaultList
	}
}

// 重置领取表
func (c *_TaskDB) setDayProgressRewardList(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.DayProgressReward = nil
	dayProgress := new(entity.ProgressList)
	dayProgress.DateStamp = c.getTodayBeginTime()
	tEntityPlayer.DayProgressReward = append(tEntityPlayer.DayProgressReward, *dayProgress)
	tEntityPlayer.DayProgressReward[0].ProgressRewardList = make([]entity.ProgressReward, 0)
	tEntityPlayer.DayProgressReward[0].ProgressRewardList = c.DayProgressRewardList

	tEntityPlayer.TaskResetDate.DayDate = c.getNowDateString()
}

// 重置领取表
func (c *_TaskDB) setWeekProgressRewardList(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.WeekProgressValue = 0
	tEntityPlayer.WeekProgressReward = nil
	dayProgress := new(entity.ProgressList)
	dayProgress.DateStamp = c.getThisWeekFirstTime()
	tEntityPlayer.WeekProgressReward = append(tEntityPlayer.WeekProgressReward, *dayProgress)
	tEntityPlayer.WeekProgressReward[0].ProgressRewardList = make([]entity.ProgressReward, 0)
	tEntityPlayer.WeekProgressReward[0].ProgressRewardList = c.WeekProgressRewardList

	tEntityPlayer.TaskResetDate.WeekDate = c.getThisWeekFirstDateString()
}

// 0点重置任务
func (c *_TaskDB) tick() {
	leftSecond := time.Duration(tools.GetLeftSecondByTomorrow()) * time.Second
	time.AfterFunc(leftSecond, func() {
		go c.ResetPlayerDbDayTaskList()
		c.tick()
	})
}

// 当天0点的时间戳
func (c *_TaskDB) getTodayBeginTime() int64 {
	return tools.GetTodayBeginTime()
}

// 获取本周周1的时间戳
func (c *_TaskDB) getThisWeekFirstTime() int64 {
	return tools.GetThisWeekFirstDate()
}

// 判断是否周1
func (c *_TaskDB) MonDay() int {
	return 1
}

// 判断是否周6
func (c *_TaskDB) Saturday() int {
	return 6
}

// 获取当前时间的日期格式
func (c *_TaskDB) getNowDateString() string {
	return tools.GetNowDateString()
}

// 获取本周周一的日期格式
func (c *_TaskDB) getThisWeekFirstDateString() string {
	return tools.GetThisWeekFirstDateString()
}

// 返回布尔
func (c *_TaskDB) isNowDateString(date string) bool {
	return date == c.getNowDateString()
}

// 返回布尔
func (c *_TaskDB) isThisWeekFirstDateString(date string) bool {
	return date == c.getThisWeekFirstDateString()
}

// 12点定时重置任务
func (c *_TaskDB) ResetPlayerDbDayTaskList() {
	log.Info("-->ResetPlayerDbDayTaskList-->begin-->")
	c.setDayTaskList()
	log.Info("TaskDefaultList", c.TaskDefaultList)
	log.Info("TaskLessClub", c.TaskLessClub)
	log.Info("-->ResetPlayerDbDayTaskList->Success!")
}

// 处理游戏服的任务数据
func (c *_TaskDB) OnSyncConditionalToDB(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SyncConditional{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	for _, val := range msgBody.Cond {
		if val.Progress == 0 && (val.ConditionalID == conf.FansNum || val.ConditionalID == conf.TotalMyFriends) {
			val.Progress = tEntityPlayer.FansNum
		}

		c.UpdateTaskFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress)
		c.UpdateCollectFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress, val.IsTotal)
		c.UpdateAchievementFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress, val.IsTotal)
	}

	tEntityPlayer.FlagChang()
}

func (c *_TaskDB) updateConditional(EntityID uint32, data []conf.ConditionData) {
	tEntity := Entity.EmEntityPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	for _, val := range data {
		if val.Progress == 0 && (val.ConditionalID == conf.FansNum || val.ConditionalID == conf.TotalMyFriends) {
			val.Progress = tEntityPlayer.FansNum
		}
		c.UpdateTaskFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress)
		c.UpdateCollectFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress, val.IsTotal)
		c.UpdateAchievementFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress, val.IsTotal)
	}

	tEntityPlayer.FlagChang()
}

// 更新任务
func (c *_TaskDB) UpdateTaskFromConditionID(tEntityPlayer *entity.EntityPlayer, conditional, progress uint32) {
	tEntityPlayer.UpdateTaskFromConditionID(conditional, progress)
	if conditional == conf.ClubTaskID {
		tEntityPlayer.UpdateClubTags()
	}
}

// 更新称号
func (c *_TaskDB) UpdateCollectFromConditionID(tEntityPlayer *entity.EntityPlayer, conditional, progress uint32, isTotal bool) {
	tEntityPlayer.UpdateCollectFromConditionID(conditional, progress, isTotal)
}

// 更新成就
func (c *_TaskDB) UpdateAchievementFromConditionID(tEntityPlayer *entity.EntityPlayer, conditional, progress uint32, isTotal bool) {
	childID := tEntityPlayer.UpdateAchievementFromConditionID(conditional, progress, isTotal)
	if len(childID) > 0 {
		// 查询获得的积分
		for _, child := range childID {
			score := Table.GetAchievementElementCfgScore(child)
			// 加角色积分
			tEntityPlayer.UpdatePlayerAchievement(score)
			isUp := Table.IsUpgradeAchievementLV(tEntityPlayer.AchievementScore, tEntityPlayer.AchievementLV)
			if isUp {
				tEntityPlayer.UpgradeAchievementLV()
			}
		}
	}
}

func (c *_TaskDB) DbServerStartUpRequest() {
	req := &gmsg.DbServerStartUpRequest{}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_DBServerStartUp), req, network.ServerType_Game)
}
