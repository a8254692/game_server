package logic

import (
	"reflect"

	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
)

type _AverageLink struct {
}

var AverageLink _AverageLink

func (this *_AverageLink) Init() {
	//注册逻辑业务事件
	event.On("Msg_MultiNinjaPointWarEnemyTeam", reflect.ValueOf(TeamRequest))
}
func TeamRequest(msgEV *network.MsgBodyEvent) {

}
