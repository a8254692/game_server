package consts

// 成就类型
// 1:通用
// 2：对局
// 3：闯关
// 4：全部
const (
	AchievementCommon = iota + 1
	AchievementBattle
	AchievementBreakThrough
	AchievementAll
)

const MaxAchievementLV uint32 = 30
