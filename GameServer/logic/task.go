package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"reflect"
	"sort"
	"time"
)

/***
 *@disc: 每日任务
 *@author: lsj
 *@date: 2023/10/8
 */

type _Task struct {
	TaskDefaultList        []entity.Task
	TaskLessClub           []entity.Task
	DayProgressRewardList  []entity.ProgressReward
	WeekProgressRewardList []entity.ProgressReward
}

var Task _Task

func (c *_Task) Init() {
	c.TaskDefaultList = make([]entity.Task, 0)
	c.TaskLessClub = make([]entity.Task, 0)
	c.DayProgressRewardList = make([]entity.ProgressReward, 0)
	c.WeekProgressRewardList = make([]entity.ProgressReward, 0)
	c.setDayTaskList()
	c.initProgressRewardList()
	TickTaskMr.tick()
	event.OnNet(gmsg.MsgTile_Task_ProgressClaimRewardRequest, reflect.ValueOf(c.OnTaskProgressClaimRewardRequest))
	event.OnNet(gmsg.MsgTile_Task_ListClaimRewardRequest, reflect.ValueOf(c.OnTaskListClaimRewardRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_DBServerStartUp), reflect.ValueOf(c.DbServerStartUpRequest))
}

func (c *_Task) DbServerStartUpRequest(msgEV *network.MsgBodyEvent) {
	GiftsMr.SendSyncMsgToDbForPopData()
	RobotMr.FirstSyncRobotFromDb()
	ClubManager.SyncEntityClub()
	log.Info("--->Db启动成功-->人气排行榜、机器人、俱乐部同步数据开始。")
}

// 设置每日任务
func (c *_Task) setDayTaskList() {
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
func (c *_Task) initProgressRewardList() {
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

func (c *_Task) setPlayerDailyTaskList(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.DayProgressValue = 0
	tEntityPlayer.TaskList = nil
	if tEntityPlayer.ClubTags {
		tEntityPlayer.TaskList = c.TaskLessClub
	} else {
		tEntityPlayer.TaskList = c.TaskDefaultList
	}
}

// 重置领取表
func (c *_Task) setDayProgressRewardList(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.DayProgressReward = nil
	dayProgress := new(entity.ProgressList)
	dayProgress.DateStamp = c.getTodayBeginTime()
	tEntityPlayer.DayProgressReward = append(tEntityPlayer.DayProgressReward, *dayProgress)
	tEntityPlayer.DayProgressReward[0].ProgressRewardList = make([]entity.ProgressReward, 0)
	tEntityPlayer.DayProgressReward[0].ProgressRewardList = c.DayProgressRewardList

	tEntityPlayer.TaskResetDate.DayDate = c.getNowDateString()
}

// 重置领取表
func (c *_Task) setWeekProgressRewardList(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.WeekProgressValue = 0
	tEntityPlayer.WeekProgressReward = nil
	dayProgress := new(entity.ProgressList)
	dayProgress.DateStamp = c.getThisWeekFirstTime()
	tEntityPlayer.WeekProgressReward = append(tEntityPlayer.WeekProgressReward, *dayProgress)
	tEntityPlayer.WeekProgressReward[0].ProgressRewardList = make([]entity.ProgressReward, 0)
	tEntityPlayer.WeekProgressReward[0].ProgressRewardList = c.WeekProgressRewardList

	tEntityPlayer.TaskResetDate.WeekDate = c.getThisWeekFirstDateString()
}

func (c *_Task) getTodayBeginTime() int64 {
	return tools.GetTodayBeginTime()
}

// 判断是否周1
func (c *_Task) isMonDay() bool {
	return tools.GetWeekDay() == 1
}

// 获取当前时间的日期格式
func (c *_Task) getNowDateString() string {
	return tools.GetNowDateString()
}

// 获取本周周一的日期格式
func (c *_Task) getThisWeekFirstDateString() string {
	return tools.GetThisWeekFirstDateString()
}

// 获取本周周1的时间戳
func (c *_Task) getThisWeekFirstTime() int64 {
	return tools.GetThisWeekFirstDate()
}

// 判断时间戳是否达到重置
// 1.超过周六比重置
// 2.超过7天了比重置
func (c *_Task) isPassThisWeekSaturday(sec int64) bool {
	weekSarSec := tools.GetThisWeekSaturday()
	return (time.Now().Unix() <= weekSarSec && weekSarSec-sec >= 7*86400) || (time.Now().Unix() >= weekSarSec && sec <= weekSarSec)
}

// 活跃进度表领取奖励
func (c *_Task) OnTaskProgressClaimRewardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.TaskProgressClaimRewardRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}
	log.Info("OnTaskProgressClaimRewardRequest", msgBody)
	msgResponse := &gmsg.TaskProgressClaimRewardResponse{}
	msgResponse.Code = 2
	msgResponse.TaskProgressKey = msgBody.TaskProgressKey

	if msgBody.TaskProgressKey > conf.DayProgress*100 && msgBody.TaskProgressKey < conf.WeekProgress*100 {
		if tools.GetNowDateString() == tEntityPlayer.TaskResetDate.DayDate {
			c.taskDayProgressClaimReward(tEntityPlayer, msgBody.TaskProgressKey)
			return
		}
	} else if msgBody.TaskProgressKey > conf.WeekProgress*100 {
		if tools.GetThisWeekFirstDateString() == tEntityPlayer.TaskResetDate.WeekDate {
			c.taskWeekProgressClaimReward(tEntityPlayer, msgBody.TaskProgressKey)
			return
		}
	} else {
		log.Error("非法的key", msgBody.TaskProgressKey)
		return
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Task_ProgressClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 0分数达到并成功，1分数未达到，2领取异常(可能是跨天领取)
func (c *_Task) taskDayProgressClaimReward(tEntityPlayer *entity.EntityPlayer, taskProgressKey uint32) {
	msgResponse := &gmsg.TaskProgressClaimRewardResponse{}
	msgResponse.Code = 1
	msgResponse.TaskProgressKey = taskProgressKey
	taskProgressCfg := Table.GetTaskProgressCfg(taskProgressKey)
	if taskProgressCfg == nil {
		log.Error("非法的key", taskProgressKey)
		return
	}

	if tEntityPlayer.IsInDayProgressRewardList(taskProgressKey) {
		msgResponse.Code = 3
	}

	if tEntityPlayer.TaskDayProgressToValue(uint32(taskProgressCfg.Progress)) && !tEntityPlayer.IsInDayProgressRewardList(taskProgressKey) {
		msgResponse.Code = 0
		tEntityPlayer.TaskDayProgressClaimReward(taskProgressKey)
		tEntityPlayer.SyncEntity(1)
		c.sendCommonReward(tEntityPlayer.EntityID, [][]uint32{taskProgressCfg.Rewards})
	}
	log.Info("-->taskDayProgressClaimReward-->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Task_ProgressClaimRewardResponse, msgResponse, []uint32{tEntityPlayer.EntityID})
}

// 0分数达到并成功，1分数未达到，2领取异常(可能是跨周领取)
func (c *_Task) taskWeekProgressClaimReward(tEntityPlayer *entity.EntityPlayer, taskProgressKey uint32) {
	msgResponse := &gmsg.TaskProgressClaimRewardResponse{}
	msgResponse.Code = 1
	msgResponse.TaskProgressKey = taskProgressKey
	taskProgressCfg := Table.GetTaskProgressCfg(taskProgressKey)
	if taskProgressCfg == nil {
		log.Error("非法的key", taskProgressKey)
		return
	}

	if tEntityPlayer.IsInWeekProgressRewardList(taskProgressKey) {
		msgResponse.Code = 3
	}

	if tEntityPlayer.TaskWeekProgressToValue(uint32(taskProgressCfg.Progress)) && !tEntityPlayer.IsInWeekProgressRewardList(taskProgressKey) {
		msgResponse.Code = 0
		tEntityPlayer.TaskWeekProgressClaimReward(taskProgressKey)
		tEntityPlayer.SyncEntity(1)
		c.sendCommonReward(tEntityPlayer.EntityID, [][]uint32{taskProgressCfg.Rewards})
	}
	log.Info("-->taskWeekProgressClaimReward-->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Task_ProgressClaimRewardResponse, msgResponse, []uint32{tEntityPlayer.EntityID})
}

// 调用通用奖励接口
func (c *_Task) sendCommonReward(entityID uint32, reward [][]uint32) {
	if len(reward) == 0 {
		return
	}
	rewardList := make([]entity.RewardEntity, 0)
	for _, val := range reward {
		rewardEntity := new(entity.RewardEntity)
		rewardEntity.ItemTableId = val[0]
		rewardEntity.Num = val[1]
		rewardEntity.ExpireTimeId = 0
		rewardList = append(rewardList, *rewardEntity)
	}

	resParam := GetResParam(conf.SYSTEM_ID_TASK, conf.Reward)
	Backpack.BackpackAddItemListAndSave(entityID, rewardList, *resParam)
}

// 任务表领取奖励
func (c *_Task) OnTaskListClaimRewardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.TaskListClaimRewardRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}
	msgResponse := &gmsg.TaskListClaimRewardResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = 1

	if tools.GetNowDateString() != tEntityPlayer.TaskResetDate.DayDate {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Task_ListClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	task, index := tEntityPlayer.IsInTaskList(msgBody.TaskID)
	if index >= 0 {
		if task.State == conf.TaskState && task.StateReward == conf.TaskUnStateReward {
			tEntityPlayer.ClaimTaskListReward(index)
			dayValue, weekValue := tEntityPlayer.AddTaskProgressValue(conf.TaskProgressValue)
			tEntityPlayer.SyncEntity(1)
			msgResponse.Code = 0
			msgResponse.Task = new(gmsg.TaskInfo)
			msgResponse.DayProgress = dayValue
			msgResponse.WeekProgress = weekValue
			stack.SimpleCopyProperties(msgResponse.Task, tEntityPlayer.TaskList[index])
			taskCfg := Table.GetTaskCfg(msgBody.TaskID)
			if taskCfg == nil {
				return
			}
			c.sendCommonReward(msgBody.EntityID, taskCfg.Rewards)
			ConditionalMr.SyncConditional(msgBody.EntityID, []conf.ConditionData{{conf.TaskTimes, 1, false}})
		}
	} else {
		log.Error("taskID不存在", msgBody.TaskID)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Task_ListClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Task) TaskListResetSync(entityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	msgResponse := &gmsg.TaskListResetSync{}
	msgResponse.EntityID = entityID
	msgResponse.DayProgress = tEntityPlayer.DayProgressValue
	msgResponse.WeekProgress = tEntityPlayer.WeekProgressValue
	msgResponse.TaskList = c.getTaskList(tEntityPlayer.TaskList)
	msgResponse.DayProgressList = c.getProgressRewardList(tEntityPlayer.DayProgressReward[0].ProgressRewardList)
	msgResponse.WeekProgressList = c.getProgressRewardList(tEntityPlayer.WeekProgressReward[0].ProgressRewardList)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_TaskListResetSync, msgResponse, []uint32{entityID})
}

func (c *_Task) ResetTaskTest(entityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	c.setPlayerDailyTaskList(tEntityPlayer)
	c.setDayProgressRewardList(tEntityPlayer)
	c.setWeekProgressRewardList(tEntityPlayer)
	tEntityPlayer.SyncEntity(1)
	c.TaskListResetSync(entityID)
}

func (c *_Task) LoginTaskDailyReset(entityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if !TickTaskMr.isNowDateString(tEntityPlayer.TaskResetDate.DayDate) {
		tEntityPlayer.ResetGiveGoldList()
		//重置任务函数
		c.setPlayerDailyTaskList(tEntityPlayer)
		//重置领取表
		c.setDayProgressRewardList(tEntityPlayer)
		// 重置俱乐部任务
		// 如果已过另一周，直接重置全部
		if tEntityPlayer.ClubId > 0 {
			if c.isPassThisWeekSaturday(tEntityPlayer.ClubAttribute.ClubReFreshUnix) {
				log.Info("-->LoginTaskDailyReset,用户俱乐部任务超过周六或者超过7天,entityid:", entityID)
				tEntityPlayer.ReSetClubAttribute()
				tEntityPlayer.ResetEntityClubShop()
			} else {
				log.Info("-->LoginTaskDailyReset,用户俱乐部任务每天重置,entityid:", entityID)
				tEntityPlayer.DailyReSetClubAttribute()
			}
		}
		//重置登录奖励数据
		LoginReward.updateLoginReward(tEntityPlayer)
	}
	if !TickTaskMr.isThisWeekFirstDateString(tEntityPlayer.TaskResetDate.WeekDate) {
		c.setWeekProgressRewardList(tEntityPlayer)
	}
	tEntityPlayer.SyncEntity(1)
}

func (c *_Task) getPlayerTask(entityID uint32) (uint32, uint32, []*gmsg.Progress, []*gmsg.Progress, []*gmsg.TaskInfo) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return 0, 0, nil, nil, nil
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	dayProgressRewardList, weekProgressRewardList, taskList := make([]*gmsg.Progress, 0), make([]*gmsg.Progress, 0), make([]*gmsg.TaskInfo, 0)
	dayProgressRewardList = c.getProgressRewardList(tEntityPlayer.DayProgressReward[0].ProgressRewardList)
	weekProgressRewardList = c.getProgressRewardList(tEntityPlayer.WeekProgressReward[0].ProgressRewardList)
	taskList = c.getTaskList(tEntityPlayer.TaskList)

	return tEntityPlayer.DayProgressValue, tEntityPlayer.WeekProgressValue, dayProgressRewardList, weekProgressRewardList, taskList
}

func (c *_Task) getTaskList(task []entity.Task) []*gmsg.TaskInfo {
	list := make([]*gmsg.TaskInfo, 0)
	for _, vl := range task {
		taskInfo := new(gmsg.TaskInfo)
		stack.SimpleCopyProperties(taskInfo, vl)
		list = append(list, taskInfo)
	}

	return list
}

func (c *_Task) getProgressRewardList(rewardList []entity.ProgressReward) []*gmsg.Progress {
	progressList := make([]*gmsg.Progress, 0)
	for _, vl := range rewardList {
		progress := new(gmsg.Progress)
		stack.SimpleCopyProperties(progress, vl)
		progressList = append(progressList, progress)
	}
	return progressList
}
