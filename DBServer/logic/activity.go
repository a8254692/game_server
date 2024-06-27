package logic

import (
	"BilliardServer/DBServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/db/collection"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
	"reflect"
	"time"
)

type _Activity struct{}

var Activity _Activity

func (s *_Activity) Init() {
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncActivityListToGameRequest), reflect.ValueOf(s.SyncActivityListToGame))
}

func (s *_Activity) SyncActivityListToGame(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InActivityListToDbRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	coll := new(collection.Activity)
	coll.SetDBConnect(consts.COLLECTION_Activity)
	list := coll.GetDataAfterTime(DBConnect)

	respList := make([]*gmsg.InActivityInfo, 0)
	if len(list) > 0 {
		now := time.Now().Unix()
		for _, v := range list {
			if v.TimeType == 1 {
				if v.EndTime < now {
					continue
				}
			}

			respList = append(respList, &gmsg.InActivityInfo{
				ActivityId:    v.ActivityId,
				TimeType:      v.TimeType,
				StartTime:     v.StartTime,
				EndTime:       v.EndTime,
				AType:         v.AType,
				SubType:       v.SubType,
				ActivityName:  v.ActivityName,
				PlatformLimit: v.PlatformLimit,
				VipLimit:      v.VipLimit,
				Config:        v.Config,
			})
		}
	}

	resp := &gmsg.InActivityList{
		IsUpdate: req.IsUpdate,
		List:     respList,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncActivityListToGameResponse), resp, network.ServerType_Game)
	return
}
