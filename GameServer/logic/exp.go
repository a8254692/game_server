package logic

import (
	"BilliardServer/Common/table"
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/log"
	"errors"
	"sort"
	"strconv"
)

var Exp _Exp

type _Exp struct {
	maxLvExp    uint32
	expConfList []*table.PlayerLevelCfg
}

func (s *_Exp) Init() {
	s.expConfList = make([]*table.PlayerLevelCfg, 0)
	_ = s.SetExpConf()
}

// 设置配置列表(复制数据，避免使用者修改时将这里的缓存数据也给改了)
func (s *_Exp) SetExpConf() error {
	list := Table.GetPlayerLevelCfgMap()
	if list == nil {
		log.Error("-->logic--_Exp--SetExpConf--list == nil")
		return errors.New("-->logic--_Exp--SetExpConf--list == nil")
	}

	if len(list) <= 0 {
		log.Error("-->logic--_Exp--SetExpConf--len(list) <= 0")
		return errors.New("-->logic--_Exp--SetExpConf--len(list) <= 0")
	}

	arr := make([]*table.PlayerLevelCfg, 0)
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

	var maxLvExp uint32
	if len(keys) > 0 {
		sort.Ints(keys)

		for _, v := range keys {
			vStr := strconv.Itoa(v)
			arr = append(arr, list[vStr])

			maxLvExp = list[vStr].Exp
		}
	}

	s.maxLvExp = maxLvExp
	s.expConfList = arr
	return nil
}

// 经验转等级
func (s *_Exp) Exp2Level(userExp uint32) uint32 {
	var lvl uint32
	for _, v := range s.expConfList {
		lvl = v.Level

		if userExp < v.Exp {
			break
		}
	}
	return lvl
}

// 获取等级开始的经验值
func (s *_Exp) GetLevelBeginExp(level uint32) uint32 {
	if level < consts.MIN_USER_LV {
		return 0
	}

	var expStart uint32
	if len(s.expConfList) > 0 {
		//查不到默认2开始
		lvKey := 1
		for k, v := range s.expConfList {
			if level == v.Level {
				lvKey = k
			}
		}

		expStart = s.expConfList[lvKey-1].Exp
	}
	return expStart
}

// 获取展示的经验值
func (s *_Exp) GetLvShowExp(level uint32, allExp uint32) uint32 {
	if level < consts.MIN_USER_LV {
		return 0
	}

	startLvExp := s.GetLevelBeginExp(level)
	if allExp < startLvExp {
		return 0
	}

	return allExp - startLvExp
}
