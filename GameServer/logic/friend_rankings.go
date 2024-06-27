package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/tools"
	"reflect"
	"strconv"
	"time"
)

const (
	FRIEND_CUE_RANK_TYPE        = 1
	FRIEND_WEALTH_RANK_TYPE     = 2
	FRIEND_PEAK_RANK_TYPE       = 3
	FRIEND_CELEBRITY_RANK_TYPE  = 4
	FRIEND_POPULARITY_RANK_TYPE = 5
)

var FriendRankings _FriendRankings

type _FriendRankings struct {
	cue              map[uint32][]*rankInfo
	wealth           map[uint32][]*rankInfo
	peakRank         map[uint32][]*rankInfo
	celebrity        map[uint32][]*rankInfo
	popularity       map[uint32][]*rankInfo
	entityUpdateTime map[uint32]int64
}

func (s *_FriendRankings) Init() {
	s.cue = make(map[uint32][]*rankInfo)
	s.wealth = make(map[uint32][]*rankInfo)
	s.peakRank = make(map[uint32][]*rankInfo)
	s.celebrity = make(map[uint32][]*rankInfo)
	s.popularity = make(map[uint32][]*rankInfo)
	s.entityUpdateTime = make(map[uint32]int64)

	//注册逻辑业务事件
	event.OnNet(gmsg.MsgTile_Rankings_FriendCueListRequest, reflect.ValueOf(s.GetFriendCueRankList))
	event.OnNet(gmsg.MsgTile_Rankings_FriendWealthListRequest, reflect.ValueOf(s.GetFriendWealthRankList))
	event.OnNet(gmsg.MsgTile_Rankings_FriendPeakRankListRequest, reflect.ValueOf(s.GetFriendPeakRankRankList))
	event.OnNet(gmsg.MsgTile_Rankings_FriendCelebrityListRequest, reflect.ValueOf(s.GetFriendCelebrityRankList))
	event.OnNet(gmsg.MsgTile_Rankings_FriendPopularityListRequest, reflect.ValueOf(s.GetFriendPopularityRankList))

	//内部通信
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Response), reflect.ValueOf(s.SetFriendRankListData))
}

// 定时通知DB同步排行榜数据
func (s *_FriendRankings) SendSyncMsgToDbForRankData(entityID uint32) {
	if entityID <= 0 {
		return
	}

	//开始初始化桌面信息
	respCue := &gmsg.InRankingsFriendDbDataRequest{}
	respCue.RankType = FRIEND_CUE_RANK_TYPE
	respCue.EntityID = entityID
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Request), respCue, network.ServerType_DB)

	respWealth := &gmsg.InRankingsFriendDbDataRequest{}
	respWealth.RankType = FRIEND_WEALTH_RANK_TYPE
	respWealth.EntityID = entityID
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Request), respWealth, network.ServerType_DB)

	respPeakRank := &gmsg.InRankingsFriendDbDataRequest{}
	respPeakRank.RankType = FRIEND_PEAK_RANK_TYPE
	respPeakRank.EntityID = entityID
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Request), respPeakRank, network.ServerType_DB)

	respCelebrity := &gmsg.InRankingsFriendDbDataRequest{}
	respCelebrity.RankType = FRIEND_CELEBRITY_RANK_TYPE
	respCelebrity.EntityID = entityID
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Request), respCelebrity, network.ServerType_DB)

	popularity := &gmsg.InRankingsFriendDbDataRequest{}
	popularity.RankType = FRIEND_POPULARITY_RANK_TYPE
	popularity.EntityID = entityID
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Request), popularity, network.ServerType_DB)
}

// 设置排行榜数据 DB->GAME
func (s *_FriendRankings) SetFriendRankListData(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InRankingsFriendDbDataResponse{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	entityID := req.EntityID
	if entityID <= 0 {
		return
	}

	if len(req.List) > 0 {
		var list []*rankInfo
		for _, v := range req.List {
			//itemId存放的是所有球杆的id，在这里拿出魅力值最大两个
			var itemIds []uint32
			if len(v.ItemIds) > 0 {
				itemIds = s.getItemIdsFromBagList(v.ItemIds)
			}

			info := rankInfo{
				EntityID:     v.EntityID,
				Name:         v.Name,
				Icon:         v.Icon,
				IconFrame:    v.IconFrame,
				Sex:          v.Sex,
				Num:          v.Num,
				VipLv:        v.VipLv,
				PeakRankLv:   v.PeakRankLv,
				PeakRankStar: v.PeakRankStar,
				ItemIds:      itemIds,
			}

			list = append(list, &info)
		}

		if len(list) > 0 {
			switch req.RankType {
			case FRIEND_CUE_RANK_TYPE:
				s.cue[entityID] = list
			case FRIEND_WEALTH_RANK_TYPE:
				s.wealth[entityID] = list
			case FRIEND_PEAK_RANK_TYPE:
				s.peakRank[entityID] = list
			case FRIEND_CELEBRITY_RANK_TYPE:
				s.celebrity[entityID] = list
			case POPULARITY_RANK_TYPE:
				s.popularity[entityID] = list
			}
		}

	}
	return
}

// 获取球杆排行榜
func (s *_FriendRankings) GetFriendCueRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetFriendCueRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	var list []*gmsg.RankInfo
	var user *gmsg.RankInfo
	//TODO:需要一个间隔更新
	if s.entityUpdateTime[req.EntityID] > 0 {
		list = s.getRankList(FRIEND_CUE_RANK_TYPE, req.EntityID)
		user = s.getRankUserInfo(FRIEND_CUE_RANK_TYPE, req.EntityID)
	} else {
		now := time.Now().Unix()
		s.entityUpdateTime[req.EntityID] = now

		//更新好友排行榜
		s.SendSyncMsgToDbForRankData(req.EntityID)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetFriendCueRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_FriendRankings--GetFriendCueRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_FriendCueListResponse, resp, targetEntityIDs)
	return
}

// 获取财富排行榜
func (s *_FriendRankings) GetFriendWealthRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetFriendWealthRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	var list []*gmsg.RankInfo
	var user *gmsg.RankInfo
	if s.entityUpdateTime[req.EntityID] > 0 {
		list = s.getRankList(FRIEND_WEALTH_RANK_TYPE, req.EntityID)
		user = s.getRankUserInfo(FRIEND_WEALTH_RANK_TYPE, req.EntityID)
	} else {
		now := time.Now().Unix()
		s.entityUpdateTime[req.EntityID] = now

		//更新好友排行榜
		s.SendSyncMsgToDbForRankData(req.EntityID)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetFriendWealthRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_FriendRankings--GetFriendWealthRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_FriendWealthListResponse, resp, targetEntityIDs)
	return
}

// 获取段位排行榜
func (s *_FriendRankings) GetFriendPeakRankRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetFriendPeakRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	var list []*gmsg.RankInfo
	var user *gmsg.RankInfo
	if s.entityUpdateTime[req.EntityID] > 0 {
		list = s.getRankList(FRIEND_PEAK_RANK_TYPE, req.EntityID)
		user = s.getRankUserInfo(FRIEND_PEAK_RANK_TYPE, req.EntityID)
	} else {
		now := time.Now().Unix()
		s.entityUpdateTime[req.EntityID] = now

		//更新好友排行榜
		s.SendSyncMsgToDbForRankData(req.EntityID)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetFriendPeakRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_FriendRankings--GetFriendPeakRankRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_FriendPeakRankListResponse, resp, targetEntityIDs)
	return
}

// 获取名人排行榜
func (s *_FriendRankings) GetFriendCelebrityRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetFriendCelebrityRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	var list []*gmsg.RankInfo
	var user *gmsg.RankInfo
	if s.entityUpdateTime[req.EntityID] > 0 {
		list = s.getRankList(FRIEND_CELEBRITY_RANK_TYPE, req.EntityID)
		user = s.getRankUserInfo(FRIEND_CELEBRITY_RANK_TYPE, req.EntityID)
	} else {
		now := time.Now().Unix()
		s.entityUpdateTime[req.EntityID] = now

		//更新好友排行榜
		s.SendSyncMsgToDbForRankData(req.EntityID)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetFriendCelebrityRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_FriendRankings--GetFriendCelebrityRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_FriendCelebrityListResponse, resp, targetEntityIDs)
	return
}

// 获取人气排行榜
func (s *_FriendRankings) GetFriendPopularityRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetFriendPopularityRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	var list []*gmsg.RankInfo
	var user *gmsg.RankInfo
	if s.entityUpdateTime[req.EntityID] > 0 {
		list = s.getRankList(FRIEND_POPULARITY_RANK_TYPE, req.EntityID)
		user = s.getRankUserInfo(FRIEND_POPULARITY_RANK_TYPE, req.EntityID)
	} else {
		now := time.Now().Unix()
		s.entityUpdateTime[req.EntityID] = now

		//更新好友排行榜
		s.SendSyncMsgToDbForRankData(req.EntityID)
	}

	//开始初始化桌面信息
	resp := &gmsg.GetFriendPopularityRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_FriendRankings--GetFriendCelebrityRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_FriendPopularityListResponse, resp, targetEntityIDs)
	return
}

func (s *_FriendRankings) getItemIdsFromBagList(itemIds []uint32) []uint32 {
	var resp []uint32

	charmScores := make(map[uint32]uint32)
	for _, vi := range itemIds {
		if vi <= 0 {
			continue
		}

		cueInfo := Table.GetCueCfg(strconv.Itoa(int(vi)))
		if cueInfo == nil {
			continue
		}
		charmScores[vi] = cueInfo.CharmScore
	}

	if len(charmScores) > 0 {
		rCharmScores := tools.RankByCount(charmScores)
		i := 0
		for _, vrcs := range rCharmScores {
			if i > 2 {
				break
			}

			resp = append(resp, vrcs.Key)
			i++
		}
	}

	return resp
}

// 获取排行榜数据
func (s *_FriendRankings) getRankList(typer uint32, entityID uint32) []*gmsg.RankInfo {
	list := make([]*rankInfo, 0)
	var respList []*gmsg.RankInfo

	if entityID <= 0 {
		return respList
	}

	switch typer {
	case FRIEND_CUE_RANK_TYPE:
		list = s.cue[entityID]
	case FRIEND_WEALTH_RANK_TYPE:
		list = s.wealth[entityID]
	case FRIEND_PEAK_RANK_TYPE:
		list = s.peakRank[entityID]
	case FRIEND_CELEBRITY_RANK_TYPE:
		list = s.celebrity[entityID]
	case FRIEND_POPULARITY_RANK_TYPE:
		list = s.popularity[entityID]
	}

	if len(list) <= 0 {
		return respList
	}

	for k, v := range list {
		peakRankLv := uint32(1)
		if v.PeakRankLv > 0 {
			peakRankLv = v.PeakRankLv
		}

		num := v.Num
		var peakRankStar uint32
		if typer == FRIEND_PEAK_RANK_TYPE {
			start := PeakRankExp.GetLevelBeginExp(peakRankLv)

			if start > 0 && v.PeakRankStar > 0 && v.PeakRankStar > start {
				peakRankStar = v.PeakRankStar - start
				num = peakRankStar
			}
		}

		info := &gmsg.RankInfo{
			RankNum:      strconv.Itoa(k + 1),
			EntityID:     v.EntityID,
			Name:         v.Name,
			Icon:         v.Icon,
			IconFrame:    v.IconFrame,
			Sex:          v.Sex,
			Num:          num,
			VipLv:        v.VipLv,
			PeakRankLv:   peakRankLv,
			PeakRankStar: peakRankStar,
			ItemIds:      v.ItemIds,
		}
		respList = append(respList, info)
	}

	return respList
}

// 获取排行榜个人数据
func (s *_FriendRankings) getRankUserInfo(typer uint32, entityID uint32) *gmsg.RankInfo {
	if entityID <= 0 {
		return nil
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return nil
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	var itemIds []uint32
	if len(tEntityPlayer.BagList) > 0 {
		for _, v := range tEntityPlayer.BagList {
			itemIds = append(itemIds, v.TableID)
		}
	}

	var randItemIds []uint32
	if len(itemIds) > 0 {
		randItemIds = s.getItemIdsFromBagList(itemIds)
	}

	userInfo := &gmsg.RankInfo{}
	userInfo.EntityID = tEntityPlayer.EntityID
	userInfo.Name = tEntityPlayer.PlayerName
	userInfo.Icon = tEntityPlayer.PlayerIcon
	userInfo.IconFrame = tEntityPlayer.IconFrame
	userInfo.Sex = tEntityPlayer.Sex
	userInfo.VipLv = tEntityPlayer.VipLv
	userInfo.PeakRankLv = tEntityPlayer.PeakRankLv
	userInfo.PeakRankStar = tEntityPlayer.PeakRankExp
	userInfo.ItemIds = randItemIds

	list := make([]*rankInfo, 0)
	switch typer {
	case FRIEND_CUE_RANK_TYPE:
		userInfo.Num = tEntityPlayer.CharmNum
		list = s.cue[entityID]
	case FRIEND_WEALTH_RANK_TYPE:
		userInfo.Num = tEntityPlayer.NumGold
		list = s.wealth[entityID]
	case FRIEND_PEAK_RANK_TYPE:
		userInfo.Num = 0
		list = s.peakRank[entityID]
	case FRIEND_CELEBRITY_RANK_TYPE:
		userInfo.Num = tEntityPlayer.FansNum
		list = s.celebrity[entityID]
	case FRIEND_POPULARITY_RANK_TYPE:
		userInfo.Num = uint32(tEntityPlayer.PopularityValue)
		list = s.popularity[entityID]
	}

	var rankNum uint32
	if len(list) > 0 {
		for k, v := range list {
			if v.EntityID == entityID {
				rankNum = uint32(k + 1)
			}
		}
	}

	if rankNum > 0 {
		userInfo.RankNum = strconv.Itoa(int(rankNum))
	} else {
		userInfo.RankNum = "50+"
	}

	return userInfo
}
