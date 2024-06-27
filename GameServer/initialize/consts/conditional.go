package consts

// 根据条件表定义
const (
	AccLV                  uint32 = iota + 100
	AddFriend                     //101
	GiveFriendshipPoint           //102
	BuyByDiscountStore            //103
	GetCueXY                      //104
	BuyTimesInStore               //105
	BattleConnectingCue           //106
	FansNum                       //107
	CueNum                        //108
	CharmNum                      //109
	VipLV                         //110
	TotalGold                     //111
	BattleConnectingWin           //112
	BattleTotalFillingBall        //113 和116有歧义
	BattleOneCleaning             //114
	BattleOneCueXGoal             //115
	BattleTotalGoal               //116
	BattleWinTimes                //117
	GiveGoldFriendDays            //118
	TaskTimes                     //119
	SignDaysConnecting            //120
	BreakThroughX                 //121
	BattleBaseScoreTimes          //122
	PlayerLV                      //123
	BattleTotalUnbill             //124
	BattleCombinationGoal         //125
	BattleTotalBorrowGoal         //126
	TotalBreakThrough             //127
	XYCue                         //128
	LotteryDrawTimes              //129
	WatchVideoTimes               //130
	JoinOrCreateClub              //131
	FirstWinX                     //132
	RechargeAmount                //133
	LuckyHitTimes                 //134
	ExtremeChallenge              //135
	RealPlayerBattle              //136
	FriendBattleTimes             //137
	BuyCommodityTimes             //138
	GiveFriendGold                //139
	BreakThroughMode              //140
	WatchVideoGetDiamond          //141
	SignInDayTimes                //142
	OneCleaningTimes              //143
	TotalGoals                    //144
	ContinuousWins                //145
	CharmRating                   //146
	TotalMyFriends                //147,粉丝数
	TotalRechargeAmount           //148
	ApprenticeNum                 //149
	BecomingApprentice            //150
	BecomingCompanion             //151
	TotalPopularityValue          //152
	ATop10Club                    //153
	SClub                         //154
	SMaxClub                      //155
)

const (
	XYCueS   uint32 = 167
	XYCueSs  uint32 = 168
	XYCueSss uint32 = 169
)

type ConditionData struct {
	ConditionalID uint32
	Progress      uint32 //
	IsTotal       bool   //默认false，更新增量；true，更新全量
}
