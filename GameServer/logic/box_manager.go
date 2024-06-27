package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"math/rand"
	"reflect"
	"time"
)

/***
 *@disc: 宝箱系统
 *@author: lsj
 *@date: 2023/10/31
 */

type _Box struct {
}

var BoxMr _Box

func (c *_Box) Init() {
	event.OnNet(gmsg.MsgTile_Hall_BoxListRequest, reflect.ValueOf(c.OnBoxListRequest))
	event.OnNet(gmsg.MsgTile_Hall_BoxUnlockRequest, reflect.ValueOf(c.OnBoxUnlockRequest))
	event.OnNet(gmsg.MsgTile_Hall_BoxOpenRequest, reflect.ValueOf(c.OnBoxOpenRequest))
	event.OnNet(gmsg.MsgTile_Hall_BoxFastForwardRequest, reflect.ValueOf(c.OnBoxFastForwardRequest))
	event.OnNet(gmsg.MsgTile_Hall_BoxClaimRewardRequest, reflect.ValueOf(c.OnBoxClaimRewardRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClaimMagicBoxRequest, reflect.ValueOf(c.OnClaimMagicBoxRequest))
}

// 结算获得宝箱
func (c *_Box) SettlementAddBox(data UpdateClubData, resParam entity.ResParam) {
	tEntity := Entity.EmPlayer.GetEntityByID(data.EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	if tEntityPlayer.GetEmptyBoxNum() > 0 {
		if data.Result == consts.RESULT_VICTORY {
			boxID := c.getRandBox()
			if boxID == 0 {
				log.Error("box is empty", boxID)
				return
			}
			tEntityPlayer.AddBox(boxID, data.GameType, data.RoomType)
			c.BoxListSync(tEntityPlayer)
			//更新产出
			SendProductionResourceLogToDb(resParam.Uuid, data.EntityID, consts.Box, 0, boxID, consts.RES_TYPE_INCR, uint64(1), uint32(tEntityPlayer.GetBoxNum()), resParam.SysID, resParam.ActionID)
		}
	}
	log.Info("-->SettlementAddBox-->end->", data)
}

// 所有宝箱权重为100，然后进行随机
func (c *_Box) getRandBox() (boxID uint32) {
	rander := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := rander.Intn(100)

	j := uint32(0)

	for _, vl := range Table.BoxCfgList {
		j += vl.Weight
		if j >= uint32(num) {
			boxID = vl.TableID
			break
		}
	}

	return
}

// 宝箱列表
func (c *_Box) OnBoxListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.BoxListRequest{}
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

	msgResponse := &gmsg.BoxListResponse{}
	msgResponse.BoxList = make([]*gmsg.Box, 0)
	for _, vl := range tEntityPlayer.GetBoxList() {
		box := new(gmsg.Box)
		stack.SimpleCopyProperties(box, vl)
		box.ID = vl.ObjID.Hex()
		msgResponse.BoxList = append(msgResponse.BoxList, box)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 解锁宝箱
func (c *_Box) OnBoxUnlockRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.BoxUnlockRequest{}
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

	msgResponse := &gmsg.BoxUnlockResponse{}
	msgResponse.ID = msgBody.ID
	msgResponse.BoxID = msgBody.BoxID
	msgResponse.Code = uint32(1)
	boxCfg := Table.GetBoxCfg(msgBody.BoxID)
	if boxCfg == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxUnlockResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	resLifeTime := tEntityPlayer.BoxUnlock(msgBody.ID, msgBody.BoxID, int64(boxCfg.LifeTime))
	if resLifeTime > 0 {
		msgResponse.Code = uint32(0)
		tEntityPlayer.SyncEntity(1)
		box := new(gmsg.Box)
		b := tEntityPlayer.GetBox(msgBody.ID)
		stack.SimpleCopyProperties(box, b)
		box.ID = msgBody.ID
		msgResponse.UnlockTimeStamp = resLifeTime
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxUnlockResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 打开宝箱
func (c *_Box) OnBoxOpenRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.BoxOpenRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnBoxOpenRequest-->begin-->", msgBody)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	msgResponse := &gmsg.BoxOpenResponse{}
	msgResponse.Code = uint32(1)

	box, boxCfg := tEntityPlayer.GetBox(msgBody.ID), Table.GetBoxCfg(msgBody.BoxID)

	if box == nil || boxCfg == nil || len(boxCfg.Cost) == 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	//todo 后面拓展
	if tEntityPlayer.NumStone < boxCfg.Cost[1] {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	resParam := GetResParam(consts.SYSTEM_ID_BOX, consts.OpenBox)

	//领取宝箱
	msgResponse.BoxItem = new(gmsg.Box)
	stack.SimpleCopyProperties(msgResponse.BoxItem, box)
	msgResponse.BoxItem.ID = msgBody.ID
	resBox := tEntityPlayer.BoxClaim(msgBody.ID, msgBody.BoxID)
	if resBox != nil {
		// 扣减钻石
		Player.UpdatePlayerPropertyItem(msgBody.EntityID, consts.Diamond, int32(-boxCfg.Cost[1]), *resParam)
		msgResponse.Code = uint32(0)
		msgResponse.Uuid = resParam.Uuid
		//发送奖励
		msgResponse.RewardList = c.sendRewardItem(msgBody.EntityID, boxCfg.FixedReward, boxCfg.RandomReward, *resParam)
	}

	tEntityPlayer.SyncEntity(1)
	log.Info("-->OnBoxOpenRequest-->end-->", msgResponse)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxOpenResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Box) sendRewardItem(EntityID uint32, fixedReward [][]uint32, randReward []uint32, resParam entity.ResParam) []*gmsg.RewardInfo {
	rewardList := make([]*gmsg.RewardInfo, 0)

	//随机奖励
	if len(randReward) > 0 {
		rand := new(gmsg.RewardInfo)
		randomRewardNum := Backpack.BackpackAddOneItemSaveFromRandReward(EntityID, [][]uint32{randReward}, resParam)
		rand.ItemTableId = randReward[0]
		rand.ExpireTimeId = 0
		rand.Num = randomRewardNum
		rewardList = append(rewardList, rand)
	}
	//固定奖励
	if len(fixedReward) > 0 {
		rewardEntityList := make([]entity.RewardEntity, 0)
		for _, vl := range fixedReward {
			rewardEntity := new(entity.RewardEntity)
			rewardEntity.ItemTableId = vl[0]
			rewardEntity.Num = vl[1]
			rewardEntity.ExpireTimeId = 0
			fixe := new(gmsg.RewardInfo)
			stack.SimpleCopyProperties(fixe, rewardEntity)
			rewardEntityList = append(rewardEntityList, *rewardEntity)
			rewardList = append(rewardList, fixe)
		}
		//发放奖励
		err, _ := Backpack.BackpackAddItemListAndUpdateItemSync(EntityID, rewardEntityList, resParam)
		if err != nil {
			log.Error(err)
		}
	}

	return rewardList
}

// 打开神秘奖励
func (c *_Box) OnClaimMagicBoxRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClaimMagicBoxRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.ClaimMagicBoxResponse{}
	msgResponse.Code = uint32(1)

	boxCfg := Table.GetBoxCfg(msgBody.BoxID)

	if len(boxCfg.SecretReward) == 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClaimMagicBoxResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	resParam := &entity.ResParam{Uuid: msgBody.Uuid, SysID: consts.SYSTEM_ID_BOX, ActionID: consts.MagicBoxReward}

	//神秘奖励
	if len(boxCfg.SecretReward) > 0 {
		msgResponse.Code = uint32(0)
		rewardEntity := new(entity.RewardEntity)
		rewardEntity.ItemTableId = boxCfg.SecretReward[0]
		rewardEntity.Num = boxCfg.SecretReward[1]
		rewardEntity.ExpireTimeId = 0
		//发放奖励
		err, _, _ := Backpack.BackpackAddOneItemAndSave(msgBody.EntityID, *rewardEntity, *resParam)
		if err != nil {
			log.Error(err)
		}
	}

	tEntityPlayer.SyncEntity(1)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClaimMagicBoxResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 加速宝箱
func (c *_Box) OnBoxFastForwardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.BoxFastForwardRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnBoxFastForwardRequest-->begin-->", msgBody)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	msgResponse := &gmsg.BoxFastForwardResponse{}
	msgResponse.Code = uint32(1)

	box, boxCfg := tEntityPlayer.GetBox(msgBody.ID), Table.GetBoxCfg(msgBody.BoxID)

	if box == nil || boxCfg == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxFastForwardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	interval := boxCfg.Interval * 3600
	//宝箱加速
	resbox := tEntityPlayer.BoxFastReward(msgBody.ID, msgBody.BoxID, interval)
	if resbox == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxFastForwardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	msgResponse.ID = msgBody.ID
	msgResponse.ReduceTime = resbox.ReduceTime
	msgResponse.UnlockTimeStamp = resbox.UnlockTimeStamp
	msgResponse.BoxID = msgBody.BoxID
	msgResponse.Code = uint32(0)
	log.Info("-->OnBoxFastForwardRequest-->end-->", msgResponse)
	tEntityPlayer.SyncEntity(1)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxFastForwardResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 领取宝箱
func (c *_Box) OnBoxClaimRewardRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.BoxClaimRewardRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnBoxClaimRewardRequest-->begin-->", msgBody)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	msgResponse := &gmsg.BoxClaimRewardResponse{}
	msgResponse.Code = uint32(1)

	box, boxCfg := tEntityPlayer.GetBox(msgBody.ID), Table.GetBoxCfg(msgBody.BoxID)

	if box == nil || boxCfg == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if (box.ReduceTime == 0 && box.UnlockTimeStamp > time.Now().Unix()) || (box.ReduceTime > 0 && int64(box.ReduceTime)+time.Now().Unix() < box.UnlockTimeStamp) {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	//领取宝箱
	msgResponse.BoxItem = new(gmsg.Box)
	stack.SimpleCopyProperties(msgResponse.BoxItem, box)
	msgResponse.BoxItem.ID = msgBody.ID
	resBox := tEntityPlayer.BoxClaim(msgBody.ID, msgBody.BoxID)
	if resBox != nil {
		resParam := GetResParam(consts.SYSTEM_ID_BOX, consts.OpenBox)
		//更新消耗
		SendConsumeResourceLogToDb(resParam.Uuid, tEntityPlayer.EntityID, consts.Box, 0, msgBody.BoxID, consts.RES_TYPE_DECR, uint64(1), uint32(tEntityPlayer.GetBoxNum()), resParam.SysID, resParam.ActionID)
		msgResponse.Code = uint32(0)
		msgResponse.Uuid = resParam.Uuid
		//发送奖励
		msgResponse.RewardList = c.sendRewardItem(msgBody.EntityID, boxCfg.FixedReward, boxCfg.RandomReward, *resParam)
	}

	log.Info("-->OnBoxClaimRewardRequest-->end-->", msgResponse)
	tEntityPlayer.SyncEntity(1)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxClaimRewardResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Box) getBoxList(tEntityPlayer *entity.EntityPlayer) (boxList []*gmsg.Box) {
	for _, vl := range tEntityPlayer.GetBoxList() {
		box := new(gmsg.Box)
		stack.SimpleCopyProperties(box, vl)
		box.ID = vl.ObjID.Hex()
		boxList = append(boxList, box)
	}
	return
}

func (c *_Box) BoxListSync(tEntityPlayer *entity.EntityPlayer) {
	msgResponse := &gmsg.BoxListResponse{}
	msgResponse.BoxList = make([]*gmsg.Box, 0)
	for _, vl := range tEntityPlayer.GetBoxList() {
		box := new(gmsg.Box)
		stack.SimpleCopyProperties(box, vl)
		box.ID = vl.ObjID.Hex()
		msgResponse.BoxList = append(msgResponse.BoxList, box)
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BoxListResponse, msgResponse, []uint32{tEntityPlayer.EntityID})
}
