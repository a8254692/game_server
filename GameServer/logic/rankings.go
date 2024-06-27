package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"reflect"
	"strconv"
)

const (
	CUE_RANK_TYPE        = 1
	WEALTH_RANK_TYPE     = 2
	PEAK_RANK_TYPE       = 3
	CELEBRITY_RANK_TYPE  = 4
	POPULARITY_RANK_TYPE = 5
)

type rankInfo struct {
	EntityID     uint32
	Name         string
	Icon         uint32
	IconFrame    uint32
	Sex          uint32
	Num          uint32
	VipLv        uint32
	PeakRankLv   uint32
	PeakRankStar uint32
	ItemIds      []uint32
}

var Rankings _Rankings

type _Rankings struct {
	cue        []*rankInfo
	wealth     []*rankInfo
	peakRank   []*rankInfo
	celebrity  []*rankInfo
	popularity []*rankInfo
}

func (s *_Rankings) Init() {
	//创建对战
	s.cue = make([]*rankInfo, 0)
	s.wealth = make([]*rankInfo, 0)
	s.peakRank = make([]*rankInfo, 0)
	s.celebrity = make([]*rankInfo, 0)
	s.popularity = make([]*rankInfo, 0)

	timer.AddTimer(s, "SendSyncMsgToDbForRankData", 300000, true)

	//注册逻辑业务事件
	event.OnNet(gmsg.MsgTile_Rankings_CueListRequest, reflect.ValueOf(s.GetCueRankList))
	event.OnNet(gmsg.MsgTile_Rankings_WealthListRequest, reflect.ValueOf(s.GetWealthRankList))
	event.OnNet(gmsg.MsgTile_Rankings_PeakRankListRequest, reflect.ValueOf(s.GetPeakRankRankList))
	event.OnNet(gmsg.MsgTile_Rankings_CelebrityListRequest, reflect.ValueOf(s.GetCelebrityRankList))
	event.OnNet(gmsg.MsgTile_Rankings_PopularityListRequest, reflect.ValueOf(s.GetPopularityRankList))

	//内部通信
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Response), reflect.ValueOf(s.SetRankListData))
}

// 定时通知DB同步排行榜数据
func (s *_Rankings) SendSyncMsgToDbForRankData() {
	//开始初始化桌面信息
	respCue := &gmsg.InRankingsDbDataRequest{}
	respCue.RankType = CUE_RANK_TYPE
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Request), respCue, network.ServerType_DB)

	respWealth := &gmsg.InRankingsDbDataRequest{}
	respWealth.RankType = WEALTH_RANK_TYPE
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Request), respWealth, network.ServerType_DB)

	resppeakRank := &gmsg.InRankingsDbDataRequest{}
	resppeakRank.RankType = PEAK_RANK_TYPE
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Request), resppeakRank, network.ServerType_DB)

	respCelebrity := &gmsg.InRankingsDbDataRequest{}
	respCelebrity.RankType = CELEBRITY_RANK_TYPE
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Request), respCelebrity, network.ServerType_DB)

	popularity := &gmsg.InRankingsDbDataRequest{}
	popularity.RankType = POPULARITY_RANK_TYPE
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Request), popularity, network.ServerType_DB)
}

// 设置排行榜数据 DB->GAME
func (s *_Rankings) SetRankListData(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InRankingsDbDataResponse{}
	err := msgEV.Unmarshal(req)
	if err != nil {
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
			case CUE_RANK_TYPE:
				s.cue = list
			case WEALTH_RANK_TYPE:
				s.wealth = list
			case PEAK_RANK_TYPE:
				s.peakRank = list
			case CELEBRITY_RANK_TYPE:
				s.celebrity = list
			case POPULARITY_RANK_TYPE:
				s.popularity = list
			}
		}
	}
	return
}

// 获取球杆排行榜
func (s *_Rankings) GetCueRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetCueRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	list := s.getRankList(CUE_RANK_TYPE)
	user := s.getRankUserInfo(CUE_RANK_TYPE, req.EntityID)

	//开始初始化桌面信息
	resp := &gmsg.GetCueRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_Rankings--GetCueRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_CueListResponse, resp, targetEntityIDs)
	return
}

// 获取财富排行榜
func (s *_Rankings) GetWealthRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetWealthRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	list := s.getRankList(WEALTH_RANK_TYPE)
	user := s.getRankUserInfo(WEALTH_RANK_TYPE, req.EntityID)

	//开始初始化桌面信息
	resp := &gmsg.GetWealthRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_Rankings--GetWealthRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_WealthListResponse, resp, targetEntityIDs)
	return
}

// 获取段位排行榜
func (s *_Rankings) GetPeakRankRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetPeakRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	list := s.getRankList(PEAK_RANK_TYPE)
	user := s.getRankUserInfo(PEAK_RANK_TYPE, req.EntityID)

	//开始初始化桌面信息
	resp := &gmsg.GetPeakRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_Rankings--GetPeakRankRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_PeakRankListResponse, resp, targetEntityIDs)
	return
}

// 获取名人排行榜
func (s *_Rankings) GetCelebrityRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetCelebrityRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	list := s.getRankList(CELEBRITY_RANK_TYPE)
	user := s.getRankUserInfo(CELEBRITY_RANK_TYPE, req.EntityID)

	//开始初始化桌面信息
	resp := &gmsg.GetCelebrityRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_Rankings--GetCelebrityRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_CelebrityListResponse, resp, targetEntityIDs)
	return
}

// 获取人气排行榜
func (s *_Rankings) GetPopularityRankList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.GetPopularityRankListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	list := s.getRankList(POPULARITY_RANK_TYPE)
	user := s.getRankUserInfo(POPULARITY_RANK_TYPE, req.EntityID)

	//开始初始化桌面信息
	resp := &gmsg.GetPopularityRankListResponse{}
	resp.Code = resp_code.CODE_SUCCESS
	resp.List = list
	resp.User = user

	log.Info("-->logic--_Rankings--GetCelebrityRankList--Resp:", resp)

	targetEntityIDs := []uint32{req.EntityID}

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Rankings_PopularityListResponse, resp, targetEntityIDs)
	return
}

func (s *_Rankings) getItemIdsFromBagList(itemIds []uint32) []uint32 {
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
func (s *_Rankings) getRankList(typer uint32) []*gmsg.RankInfo {
	list := make([]*rankInfo, 0)
	var respList []*gmsg.RankInfo

	switch typer {
	case CUE_RANK_TYPE:
		list = s.cue
	case WEALTH_RANK_TYPE:
		list = s.wealth
	case PEAK_RANK_TYPE:
		list = s.peakRank
	case CELEBRITY_RANK_TYPE:
		list = s.celebrity
	case POPULARITY_RANK_TYPE:
		list = s.popularity
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
		if typer == PEAK_RANK_TYPE {
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
func (s *_Rankings) getRankUserInfo(typer uint32, entityID uint32) *gmsg.RankInfo {
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
	case CUE_RANK_TYPE:
		userInfo.Num = tEntityPlayer.CharmNum
		list = s.cue
	case WEALTH_RANK_TYPE:
		userInfo.Num = tEntityPlayer.NumGold
		list = s.wealth
	case PEAK_RANK_TYPE:
		userInfo.Num = 0
		list = s.peakRank
	case CELEBRITY_RANK_TYPE:
		userInfo.Num = tEntityPlayer.FansNum
		list = s.celebrity
	case POPULARITY_RANK_TYPE:
		userInfo.Num = uint32(tEntityPlayer.PopularityValue)
		list = s.popularity
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
