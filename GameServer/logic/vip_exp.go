package logic

import (
	"BilliardServer/Common/table"
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/log"
	"errors"
	"sort"
	"strconv"
)

var VipExp _VipExp

type _VipExp struct {
	maxLvExp    uint32
	expConfList []*table.VipCfg
}

func (s *_VipExp) Init() {
	s.expConfList = make([]*table.VipCfg, 0)
	_ = s.SetExpConf()
}

func (s *_VipExp) GetExpConfList() []*table.VipCfg {
	return s.expConfList
}

func (s *_VipExp) GetExpConf(lv uint32) *table.VipCfg {
	var resp *table.VipCfg
	if len(s.expConfList) > 0 {
		for _, v := range s.expConfList {
			if v.Level == lv {
				resp = v
			}
		}
	}

	return resp
}

// 设置配置列表(复制数据，避免使用者修改时将这里的缓存数据也给改了)
func (s *_VipExp) SetExpConf() error {
	list := Table.GetVipCfgMap()
	if list == nil {
		log.Error("-->logic--_VipExp--SetExpConf--GetVipCfgMap--len(list) <= 0")
		return errors.New("-->logic--_VipExp--SetExpConf--GetVipCfgMap--len(list) <= 0")
	}

	if len(list) <= 0 {
		log.Error("-->logic--_VipExp--SetExpConf--len(list) <= 0")
		return errors.New("-->logic--_VipExp--SetExpConf--len(list) <= 0")
	}

	arr := make([]*table.VipCfg, 0)
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
func (s *_VipExp) Exp2Level(exp uint32) uint32 {
	var lvl uint32
	for _, v := range s.expConfList {
		lvl = v.Level

		if exp < v.Exp {
			break
		}
	}
	return lvl
}

// 获取等级开始的经验值
func (s *_VipExp) GetLevelBeginExp(level uint32) uint32 {
	if level < consts.MIN_VIP_LV {
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
func (s *_VipExp) GetLvShowExp(level uint32, allExp uint32) uint32 {
	if level < consts.MIN_VIP_LV {
		return 0
	}

	startLvExp := s.GetLevelBeginExp(level)
	if allExp < startLvExp {
		return 0
	}

	return allExp - startLvExp
}
