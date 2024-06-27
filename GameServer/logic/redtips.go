package logic

type RedTips struct {
}

var RedTipsM RedTips

func (this *RedTips) Init() {
	// 相关事件注册
	//event.OnNet(msg.Player_RedTipsRequest, reflect.ValueOf(this.onRedTipsInit))
}

//func (this *RedTips) onRedTipsInit(msgEV *network.MsgBodyEvent) {
//	msgBody := &gmsg.RedTipsRequest{}
//	err := msgEV.Unmarshal(msgBody)
//	if err != nil {
//		return
//	}
//	tEntity := UnitEntity.EmPlayer.GetEntityByID(*msgBody.EntityId)
//	if tEntity == nil {
//		return
//	}
//	tEntityPlayer := tEntity.(*entity.EntityPlayer)
//
//	msgReponse := &gmsg.RedTipsResponse{}
//	msgReponse.EntityId = msgBody.EntityId
//	for _, temp := range tEntityPlayer.RedTipsList {
//		item := new(msg.RedTipItem)
//		item.RedType = proto.Uint32(temp.RedType)
//		item.RedName = proto.String(temp.Name)
//		item.State = proto.Uint32(temp.State)
//		msgReponse.RedTipsList = append(msgReponse.RedTipsList, item)
//	}
//	ConnectManager.SendMsgPbToGateBroadCast(msg.Player_RedTipsResponse, msgReponse, []uint32{*msgBody.EntityId})
//}
