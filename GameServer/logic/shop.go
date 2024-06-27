package logic

import (
	"BilliardServer/Common/table"
	"BilliardServer/GameServer/initialize/vars"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func NewShop() (s *Shop) {
	s = new(Shop)
	s.setShopList()

	return
}

type Shop struct {
	list map[string]*table.ShopCfg
}

func (s *Shop) setShopList() {
	list := Table.GetAllShopCfg()
	s.list = list
}

// 生成订单号
func (s *Shop) getOrderID(entityID uint32, itemType uint32) (string, error) {
	if entityID <= 0 || itemType < 0 {
		return "", errors.New("-->logic--Shop--getOrderID--Param Is Empty")
	}

	pre, ok := vars.SHOP_ITEM_TYPE_MAP[itemType]
	if !ok {
		return "", errors.New("-->logic--Shop--getOrderID--Pre Is Empty")
	}
	now := time.Now().Unix()
	randNum := rand.Int31n(9999) + 1000
	return fmt.Sprintf("%s-%d-%d%d", pre, entityID, now, randNum), nil
}

// 获取商城商品列表
func (s *Shop) GetShopList() (map[string]*table.ShopCfg, error) {
	if s.list == nil {
		return nil, errors.New("-->logic--Shop--GetShopList--s.list Is nil")
	}

	return s.list, nil
}

// 获取商城商品详情
func (s *Shop) GetItemByID(itemID string) (*table.ShopCfg, error) {
	if s.list == nil {
		return nil, errors.New("-->logic--Shop--GetShopList--s.list Is nil")
	}
	if itemID == "" {
		return nil, errors.New("-->logic--Shop--GetShopList--itemID Is empty")
	}
	return s.list[itemID], nil
}

// 新增订单记录
func (s *Shop) AddOrderRecord() (int64, error) {

	return 0, nil
}
