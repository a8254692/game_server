package timer

import (
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"
	"reflect"
	"time"
)

// 时间性能数据
type TimePerfromanceData struct {
	CallCount   int64 //调用次数
	CallTime    int64 //时间
	CallAvgTime int64 //平均调用时间
}

// 定时器对象
type TimerObj struct {
	NameTimer string
	Obj       interface{}   //调用对象
	FunName   string        //函数名字
	Interval  int64         //调用间隔
	Loop      bool          //是否循环
	CallBack  reflect.Value //回调函数
	SetTime   time.Time     //启动的时间

	timerList              map[interface{}]*TimerInfo
	timerChan              chan *TimerObj
	timePerfromanceDataMap map[string]*TimePerfromanceData
}

// 定时器信息
type TimerInfo struct {
	Timers map[string]*time.Timer //时间列表
}

func (this *TimerObj) Init(nameTimer string) {
	//初始化变量
	this.NameTimer = nameTimer
	this.timerList = make(map[interface{}]*TimerInfo)
	this.timePerfromanceDataMap = make(map[string]*TimePerfromanceData)
	this.timerChan = make(chan *TimerObj)
}

// 定时器开启
func (this *TimerObj) Star() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()
	if this.NameTimer == "" {
		log.Error("Timer have not Init")
		return
	}
	//注册函数
	event.Register("OnTimeEvent", reflect.ValueOf(this.OnTimeEvent))
}

// 添加一个定时器
// 间隔时间单位为毫秒
func (this *TimerObj) AddTimer(obj interface{}, funName string, interval int64, loop bool) {
	//判断时间是否为负数
	if interval < 0 {
		log.Error("Timer Error : interval < 0  interval:", interval, "|funName:", funName)
		stack.PrintCallStack()
		interval = 1000
	}

	//创建定时对象
	timerObj := new(TimerObj)

	//检测函数是否存在
	objValue := reflect.ValueOf(obj)
	timerObj.CallBack = objValue.MethodByName(funName)
	if !timerObj.CallBack.IsValid() {
		log.Error("AddTimer error,对象不存在该方法 ", objValue.Elem().Type().Name(), ".", funName)
		return
	}

	//获取该对象定时器信息
	timeInfo := this.timerList[obj]
	if timeInfo == nil {
		timeInfo = new(TimerInfo)
		timeInfo.Timers = make(map[string]*time.Timer)
		this.timerList[obj] = timeInfo
	}

	//判断是否有重复订阅
	timer := timeInfo.Timers[funName]
	if timer != nil {
		log.Error("AddTimer error,重复订阅定时器 ", objValue.Elem().Type().Name(), ".", funName)
		return
	}

	//启动定时器
	timerObj.Loop = loop
	timerObj.Interval = interval
	timerObj.Obj = obj
	timerObj.FunName = funName
	timerObj.SetTime = time.Now()
	timer = time.AfterFunc(time.Millisecond*time.Duration(interval), func() {
		//判断是否提前调用
		timeCure := time.Now().UnixNano()
		callNano := timerObj.SetTime.Add(time.Millisecond * time.Duration(timerObj.Interval)).UnixNano()

		if callNano-timeCure > 0 {
			time.Sleep(time.Duration(callNano - timeCure))
		}

		event.Fire("OnTimeEvent", timerObj)
	})

	timeInfo.Timers[funName] = timer
}

// 删除定时器
func (this *TimerObj) DellTimer(obj interface{}, funName string) {
	//查找对象定时信息
	timeInfo := this.timerList[obj]
	if timeInfo == nil {
		return
	}

	timer := timeInfo.Timers[funName]
	if timer != nil {
		timer.Stop()
	}
	delete(timeInfo.Timers, funName)

	//判断长度
	if len(timeInfo.Timers) == 0 {
		delete(this.timerList, obj)
		return
	}
}

// 删除某个对象所有的定时器
func (this *TimerObj) DellObjAllTimer(obj interface{}) {
	//查找对象定时信息
	timeInfo := this.timerList[obj]
	if timeInfo == nil {
		return
	}

	//结束所有定时器
	for _, timer := range timeInfo.Timers {
		if timer != nil {
			timer.Stop()
		}
	}

	//删除列表
	delete(this.timerList, obj)
}

// 定时器事件函数
func (this *TimerObj) OnTimeEvent(obj *TimerObj) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			stack.PrintCallStack()
		}
	}()

	if obj == nil {
		return
	}

	timeInfo := this.timerList[obj.Obj]
	if timeInfo == nil {
		return
	}
	timer := timeInfo.Timers[obj.FunName]
	if timer == nil {
		return
	}

	//先删除定时器
	this.DellTimer(obj.Obj, obj.FunName)

	timeCure := time.Now().UnixNano()

	//记录最后调用的定时器
	event.SetLastTimeEvent(obj.FunName)
	//调用函数
	obj.CallBack.Call(nil)
	callTime := time.Now().UnixNano() - timeCure

	if this.timePerfromanceDataMap[obj.FunName] != nil {
		data := this.timePerfromanceDataMap[obj.FunName]
		data.CallCount++
		data.CallTime += callTime
		data.CallAvgTime = data.CallTime / data.CallCount / 1e6
	} else {
		data := new(TimePerfromanceData)
		data.CallCount++
		data.CallTime += callTime
		data.CallAvgTime = data.CallTime / data.CallCount / 1e6
		this.timePerfromanceDataMap[obj.FunName] = data
	}

	//判断是否循环需要继续调用
	if obj.Loop {
		this.AddTimer(obj.Obj, obj.FunName, obj.Interval, obj.Loop)
	}
}

// 获取当前到特定时间的时间差（单位是毫秒）
func (this *TimerObj) GetMsFromNow(year, month, day, hour, min, sec int) int64 {
	now := time.Now()
	data := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
	return (data.Unix() - now.Unix()) * 1000
}

// 打印定时性能信息
func (this *TimerObj) PrintPerfromance() {
	log.Info("-----------------TimePerfromance-----------------")
	for funcName, data := range this.timePerfromanceDataMap {
		log.Info(funcName, ", Count = ", data.CallCount, ", Time = ", data.CallTime, ", Avg = ", data.CallAvgTime)
	}
	log.Info("-------------------------------------------------")
}

// 获取性能信息
func (this *TimerObj) GetPerfromance() map[string]*TimePerfromanceData {
	return this.timePerfromanceDataMap
}
