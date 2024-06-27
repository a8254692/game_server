package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Util/log"
	"BilliardServer/Util/tools"
	"time"
)

/***
 *@disc: 任务定时器
 *@author: lsj
 *@date: 2023/11/2
 */
//todo gameServer和dbServer都有重置，gameServer只处理在线用户并通知dbServer同步，dbServer处理所有用户（gameServer已重置过的，db会跳过）

/*
1.用户每天任务，0点重置
2.俱乐部评级，每周1的0点
3.俱乐部任务，每周6的0点
*/
type _TickTask struct {
}

var TickTaskMr _TickTask

// 每天0点重置
func (c *_TickTask) tick() {
	leftSecond := time.Duration(tools.GetLeftSecondByTomorrow()) * time.Second
	time.AfterFunc(leftSecond, func() {
		go c.ResetGameDayTask()
		c.tick()
	})
}

// 当天0点的时间戳
func (c *_TickTask) getTodayBeginTime() int64 {
	return tools.GetTodayBeginTime()
}

// 获取本周周1的时间戳
func (c *_TickTask) getThisWeekFirstTime() int64 {
	return tools.GetThisWeekFirstDate()
}

// 判断是否周1
func (c *_TickTask) MonDay() int {
	return 1
}

// 判断是否周6
func (c *_TickTask) Saturday() int {
	return 6
}

// 获取当前时间的日期格式
func (c *_TickTask) getNowDateString() string {
	return tools.GetNowDateString()
}

// 获取本周周一的日期格式
func (c *_TickTask) getThisWeekFirstDateString() string {
	return tools.GetThisWeekFirstDateString()
}

// 返回布尔
func (c *_TickTask) isNowDateString(date string) bool {
	return date == c.getNowDateString()
}

// 返回布尔
func (c *_TickTask) isThisWeekFirstDateString(date string) bool {
	return date == c.getThisWeekFirstDateString()
}

// 12点定时重置任务
func (c *_TickTask) ResetGameDayTask() {
	log.Info("-->ResetGameDayTask-->begin-->")
	Task.setDayTaskList()
	log.Info("TaskDefaultList", Task.TaskDefaultList)
	log.Info("TaskLessClub", Task.TaskLessClub)
	weekDay := tools.GetWeekDay()
	//刷新俱乐部任务和商店
	if weekDay == c.Saturday() {
		ClubManager.clubShopCfg()
		ClubManager.resetClubTaskAndShop()
	}
	//刷新在线玩家任务和数据
	for _, emPlayer := range Entity.EmPlayer.EntityMap {
		p := emPlayer.(*entity.EntityPlayer)
		if p.IsRobot {
			continue
		}
		if !c.isNowDateString(p.TaskResetDate.DayDate) {
			p.ResetGiveGoldList()
			//重置任务函数
			Task.setPlayerDailyTaskList(p)
			//重置领取表
			Task.setDayProgressRewardList(p)
			// 重置俱乐部任务
			if p.ClubId > 0 {
				if weekDay == c.Saturday() {
					p.ReSetClubAttribute()
					p.ResetEntityClubShop()
				} else {
					p.DailyReSetClubAttribute()
				}
				//推送俱乐部任务
				ClubManager.SyncClubTaskList(p)
			}
			//重置登录奖励数据
			LoginReward.updateLoginReward(p)
		}
		if weekDay == c.MonDay() && !c.isThisWeekFirstDateString(p.TaskResetDate.WeekDate) {
			Task.setWeekProgressRewardList(p)
		}

		p.SyncEntity(1)
		Task.TaskListResetSync(p.EntityID)
		SocialManager.MyFriendListSync(p)
		WelfareMr.LoginPlayerSinInListSync(p.EntityID)
	}

	log.Info("-->ResetGameDayTask->end-->")
}
