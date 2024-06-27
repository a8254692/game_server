package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/DBServer/initialize/consts"
	"BilliardServer/Util/db/collection"
	"BilliardServer/Util/log"
	"reflect"

	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
)

type _Chat struct {
}

var Chat _Chat

func (s *_Chat) Init() {
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SavePrivateFriendsListRequest), reflect.ValueOf(s.OnSavePrivateFriendsListRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_GetOfflinePrivateFriendsRequest), reflect.ValueOf(s.OnGetOfflinePrivateFriendsRequest))
}

func (s *_Chat) SyncChatFriendsListToGame(entityId uint32) {
	if entityId <= 0 {
		return
	}

	chat := new(collection.Chat)
	chat.SetDBConnect(consts.COLLECTION_CHART)
	chat.InitFormDB(entityId, DBConnect)

	if len(chat.FriendsChatList) <= 0 {
		return
	}

	friendsList := make([]*gmsg.InFriendsInfo, 0)
	for _, v := range chat.FriendsChatList {
		friendsList = append(friendsList, &gmsg.InFriendsInfo{
			EntityID:   v.EntityID,
			PlayerName: v.PlayerName,
			Sex:        v.Sex,
			PlayerIcon: v.PlayerIcon,
			IconFrame:  v.IconFrame,
			VipLv:      v.VipLv,
		})
	}

	resp := &gmsg.InFriendsList{
		EntityID:    entityId,
		FriendsList: friendsList,
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPrivateFriendsListRequest), resp, network.ServerType_Game)
	return
}

// 同步用户私聊列表(游戏服->DB服)
func (s *_Chat) OnSavePrivateFriendsListRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InSavePrivateFriendsListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	log.Info("-->logic--_Chat--OnSavePrivateFriendsListRequest--req--", req)

	if len(req.List) <= 0 {
		log.Waring("-->logic--_Chat--OnSavePrivateFriendsListRequest--len(req.List) <= 0")
		return
	}

	for _, v := range req.List {
		if v.EntityID <= 0 || len(v.FriendsList) <= 0 {
			continue
		}

		friendsChatList := make([]collection.PrivateChatEntity, 0)
		for _, fv := range v.FriendsList {
			friendsChatList = append(friendsChatList, collection.PrivateChatEntity{
				EntityID:   fv.EntityID,
				PlayerName: fv.PlayerName,
				Sex:        fv.Sex,
				PlayerIcon: fv.PlayerIcon,
				IconFrame:  fv.IconFrame,
				VipLv:      fv.VipLv,
			})
		}

		info := new(collection.Chat)
		info.InitByFirst(consts.COLLECTION_CHART, v.EntityID)
		dbInfo := info.GetDataByEntityId(v.EntityID, DBConnect)
		if dbInfo != nil && dbInfo.EntityId > 0 && dbInfo.ObjID != "" {
			info.ObjID = dbInfo.ObjID
		}

		info.FriendsChatList = friendsChatList

		_ = info.Save(DBConnect)
	}

	return
}

func (s *_Chat) OnGetOfflinePrivateFriendsRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InGetOfflinePrivateFriendsRequest{}
	if err := msgEV.Unmarshal(req); err != nil {
		log.Waring("-->logic--_Chat--OnGetOfflinePrivateFriendsRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_Chat--OnGetOfflinePrivateFriendsRequest--req--", req)

	if req.EntityID <= 0 || req.PrivateEntityID <= 0 {
		log.Waring("-->logic--_Chat--OnGetOfflinePrivateFriendsRequest--req.EntityID <= 0 || req.PrivateEntityID <= 0")
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(req.PrivateEntityID)
	if tEntity == nil {
		log.Waring("-->logic--_Chat--OnGetOfflinePrivateFriendsRequest--GetEntityByID == nil")
		return
	}
	tEntityPlayer, ok := tEntity.(*entity.EntityPlayer)
	if !ok {
		log.Waring("-->logic--_Chat--OnGetOfflinePrivateFriendsRequest-- tEntity.(*entity.EntityPlayer)--!ok")
		return
	}

	resp := &gmsg.InGetOfflinePrivateFriendsResponse{
		EntityID: req.EntityID,
		FriendsInfo: &gmsg.InFriendsInfo{
			EntityID:   tEntityPlayer.EntityID,
			PlayerName: tEntityPlayer.PlayerName,
			Sex:        tEntityPlayer.Sex,
			PlayerIcon: tEntityPlayer.PlayerIcon,
			IconFrame:  tEntityPlayer.IconFrame,
			VipLv:      tEntityPlayer.VipLv,
		},
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_GetOfflinePrivateFriendsResponse), resp, network.ServerType_Game)

	return
}
