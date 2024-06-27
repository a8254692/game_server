package logic

import (
	"BilliardServer/GameServer/initialize/vars"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"reflect"
	"time"
)

type _LoginNotice struct {
	List []vars.LoginNotice
}

var LoginNotice _LoginNotice

func (s *_LoginNotice) Init() {
	s.List = make([]vars.LoginNotice, 0)

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncLoginNoticeToGameResponse), reflect.ValueOf(s.SyncLoginNoticeListFromDb))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_LoginNoticeOtherToGameSync), reflect.ValueOf(s.AdminChangeLoginNoticeList))

	time.AfterFunc(time.Millisecond*1000, s.SyncLoginNoticeListToDb)
}

func (s *_LoginNotice) SyncLoginNoticeListFromDb(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InLoginNoticeList{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_LoginNotice--SyncLoginNoticeListFromDb--msgEV.Unmarshal(req) err:", err)
		return
	}

	respList := make([]vars.LoginNotice, 0)
	if len(req.List) > 0 {
		for _, v := range req.List {
			if v.Context == "" {
				continue
			}

			info := vars.LoginNotice{
				LoginNoticeId: v.LoginNoticeId,
				Title:         v.Title,
				Name:          v.Name,
				Context:       v.Context,
				StartTime:     v.StartTime,
				EndTime:       v.EndTime,
				PlatformLimit: v.PlatformLimit,
				VipLimit:      v.VipLimit,
			}

			respList = append(respList, info)
		}
	}

	s.List = respList
	return
}

func (s *_LoginNotice) AdminChangeLoginNoticeList(msgEV *network.MsgBodyEvent) {
	req := &gmsg.InAdminLoginNoticeListSync{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_LoginNotice--AdminChangeLoginNoticeList--msgEV.Unmarshal(req) err:", err)
		return
	}

	log.Info("-->logic--_LoginNotice--AdminChangeLoginNoticeList--req", req)

	toReq := &gmsg.InLoginNoticeToDbRequest{}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncLoginNoticeToGameRequest), toReq, network.ServerType_DB)
	return
}

func (s *_LoginNotice) SyncLoginNoticeListToDb() {
	//开始初始化桌面信息
	req := &gmsg.InLoginNoticeToDbRequest{}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SyncLoginNoticeToGameRequest), req, network.ServerType_DB)
	return
}

func (s *_LoginNotice) GetLoginNoticeListRequest(entityId uint32) []*gmsg.LoginNoticeInfo {
	resp := make([]*gmsg.LoginNoticeInfo, 0)

	if len(s.List) <= 0 {
		return resp
	}

	//tEntityPlayer, err := GetEntityPlayerById(entityId)
	//if err != nil {
	//	log.Waring("-->logic--_LoginNotice--GetLoginNoticeListRequest--GetEntityPlayerById--err--", err)
	//	return resp
	//}

	for _, v := range s.List {
		info := gmsg.LoginNoticeInfo{
			LoginNoticeId: v.LoginNoticeId,
			Title:         v.Title,
			Name:          v.Name,
			Context:       v.Context,
			StartTime:     v.StartTime,
			EndTime:       v.EndTime,
		}
		resp = append(resp, &info)
	}

	log.Info("-->logic--_LoginNotice--GetLoginNoticeListRequest--Resp:", resp)

	return resp
}
