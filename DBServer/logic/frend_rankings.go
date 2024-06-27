package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
	"reflect"
	"sort"
)

const (
	FRIEND_CUE_RANK_TYPE        = 1
	FRIEND_WEALTH_RANK_TYPE     = 2
	FRIEND_PEAK_RANK_TYPE       = 3
	FRIEND_CELEBRITY_RANK_TYPE  = 4
	FRIEND_POPULARITY_RANK_TYPE = 5
)

type _FriendRankings struct {
}

var FriendRankings _FriendRankings

func (s *_FriendRankings) Init() {
	//注册逻辑业务事件
	//event.On("Msg_MultiNinjaPointWarEnemyTeam", reflect.ValueOf(TeamRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Request), reflect.ValueOf(s.OnFriendRankDataSortRequest))
}

// 查询好友排行榜数据 DB服->游戏服
func (s *_FriendRankings) OnFriendRankDataSortRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InRankingsFriendDbDataRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	player := Entity.EmEntityPlayer.GetEntityByID(req.EntityID)
	tPlayer := player.(*entity.EntityPlayer)
	fList := tPlayer.GetMyFriends()
	if len(fList) <= 0 {
		return
	}

	tEntityPlayerArgs := make([]*entity.EntityPlayer, 0)
	for _, v := range fList {
		if !tPlayer.IsInFansList(v.EntityID) {
			continue
		}

		player := Entity.EmEntityPlayer.GetEntityByID(v.EntityID)
		tPlayer := player.(*entity.EntityPlayer)
		tEntityPlayerArgs = append(tEntityPlayerArgs, tPlayer)
	}

	var respList []*gmsg.InRankInfo
	switch req.RankType {
	case FRIEND_CUE_RANK_TYPE:

		sort.Sort(CharmNumSlice(tEntityPlayerArgs))
		var incrTotal int32
		if len(tEntityPlayerArgs) > 0 {
			if incrTotal > 49 {
				break
			}
			incrTotal++

			for _, v := range tEntityPlayerArgs {
				//拿出背包所有的球杆
				var allCueItemIds []uint32
				if len(v.BagList) > 0 {
					for _, vb := range v.BagList {
						if vb.ItemType == 1 {
							allCueItemIds = append(allCueItemIds, vb.TableID)
						}
					}
				}

				info := &gmsg.InRankInfo{
					EntityID:     v.EntityID,
					Name:         v.PlayerName,
					Icon:         v.PlayerIcon,
					IconFrame:    v.IconFrame,
					Sex:          v.Sex,
					Num:          v.CharmNum,
					VipLv:        v.VipLv,
					PeakRankLv:   0,
					PeakRankStar: 0,
					ItemIds:      allCueItemIds,
				}

				respList = append(respList, info)
			}
		}
	case FRIEND_WEALTH_RANK_TYPE:

		sort.Sort(NumGoldSlice(tEntityPlayerArgs))

		var incrTotal int32
		if len(tEntityPlayerArgs) > 0 {
			if incrTotal > 49 {
				break
			}
			incrTotal++

			for _, v := range tEntityPlayerArgs {
				info := &gmsg.InRankInfo{
					EntityID:     v.EntityID,
					Name:         v.PlayerName,
					Icon:         v.PlayerIcon,
					IconFrame:    v.IconFrame,
					Sex:          v.Sex,
					Num:          v.NumGold,
					VipLv:        v.VipLv,
					PeakRankLv:   0,
					PeakRankStar: 0,
					ItemIds:      nil,
				}

				respList = append(respList, info)
			}
		}
	case FRIEND_PEAK_RANK_TYPE:

		sort.Sort(PeakRankExpSlice(tEntityPlayerArgs))

		var incrTotal int32
		if len(tEntityPlayerArgs) > 0 {
			if incrTotal > 49 {
				break
			}
			incrTotal++

			for _, v := range tEntityPlayerArgs {
				info := &gmsg.InRankInfo{
					EntityID:     v.EntityID,
					Name:         v.PlayerName,
					Icon:         v.PlayerIcon,
					IconFrame:    v.IconFrame,
					Sex:          v.Sex,
					Num:          0,
					VipLv:        v.VipLv,
					PeakRankLv:   v.PeakRankLv,
					PeakRankStar: v.PeakRankExp,
					ItemIds:      nil,
				}

				respList = append(respList, info)
			}
		}
	case FRIEND_CELEBRITY_RANK_TYPE:

		sort.Sort(FansNumSlice(tEntityPlayerArgs))

		var incrTotal int32
		if len(tEntityPlayerArgs) > 0 {
			if incrTotal > 49 {
				break
			}
			incrTotal++

			for _, v := range tEntityPlayerArgs {
				info := &gmsg.InRankInfo{
					EntityID:     v.EntityID,
					Name:         v.PlayerName,
					Icon:         v.PlayerIcon,
					IconFrame:    v.IconFrame,
					Sex:          v.Sex,
					Num:          v.FansNum,
					VipLv:        v.VipLv,
					PeakRankLv:   0,
					PeakRankStar: 0,
					ItemIds:      nil,
				}

				respList = append(respList, info)
			}
		}
	case FRIEND_POPULARITY_RANK_TYPE:

		sort.Sort(PopularityNumSlice(tEntityPlayerArgs))

		var incrTotal int32
		if len(tEntityPlayerArgs) > 0 {
			if incrTotal > 49 {
				break
			}
			incrTotal++

			for _, v := range tEntityPlayerArgs {
				info := &gmsg.InRankInfo{
					EntityID:     v.EntityID,
					Name:         v.PlayerName,
					Icon:         v.PlayerIcon,
					IconFrame:    v.IconFrame,
					Sex:          v.Sex,
					Num:          uint32(v.PopularityValue),
					VipLv:        v.VipLv,
					PeakRankLv:   0,
					PeakRankStar: 0,
					ItemIds:      nil,
				}

				respList = append(respList, info)
			}
		}
	}

	resp := &gmsg.InRankingsFriendDbDataResponse{
		Code:     resp_code.CODE_SUCCESS,
		RankType: req.RankType,
		EntityID: req.EntityID,
		List:     respList,
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Friend_Db_Data_Response), resp, network.ServerType_Game)
}

// 按照 PeakRankExp 从大到小排序
type PeakRankExpSlice []*entity.EntityPlayer

func (a PeakRankExpSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a PeakRankExpSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a PeakRankExpSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].PeakRankExp < a[i].PeakRankExp
}

// 按照 NumGoldSlice 从大到小排序
type NumGoldSlice []*entity.EntityPlayer

func (a NumGoldSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a NumGoldSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a NumGoldSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].NumGold < a[i].NumGold
}

// 按照 CharmNumSlice 从大到小排序
type CharmNumSlice []*entity.EntityPlayer

func (a CharmNumSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a CharmNumSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a CharmNumSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].CharmNum < a[i].CharmNum
}

// 按照 FansNum 从大到小排序
type FansNumSlice []*entity.EntityPlayer

func (a FansNumSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a FansNumSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a FansNumSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].FansNum < a[i].FansNum
}

// 按照 Popularity 从大到小排序
type PopularityNumSlice []*entity.EntityPlayer

func (a PopularityNumSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a PopularityNumSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a PopularityNumSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].PopularityValue < a[i].PopularityValue
}
