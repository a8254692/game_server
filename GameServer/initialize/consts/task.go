package consts

/*
*
俱乐部任务id
*/
const ClubTaskID = 131

const ProgressKey = 100

// 活跃度配置10
const TaskProgressValue = 10

// 1为日进度，2为周进度
const (
	DayProgress = iota + 1
	WeekProgress
)

const (
	TaskUnState = 0
	TaskState   = 1
)

const (
	TaskUnStateReward = 0
	TaskStateReward   = 1
)

/*
类型0：无参数条件
类型1：达到条件。参数数量，大于等于完成
类型2：累计条件。参数数量，大于等于完成
类型3：每日达到类型，参数等于，数量大于等于完成
类型4：排名类，第一个参数是排行类型，相等完成。后面参数小于等于完成，数量累计
类型5：开服天数达到等于
类型6：一人完成，所有人都完成。完成方式和3相同
*/

const (
	Conditional_0 = iota
	Conditional_1
	Conditional_2
	Conditional_3
	Conditional_4
	Conditional_5
	Conditional_6
)
