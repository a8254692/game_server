package timer

var TimerSign struct {
	timerName    string
	timerPointer *TimerObj
}

// 初始化名称
func Init(nameTimer string) {
	//初始化变量
	TimerSign.timerPointer = new(TimerObj)
	TimerSign.timerName = nameTimer
	TimerSign.timerPointer.Init(nameTimer)
}

// 开始
func Star() {
	TimerSign.timerPointer.Star()
}

// 添加一个定时器,间隔时间单位为毫秒
func AddTimer(obj interface{}, funName string, interval int64, loop bool) {
	TimerSign.timerPointer.AddTimer(obj, funName, interval, loop)
}

// 删除计时器
func DellTimer(obj interface{}, funName string) {
	TimerSign.timerPointer.DellTimer(obj, funName)
}

// 删除所有计时器
func DellObjAllTimer(obj interface{}) {
	TimerSign.timerPointer.DellObjAllTimer(obj)
}

// 打印定时器性能信息
func PrintPerfromance() {
	TimerSign.timerPointer.PrintPerfromance()
}

// 获得计性能信息
func GetPerfromance() map[string]*TimePerfromanceData {
	return TimerSign.timerPointer.GetPerfromance()
}
