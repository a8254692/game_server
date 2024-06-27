package consts

const (
	SYSTEM_ID_C8_BATTLE = iota + 1
	SYSTEM_ID_TASK
	SYSTEM_ID_BAG
	SYSTEM_ID_BOX
	SYSTEM_ID_SHOP
	SYSTEM_ID_SPECIAL_SHOP
	SYSTEM_ID_ACTIVITY
	SYSTEM_ID_WELFARE
	SYSTEM_ID_CLUB_TASk
	SYSTEM_ID_REWARD
	SYSTEM_ID_RED_Envelope
	SYSTEM_ID_CLUB_SHOP
	SYSTEM_ID_EMAIL
	SYSTEM_ID_VIP
	SYSTEM_ID_CUE_HAND_BOOK
)

const (
	Buy              uint32 = iota + 1 //购买
	Reward                             //奖励
	CueUpgrade                         //球杆升星
	ComposeItem                        //道具合成
	RedEnvelopeOpen                    //领取红包
	SendRedEnvelope                    //发红包
	OpenBox                            //领取宝箱
	ClubDailySign                      //俱乐部打卡任务
	ClubBattleTask                     //俱乐部对战任务
	ClubConsumeTask                    //俱乐部消费任务
	ClubSupportFunds                   //俱乐部赞助资金任务
	BagGift                            //背包礼盒
	BuyVipGift                         //购买VIP礼包
	MagicBoxReward                     //神秘宝箱
	AchievementLvReward
	TaskDayReward
	TaskWeekReward
	ChangeSex       // 切换性别消耗
	FreeShopRefresh //免费商店刷新消耗
	FreeShopBuy
	GiveGifts
	LoginReward
	ActivateReward
	GM = 99 //GM命令
)

const (
	RES_TYPE_INCR = 1 //增加
	RES_TYPE_DECR = 2 //减少
)
