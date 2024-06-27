package logic

import (
	"BilliardServer/Common/entity"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"math/rand"
	"time"
)

/***
 *@disc: 内部测试
 *@author: lsj
 *@date: 2023/12/20
 */

func (c *_Club) OnBatchReqClubRequest(level uint32) {
	resBody := &gmsg.BatchCreateClubRequest{}
	resBody.RegClubLevel = level
	resBody.RegNum = 10
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_BatchCreateClubRequest), resBody, network.ServerType_DB)
}

func (c *_Club) OnBatchReqClubResponse(msgEV *network.MsgBodyEvent) {
	resBody := &gmsg.BatchCreateClubResponse{}
	if err := msgEV.Unmarshal(resBody); err != nil {
		return
	}

	for _, val := range resBody.ClubID {
		emClub := Entity.EmClub.GetEntityByID(val)
		club := emClub.(*entity.Club)
		club.ShopList = c.ClubShopList
		club.SyncEntity(1)
	}

	log.Info("-->OnBatchReqClubResponse->end.")
}

func (c *_Club) OnAddClubData(cid, param uint32) {
	emClub := Entity.EmClub.GetEntityByID(cid)
	club := emClub.(*entity.Club)
	switch param {
	case 1:
		club.ClubScore += 1000
		club.TotalScore += 1000
	case 2:
		club.ClubScore += 10000
		club.TotalScore += 10000
	case 3:
		club.ClubScore += 100000
		club.TotalScore += 100000
	}
	club.SyncEntity(1)
	log.Info(">>>>>>>>>OnAddClubData>>>id-->", cid)
}

func (c *_Club) OnAddAllClubData() {
	end, start := 210000, 1000
	for id, _ := range Entity.EmClub.EntityMap {
		emClub := Entity.EmClub.GetEntityByID(id)
		club := emClub.(*entity.Club)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		//生成随机数
		num := uint32(r.Intn((end - start)) + start)
		club.ClubScore += num
		club.TotalScore += num
		club.SyncEntity(1)
	}
}
