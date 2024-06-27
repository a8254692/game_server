package logic

import (
	"math/rand"
	"sync"
	"time"
)

/***
 *@disc: 红包逻辑
 *@author: lsj
 *@date: 2023/11/9
 */

type RedPack struct {
	lock            sync.RWMutex
	Id              string
	Num             int   // 红包个数
	NumDelivered    int   // 已拆包数量
	Amount          int   // 红包金额
	AmountDelivered int   // 已发放金额
	SendTime        int64 // 发送时间
}

func NewRedPack(id string, num int, amount int, sendTime int64) *RedPack {
	return &RedPack{Id: id, Num: num, Amount: amount, NumDelivered: 0, AmountDelivered: 0, SendTime: sendTime}
}

func (pack *RedPack) OpenRedPack() (int, int, int) {
	pack.lock.Lock()
	defer pack.lock.Unlock()

	if pack.NumDelivered == pack.Num {
		return 0, 0, 0
	}

	// 最后一个红包直接返回
	if pack.NumDelivered == pack.Num-1 {
		amount := pack.Amount - pack.AmountDelivered
		pack.NumDelivered += 1
		pack.AmountDelivered = pack.Amount
		return amount, pack.NumDelivered, pack.AmountDelivered
	}
	// 动态计算红包平均值
	avg := (pack.Amount - pack.AmountDelivered) / (pack.Num - pack.NumDelivered)
	// 随机计算红包金额(1到2倍均值)
	rand.Seed(time.Now().UnixNano())
	calAmount := rand.Intn(2 * avg)
	// 红包金额最少1分
	if calAmount == 0 {
		calAmount = 1
	}
	// 保证后续每个人至少有1分
	if (pack.Amount - pack.AmountDelivered - calAmount) < (pack.Num - pack.NumDelivered - 1) {
		calAmount = pack.Amount - pack.AmountDelivered - (pack.Num - pack.NumDelivered - 1)
	}

	pack.NumDelivered += 1
	pack.AmountDelivered += calAmount
	return calAmount, pack.NumDelivered, pack.AmountDelivered
}
