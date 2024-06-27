// 包说明：管理多个定时器组成的map定时器，即MpTimers
package xtimer

import (
	"errors"
	"sync"
	systime "time"
)

// 定义全局错误变量
var (
	TIMER_EXISTS     = errors.New("timer exists!")
	TIMER_NOT_EXISTS = errors.New("timer not exists!")
	TIMER_NEW_ERROR  = errors.New("timer new error!")
)

// 定义Fn的回调方法，可通过合适的回调方法进行通知
type FnCallBack func(info interface{})

// 定时器组的抽象
type MpTimersIntf interface {
	//开启某个特定的定时器
	StartOneTimer(timeStr string, timeout systime.Duration, fn FnCallBack, info interface{}) error
	//关闭某定时器
	CloseOneTimer(timeStr string) error
	//清理掉所有的定时器
	ClearAll() error
}

// ==============================定时器组接口的实现
type MpTimers struct {
	timers map[string]*timeElem
	mtx    sync.RWMutex
}

// 创建一个定时器管理器对象
func NewMpTimers() MpTimersIntf {
	obj := new(MpTimers)
	obj.timers = make(map[string]*timeElem)
	obj.mtx = sync.RWMutex{}
	return obj
}

// 开启某个特定的定时器
func (mps *MpTimers) StartOneTimer(timeStr string, timeout systime.Duration, fn FnCallBack, info interface{}) error {
	mps.mtx.Lock()
	defer mps.mtx.Unlock()
	//创建一个定时器并加入到map中
	timer := mps.newAndStartTimer(timeStr, timeout, fn, info)
	if timer == nil {
		return TIMER_NEW_ERROR
	}
	return nil
}

// 关闭某个特定的定时器
func (mps *MpTimers) CloseOneTimer(timeStr string) error {
	mps.mtx.Lock()
	defer mps.mtx.Unlock()
	timer, ok := mps.timers[timeStr]
	if ok {
		timer.close()
		//删除该定时器
		delete(mps.timers, timeStr)
		return nil
	}
	return TIMER_NOT_EXISTS
}

// 清理掉所有的定时器
func (mps *MpTimers) ClearAll() error {
	mps.mtx.Lock()
	defer mps.mtx.Unlock()
	for k, v := range mps.timers {
		v.close()
		delete(mps.timers, k)
	}
	return nil
}

// 内部使用，创建一个timer，开始计时并加入到map中
func (mps *MpTimers) newAndStartTimer(timeStr string, timeout systime.Duration, fn FnCallBack, info interface{}) *timeElem {
	cct := new(timeElem)
	cct.initWithFn(timeout, fn, info)
	cct.startTimer(timeStr, mps)
	mps.timers[timeStr] = cct
	return cct
}

// 一个定时器结构体
type timeElem struct {
	fn      FnCallBack  //通知的方式二：方法回调
	info    interface{} //info是配合fn回调使用，也就是这个定时器在时间到的时候传入回调方法的参数
	tm      *systime.Timer
	timeout systime.Duration
	isclose bool
}

// 内部使用，初始化一个定时器的时候调用！
func (t *timeElem) initWithFn(timeout systime.Duration, fn FnCallBack, info interface{}) {
	t.fn = fn
	t.info = info
	t.timeout = timeout
	t.isclose = false
}

func (t *timeElem) startTimer(timeStr string, mps *MpTimers) {
	t.tm = systime.NewTimer(t.timeout)
	t.isclose = false
	go func() {
		select {
		case <-t.tm.C:
			//mps.mtx.Lock()
			//defer mps.mtx.Unlock()
			//delete(mps.timers, timeStr)
			//状态置为关闭
			if t.isclose == true {
				return
			} else {
				t.isclose = true
			}

			//时间到了，通知关心（持有）方
			if t.fn != nil {
				t.fn(t.info)
			}
		}
	}()
}

// 关闭一个定时器
func (t *timeElem) close() {
	if t.isclose == false {
		//todo:Q:isclose的值有可能并发的修改
		t.isclose = true
		//why zero
		t.tm.Reset(0 * systime.Millisecond)
	}
}
