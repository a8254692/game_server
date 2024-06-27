package logic

import (
	"BilliardServer/OtherServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"encoding/json"
	"strconv"
	"strings"
)

type _AdminMsg struct {
	startMsgChan chan bool
}

var AdminMsg _AdminMsg

func (s *_AdminMsg) Init() {
	s.startMsgChan = make(chan bool)

	GoFunc(s.StartReceiveMqMsg)
}

func (s *_AdminMsg) StartReceiveMqMsg() {
	if Mq == nil {
		log.Waring("-->logic--_AdminMsg--StartReceiveMqMsg--Mq == nil")
		return
	}

	msgChan, err := Mq.ConsumeSimple()
	if err != nil {
		log.Waring("-->logic--_AdminMsg--StartReceiveMqMsg--Mq.ConsumeSimple--err", err)
		return
	}

	for d := range msgChan {
		if len(d.Body) > 0 {
			s.distributeMqMsg(d.Body)
			_ = d.Ack(true)
		}
	}

	<-s.startMsgChan
}

type UserChangeMsg struct {
	MsgType uint32 `json:"msgType"` //消息类型
	Data    string `json:"data"`    //消息类型
}

type SendUserStatusChangeRequest struct {
	EntityID uint32 `json:"entityID"` //用户id
	OType    uint32 `json:"oType"`    //操作类型 1下线 2禁言 3封账号 4封ip
}

type SendUserAttrChangeRequest struct {
	EntityID uint32 `json:"entityID"` //用户id
	OType    uint32 `json:"oType"`    //操作类型 1加钱 2加钻石 3加经验
	Num      uint32 `json:"num"`      //数量
}

type MarqueeMsgSync struct {
	MarqueeType uint32 `json:"marqueeType"`
	Context     string `json:"context"`
}

type SendEmailRequest struct {
	EntityId uint32 `json:"entityId"` //接收人的id
	Title    string `json:"title"`    //邮件标题
	Context  string `json:"context"`  //邮件内容
	Annex    string `json:"annex"`    //附件(`|`分割物品 `;`分割属性 id;数量;时效 例:60000006;1|40000001;2)
}

type MqActivityData struct {
	ActivityId string `json:"activity_id"` //活动唯一ID
	IsRelease  bool   `json:"is_release"`  //是否发布
}

type MqLoginNoticeData struct {
	LoginNoticeId string `json:"login_notice_id"` //登录公告唯一ID
}

type MqPointsMallData struct {
	PointsMallId string `json:"points_mall_id"` //活动唯一ID
	IsRelease    bool   `json:"is_release"`     //是否发布
}

func (s *_AdminMsg) distributeMqMsg(msg []byte) {
	msgData := UserChangeMsg{}
	err := json.Unmarshal(msg, &msgData)
	//解析失败会报错，如json字符串格式不对，缺"号，缺}等。
	if err != nil {
		log.Waring("-->logic--_AdminMsg--distributeMqMsg-- json.Unmarshal--err", err, string(msg))
		return
	}
	if msgData.Data == "" {
		log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.Data == nil")
		return
	}

	switch msgData.MsgType {
	case uint32(consts.USER_STATUS_CHANGE):
		data := SendUserStatusChangeRequest{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--USER_STATUS_CHANGE", msgData)
			return
		}

		if data.EntityID <= 0 || data.OType <= 0 {
			return
		}

		req := &gmsg.InEditUserStatusRequest{
			EntityID: data.EntityID,
			OType:    data.OType,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Edit_User_Status_Request), req, network.ServerType_Game)

	case uint32(consts.USER_ATTR_CHANGE):
		data := SendUserAttrChangeRequest{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--USER_ATTR_CHANGE", msgData)
			return
		}

		if data.EntityID <= 0 || data.OType <= 0 || data.Num <= 0 {
			return
		}

		req := &gmsg.InEditUserAttrRequest{
			EntityID: data.EntityID,
			OType:    data.OType,
			Param:    data.Num,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Edit_User_Attr_Request), req, network.ServerType_Game)

	case uint32(consts.ADMIN_MARQUEE_MSG):
		data := MarqueeMsgSync{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--USER_ATTR_CHANGE", msgData)
			return
		}

		if data.Context == "" {
			return
		}

		req := &gmsg.InMarqueeMsgSync{
			MarqueeType: data.MarqueeType,
			Context:     data.Context,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SendMarqueeMsgSync), req, network.ServerType_Game)

	case uint32(consts.ADMIN_SEND_EMAIL):
		data := SendEmailRequest{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--ADMIN_SEND_EMAIL", msgData)
			return
		}

		if data.EntityId <= 0 || data.Title == "" {
			return
		}

		rList := make([]*gmsg.InRewardInfo, 0)
		if data.Annex != "" {
			rList = s.SplitRewardStr(data.Annex)
		}

		emailInfo := &gmsg.InEmail{
			RewardList: rList,
			Tittle:     data.Title,
			Content:    data.Context,
		}
		req := &gmsg.InAddEmailRequest{
			EntityID: data.EntityId,
			Email:    emailInfo,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_SendEmailRequest), req, network.ServerType_Game)

	case uint32(consts.ADMIN_SYNC_ACTIVITY):
		data := MqActivityData{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--ADMIN_SYNC_ACTIVITY", msgData)
			return
		}

		log.Info("-->logic--_AdminMsg--distributeMqMsg--ADMIN_SYNC_ACTIVITY--", data)

		req := &gmsg.InAdminActivityListSync{
			ActivityId: data.ActivityId,
			IsRelease:  data.IsRelease,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_ActivityOtherToGameSync), req, network.ServerType_Game)

	case uint32(consts.ADMIN_SYNC_LOGIN_NOTICE):
		data := MqLoginNoticeData{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--ADMIN_SYNC_LOGIN_NOTICE", msgData)
			return
		}

		log.Info("-->logic--_AdminMsg--distributeMqMsg--ADMIN_SYNC_LOGIN_NOTICE--", data)

		req := &gmsg.InAdminLoginNoticeListSync{
			ActivityId: data.LoginNoticeId,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_LoginNoticeOtherToGameSync), req, network.ServerType_Game)

	case uint32(consts.ADMIN_SYNC_POINTS_MALL):
		data := MqPointsMallData{}
		err := json.Unmarshal([]byte(msgData.Data), &data)
		if err != nil {
			log.Waring("-->logic--_AdminMsg--distributeMqMsg--msgData.MsgType--ADMIN_SYNC_POINTS_MALL", msgData)
			return
		}

		log.Info("-->logic--_AdminMsg--distributeMqMsg--ADMIN_SYNC_POINTS_MALL--", data)

		req := &gmsg.InAdminPointsShopListSync{
			PointsMallId: data.PointsMallId,
		}
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_PointsShopOtherToGameSync), req, network.ServerType_Game)
	}

	return
}

func (s *_AdminMsg) SplitRewardStr(c string) []*gmsg.InRewardInfo {
	if c == "" {
		return nil
	}

	resp := make([]*gmsg.InRewardInfo, 0)
	fs := strings.Split(c, "|")
	if len(fs) > 0 {
		for _, v := range fs {
			attrArr := strings.Split(v, ";")
			if len(attrArr) > 0 {
				var itemId uint32
				var num uint32
				var expTime uint32

				if len(attrArr) > 0 {
					itemIdInt, err := strconv.Atoi(attrArr[0])
					if err != nil {
						continue
					}
					itemId = uint32(itemIdInt)
				}

				if len(attrArr) > 1 {
					numInt, err := strconv.Atoi(attrArr[1])
					if err != nil {
						continue
					}
					num = uint32(numInt)
				}

				if len(attrArr) > 2 {
					expTimeInt, err := strconv.Atoi(attrArr[2])
					if err != nil {
						continue
					}
					expTime = uint32(expTimeInt)
				}

				resp = append(resp, &gmsg.InRewardInfo{
					ItemTableId:  itemId,
					Num:          num,
					ExpireTimeId: expTime,
				})
			}
		}
	}

	return resp
}
