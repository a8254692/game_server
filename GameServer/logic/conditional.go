package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
)

/***
 *@disc:条件相关
 *@author: lsj
 *@date: 2023/10/9
 */

/*
类型0：无参数条件
类型1：达到条件。参数数量，大于等于完成
类型2：累计条件。参数数量，大于等于完成
类型3：每日达到类型，参数等于，数量大于等于完成
类型4：排名类，第一个参数是排行类型，相等完成。后面参数小于等于完成，数量累计
类型5：开服天数达到等于
类型6：一人完成，所有人都完成。完成方式和3相同
*/

type Conditional struct {
}

var ConditionalMr Conditional

func (c *Conditional) IsToCondition(conditionID uint32, condVl, value uint32) bool {
	cfg := Table.GetConditionalCfg(conditionID)
	if cfg == nil {
		return false
	}
	return c.enableCondition(cfg.TypeN, condVl, value)
}

func (c *Conditional) enableCondition(typeN, condVl, value uint32) bool {
	switch typeN {
	case conf.Conditional_0:
		return value > 0
	case conf.Conditional_1:
		return value >= condVl
	}
	return false
}

// 更新条件关系
func (c *Conditional) SyncConditional(EntityID uint32, cond []conf.ConditionData) {
	log.Info("-->SyncConditional-->begin-->", EntityID, "-->conditional-->", cond)
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		// 通知DB处理
		c.syncConditionalToDB(EntityID, cond)
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	for _, val := range cond {
		TaskManger.UpdateTaskFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress)
		CollectMr.UpdateCollectFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress, val.IsTotal)
		AchievementMr.UpdateAchievementFromConditionID(tEntityPlayer, val.ConditionalID, val.Progress, val.IsTotal)
	}

	tEntityPlayer.SyncEntity(1)
}

// 发送到db处理
func (c *Conditional) syncConditionalToDB(EntityID uint32, cond []conf.ConditionData) {
	msgBody := &gmsg.SyncConditional{}
	msgBody.EntityID = EntityID
	msgBody.Cond = make([]*gmsg.ConditionData, 0)
	for _, v := range cond {
		conds := new(gmsg.ConditionData)
		stack.SimpleCopyProperties(conds, &v)
		msgBody.Cond = append(msgBody.Cond, conds)
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncConditionalToDB), msgBody, network.ServerType_DB)
}

// 通过统计数据更新成就
func (c *Conditional) SyncConditionalStatics(data *DataStatistics, result uint32) {
	condData := TaskManger.UpdateBattleTask(result)
	cond := make([]conf.ConditionData, 0)
	cond = append(cond,
		conf.ConditionData{conf.TotalGold, data.GetAccumulateGold(), true},                   //累计获取金币
		conf.ConditionData{conf.BattleTotalGoal, data.GetAccumulateGoal(), true},             //累计进球个数
		conf.ConditionData{conf.BattleWinTimes, data.GetC8WinNum(), true},                    //胜利次数
		conf.ConditionData{conf.BattleOneCleaning, data.GetOneCueClear(), true},              //一杆清台
		conf.ConditionData{conf.BattleConnectingWin, data.GetC8MaxContinuousWin(), true},     //最大连胜次数
		conf.ConditionData{conf.BattleBaseScoreTimes, data.GetIncrBindNum(), true},           //加注次数
		conf.ConditionData{conf.ContinuousWins, data.GetC8ContinuousWin(), true},             //连胜次数(输了清零)
		conf.ConditionData{conf.BattleOneCueXGoal, data.GetC8MaxOneCueGoal(), true},          //最大一杆进球数
		conf.ConditionData{conf.OneCleaningTimes, data.GetOneCueClear(), true},               //清台次数
		conf.ConditionData{conf.BattleConnectingCue, data.GetC8MaxContinuousGoalNum(), true}, //对局完成{X}连杆
		//对局累计灌球{X}次，对局累计解球{X}次，对局累计组合球{X}次，对局累计借球{X}次
	)

	for _, val := range condData {
		cond = append(cond, val)
	}

	c.SyncConditional(data.entityId, cond)
}

func (c *Conditional) SyncConditionalPlayerLv(entityID, playerLV uint32) {
	c.SyncConditional(entityID, []conf.ConditionData{{conf.PlayerLV, playerLV, true}})
}

func (c *Conditional) SyncConditionalVipLv(entityID, vipLV uint32) {
	c.SyncConditional(entityID, []conf.ConditionData{{conf.VipLV, vipLV, true}})
}

func (c *Conditional) SyncConditionalRecharge(entityID, num uint32) {
	c.SyncConditional(entityID, []conf.ConditionData{{conf.RechargeAmount, num, false}})
}
