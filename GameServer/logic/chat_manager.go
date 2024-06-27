package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/timer"
	"container/list"
	"reflect"
	"time"
)

type msgInfo struct {
	entityID   uint32
	mType      uint32
	content    string
	creatTime  int64
	playerName string
	sex        uint32
	playerIcon uint32
	iconFrame  uint32
	vipLv      uint32
	collectId  uint32
	chatBox    uint32
}

type privateChatEntity struct {
	entityID   uint32
	playerName string
	sex        uint32
	playerIcon uint32
	iconFrame  uint32
	vipLv      uint32
}

var ChatMgr _ChatMgr

type _ChatMgr struct {
	worldList             *list.List
	clubList              map[uint32]*list.List
	friendsList           map[uint32]*list.List //私聊好友列表
	friendsMsgList        map[uint32]*list.List
	friendsRedDotList     map[uint32]map[uint32]uint32 //私聊红点记录
	friendsListUpdateTime map[uint32]int64             //私聊红点记录
}

// TODO:控制私聊记录长度
func (s *_ChatMgr) Init() {
	s.worldList = list.New()
	s.clubList = make(map[uint32]*list.List)
	s.friendsList = make(map[uint32]*list.List)
	s.friendsMsgList = make(map[uint32]*list.List)
	s.friendsRedDotList = make(map[uint32]map[uint32]uint32)
	s.friendsListUpdateTime = make(map[uint32]int64)

	event.OnNet(gmsg.MsgTile_Hall_SendWorldMsgRequest, reflect.ValueOf(s.OnSendWorldMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_SendClubMsgRequest, reflect.ValueOf(s.OnSendClubMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_SendPrivateChatMsgRequest, reflect.ValueOf(s.OnSendPrivateChatMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_GetWorldMsgRequest, reflect.ValueOf(s.OnGetWorldMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_GetClubMsgRequest, reflect.ValueOf(s.OnGetClubMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_GetPrivateChatEntityRequest, reflect.ValueOf(s.OnGetPrivateChatEntityRequest))
	event.OnNet(gmsg.MsgTile_Hall_GetPrivateChatMsgRequest, reflect.ValueOf(s.OnGetPrivateChatMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_SeePrivateChatMsgRequest, reflect.ValueOf(s.OnSeePrivateChatMsgRequest))
	event.OnNet(gmsg.MsgTile_Hall_DelPrivateChatEntityRequest, reflect.ValueOf(s.OnDelPrivateChatEntityRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncPrivateFriendsListRequest), reflect.ValueOf(s.OnSyncPrivateFriendsListRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_GetOfflinePrivateFriendsResponse), reflect.ValueOf(s.OnGetOfflinePrivateFriendsResponse))

	timer.AddTimer(s, "SyncFriendsListToDb", 10000, true)
}

func (s *_ChatMgr) SyncFriendsListToDb() {
	if len(s.friendsListUpdateTime) <= 0 {
		return
	}

	reqList := make([]*gmsg.InFriendsList, 0)

	for k, v := range s.friendsListUpdateTime {
		if k <= 0 && v > 0 {
			continue
		}

		friendsList := make([]*gmsg.InFriendsInfo, 0)
		if s.friendsList[k].Len() > 0 {
			for e := s.friendsList[k].Front(); e != nil; e = e.Next() {
				ev, ok := e.Value.(privateChatEntity)
				if !ok {
					continue
				}

				friendsList = append(friendsList, &gmsg.InFriendsInfo{
					EntityID:   ev.entityID,
					PlayerName: ev.playerName,
					Sex:        ev.sex,
					PlayerIcon: ev.playerIcon,
					IconFrame:  ev.iconFrame,
					VipLv:      ev.vipLv,
				})
			}
		}

		reqList = append(reqList, &gmsg.InFriendsList{
			EntityID:    k,
			FriendsList: friendsList,
		})
	}

	if len(reqList) <= 0 {
		return
	}

	//重置更新列表
	s.friendsListUpdateTime = make(map[uint32]int64)

	req := &gmsg.InSavePrivateFriendsListRequest{
		List: reqList,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SavePrivateFriendsListRequest), req, network.ServerType_DB)
	return
}

func (s *_ChatMgr) GetPrivateChatRedDotNum(entityID uint32, friendEntityID uint32) uint32 {
	var resp uint32
	if entityID <= 0 || friendEntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--GetPrivateChatRedDotList--entityID <= 0 || friendEntityID <= 0")
		return resp
	}

	if s.friendsRedDotList[entityID] == nil && len(s.friendsRedDotList[entityID]) <= 0 {
		return resp
	}

	for k, v := range s.friendsRedDotList[entityID] {
		if k == friendEntityID {
			resp = v
			break
		}
	}

	return resp
}

func (s *_ChatMgr) GetPrivateChatAllRedDotNum(entityID uint32) uint32 {
	var resp uint32
	if entityID <= 0 {
		log.Waring("-->logic--_ChatMgr--GetPrivateChatRedDotList--entityID <= 0")
		return resp
	}

	if s.friendsRedDotList[entityID] != nil && len(s.friendsRedDotList[entityID]) <= 0 {
		return resp
	}

	for _, v := range s.friendsRedDotList[entityID] {
		resp += v
	}

	return resp
}

func (s *_ChatMgr) InnerSendWorldMsg(mType uint32, entityID uint32, context string) {
	if mType <= 0 || context == "" {
		log.Waring("-->logic--_ChatMgr--InnerSendWorldMsg--req.EntityID <= 0 || req.Context == nil")
		return
	}

	now := time.Now().Unix()
	sendMsgInfo := new(gmsg.MsgInfo)
	var mi msgInfo
	if mType == uint32(gmsg.MsgType_MtSystem) {
		sendMsgInfo = &gmsg.MsgInfo{
			MType:     mType,
			EntityID:  entityID,
			Content:   context,
			CreatTime: now,
		}

		mi = msgInfo{
			entityID:  entityID,
			mType:     mType,
			content:   context,
			creatTime: now,
		}

	} else {
		if entityID <= 0 {
			log.Waring("-->logic--_ChatMgr--InnerSendWorldMsg--tEntity <= 0")
			return
		}

		tEntityPlayer, err := GetEntityPlayerById(entityID)
		if err != nil {
			log.Waring("-->logic--_ChatMgr--InnerSendWorldMsg--GetEntityPlayerById--err--", err)
			return
		}

		//禁言用户
		if tEntityPlayer.State == consts.USER_STATUS_PROHIBITION {
			log.Waring("-->logic--_ChatMgr--InnerSendWorldMsg--State == consts.USER_STATUS_PROHIBITION", entityID)
			return
		}

		sendMsgInfo = &gmsg.MsgInfo{
			MType:           mType,
			EntityID:        entityID,
			EntityName:      tEntityPlayer.PlayerName,
			EntitySex:       tEntityPlayer.Sex,
			EntityIcon:      tEntityPlayer.PlayerIcon,
			EntityIconFrame: tEntityPlayer.IconFrame,
			EntityVipLv:     tEntityPlayer.VipLv,
			ChatBox:         tEntityPlayer.ClothingBubble,
			Content:         context,
			Designation:     tEntityPlayer.CollectId,
			CreatTime:       now,
		}

		//消息加入队列
		mi = msgInfo{
			entityID:   entityID,
			mType:      mType,
			content:    context,
			creatTime:  now,
			playerName: tEntityPlayer.PlayerName,
			sex:        tEntityPlayer.Sex,
			playerIcon: tEntityPlayer.PlayerIcon,
			iconFrame:  tEntityPlayer.IconFrame,
			vipLv:      tEntityPlayer.VipLv,
			collectId:  tEntityPlayer.CollectId,
			chatBox:    tEntityPlayer.ClothingBubble,
		}
	}

	s.worldList.PushFront(mi)
	if s.worldList.Len() > consts.MSGMAXNUM {
		s.worldList.Remove(s.worldList.Back())
	}

	//开始初始化桌面信息
	resp := &gmsg.SendWorldMsgSync{
		Code:     0,
		EntityID: entityID,
		Msg:      sendMsgInfo,
	}

	log.Info("-->logic--_BattleC8Mgr--InnerSendWorldMsg--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCastAll(gmsg.MsgTile_Hall_SendWorldMsgSync, resp)
	return
}

func (s *_ChatMgr) InnerSendClubMsg(mType uint32, entityID uint32, clubID uint32, context string) {
	if mType <= 0 || clubID <= 0 || context == "" {
		log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--mType <= 0 || clubID <= 0 || context == \"\"")
		return
	}

	emClub := Entity.EmClub.GetEntityByID(clubID)
	club := emClub.(*entity.Club)
	if club == nil || club.ClubID <= 0 {
		log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--club == nil", clubID)
		return
	}

	members := club.GetMembers()
	if len(members) <= 0 {
		log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--len(members) <= 0", clubID)
		return
	}

	now := time.Now().Unix()
	sendMsgInfo := new(gmsg.MsgInfo)
	var mi msgInfo

	if mType == uint32(gmsg.MsgType_MtSystem) {
		sendMsgInfo = &gmsg.MsgInfo{
			MType:     mType,
			EntityID:  entityID,
			Content:   context,
			CreatTime: now,
		}

		mi = msgInfo{
			entityID:  entityID,
			mType:     mType,
			content:   context,
			creatTime: now,
		}

	} else {
		if entityID <= 0 {
			log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--tEntity <= 0")
			return
		}

		tEntityPlayer, err := GetEntityPlayerById(entityID)
		if err != nil {
			log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--GetEntityPlayerById--err--", err)
			return
		}

		//禁言用户
		if tEntityPlayer.State == consts.USER_STATUS_PROHIBITION {
			log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--State == consts.USER_STATUS_PROHIBITION", entityID)
			return
		}

		if tEntityPlayer.ClubId != clubID {
			log.Waring("-->logic--_ChatMgr--InnerSendClubMsg--tEntityPlayer.ClubId != req.ClubID", clubID)
			return
		}

		sendMsgInfo = &gmsg.MsgInfo{
			MType:           mType,
			EntityID:        entityID,
			EntityName:      tEntityPlayer.PlayerName,
			EntitySex:       tEntityPlayer.Sex,
			EntityIcon:      tEntityPlayer.PlayerIcon,
			EntityIconFrame: tEntityPlayer.IconFrame,
			EntityVipLv:     tEntityPlayer.VipLv,
			ChatBox:         tEntityPlayer.ClothingBubble,
			Content:         context,
			Designation:     tEntityPlayer.CollectId,
			CreatTime:       now,
		}

		//消息加入队列
		mi = msgInfo{
			entityID:   entityID,
			mType:      mType,
			content:    context,
			creatTime:  now,
			playerName: tEntityPlayer.PlayerName,
			sex:        tEntityPlayer.Sex,
			playerIcon: tEntityPlayer.PlayerIcon,
			iconFrame:  tEntityPlayer.IconFrame,
			vipLv:      tEntityPlayer.VipLv,
			collectId:  tEntityPlayer.CollectId,
			chatBox:    tEntityPlayer.ClothingBubble,
		}
	}

	if s.clubList[clubID] == nil {
		msgList := list.New()
		msgList.PushFront(mi)
		s.clubList[clubID] = msgList
	} else {
		s.clubList[clubID].PushFront(mi)
		if s.clubList[clubID].Len() > consts.MSGMAXNUM {
			s.clubList[clubID].Remove(s.clubList[clubID].Back())
		}
	}

	targetEntityIDs := make([]uint32, 0)
	for _, v := range members {
		targetEntityIDs = append(targetEntityIDs, v.EntityID)
	}

	//开始初始化桌面信息
	resp := &gmsg.SendClubMsgSync{
		Code:     0,
		EntityID: entityID,
		ClubID:   clubID,
		Msg:      sendMsgInfo,
	}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_SendClubMsgSync, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) InnerSendPrivateChatMsg(mType uint32, entityID uint32, privateEntityID uint32, context string) {
	if entityID <= 0 || privateEntityID <= 0 || mType <= 0 || context == "" {
		log.Waring("-->logic--_ChatMgr--InnerSendPrivateChatMsg--entityID <= 0 || privateEntityID <= 0 || mType <= 0 || context == ")
		return
	}

	now := time.Now().Unix()

	// 先更新关系
	s.updateUserFriendList(entityID, privateEntityID)

	sendMsgInfo := new(gmsg.MsgInfo)
	var mi msgInfo
	if mType == uint32(gmsg.MsgType_MtSystem) {
		sendMsgInfo = &gmsg.MsgInfo{
			MType:     mType,
			EntityID:  entityID,
			Content:   context,
			CreatTime: now,
		}

		mi = msgInfo{
			entityID:  entityID,
			mType:     mType,
			content:   context,
			creatTime: now,
		}
	} else {
		tEntityPlayer, err := GetEntityPlayerById(entityID)
		if err != nil {
			log.Waring("-->logic--_ChatMgr--InnerSendPrivateChatMsg--GetEntityPlayerById--err--", err)
			return
		}

		sendMsgInfo = &gmsg.MsgInfo{
			MType:           mType,
			EntityID:        entityID,
			EntityName:      tEntityPlayer.PlayerName,
			EntitySex:       tEntityPlayer.Sex,
			EntityIcon:      tEntityPlayer.PlayerIcon,
			EntityIconFrame: tEntityPlayer.IconFrame,
			EntityVipLv:     tEntityPlayer.VipLv,
			ChatBox:         tEntityPlayer.ClothingBubble,
			Content:         context,
			Designation:     tEntityPlayer.CollectId,
			CreatTime:       now,
		}

		mi = msgInfo{
			entityID:   entityID,
			mType:      mType,
			content:    context,
			creatTime:  now,
			playerName: tEntityPlayer.PlayerName,
			sex:        tEntityPlayer.Sex,
			playerIcon: tEntityPlayer.PlayerIcon,
			iconFrame:  tEntityPlayer.IconFrame,
			vipLv:      tEntityPlayer.VipLv,
			collectId:  tEntityPlayer.CollectId,
			chatBox:    tEntityPlayer.ClothingBubble,
		}
	}

	msgKey := entityID + privateEntityID
	if s.friendsMsgList[msgKey] == nil {
		msgList := list.New()
		msgList.PushFront(mi)
		s.friendsMsgList[msgKey] = msgList
	} else {
		s.friendsMsgList[msgKey].PushFront(mi)

		if s.friendsMsgList[msgKey].Len() > consts.MSGMAXNUM {
			s.friendsMsgList[msgKey].Remove(s.friendsMsgList[msgKey].Back())
		}
	}

	//开始初始化桌面信息
	resp := &gmsg.SendPrivateChatMsgSync{
		Code:            0,
		EntityID:        entityID,
		PrivateEntityID: privateEntityID,
		Msg:             sendMsgInfo,
	}

	targetEntityIDs := []uint32{entityID, privateEntityID}

	log.Info("-->logic--_BattleC8Mgr--InnerSendPrivateChatMsg--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_SendPrivateChatMsgSync, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) checkFriendList(entityID uint32, otherEntityID uint32) bool {
	var isIn bool

	if s.friendsList[entityID] != nil {
		if s.friendsList[entityID].Len() > 0 {
			for e := s.friendsList[entityID].Front(); e != nil; e = e.Next() {
				ev, ok := e.Value.(privateChatEntity)
				if !ok {
					continue
				}

				if ev.entityID == otherEntityID {
					isIn = true
					break
				}
			}
		}
	}

	return isIn
}

func (s *_ChatMgr) updateUserFriendList(entityID uint32, privateEntityID uint32) {
	if entityID <= 0 || privateEntityID <= 0 {
		return
	}

	privateEntityPlayer, err := GetEntityPlayerById(privateEntityID)
	if err == nil && privateEntityPlayer != nil {
		privateEntityChat := privateChatEntity{
			entityID:   privateEntityPlayer.EntityID,
			playerName: privateEntityPlayer.PlayerName,
			sex:        privateEntityPlayer.Sex,
			playerIcon: privateEntityPlayer.PlayerIcon,
			iconFrame:  privateEntityPlayer.IconFrame,
			vipLv:      privateEntityPlayer.VipLv,
		}
		if s.friendsList[entityID] == nil {
			fList := list.New()
			fList.PushFront(privateEntityChat)
			s.friendsList[entityID] = fList

			s.friendsListUpdateTime[entityID] = time.Now().Unix()
		} else {
			if !s.checkFriendList(entityID, privateEntityID) {
				s.friendsList[entityID].PushFront(privateEntityChat)

				s.friendsListUpdateTime[entityID] = time.Now().Unix()
			}
		}
	} else {
		req := &gmsg.InGetOfflinePrivateFriendsRequest{
			EntityID:        entityID,
			PrivateEntityID: privateEntityID,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_GetOfflinePrivateFriendsRequest), req, network.ServerType_DB)
	}

	tEntityPlayer, err := GetEntityPlayerById(entityID)
	if err == nil && tEntityPlayer != nil {
		entityChat := privateChatEntity{
			entityID:   tEntityPlayer.EntityID,
			playerName: tEntityPlayer.PlayerName,
			sex:        tEntityPlayer.Sex,
			playerIcon: tEntityPlayer.PlayerIcon,
			iconFrame:  tEntityPlayer.IconFrame,
			vipLv:      tEntityPlayer.VipLv,
		}
		if s.friendsList[privateEntityID] == nil {
			fList := list.New()
			fList.PushFront(entityChat)
			s.friendsList[privateEntityID] = fList

			s.friendsListUpdateTime[privateEntityID] = time.Now().Unix()
		} else {
			if !s.checkFriendList(privateEntityID, entityID) {
				s.friendsList[privateEntityID].PushFront(entityChat)

				s.friendsListUpdateTime[privateEntityID] = time.Now().Unix()
			}
		}
	} else {
		req := &gmsg.InGetOfflinePrivateFriendsRequest{
			EntityID:        privateEntityID,
			PrivateEntityID: entityID,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_GetOfflinePrivateFriendsRequest), req, network.ServerType_DB)
	}

	//更新红点数量
	if s.friendsRedDotList[privateEntityID] == nil {
		newMap := make(map[uint32]uint32)
		newMap[entityID] += 1
		s.friendsRedDotList[privateEntityID] = newMap
	} else {
		s.friendsRedDotList[privateEntityID][entityID] += 1
	}

	return
}

/*************************************************下面为注册的方法*************************************************/

func (s *_ChatMgr) OnSyncPrivateFriendsListRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InFriendsList{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnSyncPrivateFriendsListRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || len(req.FriendsList) <= 0 {
		log.Waring("-->logic--_ChatMgr--OnSyncPrivateFriendsListRequest--req.EntityID <= 0 || len(req.FriendsList) <= 0")
		return
	}

	if s.friendsList[req.EntityID] == nil {
		s.friendsList[req.EntityID] = list.New()

		for _, v := range req.FriendsList {
			info := privateChatEntity{
				entityID:   v.EntityID,
				playerName: v.PlayerName,
				sex:        v.Sex,
				playerIcon: v.PlayerIcon,
				iconFrame:  v.IconFrame,
				vipLv:      v.VipLv,
			}
			s.friendsList[req.EntityID].PushFront(info)
		}
	}

	return
}

func (s *_ChatMgr) OnSendWorldMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.SendWorldMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnSendWorldMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.Context == "" {
		log.Waring("-->logic--_ChatMgr--OnSendWorldMsgRequest--req.EntityID <= 0 || req.Context == nil")
		return
	}

	s.InnerSendWorldMsg(uint32(req.MType), req.EntityID, req.Context)
	return
}

func (s *_ChatMgr) OnSendClubMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.SendClubMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnSendClubMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.ClubID <= 0 || req.Context == "" {
		log.Waring("-->logic--_ChatMgr--OnSendClubMsgRequest--req.EntityID <= 0 || req.Context == nil")
		return
	}

	s.InnerSendClubMsg(uint32(req.MType), req.EntityID, req.ClubID, req.Context)
	return
}

func (s *_ChatMgr) OnSendPrivateChatMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.SendPrivateChatMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnSendPrivateChatMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.PrivateEntityID <= 0 || req.Context == "" {
		log.Waring("-->logic--_ChatMgr--OnSendPrivateChatMsgRequest--req.EntityID <= 0 || req.Context == nil")
		return
	}

	s.InnerSendPrivateChatMsg(uint32(req.MType), req.EntityID, req.PrivateEntityID, req.Context)
	return
}

func (s *_ChatMgr) OnGetWorldMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetWorldMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnGetWorldMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnGetWorldMsgRequest--req.EntityID <= 0")
		return
	}

	respList := make([]*gmsg.MsgInfo, 0)
	if s.worldList.Len() > 0 {
		var count uint32
		// 遍历输出链表内容
		for e := s.worldList.Front(); e != nil; e = e.Next() {
			if count >= consts.MSGMAXNUM {
				break
			}
			count++

			ev, ok := e.Value.(msgInfo)
			if !ok {
				continue
			}

			respList = append(respList, &gmsg.MsgInfo{
				MType:           ev.mType,
				EntityID:        ev.entityID,
				EntityName:      ev.playerName,
				EntitySex:       ev.sex,
				EntityIcon:      ev.playerIcon,
				EntityIconFrame: ev.iconFrame,
				EntityVipLv:     ev.vipLv,
				ChatBox:         ev.chatBox,
				Content:         ev.content,
				Designation:     ev.collectId,
				CreatTime:       ev.creatTime,
			})
		}
	}

	//开始初始化桌面信息
	resp := &gmsg.GetWorldMsgResponse{
		Code:     0,
		EntityID: req.EntityID,
		List:     respList,
	}

	targetEntityIDs := []uint32{req.EntityID}

	log.Info("-->logic--_BattleC8Mgr--OnGetWorldMsgRequest--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_GetWorldMsgResponse, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) OnGetClubMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetClubMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnGetClubMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_ChatMgr--OnGetClubMsgRequest--req--", req)

	if req.EntityID <= 0 || req.ClubID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnGetClubMsgRequest--req.EntityID <= 0 || req.ClubID <= 0")
		return
	}

	emClub := Entity.EmClub.GetEntityByID(req.ClubID)
	if emClub == nil {
		log.Waring("-->logic--_ChatMgr--OnGetClubMsgRequest--emClub == nil")
		return
	}

	club := emClub.(*entity.Club)
	if club == nil || club.ClubID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnGetClubMsgRequest--club == nil", req.ClubID)
		return
	}

	respList := make([]*gmsg.MsgInfo, 0)
	if s.clubList[req.ClubID] != nil && s.clubList[req.ClubID].Len() > 0 {
		var count uint32
		// 遍历输出链表内容
		for e := s.clubList[req.ClubID].Front(); e != nil; e = e.Next() {
			if count >= consts.MSGMAXNUM {
				break
			}
			count++

			ev, ok := e.Value.(msgInfo)
			if !ok {
				continue
			}

			respList = append(respList, &gmsg.MsgInfo{
				MType:           ev.mType,
				EntityID:        ev.entityID,
				EntityName:      ev.playerName,
				EntitySex:       ev.sex,
				EntityIcon:      ev.playerIcon,
				EntityIconFrame: ev.iconFrame,
				EntityVipLv:     ev.vipLv,
				ChatBox:         ev.chatBox,
				Content:         ev.content,
				Designation:     ev.collectId,
				CreatTime:       ev.creatTime,
			})
		}
	}

	//开始初始化桌面信息
	resp := &gmsg.GetClubMsgResponse{
		Code:     0,
		EntityID: req.EntityID,
		ClubID:   req.ClubID,
		List:     respList,
	}

	targetEntityIDs := []uint32{req.EntityID}

	log.Info("-->logic--_BattleC8Mgr--OnGetClubMsgRequest--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_GetClubMsgResponse, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) OnGetPrivateChatEntityRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetPrivateChatEntityRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnGetPrivateChatEntityRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_ChatMgr--OnGetPrivateChatEntityRequest--req--", req)

	if req.EntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnGetPrivateChatEntityRequest--req.EntityID <= 0")
		return
	}

	respList := make([]*gmsg.PrivateChatEntity, 0)
	if s.friendsList[req.EntityID] != nil && s.friendsList[req.EntityID].Len() > 0 {
		for e := s.friendsList[req.EntityID].Front(); e != nil; e = e.Next() {
			ev, ok := e.Value.(privateChatEntity)
			if !ok {
				continue
			}

			respList = append(respList, &gmsg.PrivateChatEntity{
				EntityID:        ev.entityID,
				EntityName:      ev.playerName,
				EntitySex:       ev.sex,
				EntityIcon:      ev.playerIcon,
				EntityIconFrame: ev.iconFrame,
				EntityVipLv:     ev.vipLv,
				RedDotNum:       s.GetPrivateChatRedDotNum(req.EntityID, ev.entityID),
			})
		}
	}

	//开始初始化桌面信息
	resp := &gmsg.GetPrivateChatEntityResponse{
		Code:     0,
		EntityID: req.EntityID,
		List:     respList,
	}

	targetEntityIDs := []uint32{req.EntityID}

	log.Info("-->logic--_BattleC8Mgr--OnGetPrivateChatEntityRequest--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_GetPrivateChatEntityResponse, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) OnGetPrivateChatMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetPrivateChatMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnGetPrivateChatMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_ChatMgr--OnGetPrivateChatMsgRequest--req--", req)

	if req.EntityID <= 0 || req.PrivateEntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnGetPrivateChatMsgRequest--req.EntityID <= 0 || req.PrivateEntityID <= 0")
		return
	}

	respList := make([]*gmsg.MsgInfo, 0)

	msgKey := req.EntityID + req.PrivateEntityID
	if s.friendsMsgList[msgKey] != nil && s.friendsMsgList[msgKey].Len() > 0 {
		var count uint32
		// 遍历输出链表内容
		for e := s.friendsMsgList[msgKey].Front(); e != nil; e = e.Next() {
			if count >= consts.MSGMAXNUM {
				break
			}
			count++

			ev, ok := e.Value.(msgInfo)
			if !ok {
				continue
			}

			respList = append(respList, &gmsg.MsgInfo{
				MType:           ev.mType,
				EntityID:        ev.entityID,
				EntityName:      ev.playerName,
				EntitySex:       ev.sex,
				EntityIcon:      ev.playerIcon,
				EntityIconFrame: ev.iconFrame,
				EntityVipLv:     ev.vipLv,
				ChatBox:         ev.chatBox,
				Content:         ev.content,
				Designation:     ev.collectId,
				CreatTime:       ev.creatTime,
			})
		}
	}

	//开始初始化桌面信息
	resp := &gmsg.GetPrivateChatMsgResponse{
		Code:            0,
		EntityID:        req.EntityID,
		PrivateEntityID: req.PrivateEntityID,
		List:            respList,
	}

	targetEntityIDs := []uint32{req.EntityID}

	log.Info("-->logic--_BattleC8Mgr--OnGetPrivateChatMsgRequest--Resp:", resp)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_GetPrivateChatMsgResponse, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) OnSeePrivateChatMsgRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.SeePrivateChatMsgRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnSeePrivateChatMsgRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.PrivateEntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnSeePrivateChatMsgRequest--req.PrivateEntityID <= 0 || req.EntityID <= 0")
		return
	}

	if s.friendsRedDotList[req.EntityID] != nil {
		s.friendsRedDotList[req.EntityID][req.PrivateEntityID] = 0
	}

	//开始初始化桌面信息
	resp := &gmsg.SeePrivateChatMsgResponse{
		Code:            0,
		EntityID:        req.EntityID,
		PrivateEntityID: req.PrivateEntityID,
	}
	targetEntityIDs := []uint32{req.EntityID}
	log.Info("-->logic--_BattleC8Mgr--OnSeePrivateChatMsgRequest--Resp:", resp)
	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_SeePrivateChatMsgResponse, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) OnDelPrivateChatEntityRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.DelPrivateChatEntityRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnDelPrivateChatEntityRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.PrivateEntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnDelPrivateChatEntityRequest--req.PrivateEntityID <= 0 || req.EntityID <= 0")
		return
	}

	if s.friendsList[req.EntityID] != nil {
		if s.friendsList[req.EntityID].Len() > 0 {
			for e := s.friendsList[req.EntityID].Front(); e != nil; e = e.Next() {
				ev, ok := e.Value.(privateChatEntity)
				if !ok {
					continue
				}

				if ev.entityID == req.PrivateEntityID {
					s.friendsList[req.EntityID].Remove(e)
					break
				}
			}
		}
	}

	//开始初始化桌面信息
	resp := &gmsg.DelPrivateChatEntityResponse{
		Code:            0,
		EntityID:        req.EntityID,
		PrivateEntityID: req.PrivateEntityID,
	}
	targetEntityIDs := []uint32{req.EntityID}
	log.Info("-->logic--_BattleC8Mgr--OnDelPrivateChatEntityRequest--Resp:", resp)
	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_DelPrivateChatEntityResponse, resp, targetEntityIDs)
	return
}

func (s *_ChatMgr) OnGetOfflinePrivateFriendsResponse(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InGetOfflinePrivateFriendsResponse{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_ChatMgr--OnGetOfflinePrivateFriendsResponse--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 || req.FriendsInfo.EntityID <= 0 {
		log.Waring("-->logic--_ChatMgr--OnGetOfflinePrivateFriendsResponse--req.EntityID <= 0 || len(req.FriendsList) <= 0")
		return
	}

	entityChat := privateChatEntity{
		entityID:   req.FriendsInfo.EntityID,
		playerName: req.FriendsInfo.PlayerName,
		sex:        req.FriendsInfo.Sex,
		playerIcon: req.FriendsInfo.PlayerIcon,
		iconFrame:  req.FriendsInfo.IconFrame,
		vipLv:      req.FriendsInfo.VipLv,
	}
	if s.friendsList[req.EntityID] == nil {
		fList := list.New()
		fList.PushFront(entityChat)
		s.friendsList[req.EntityID] = fList

		s.friendsListUpdateTime[req.EntityID] = time.Now().Unix()
	} else {
		if !s.checkFriendList(req.EntityID, req.FriendsInfo.EntityID) {
			s.friendsList[req.EntityID].PushFront(entityChat)

			s.friendsListUpdateTime[req.EntityID] = time.Now().Unix()
		}
	}

	return
}
