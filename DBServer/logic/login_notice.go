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

type _LoginNotice struct{}

var LoginNotice _LoginNotice

func (s *_LoginNotice) Init() {
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncLoginNoticeToGameRequest), reflect.ValueOf(s.SyncLoginNoticeListToGame))
}

func (s *_LoginNotice) SyncLoginNoticeListToGame(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InLoginNoticeToDbRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	coll := new(collection.LoginNotice)
	coll.SetDBConnect(consts.COLLECTION_LOGINNOTICE)
	list := coll.GetDataOfQuery(DBConnect)

	respList := make([]*gmsg.InLoginNoticeInfo, 0)
	if len(list) > 0 {
		now := time.Now().Unix()
		for _, v := range list {
			if v.EndTime < now {
				continue
			}

			respList = append(respList, &gmsg.InLoginNoticeInfo{
				LoginNoticeId: v.LoginNoticeId,
				Title:         v.Title,
				Name:          v.Name,
				Context:       v.Context,
				StartTime:     v.StartTime,
				EndTime:       v.EndTime,
				PlatformLimit: v.PlatformLimit,
				VipLimit:      v.VipLimit,
			})
		}
	}

	resp := &gmsg.InLoginNoticeList{
		List: respList,
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncLoginNoticeToGameResponse), resp, network.ServerType_Game)
	return
}
