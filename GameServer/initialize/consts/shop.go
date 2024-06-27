package consts

const (
	ITEM_TYPE_UNKNOWN = 1 //未知
	ITEM_TYPE_GOLD    = 2 //金币
	ITEM_TYPE_DIAMOND = 3 //钻石
	ITEM_TYPE_VIP     = 4 //VIP
	ITEM_TYPE_CUE     = 5 //球杆
	ITEM_TYPE_DRESS   = 6 //服饰
	ITEM_TYPE_EFFECT  = 7 //效果
	ITEM_TYPE_PROP    = 8 //道具
)

const (
	DefaultRecharge uint32 = iota
	FirstRecharge
	ActivityRecharge
)

const (
	Shop uint32 = iota + 1
	SpecialShop
)
