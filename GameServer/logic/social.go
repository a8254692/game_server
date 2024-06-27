package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"reflect"
	"strconv"
	"sync"
	"time"
)

/***
 *@disc: 社交
 *@author: lsj
 *@date: 2023/9/22
 */

type _Social struct {
	WeekRank       []*gmsg.PopRanks
	TotalRank      []*gmsg.PopRanks
	NearPlayerData []*gmsg.NearPlayerList
	Total          int
}

var SocialManager _Social

var SocialMutex sync.Mutex

func (c *_Social) Init() {
	c.WeekRank = make([]*gmsg.PopRanks, 0)
	c.TotalRank = make([]*gmsg.PopRanks, 0)
	c.NearPlayerData = make([]*gmsg.NearPlayerList, 0)
	c.Total = 21

	timer.AddTimer(c, "LoadNearPlayerList", 300*1000, true)

	event.OnNet(gmsg.MsgTile_Hall_PopularityRankRequest, reflect.ValueOf(c.OnPopularityRankRequest))
	event.OnNet(gmsg.MsgTile_Hall_MyFriendsListRequest, reflect.ValueOf(c.OnMyFriendsListRequest))
	event.OnNet(gmsg.MsgTile_Hall_MyFansListRequest, reflect.ValueOf(c.OnMyFansListRequest))
	event.OnNet(gmsg.MsgTile_Hall_AddMyFriendsRequest, reflect.ValueOf(c.OnAddMyFriendsRequest))
	event.OnNet(gmsg.MsgTile_Hall_AddMyFriendsResponse, reflect.ValueOf(c.OnAddMyFriendsResponse))
	event.OnNet(gmsg.MsgTile_Hall_CancelMyFriendsRequest, reflect.ValueOf(c.OnCancelMyFriendsRequest))
	event.OnNet(gmsg.MsgTile_Hall_NearbyPlayerListRequest, reflect.ValueOf(c.OnNearByPlayerListRequest))
	event.OnNet(gmsg.MsgTile_Hall_NearbyPlayerListResponse, reflect.ValueOf(c.OnNearByPlayerListResponse))
	event.OnNet(gmsg.MsgTile_Hall_SearchPlayerFromIDRequest, reflect.ValueOf(c.OnSearchPlayerFromIDRequest))
	event.OnNet(gmsg.MsgTile_Hall_SearchPlayerFromIDResponse, reflect.ValueOf(c.OnSearchPlayerFromIDResponse))
	event.OnNet(gmsg.MsgTile_Hall_AddGoldToMyFriendsRequest, reflect.ValueOf(c.OnAddGoldToMyFriendsRequest))
	event.OnNet(gmsg.MsgTile_Hall_UpdateFansUnixSec, reflect.ValueOf(c.OnUpdateFansUnixSec))
}

// 获取人气周榜和总榜
func (c *_Social) OnPopularityRankRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PopularityRankRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.PopularityRankResponse{}
	msgResponse.WeekList = c.WeekRank
	msgResponse.TotalList = c.TotalRank

	msgResponse.Code = resp_code.CODE_SUCCESS
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_PopularityRankResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 加载附近的在线20人
func (c *_Social) LoadNearPlayerList() {
	list := make([]*gmsg.NearPlayerList, 0)
	num, count := 0, Entity.EmPlayer.EntityCount
	for _, emPlayer := range Entity.EmPlayer.EntityMap {
		count--
		player := emPlayer.(*entity.EntityPlayer)
		if player.IsRobot && num < c.Total && count-num > c.Total {
			continue
		}
		nearOne := new(gmsg.NearPlayerList)
		stack.SimpleCopyProperties(nearOne, player)
		nearOne.PlayerType = Player.GetBehaviorStatus(player.EntityID)
		list = append(list, nearOne)
		num++
		if num >= c.Total {
			break
		}
	}

	c.NearPlayerData = list
}

// 我的关注列表
func (c *_Social) OnMyFriendsListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.MyFriendsListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.MyFriendsListResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.List = make([]*gmsg.FriendList, 0)
	msgResponse.List = c.getFriendsList(tEntityPlayer)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_MyFriendsListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 我的粉丝列表
func (c *_Social) OnMyFansListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.FansListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.FansListResponse{}

	for _, v := range tEntityPlayer.FansList.List {
		fans := new(gmsg.FansList)
		if playerBase, ok := Player.PlayerBaseList[v.EntityID]; ok {
			stack.SimpleCopyProperties(fans, playerBase)
			msgResponse.List = append(msgResponse.List, fans)
		}
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_MyFansListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 添加关注
func (c *_Social) OnAddMyFriendsRequest(msgEV *network.MsgBodyEvent) {
	SocialMutex.Lock()
	defer SocialMutex.Unlock()
	msgBody := &gmsg.AddMyFriendsRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.AddMyFriendsResponse{}
	msgResponse.Code = resp_code.CODE_ERR
	msgResponse.EntityID = msgBody.EntityID

	if msgBody.EntityID == msgBody.AddEntityID {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddMyFriendsResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if !tEntityPlayer.IsInMyFriends(msgBody.AddEntityID) {
		playerBase, ok := Player.PlayerBaseList[msgBody.AddEntityID]
		if !ok {
			ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_AddMyFriendsRequest, msgBody, network.ServerType_DB)
			return
		}
		resFriend := tEntityPlayer.AddMyFriends(msgBody.AddEntityID)
		tEntityPlayer.SyncEntity(1)

		msgResponse.Code = resp_code.CODE_SUCCESS
		msgResponse.MyFriendsNum = uint32(len(tEntityPlayer.MyFriends))
		msgResponse.AddEntityID = msgBody.AddEntityID

		friend := c.getFriendList(*resFriend)
		stack.SimpleCopyProperties(friend, playerBase)
		msgResponse.AddList = append(msgResponse.AddList, friend)

		c.AddFansList(msgBody.AddEntityID, msgBody.EntityID)
		ChatMgr.InnerSendPrivateChatMsg(uint32(gmsg.MsgType_MtSystem), msgBody.EntityID, msgBody.AddEntityID, consts.PRIVATELY_BUY_NOTICE)
	}
	log.Info("-->OnAddMyFriendsRequest-->end->", msgResponse)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddMyFriendsResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 添加关注db返回
func (c *_Social) OnAddMyFriendsResponse(msgEV *network.MsgBodyEvent) {
	SocialMutex.Lock()
	defer SocialMutex.Unlock()
	msgResponse := &gmsg.AddMyFriendsResponse{}
	if err := msgEV.Unmarshal(msgResponse); err != nil {
		return
	}
	if msgResponse.Code != uint32(0) {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddMyFriendsResponse, msgResponse, []uint32{msgResponse.EntityID})
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgResponse.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	resFriend := tEntityPlayer.AddMyFriends(msgResponse.AddEntityID)
	tEntityPlayer.SyncEntity(1)

	res := msgResponse.AddList[0]
	res.AddTime = resFriend.AddTime
	res.Online = Player.GetGamePlayerOnline(msgResponse.AddEntityID)
	res.IsGiveGold = c.isHaveGiveGold(resFriend.GiveGoldSec)
	msgResponse.AddList[0] = res
	msgResponse.MyFriendsNum = uint32(len(tEntityPlayer.MyFriends))

	log.Info("-->OnAddMyFriendsResponse-->end->", msgResponse)
	c.AddFansList(msgResponse.AddEntityID, msgResponse.EntityID)
	ChatMgr.InnerSendPrivateChatMsg(uint32(gmsg.MsgType_MtSystem), msgResponse.EntityID, msgResponse.AddEntityID, consts.PRIVATELY_BUY_NOTICE)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddMyFriendsResponse, msgResponse, []uint32{msgResponse.EntityID})
	return
}

func (c *_Social) AddFansList(EntityID, AddEntityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		//todo 用户不在线，去处理
		msgBody := &gmsg.AddMyFansRequest{}
		msgBody.EntityID = EntityID
		msgBody.AddEntityID = AddEntityID
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_AddFansFromGameToDB), msgBody, network.ServerType_DB)
		return
	}

	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if tEntityPlayer.IsInFansList(AddEntityID) {
		return
	}

	tEntityPlayer.AddFansList(AddEntityID)
	tEntityPlayer.SyncEntity(1)
	conds := make([]consts.ConditionData, 0)
	conds = append(conds, consts.ConditionData{consts.TotalMyFriends, tEntityPlayer.FansNum, true}, consts.ConditionData{consts.FansNum, tEntityPlayer.FansNum, true})
	ConditionalMr.SyncConditional(EntityID, conds)
	c.AddFansSync(EntityID, AddEntityID)
}

func (c *_Social) getFriendList(f entity.Friend) *gmsg.FriendList {
	friend := new(gmsg.FriendList)
	friend.EntityID = f.EntityID
	friend.AddTime = f.AddTime
	friend.Online = Player.GetGamePlayerOnline(f.EntityID)
	friend.IsGiveGold = c.isHaveGiveGold(f.GiveGoldSec)

	return friend
}

// 取消关注 游戏->db
func (c *_Social) OnCancelMyFriendsRequest(msgEV *network.MsgBodyEvent) {
	SocialMutex.Lock()
	defer SocialMutex.Unlock()
	msgBody := &gmsg.CancelMyFriendsRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.CancelMyFriendsResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = resp_code.CODE_ERR
	msgResponse.DelEntityID = msgBody.DelEntityID

	if msgBody.EntityID == msgBody.DelEntityID {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_CancelMyFriendsResponse, msgResponse, []uint32{msgResponse.EntityID})
		return
	}

	if tEntityPlayer.IsInMyFriends(msgBody.DelEntityID) {
		cancelFriend := tEntityPlayer.CancelMyFriends(msgBody.DelEntityID)

		tEntityPlayer.SyncEntity(1)

		msgResponse.Code = resp_code.CODE_SUCCESS
		msgResponse.MyFriendsNum = uint32(len(tEntityPlayer.MyFriends))
		friend := c.getFriendList(cancelFriend)
		msgResponse.CancelList = append(msgResponse.CancelList, friend)

		c.delFansList(msgBody.DelEntityID, msgBody.EntityID)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_CancelMyFriendsResponse, msgResponse, []uint32{msgResponse.EntityID})
}

func (c *_Social) delFansList(EntityID, DelEntityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		//todo 用户不在线，去处理
		msgBody := &gmsg.DelMyFansRequest{}
		msgBody.EntityID = EntityID
		msgBody.DelEntityID = DelEntityID
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_DelFansFromGameToDB), msgBody, network.ServerType_DB)
		return
	}

	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if !tEntityPlayer.IsInFansList(DelEntityID) {
		return
	}

	tEntityPlayer.DelFans(DelEntityID)
	tEntityPlayer.SyncEntity(1)

	c.ReduceFansSync(EntityID, DelEntityID)
}

// 附近玩家的20位玩家
func (c *_Social) OnNearByPlayerListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.NearByPlayerListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	if len(c.NearPlayerData) == 0 {
		c.LoadNearPlayerList()
	}

	msgResponse := &gmsg.NearByPlayerListResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.List = make([]*gmsg.NearPlayerList, 0)
	num, max := 0, 20
	for _, val := range c.NearPlayerData {
		if val.EntityID == msgBody.EntityID {
			continue
		}
		val.PlayerType = Player.GetBehaviorStatus(val.EntityID)
		msgResponse.List = append(msgResponse.List, val)
		num++
		if num >= max {
			break
		}
	}
	//log.Info("OnNearByPlayerListRequest", msgResponse)
	if len(msgResponse.List) >= max {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_NearbyPlayerListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_NearbyPlayerListRequest, msgBody, network.ServerType_DB)
}

// 附近玩家的20位玩家响应
func (c *_Social) OnNearByPlayerListResponse(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.NearByPlayerListResponse{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_NearbyPlayerListResponse, msgBody, []uint32{msgBody.EntityID})
}

// 搜索玩家请求
func (c *_Social) OnSearchPlayerFromIDRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SearchPlayerFromIDRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnSearchPlayerFromIDRequest-->begin:", msgBody)

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_SearchPlayerFromIDRequest, msgBody, network.ServerType_DB)
}

// 搜索玩家响应
func (c *_Social) OnSearchPlayerFromIDResponse(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SearchPlayerFromIDResponse{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("-->OnSearchPlayerFromIDResponse-->end:", msgBody)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_SearchPlayerFromIDResponse, msgBody, []uint32{msgBody.EntityID})
}

// 赠送金币 请求
func (c *_Social) OnAddGoldToMyFriendsRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.AddGoldToMyFriendsRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	log.Info("-->OnAddGoldToMyFriendsRequest-->begin--->", msgBody)
	msgResponse := &gmsg.AddGoldToMyFriendsResponse{}
	msgResponse.Code = 1

	if tEntityPlayer.GiveGoldNum() >= consts.DayGiveGoldTimes {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddGoldToMyFriendsResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	// 必须互关才能赠送
	if !tEntityPlayer.IsHaveFriend(msgBody.AddEntityID) {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddGoldToMyFriendsResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	giveSec := tEntityPlayer.IsGiveGold(msgBody.AddEntityID)
	if giveSec == int64(0) {
		resFriend := tEntityPlayer.AddGoldToFriend(msgBody.AddEntityID, consts.GoldAmount)
		if !tEntityPlayer.IsDailyGiveGold() {
			tEntityPlayer.AddGiveGoldDataDate()
			go c.GetPlayerGiveGoldDays(tEntityPlayer)
		}
		tEntityPlayer.SyncEntity(1)
		msgResponse.Code = 0
		friend := c.getFriendList(*resFriend)
		isFriend := true
		friend.IsFriend = &isFriend
		playerBase, ok := Player.PlayerBaseList[msgBody.AddEntityID]
		if !ok {
			return
		}
		stack.SimpleCopyProperties(friend, playerBase)
		msgResponse.Friend = friend
		c.addGoldEmailToMyFriends(msgBody.AddEntityID, tEntityPlayer.PlayerName)
		ConditionalMr.SyncConditional(msgBody.EntityID, []consts.ConditionData{{consts.GiveFriendGold, 1, false}})
	} else {
		msgResponse.Code = uint32(3)
	}

	log.Info("-->OnAddGoldToMyFriendsRequest-->end-->", msgResponse)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddGoldToMyFriendsResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 获取汇总签到的数据
func (c *_Social) GetPlayerGiveGoldDays(tEntityPlayer *entity.EntityPlayer) {
	continueNDay := uint32(0)
	// 获取75天前的时间,包含当天
	bfNDays := tools.GetBeforeNDayString(consts.ContinueNDays75)

	for _, value := range bfNDays {
		month, _ := strconv.Atoi(value[5:7])
		day, _ := strconv.Atoi(value[8:10])

		summaryDailyLog := tEntityPlayer.GetMonthGiveGoldDays(month)
		if summaryDailyLog == nil {
			return
		}
		if summaryDailyLog.Test(uint(day)) {
			continueNDay++
		} else {
			break
		}
	}
	if continueNDay == 0 {
		return
	}
	ConditionalMr.SyncConditional(tEntityPlayer.EntityID, []consts.ConditionData{{consts.GiveGoldFriendDays, continueNDay, true}})
}

// 赠送金币 通知用户添加邮件
func (c *_Social) addGoldEmailToMyFriends(toEntityID uint32, playerName string) {
	//预留发送邮件
	Tittle := tools.StringReplace(Table.GetConstTextFromID(6, DefaultText), "s", playerName)
	Content := tools.StringReplace(Table.GetConstTextFromID(5, DefaultText), "s", playerName)
	email := new(gmsg.Email)
	email.EmailID = Player.GetMaxUuid(toEntityID)
	email.Date = tools.GetTimeByTimeStamp(time.Now().Unix())
	email.StateReward = false
	email.Tittle = Tittle
	email.Content = Content
	email.IsRewardEmail = true
	email.RewardList = make([]*gmsg.RewardInfo, 0)

	emailRewardEntity := new(gmsg.RewardInfo)
	emailRewardEntity.ItemTableId = consts.Gold
	emailRewardEntity.Num = consts.GoldAmount
	email.RewardList = append(email.RewardList, emailRewardEntity)
	Email.AddEmail(toEntityID, email)
}

// 添加粉丝推送
func (c *_Social) AddFansSync(entityID, fansEntityID uint32) {
	msgResponse := &gmsg.AddFansSync{}
	msgResponse.Fans = new(gmsg.FansList)
	tEntity := Entity.EmPlayer.GetEntityByID(fansEntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	stack.SimpleCopyProperties(msgResponse.Fans, tEntityPlayer)
	msgResponse.EntityID = entityID

	log.Info("AddFansSync,发送成功，", entityID)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_AddFansSync, msgResponse, []uint32{entityID})
}

// 减少粉丝推送
func (c *_Social) ReduceFansSync(entityID, fansEntityID uint32) {
	msgResponse := &gmsg.ReduceFansSync{}
	msgResponse.Fans = new(gmsg.FansList)
	msgResponse.Fans.EntityID = fansEntityID
	msgResponse.EntityID = entityID

	log.Info("ReduceFansSync,发送成功，", entityID)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ReduceFansSync, msgResponse, []uint32{entityID})
}

func (c *_Social) getFriendsList(tEntityPlayer *entity.EntityPlayer) []*gmsg.FriendList {
	friends := make([]*gmsg.FriendList, 0)
	for _, vl := range tEntityPlayer.MyFriends {
		Friend := &gmsg.FriendList{}
		Friend.EntityID = vl.EntityID
		Friend.AddTime = vl.AddTime
		Friend.Online = Player.GetGamePlayerOnline(vl.EntityID)
		Friend.IsGiveGold = c.isHaveGiveGold(vl.GiveGoldSec)
		for _, val := range tEntityPlayer.GiveGoldList {
			if vl.EntityID == val.EntityID {
				Friend.IsGiveGold = true
			}
		}
		if playerBase, ok := Player.PlayerBaseList[vl.EntityID]; ok {
			stack.SimpleCopyProperties(Friend, playerBase)
		}
		friends = append(friends, Friend)
	}

	return friends
}

func (c *_Social) getFansList(tEntityPlayer *entity.EntityPlayer) (fansList []*gmsg.FansList, fanNum uint32) {
	for _, vl := range tEntityPlayer.FansList.List {
		if vl.AddTime >= uint64(tEntityPlayer.FansList.FansUnixSec) {
			fanNum++
		}
		fan := &gmsg.FansList{}
		if playerBase, ok := Player.PlayerBaseList[vl.EntityID]; ok {
			stack.SimpleCopyProperties(fan, playerBase)
		}
		fansList = append(fansList, fan)
	}
	return fansList, fanNum
}

func (c *_Social) OnUpdateFansUnixSec(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.UpdateFansUnixSec{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.UpdateFansUnixSec()
	tEntityPlayer.SyncEntity(1)
}

func (c *_Social) MyFriendListSync(tEntityPlayer *entity.EntityPlayer) {
	msgResponse := &gmsg.MyFriendsListResponse{}
	msgResponse.EntityID = tEntityPlayer.EntityID
	msgResponse.List = make([]*gmsg.FriendList, 0)
	msgResponse.List = c.getFriendsList(tEntityPlayer)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_MyFriendsListResponse, msgResponse, []uint32{msgResponse.EntityID})
}

func (c *_Social) isHaveGiveGold(sec int64) bool {
	return sec > 0 && sec >= tools.GetTodayBeginTime()
}
