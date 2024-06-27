package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/table"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"reflect"
	"time"
)

/***
 *@disc:图鉴
 *@author: lsj
 *@date: 2023/11/21
 */

type _CueHandBook struct {
	CueHandBookCfg map[uint32]*table.CueHandbookCfg
}

var CueHandBookMr _CueHandBook

func (c *_CueHandBook) Init() {
	c.initCueHandBookCfg()
	event.OnNet(gmsg.MsgTile_Player_CueHandBookActivateRequest, reflect.ValueOf(c.OnCueHandBookActivateRequest))
}

func (c *_CueHandBook) initCueHandBookCfg() {
	c.CueHandBookCfg = make(map[uint32]*table.CueHandbookCfg, 0)
	for _, val := range Table.GetCueHandbookCfg() {
		c.CueHandBookCfg[val.CueID] = val
	}
}

// 更新图鉴列表
func (c *_CueHandBook) UpdateCueHandBook(entityID, cueId uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.CueHandBookSync{}
	msgResponse.ElemBook = new(gmsg.ElemBook)
	msgResponse.EntityID = entityID
	cueQuality, _, cueIdKey := Backpack.getCueQualityAndStarByTableId(cueId)
	for key, val := range tEntityPlayer.CueHandBook {
		cueHandBookQuality, _, cueHandBookKey := Backpack.getCueQualityAndStarByTableId(val.CueID)
		if cueHandBookKey == cueIdKey && val.State == 0 && cueQuality == cueHandBookQuality {
			cue := tEntityPlayer.CueHandBook[key]
			cue.State = 1
			cue.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			tEntityPlayer.CueHandBook[key] = cue
			stack.SimpleCopyProperties(msgResponse.ElemBook, &cue)
		}
	}
	log.Info("--UpdateCueHandBook-->", msgResponse)
	tEntityPlayer.SyncEntity(1)
	c.syncCueHandBook(entityID, msgResponse.ElemBook)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookSync, msgResponse, []uint32{msgResponse.EntityID})
}

func (c *_CueHandBook) getCueHandBookList(tEntityPlayer *entity.EntityPlayer) (bookList []*gmsg.ElemBook) {
	for _, vl := range tEntityPlayer.CueHandBook {
		elem := new(gmsg.ElemBook)
		stack.SimpleCopyProperties(elem, &vl)
		bookList = append(bookList, elem)
	}
	return
}

// 激活图鉴
func (c *_CueHandBook) OnCueHandBookActivateRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.CueHandBookActivateRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("-->OnCueHandBookActivateRequest-->begin->", msgBody)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse := &gmsg.CueHandBookActivateResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = uint32(1)
	if tEntityPlayer.GetCueHandBook(msgBody.CueID) == nil || tEntityPlayer.GetCueHandBook(msgBody.CueID).State != 1 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookActivateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	res := tEntityPlayer.CueHandBookActivate(msgBody.CueID)
	if res == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookActivateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	cfg, ok := c.CueHandBookCfg[msgBody.CueID]
	if !ok || len(cfg.FixedReward) == 0 {
		log.Error("-->OnCueHandBookActivateRequest-->cfg is err")
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookActivateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	rewardEntityList := make([]entity.RewardEntity, 0)
	for _, vl := range cfg.FixedReward {
		rewardEntity := new(entity.RewardEntity)
		rewardEntity.ItemTableId = vl[0]
		rewardEntity.Num = vl[1]
		rewardEntity.ExpireTimeId = 0
		rewardEntityList = append(rewardEntityList, *rewardEntity)
	}
	resParam := GetResParam(consts.SYSTEM_ID_CUE_HAND_BOOK, consts.ActivateReward)
	//发放奖励
	err, _ := Backpack.BackpackAddItemListAndSave(msgBody.EntityID, rewardEntityList, *resParam)
	if err != nil {
		log.Error(err)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookActivateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	msgResponse.Code = uint32(0)
	msgResponse.ElemBook = new(gmsg.ElemBook)
	stack.SimpleCopyProperties(msgResponse.ElemBook, res)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookActivateResponse, msgResponse, []uint32{msgBody.EntityID})
	c.syncCueHandBook(msgBody.EntityID, msgResponse.ElemBook)
}

func (c *_CueHandBook) syncCueHandBook(entityID uint32, elemBook *gmsg.ElemBook) {
	msgResponse := &gmsg.CueHandBookSync{}
	msgResponse.ElemBook = elemBook
	msgResponse.EntityID = entityID
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Player_CueHandBookSync, msgResponse, []uint32{msgResponse.EntityID})
}
