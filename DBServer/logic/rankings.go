package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/DBServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
)

const (
	CUE_RANK_TYPE        = 1
	WEALTH_RANK_TYPE     = 2
	PEAK_RANK_TYPE       = 3
	CELEBRITY_RANK_TYPE  = 4
	POPULARITY_RANK_TYPE = 5
)

type _Rankings struct {
}

var Rankings _Rankings

func (s *_Rankings) Init() {
	//注册逻辑业务事件
	//event.On("Msg_MultiNinjaPointWarEnemyTeam", reflect.ValueOf(TeamRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Request), reflect.ValueOf(s.OnRankDataSortRequest))
}

// 查询排行榜数据 DB服->游戏服
func (s *_Rankings) OnRankDataSortRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InRankingsDbDataRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	var respList []*gmsg.InRankInfo

	switch req.RankType {
	case CUE_RANK_TYPE:
		tEntityPlayerArgs := make([]entity.EntityPlayer, 0)
		// 构建排序规则
		err = DBConnect.GetLimitDataAndSort(consts.COLLECTION_PLAYER, 100, nil, &tEntityPlayerArgs, "-CharmNum")
		if err != nil {
			log.Error("-->logic._Rankings--OnRankDataSortRequest--CUE_RANK_TYPE--err--", err)
			return
		}
		if len(tEntityPlayerArgs) > 0 {
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
	case WEALTH_RANK_TYPE:
		tEntityPlayerArgs := make([]entity.EntityPlayer, 0)
		err = DBConnect.GetLimitDataAndSort(consts.COLLECTION_PLAYER, 50, nil, &tEntityPlayerArgs, "-NumGold")
		if err != nil {
			log.Error("-->logic._Rankings--OnRankDataSortRequest--WEALTH_RANK_TYPE--err--", err)
			return
		}
		if len(tEntityPlayerArgs) > 0 {
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
	case PEAK_RANK_TYPE:
		tEntityPlayerArgs := make([]entity.EntityPlayer, 0)
		err = DBConnect.GetLimitDataAndSort(consts.COLLECTION_PLAYER, 100, nil, &tEntityPlayerArgs, "-PeakRankExp")
		if err != nil {
			log.Error("-->logic._Rankings--OnRankDataSortRequest--PEAK_RANK_TYPE--err--", err)
			return
		}
		if len(tEntityPlayerArgs) > 0 {
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
	case CELEBRITY_RANK_TYPE:
		tEntityPlayerArgs := make([]entity.EntityPlayer, 0)
		err = DBConnect.GetLimitDataAndSort(consts.COLLECTION_PLAYER, 100, nil, &tEntityPlayerArgs, "-FansNum")
		if err != nil {
			log.Error("-->logic._Rankings--OnRankDataSortRequest--CELEBRITY_RANK_TYPE--err", err)
			return
		}
		if len(tEntityPlayerArgs) > 0 {
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
	case POPULARITY_RANK_TYPE:
		tEntityPlayerArgs := make([]entity.EntityPlayer, 0)
		err = DBConnect.GetLimitDataAndSort(consts.COLLECTION_PLAYER, 100, nil, &tEntityPlayerArgs, "-PopularityValue")
		if err != nil {
			log.Error("-->logic._Rankings--OnRankDataSortRequest--POPULARITY_RANK_TYPE--err", err)
			return
		}
		if len(tEntityPlayerArgs) > 0 {
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

	resp := &gmsg.InRankingsDbDataResponse{
		Code:     resp_code.CODE_SUCCESS,
		RankType: req.RankType,
		List:     respList,
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Rankings_Db_Data_Response), resp, network.ServerType_Game)
}
