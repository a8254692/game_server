package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Util/log"
)

/***
 *@disc: 成就管理
 *@author: lsj
 *@date: 2023/10/16
 */

type _AchievementMr struct {
}

var AchievementMr _AchievementMr

func (c *_AchievementMr) UpdateAchievementFromConditionID(tEntityPlayer *entity.EntityPlayer, conditional, progress uint32, isTotal bool) {
	childID := tEntityPlayer.UpdateAchievementFromConditionID(conditional, progress, isTotal)
	if len(childID) > 0 {
		// 查询获得的积分
		for _, child := range childID {
			score := Table.GetAchievementElementCfgScore(child)
			log.Info("-->更新成就-->score-->", score, "->child->", child, "-->entityID->", tEntityPlayer.EntityID, "-->AchievementLV->", tEntityPlayer.AchievementLV, "-->AchievementScore->", tEntityPlayer.AchievementScore)
			// 加角色积分
			tEntityPlayer.UpdatePlayerAchievement(score)
			//升级成就
			isUp := Table.IsUpgradeAchievementLV(tEntityPlayer.AchievementScore, tEntityPlayer.AchievementLV)
			//升级成就推送客户端
			if isUp {
				tEntityPlayer.UpgradeAchievementLV()
				Player.PlayerAchievementLVSync(tEntityPlayer.EntityID)
			}
			//推送积分和成就等级同步
			Player.PlayerAchievementLVAndScoreSync(tEntityPlayer.EntityID)
			log.Info("-->更新成就-->UpgradeAchievementLV-->", tEntityPlayer.AchievementLV, "-->entityID->", tEntityPlayer.EntityID, "-->AchievementScore->", tEntityPlayer.AchievementScore)
		}
	}
}

// GM命令升级成就
func (c *_AchievementMr) AddAchievementFromConditionID(entityID, cond, progress uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}
	log.Info("AddAchievementFromConditionID-->entityid->", entityID, "--cond--", cond, "--progress--", progress)
	c.UpdateAchievementFromConditionID(tEntityPlayer, cond, progress, true)
	tEntityPlayer.SyncEntity(1)
}
