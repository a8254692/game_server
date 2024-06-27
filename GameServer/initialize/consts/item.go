package consts

import "fmt"

var (
	CueQualityDigit = 100000
	CueStarDigit    = 1000
	CueKeyDigit     = 1

	Gift           = uint32(1)
	ReceivingGifts = uint32(2)
)

const (
	Gold        = 60000001 //金币
	Diamond     = 60000002 //钻石
	ClubGold    = 60000003 //俱乐部币
	Exchange    = 60000004 //兑换卷
	LvExp       = 60000005
	VipLvExp    = 60000006
	ShopScore   = 60000007 //商城积分
	PeakRankExp = 60000008 //排位经验
	Popularity  = 60000009 //人气值
)

// 1：球杆,2：服装,3：特效,4：道具,5：装扮  6：属性道具，7.宝箱
const (
	Cue = iota + 1
	Dress
	Effect
	Item
	Clothing
	PropertyItem
	Box
)

// 1球桌 2击球效果 3进球效果 4主球
const (
	Effect_1 = iota + 1
	Effect_2
	Effect_3
	Effect_4
)

// 1头像，2头像框，3气泡，4倒计时，5表情，6魔法表情
const (
	Clothing_1 = iota + 1
	Clothing_2
	Clothing_3
	Clothing_4
	Clothing_5
	Clothing_6
)

// 0默认道具，1普通道具，2碎片，3强化道具，4表情，5礼包，6游戏内道具,7礼物道具
const (
	Item_0 = iota
	Item_1
	Item_2
	Item_3
	Item_4
	Item_5
	Item_6
	Item_7
)

const (
	ItemNoUse = iota + uint32(0)
	ItemUse
)

var BaseTableInfo = map[string]string{
	fmt.Sprintf("%d_0", Cue):                   "CueTableId",
	fmt.Sprintf("%d_0", Dress):                 "PlayerDress",
	fmt.Sprintf("%d_%d", Effect, Effect_1):     "TableCloth",
	fmt.Sprintf("%d_%d", Effect, Effect_2):     "BattingEffect",
	fmt.Sprintf("%d_%d", Effect, Effect_3):     "GoalInEffect",
	fmt.Sprintf("%d_%d", Effect, Effect_4):     "CueBall",
	fmt.Sprintf("%d_%d", Clothing, Clothing_1): "PlayerIcon",
	fmt.Sprintf("%d_%d", Clothing, Clothing_2): "IconFrame",
	fmt.Sprintf("%d_%d", Clothing, Clothing_3): "ClothingBubble",
	fmt.Sprintf("%d_%d", Clothing, Clothing_4): "ClothingCountDown",
}

// 球杆品质
const (
	CueQuality1 = iota + 1
	CueQuality2
	CueQuality3
	CueQuality4
	CueQuality5
	CueQuality6
)
