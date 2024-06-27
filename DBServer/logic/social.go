package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"reflect"
	"sync"
)

/***
 *@disc: 社交
 *@author: lsj
 *@date: 2023/9/22
 */

type _SocialDB struct {
}

var SocialDBManager _SocialDB

var SocialMutex sync.Mutex

func (c *_SocialDB) Init() {
	//内部协议
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_AddFansFromGameToDB), reflect.ValueOf(c.AddFansDBRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_DelFansFromGameToDB), reflect.ValueOf(c.DelFansDBRequest))

	event.OnNet(gmsg.MsgTile_Hall_AddMyFriendsRequest, reflect.ValueOf(c.AddMyFriendsDBRequest))
	event.OnNet(gmsg.MsgTile_Hall_SearchPlayerFromIDRequest, reflect.ValueOf(c.OnSearchPlayerFromIDDBRequest))
	event.OnNet(gmsg.MsgTile_Hall_NearbyPlayerListRequest, reflect.ValueOf(c.OnNearByPlayerListDBRequest))
}

// 添加关注
func (c *_SocialDB) AddMyFriendsDBRequest(msgEV *network.MsgBodyEvent) {
	SocialMutex.Lock()
	defer SocialMutex.Unlock()
	msgBody := &gmsg.AddMyFriendsRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.AddMyFriendsResponse{}
	msgResponse.Code = resp_code.CODE_ERR
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.AddEntityID = msgBody.AddEntityID
	if !tEntityPlayer.IsInMyFriends(msgBody.AddEntityID) {
		addtEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.AddEntityID)
		addtEntityPlayer := addtEntity.(*entity.EntityPlayer)
		friendList := new(gmsg.FriendList)
		stack.SimpleCopyProperties(friendList, addtEntityPlayer)
		msgResponse.Code = resp_code.CODE_SUCCESS
		msgResponse.AddList = make([]*gmsg.FriendList, 0)
		msgResponse.AddList = append(msgResponse.AddList, friendList)
		Entity.SendPlayerBaseSync(addtEntityPlayer)
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_AddMyFriendsResponse, msgResponse, network.ServerType_Game)
}

// 添加粉丝
func (c *_SocialDB) AddFansDBRequest(msgEV *network.MsgBodyEvent) {
	SocialMutex.Lock()
	defer SocialMutex.Unlock()
	msgBody := &gmsg.AddMyFansRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if !tEntityPlayer.IsInFansList(msgBody.AddEntityID) {
		tEntityPlayer.AddFansList(msgBody.AddEntityID)
		tEntityPlayer.FlagChang()
	}
	conds := make([]consts.ConditionData, 0)
	conds = append(conds, consts.ConditionData{consts.TotalMyFriends, tEntityPlayer.FansNum, true}, consts.ConditionData{consts.FansNum, tEntityPlayer.FansNum, true})
	TaskDBManger.updateConditional(tEntityPlayer.EntityID, conds)
	log.Info("-->AddFansDBRequest-->end->", msgBody)
}

// 减少粉丝
func (c *_SocialDB) DelFansDBRequest(msgEV *network.MsgBodyEvent) {
	SocialMutex.Lock()
	defer SocialMutex.Unlock()
	msgBody := &gmsg.DelMyFansRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if tEntityPlayer.IsInFansList(msgBody.DelEntityID) {
		tEntityPlayer.DelFans(msgBody.DelEntityID)
		tEntityPlayer.FlagChang()
	}
	log.Info("-->DelFansDBRequest-->end->", msgBody)
}

// 搜索用户
func (c *_SocialDB) OnSearchPlayerFromIDDBRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SearchPlayerFromIDRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.SearchPlayerFromIDResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = uint32(1)
	qid := uint32(tools.StringToInt(msgBody.QueryEntityID))
	qEntity := Entity.EmEntityPlayer.GetEntityByID(qid)
	if qEntity == nil {
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_SearchPlayerFromIDResponse, msgResponse, network.ServerType_Game)
		return
	}
	resPlayers := qEntity.(*entity.EntityPlayer)

	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if !tEntityPlayer.IsInMyFriends(qid) {
		msgResponse.ResultEntityID = qid
		msgResponse.PlayerName = resPlayers.PlayerName
		msgResponse.IconFrame = resPlayers.IconFrame
		msgResponse.Sex = resPlayers.Sex
		msgResponse.VipLv = resPlayers.VipLv
		msgResponse.PlayerIcon = resPlayers.PlayerIcon
		msgResponse.PlayerLv = resPlayers.PlayerLv
		msgResponse.PeakRankLv = resPlayers.PeakRankLv
		msgResponse.Code = uint32(0)
	} else {
		msgResponse.Code = uint32(2)
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_SearchPlayerFromIDResponse, msgResponse, network.ServerType_Game)
}

func (c *_SocialDB) OnNearByPlayerListDBRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.NearByPlayerListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.NearByPlayerListResponse{}
	msgResponse.EntityID = msgBody.EntityID
	onlineMap := ConnectManager.GetTcpMapConnect()
	num, max := 0, 20
	for key, _ := range onlineMap {
		if uint32(key) == 0 {
			continue
		}
		a := new(gmsg.NearPlayerList)
		a.EntityID = uint32(key)
		a.PlayerType = 0
		msgResponse.List = append(msgResponse.List)
		num++
		if num >= max {
			break
		}
	}

	if num < max {
		c.addNearByPlayerList(msgResponse, max, msgBody.EntityID)
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_NearbyPlayerListResponse, msgResponse, network.ServerType_Game)
}

func (c *_SocialDB) addNearByPlayerList(res *gmsg.NearByPlayerListResponse, max int, entityID uint32) {
	if len(res.List) == 0 {
		for _, emPlayer := range Entity.EmEntityPlayer.EntityMap {
			player := emPlayer.(*entity.EntityPlayer)
			if player.EntityID == entityID {
				continue
			}
			a := new(gmsg.NearPlayerList)
			a.EntityID = player.EntityID
			a.PlayerType = 2
			res.List = append(res.List, a)
			if len(res.List) >= max {
				break
			}
		}
		return
	}
	for _, vl := range res.List {
		for _, emPlayer := range Entity.EmEntityPlayer.EntityMap {
			player := emPlayer.(*entity.EntityPlayer)
			if player.EntityID == vl.EntityID || player.EntityID == entityID {
				continue
			}
			a := new(gmsg.NearPlayerList)
			a.EntityID = player.EntityID
			a.PlayerType = 2
			res.List = append(res.List, a)
			if len(res.List) >= max {
				break
			}
		}
	}
}
