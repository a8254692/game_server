package consts

// 0普通成员,1精英，2副部长,3部长
const (
	General       uint32 = iota // 普通成员
	Elite                       //精英
	Second_Master               //副部长
	Master                      //部长
)

const MaxReqClub = 5

const RedEnvelopeCommission uint32 = 100

const MaxClubRateNum = 25
const MaxPalaceClubNum = 10
const MaxProfitNum = 50
const PromoteEliteActive uint32 = 900

const (
	BattleTask   = 156
	DailyTask    = 157
	ConsumeTask  = 158
	SupportFunds = 159
)

const DeductStone = 10

const (
	ClubRateE uint32 = iota + 1
	ClubRateD
	ClubRateC
	ClubRateB
	ClubRateA
	ClubRateS
	ClubRateSPlus
)

// 俱乐部评级排行升降枚举
const (
	ClubKeepRate uint32 = iota
	ClubUpgradeRate
	ClubReduceRate
)

// 俱乐部排名奖励
const ClubItemRewardRank = 10
const ClubRateSTotal = 10
