package event

import (
	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// 事件管理器
type EventManager struct {
	NameEM    string
	HandleMap map[string]*reflect.Value
	EventChan chan Event

	mEventPerfromanceMap map[string]*EventPerfromanceData
	// mLastEvent           interface{} //最后调用的事件
	// mLastEventData       interface{} //最后调用事件的参数
	mLastEventKey string
	stopFunc      func()
	LastTimeEvent string
}

// 时间管理器性能
type EventPerfromanceData struct {
	CallCount   int64 //调用次数
	CallTime    int64 //时间
	CallAvgTime int64 //平均调用时间
}

// 初始化事件管理器
func (this *EventManager) Init(nameEm string, enentChanLength int, stopFunc func()) {
	this.NameEM = nameEm
	this.HandleMap = make(map[string]*reflect.Value)
	this.EventChan = make(chan Event, enentChanLength)
	this.stopFunc = stopFunc

	playerMsgRouter.MsgMap = make(map[string]reflect.Type)
	this.mEventPerfromanceMap = make(map[string]*EventPerfromanceData)

	this.Register("PrintEventMsg", reflect.ValueOf(this.PrintEventMsg))
	log.Info("-->事件管理器初始化完毕：" + this.NameEM)
}

// 运行事件事件管理
func (this *EventManager) Run() (result bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			result = false
		}
	}()

	for ev := range this.EventChan {
		this.OnEvent(ev)
	}
	result = true
	return
}

func (this *EventManager) OnEvent(ev Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(this.NameEM+"-->EventManager error:", r, " mLastEventKey:", this.mLastEventKey)
			stack.PrintCallStack()
		}
	}()

	fun := this.HandleMap[ev.HadleFunc]
	if fun == nil {
		log.Error("事件派发失败,HadleFunc = ", ev.HadleFunc, ev.FireFile)
		return
	}

	val := make([]reflect.Value, 1)
	val[0] = reflect.ValueOf(ev.Data)
	this.mLastEventKey = ev.HadleFunc
	//this.mLastEvent = ev
	//this.mLastEventData = ev.Data
	timeCure := time.Now().UnixNano()
	//调用
	fun.Call(val)
	//go fun.Call(val)

	callTime := time.Now().UnixNano() - timeCure

	if this.mEventPerfromanceMap[ev.HadleFunc] != nil {
		data := this.mEventPerfromanceMap[ev.HadleFunc]
		data.CallCount++
		data.CallTime += callTime
		data.CallAvgTime = data.CallTime / data.CallCount / 1e6
	} else {
		data := new(EventPerfromanceData)
		data.CallCount++
		data.CallTime += callTime
		data.CallAvgTime = data.CallTime / data.CallCount / 1e6
		this.mEventPerfromanceMap[ev.HadleFunc] = data
	}
}

// 关闭事件运行
func (this *EventManager) Stop() {
	close(this.EventChan)

	//打印事件性能分析
}

// 是否注册了此事件
func (this *EventManager) IsExist(eventName string) bool {
	funVal := this.HandleMap[eventName]
	if funVal == nil {
		return false
	}
	return true
}

// 注册一个事件响应函数
func (this *EventManager) Register(funcName string, funVal reflect.Value) {
	if this.HandleMap[funcName] != nil {
		//获取函数调用文件及调用行数
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		}

		//获取短文件名
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		log.Error("重复点阅事件,HadleFunc = ", funcName, fmt.Sprint(" file:", short, " line:", line))
		return
	}
	this.HandleMap[funcName] = &funVal
}

// 派发事件
func (this *EventManager) Fire(funcName string, data interface{}) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()
	//如果事件没有注册没直接返回
	if this.HandleMap[funcName] == nil {
		return false
	}
	//获取函数调用文件及调用行数
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}

	//获取短文件名
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	// 派发超时处理
	// 超时定时器需要手动停止，否则定时结束前不会被回收
	timeout := time.NewTimer(20e9)
	select {
	case this.EventChan <- Event{funcName, data, fmt.Sprint(" file:", short, " line:", line)}:
		timeout.Stop()
		return true
	case <-timeout.C:
		log.Error("事件派发缓冲已满，抛弃事件, funcName = ", funcName, ",", data, "----------------")
		log.Error("最后的事件名称：", this.mLastEventKey)
		log.Error("最后定时器事件:", this.LastTimeEvent)
		log.Error("---------------------------------------------------------------")
		this.Stop()
		if this.stopFunc != nil {
			this.stopFunc()
		}
		log.Error("事件派发满异常退出")
		os.Exit(1)
		return false
	}
}

// 打印最后的事件
func (this *EventManager) PrintEventMsg(commond *string) {
	log.Info("--------------------------------------------------")
	log.Waring(this.mLastEventKey)
	log.Info("--------------------------------------------------")
}
func (this *EventManager) GetMsgTileList() []uint32 {
	var tileArgs []uint32 = make([]uint32, 0)
	var tileNameArgs []string
	for name, _ := range this.HandleMap {
		if strings.Contains(name, "Msg_") {
			tileNameArgs = strings.Split(name, "_")
			msgTile, err := strconv.Atoi(tileNameArgs[1])
			if err == nil {
				tileArgs = append(tileArgs, uint32(msgTile))
			}

		}
	}
	return tileArgs
}
