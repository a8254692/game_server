package logic

import (
	"errors"
	"sync"
)

func NewBattlePlayerMgr() (um *BattlePlayerMgr, err error) {
	um = new(BattlePlayerMgr)
	um.battlePlayerList = make(map[uint32]*BattlePlayer)
	um.lock = sync.RWMutex{}

	return
}

// 所有对战用户的统一管理器
type BattlePlayerMgr struct {
	battlePlayerList map[uint32]*BattlePlayer
	lock             sync.RWMutex
}

// 用户加入管理器中
func (s *BattlePlayerMgr) SetPlayer(b *BattlePlayer) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.battlePlayerList[b.entityID] = b
	return nil
}

// 获取一个Player
func (s *BattlePlayerMgr) GetAllPlayerIds() []uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	resp := make([]uint32, 0)
	if len(s.battlePlayerList) <= 0 {
		return resp
	}

	for _, v := range s.battlePlayerList {
		resp = append(resp, v.entityID)
	}

	return resp
}

// 获取一个Player
func (s *BattlePlayerMgr) GetPlayerByID(entityID uint32) (*BattlePlayer, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	entity, ok := s.battlePlayerList[entityID]
	if !ok {
		return nil, errors.New("-->BattlePlayerMgr--GetPlayer--Entity Not In List")
	}
	return entity, nil
}

// 获取Players
func (s *BattlePlayerMgr) GetPlayerList() (map[uint32]*BattlePlayer, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.battlePlayerList, nil
}

// 从用户管理器中剔除用户
func (s *BattlePlayerMgr) DelPlayerByID(entityID uint32) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.battlePlayerList[entityID] = nil
	delete(s.battlePlayerList, entityID)
	return nil
}

// 用户是否存在
func (s *BattlePlayerMgr) IsPlayerExists(entityID uint32) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, isFound := s.battlePlayerList[entityID]
	return isFound
}

// 获取用户管理器中用户的数量
func (s *BattlePlayerMgr) GetPlayerCount() int32 {
	return int32(len(s.battlePlayerList))
}

// 清理玩家对战数据写记录
func (s *BattlePlayerMgr) Clear() {
	if len(s.battlePlayerList) > 0 {
		for _, v := range s.battlePlayerList {
			v.Clear()

			v = nil
		}
	}

	return
}
