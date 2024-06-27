package logic

import (
	"BilliardServer/Common/entity"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"
)

/***
 *@disc: 称号管理
 *@author: lsj
 *@date: 2023/10/12
 */

type _CollectMr struct {
}

var CollectMr _CollectMr

func (c *_CollectMr) UpdateCollectFromConditionID(tEntityPlayer *entity.EntityPlayer, conditional, progress uint32, isTotal bool) {
	resCollect := tEntityPlayer.UpdateCollectFromConditionID(conditional, progress, isTotal)
	if len(resCollect) > 0 {
		for _, collect := range resCollect {
			collectInfoSync := &gmsg.CollectInfoSync{}
			collectInfoSync.EntityID = tEntityPlayer.EntityID
			collectInfoSync.CollectInfo = new(gmsg.CollectInfo)
			stack.SimpleCopyProperties(collectInfoSync.CollectInfo, collect)
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CollectListSync, collectInfoSync, []uint32{collectInfoSync.EntityID})
		}
	}
}

// GM命令升级称号
func (c *_CollectMr) AddCollectFromConditionID(entityID, cond, progress uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}
	log.Info("AddCollectFromConditionID-->entityid->", entityID, "--cond--", cond, "--progress--", progress)
	c.UpdateCollectFromConditionID(tEntityPlayer, cond, progress, true)
	tEntityPlayer.SyncEntity(1)
}
