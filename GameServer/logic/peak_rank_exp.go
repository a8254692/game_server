package logic

import (
	"BilliardServer/Common/table"
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/log"
	"errors"
	"sort"
	"strconv"
)

var PeakRankExp _PeakRankExp

type _PeakRankExp struct {
	maxLvStar           uint32
	peakRankExpConfList []*table.DanCfg
}

func (s *_PeakRankExp) Init() {
	s.peakRankExpConfList = make([]*table.DanCfg, 0)
	_ = s.SetPeakRankExpConf()

	//内部通信
	//event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PeakRankExp_Db_Data_Response), reflect.ValueOf(s.SetRankListData))
}

// 设置巅峰赛配置列表(复制数据，避免使用者修改时将这里的缓存数据也给改了)
func (s *_PeakRankExp) SetPeakRankExpConf() error {
	list := Table.GetDanCfgMap()
	if list == nil {
		log.Error("-->logic--_PeakRankExp--SetPeakRankExpConf--GetDanCfgMap--len(list) <= 0")
		return errors.New("-->logic--_PeakRankExp--SetPeakRankExpConf--GetDanCfgMap--len(list) <= 0")
	}

	if len(list) <= 0 {
		log.Error("-->logic--_PeakRankExp--SetPeakRankExpConf--len(list) <= 0")
		return errors.New("-->logic--_PeakRankExp--SetPeakRankExpConf--len(list) <= 0")
	}

	arr := make([]*table.DanCfg, 0)
	keys := make([]int, 0)
	if len(list) > 0 {
		for k := range list {
			kInt, err := strconv.Atoi(k)
			if err != nil {
				continue
			}

			keys = append(keys, kInt)
		}
	}

	var maxLvStar uint32
	if len(keys) > 0 {
		sort.Ints(keys)
		for _, v := range keys {
			vStr := strconv.Itoa(v)
			arr = append(arr, list[vStr])

			maxLvStar = list[vStr].UpgradeStar
		}
	}

	s.maxLvStar = maxLvStar
	s.peakRankExpConfList = arr
	return nil
}

// 经验转等级
func (s *_PeakRankExp) Exp2Level(userExp uint32) uint32 {
	var lvl uint32
	for _, v := range s.peakRankExpConfList {
		lvl = v.Lv

		if userExp < v.UpgradeStar {
			break
		}
	}
	return lvl
}

// 获取等级开始的经验值
func (s *_PeakRankExp) GetLevelBeginExp(level uint32) uint32 {
	if level <= consts.MIN_PEAK_RANK_LV {
		return 0
	}

	var expStart uint32
	if len(s.peakRankExpConfList) > 0 {
		//查不到默认2开始
		lvKey := 1
		for k, v := range s.peakRankExpConfList {
			if level == v.Lv {
				lvKey = k
			}
		}

		expStart = s.peakRankExpConfList[lvKey-1].UpgradeStar
	}

	return expStart
}

// 获取展示的经验值
func (s *_PeakRankExp) GetLvShowExp(level uint32, allExp uint32) uint32 {
	if level < consts.MIN_PEAK_RANK_LV {
		return 0
	}

	startLvExp := s.GetLevelBeginExp(level)
	if allExp < startLvExp {
		return 0
	}

	return allExp - startLvExp
}
