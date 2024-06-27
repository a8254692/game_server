package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"errors"
	"fmt"
	"gitee.com/go-package/carbon/v2"
	"google.golang.org/protobuf/proto"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sort"
	"sync"
	"time"
)

/***
 *@disc:俱乐部管理
 *@author: lsj
 *@date: 2023/9/13
 */

type _Club struct {
	Top10          []*gmsg.ClubTop10Info
	IsSync         bool
	RankList       map[uint32][]ClubRate
	PalaceRankList []ClubRate
	ProfitGoldList []ProfitGold
	RedPack        map[string]*RedPack
	ClubShopList   []entity.ClubShopItem
}

type UpdateClubData struct {
	GameType       uint32 // 游戏类型，0：8球，1：血流，2：斯诺克
	RoomType       uint32 // 房间类型，0：新手，1初级，2中级，3高级，4巅峰
	Result         uint32 // 游戏输赢
	SettlementType uint32 // 结算类型
	EntityID       uint32 // 用户id
	Gold           uint32 //赢的金币，输的填0
}

type ClubRate struct {
	ClubID         uint32
	ClubName       string
	ClubBadge      uint32
	RateRank       uint32
	ClubScore      uint32
	ClubRate       uint32
	LastWeekRank   uint32
	MasterEntityID uint32
	TotalScore     uint32
}

type ProfitGold struct {
	ClubID     uint32
	ClubName   string
	ProfitGold uint64
}

var ClubManager _Club

var ClubMutex sync.Mutex

func (c *_Club) Init() {
	c.Top10 = make([]*gmsg.ClubTop10Info, 0)
	c.RankList = make(map[uint32][]ClubRate, 0)
	c.PalaceRankList = make([]ClubRate, 0)
	c.RedPack = make(map[string]*RedPack, 0)
	c.ClubShopList = make([]entity.ClubShopItem, 0)
	c.IsSync = false
	c.statisticsTick()
	time.AfterFunc(time.Millisecond*500, c.SyncEntityClub)
	c.clubShopCfg()

	event.OnNet(gmsg.MsgTile_Hall_ClubListRequest, reflect.ValueOf(c.OnClubListRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubListResponse, reflect.ValueOf(c.OnClubListResponse))
	event.OnNet(gmsg.MsgTile_Hall_ClubTop10Request, reflect.ValueOf(c.OnClubTop10Request))
	event.OnNet(gmsg.MsgTile_Hall_ClubCreateRequest, reflect.ValueOf(c.OnClubCreateRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubCreateResponse, reflect.ValueOf(c.OnClubCreateResponse))
	event.OnNet(gmsg.MsgTile_Hall_UpdateClubRequest, reflect.ValueOf(c.OnClubUpdateRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubRatifyListRequest, reflect.ValueOf(c.OnClubRatifyListRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubExitRequest, reflect.ValueOf(c.OnClubExitRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubRatifyJoinRequest, reflect.ValueOf(c.ClubRatifyJoinRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubHomePageRequest, reflect.ValueOf(c.ClubHomePageRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubReqJoinRequest, reflect.ValueOf(c.OnClubReqJoinRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubCancelJoinRequest, reflect.ValueOf(c.OnClubCancelJoinRequest))
	event.OnNet(gmsg.MsgTile_Hall_TransferMasterRequest, reflect.ValueOf(c.OnTransferMasterRequest))
	event.OnNet(gmsg.MsgTile_Hall_CommissionSecondMasterRequest, reflect.ValueOf(c.OnCommissionSecondMasterRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubDelMembersRequest, reflect.ValueOf(c.OnClubDelMembersRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubDelMembersResponse, reflect.ValueOf(c.OnClubDelMembersResponse))
	event.OnNet(gmsg.MsgTile_Hall_ClubTaskListRequest, reflect.ValueOf(c.OnClubTaskListRequest))

	event.OnNet(gmsg.MsgTile_Hall_ClubShopListRequest, reflect.ValueOf(c.OnClubShopListRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubShopBuyRequest, reflect.ValueOf(c.OnClubShopBuyRequest))

	event.OnNet(gmsg.MsgTile_Hall_ClubRedEnvelopeListRequest, reflect.ValueOf(c.OnClubRedEnvelopeListRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubSendRedEnvelopeRequest, reflect.ValueOf(c.OnClubSendRedEnvelopeRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubRedEnvelopeRecordListRequest, reflect.ValueOf(c.OnClubRedEnvelopeRecordListRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenRequest, reflect.ValueOf(c.OnClubRedEnvelopeOpenRequest))

	event.OnNet(gmsg.MsgTile_Hall_ClubDailySignInRequest, reflect.ValueOf(c.OnClubDailySignInRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubSupportRequest, reflect.ValueOf(c.OnClubSupportRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubClaimTaskProgressRequest, reflect.ValueOf(c.OnClubClaimTaskProgressRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClaimMyClubTaskProgressRequest, reflect.ValueOf(c.OnClaimMyClubTaskProgressRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubRateListRequest, reflect.ValueOf(c.OnClubRateListRequest))
	event.OnNet(gmsg.MsgTile_Hall_PalaceClubRateListRequest, reflect.ValueOf(c.OnPalaceClubRateListRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubProfitGoldListRequest, reflect.ValueOf(c.OnClubProfitGoldListRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_ClubTOP10_DB_Data_Response), reflect.ValueOf(c.SetMsgToDbForClubTOP10Data))
	event.OnNet(gmsg.MsgTile_Login_PlayerClubSync, reflect.ValueOf(c.OnMainPlayerClubSync))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncEntityClubDBToGame), reflect.ValueOf(c.SyncEntityClubFromDB))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubRatifyJoinToGame), reflect.ValueOf(c.ClubRatifyJoinFromDbRequest))
	timer.AddTimer(c, "SendSyncMsgToDbForClubTOP10Data", 60000, true)
	timer.AddTimer(c, "SetClubRankList", 10000, true)

	//----------------------俱乐部测试协议
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_BatchCreateClubResponse), reflect.ValueOf(c.OnBatchReqClubResponse))
}

// game启动，并延迟通知DB同步club数据
func (c *_Club) SyncEntityClub() {
	resBody := &gmsg.SyncEntityClubNoticeDB{}
	resBody.TimeStamp = uint32(carbon.Now().Timestamp())

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncEntityClub), resBody, network.ServerType_DB)
}

func (c *_Club) SetClubRankList() {
	if len(Entity.EmClub.EntityMap) == 0 {
		log.Info("没有俱乐部。")
		return
	}

	// 循环所有评级
	for m := consts.ClubRateE; m <= consts.ClubRateS; m++ {
		c.RankList[m] = c.getClubRate(m)
	}

	dataRate, dataGold := c.getPalaceRankList()
	c.PalaceRankList = dataRate
	c.ProfitGoldList = dataGold

	//log.Info("-->SetClubRankList-->", c.RankList, c.ProfitGoldList, "-->count:", len(c.RankList))
}

func (c *_Club) getClubRate(rate uint32) []ClubRate {
	data := make([]ClubRate, 0)
	for clubID, _ := range Entity.EmClub.EntityMap {
		emClub := Entity.EmClub.GetEntityByID(clubID)
		club := emClub.(*entity.Club)
		if rate < consts.ClubRateS && rate == club.ClubRate {
			clubRateInfo := new(ClubRate)
			stack.SimpleCopyProperties(clubRateInfo, club)
			data = append(data, *clubRateInfo)
		} else if rate == consts.ClubRateS && club.ClubRate >= rate {
			clubRateInfo := new(ClubRate)
			stack.SimpleCopyProperties(clubRateInfo, club)
			data = append(data, *clubRateInfo)
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].ClubScore > data[j].ClubScore || (data[i].ClubScore == data[j].ClubScore && data[i].TotalScore > data[j].TotalScore)
	})
	return data
}

func (c *_Club) getPalaceRankList() ([]ClubRate, []ProfitGold) {
	dataRate := make([]ClubRate, 0)
	dataGold := make([]ProfitGold, 0)
	for clubID, _ := range Entity.EmClub.EntityMap {
		emClub := Entity.EmClub.GetEntityByID(clubID)
		club := emClub.(*entity.Club)
		clubRateInfo := new(ClubRate)
		stack.SimpleCopyProperties(clubRateInfo, club)
		dataRate = append(dataRate, *clubRateInfo)
		gold := new(ProfitGold)
		stack.SimpleCopyProperties(gold, club)
		dataGold = append(dataGold, *gold)
	}
	sort.Slice(dataRate, func(i, j int) bool {
		return dataRate[i].TotalScore > dataRate[j].TotalScore
	})

	sort.Slice(dataGold, func(i, j int) bool {
		return dataGold[i].ProfitGold > dataGold[j].ProfitGold
	})
	return dataRate, dataGold
}

func (c *_Club) statisticsTick() {
	leftSecond := time.Duration(tools.GetLeftSecondByTomorrow()+1) * time.Second
	time.AfterFunc(leftSecond, func() {
		if tools.GetWeekDay() == 1 {
			go c.statisticsClubFunc()
		}
		c.clearExpireRedPack()
		c.statisticsTick()
	})
}

func (c *_Club) statisticsClubFunc() {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()

	// 循环所有评级
	log.Info("-->statisticsClubFunc-->begin-->")
	RankList := make(map[uint32][]ClubRate, 0)
	for m := consts.ClubRateE; m <= consts.ClubRateS; m++ {
		data := c.getClubRate(m)
		RankList[m] = data
	}

	for key, val := range RankList {
		log.Info("rate:", key)
		for k, vl := range val {
			rate := consts.ClubRateE
			if key == consts.ClubRateS {
				rate = c.statisticsClubRateS(vl.ClubRate, vl.ClubScore, k)
			} else if key < consts.ClubRateA {
				rate = c.statisticsBeforeClubRateB(vl.ClubRate, vl.ClubScore)
			} else {
				rate = c.statisticsClubRateA(vl.ClubRate, vl.ClubScore, k)
			}
			emClub := Entity.EmClub.GetEntityByID(vl.ClubID)
			club := emClub.(*entity.Club)
			club.UpgradeClubFromScore(rate, uint32(k))
			club.ReSetClubScore()
			club.SyncEntity(1)
			log.Info("club:", club.ClubID, "->ClubScore-->", vl.ClubScore, "->rate-->", rate, "->key-->", k)
		}
	}
	//重置周盈利缓存
	c.ProfitGoldList = nil
	c.SetClubRankList()
	//发放奖励
	go c.ClubSendItemRewardEmail()
	log.Info("-->statisticsClubFunc-->end-->")
}

func (c *_Club) statisticsBeforeClubRateB(clubRate uint32, clubScore uint32) (rate uint32) {
	rateCfg := Table.GetClubRateRewardCfg(clubRate)
	if rateCfg == nil {
		return clubRate
	}

	// 前4处理
	if clubRate >= consts.ClubRateE && clubRate < consts.ClubRateA {
		if clubRate >= consts.ClubRateE && clubScore >= rateCfg.Upgrade {
			//升级
			rate = clubRate + 1
		} else if clubRate > consts.ClubRateE && clubScore < rateCfg.KeepGrade {
			//降级
			rate = clubRate - 1
		} else {
			//不升不降
			rate = clubRate
		}
	}

	return rate
}

func (c *_Club) statisticsClubRateA(clubRate uint32, clubScore uint32, rank int) (rate uint32) {
	rateCfg := Table.GetClubRateRewardCfg(clubRate)
	if rateCfg == nil {
		return clubRate
	}

	// A处理，Upgrade是排名参数，不是分数
	if clubRate >= consts.ClubRateA {
		if rank < int(rateCfg.Upgrade) && clubScore >= rateCfg.KeepGrade {
			//升级
			rate = clubRate + 1
		} else if clubScore < rateCfg.KeepGrade {
			//降级
			rate = clubRate - 1
		} else {
			//不升不降
			rate = clubRate
		}
	}

	return rate
}

func (c *_Club) statisticsClubRateS(clubRate uint32, clubScore uint32, rank int) (rate uint32) {
	if clubRate == consts.ClubRateSPlus {
		clubRate -= 1
	}
	rateCfg := Table.GetClubRateRewardCfg(clubRate)
	if rateCfg == nil {
		return clubRate
	}

	// S处理,Upgrade是排名参数，不是分数
	if clubRate >= consts.ClubRateS {
		//如果评分为零，降级
		if clubScore == 0 {
			//降级
			rate = clubRate - 1
			return
		}

		if rank < int(rateCfg.Upgrade) {
			//升级
			rate = clubRate + 1
		} else if consts.ClubRateSTotal-rank <= int(rateCfg.Upgrade) && rank > int(rateCfg.Upgrade) {
			//降级
			rate = clubRate - 1
		} else {
			rate = clubRate
		}

	}

	return rate
}

// 重置俱乐部任务相关数据
func (c *_Club) resetClubTaskAndShop() {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	log.Info("-->resetClubTaskAtSaturday-->begin-->")
	for _, emClub := range Entity.EmClub.EntityMap {
		club := emClub.(*entity.Club)
		club.ResetClubMember(consts.PromoteEliteActive)
		club.ReSetClubActiveValue()
		club.ReSeProfitGold()
		club.ShopList = nil
		club.ShopList = c.ClubShopList
		club.SyncEntity(1)
	}
	log.Info("-->resetClubTaskAtSaturday-->end-->", "count:", len(Entity.EmClub.EntityMap))
}

// 同步俱乐部entity
func (c *_Club) OnMainPlayerClubSync(msgEV *network.MsgBodyEvent) {
	var tEntityClub entity.Club
	stack.BytesToStruct_Gob(msgEV.MsgBody, &tEntityClub)
	Entity.EmClub.AddEntity(&tEntityClub)
	log.Info("-->OnMainPlayerClubSync-->end", &tEntityClub)
}

// db服同步entity_club到游戏服
func (c *_Club) SyncEntityClubFromDB(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	var tEntityClubArgs []entity.Club
	c.RedPack = make(map[string]*RedPack, 0)
	stack.BytesToStruct_Gob(msgEV.MsgBody, &tEntityClubArgs)
	for _, tEntityClub := range tEntityClubArgs {
		info := tEntityClub
		Entity.EmClub.AddEntity(&info)
		c.BuildRedPackFirstTime(&info)
	}
	c.IsSync = true
	log.Info("俱乐部同步成功,SyncEntityClubFromDB,end,", len(tEntityClubArgs))
}

func (c *_Club) BuildRedPackFirstTime(club *entity.Club) {
	if len(club.RedEnvelopeList) == 0 {
		return
	}

	for _, val := range club.RedEnvelopeList {
		if val.SendTime < time.Now().Unix()-86400 || val.TotalSendNum == val.NumDelivered {
			continue
		}
		newRedPack := new(RedPack)
		newRedPack.SendTime = val.SendTime
		newRedPack.Id = val.RedEnvelopeID.Hex()
		newRedPack.Num = int(val.TotalSendNum)
		newRedPack.Amount = int(val.SendCoinNum)
		newRedPack.NumDelivered = int(val.NumDelivered)
		newRedPack.AmountDelivered = int(val.AmountDelivered)
		c.RedPack[val.RedEnvelopeID.Hex()] = newRedPack
	}

	log.Info("-->BuildRedPackFirstTime-->end-->", c.RedPack)
}

// 清空过期红包
func (c *_Club) clearExpireRedPack() {
	if len(c.RedPack) == 0 {
		return
	}

	for key, val := range c.RedPack {
		if val.SendTime < time.Now().Unix()-86400 {
			delete(c.RedPack, key)
		}
	}

	log.Info("-->clearExpireRedPack-->end")
}

// 查询俱乐部列表 游戏->db服
func (c *_Club) OnClubListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	if *msgBody.IsJoinLevel {
		tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
		if tEntity == nil {
			return
		}
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		msgBody.PlayerLV = &tEntityPlayer.PlayerLv
	} else {
		msgBody.PlayerLV = proto.Uint32(0)
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_ClubListRequest, msgBody, network.ServerType_DB)
}

func (c *_Club) OnClubListResponse(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubListResponse{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubListResponse, msgBody, []uint32{msgBody.EntityID})
}

// 定时通知DB同步TOP10数据
func (c *_Club) SendSyncMsgToDbForClubTOP10Data() {
	request := &gmsg.ClubTop10DBRequest{}
	request.TimeStamp = uint32(carbon.Now().Timestamp())
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_ClubTOP10_DB_Data_Request), request, network.ServerType_DB)
}

// 定时同步TOP10数据
func (c *_Club) SetMsgToDbForClubTOP10Data(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubTop10DBResponse{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	if len(msgBody.List) == 0 {
		return
	}

	list := make([]*gmsg.ClubTop10Info, 0)
	for _, vl := range msgBody.List {
		club := new(gmsg.ClubTop10Info)
		stack.SimpleCopyProperties(club, vl)
		list = append(list, club)
	}

	c.Top10 = list

	//log.Info("SetMsgToDbForClubTOP10Data,Complete", c.Top10, msgBody.TimeStamp)
}

// TOP10俱乐部排名 查询内存->前端
func (c *_Club) OnClubTop10Request(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubTop10Response{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.ClubTop10Response{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.List = c.Top10
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubTop10Response, msgResponse, []uint32{msgBody.EntityID})
}

// 创建俱乐部请求
func (c *_Club) OnClubCreateRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubCreateRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse := &gmsg.ClubCreateResponse{}
	if tEntityPlayer.GetClubID() > 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubCreateResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	//if !tools.LimitChinese(msgBody.ClubName) || msgBody.ClubNotice == "" {
	//	msgResponse := &gmsg.ClubCreateResponse{}
	//	msgResponse.Code = 1
	//	msgResponse.MasterEntityID = msgBody.EntityID
	//	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubCreateResponse, msgResponse, []uint32{msgBody.EntityID})
	//	return
	//}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_ClubCreateRequest, msgBody, network.ServerType_DB)
}

// 创建俱乐部返回
func (c *_Club) OnClubCreateResponse(msgEV *network.MsgBodyEvent) {
	msgResponse := &gmsg.ClubCreateResponse{}
	if err := msgEV.Unmarshal(msgResponse); err != nil {
		return
	}

	if msgResponse.ClubID > 0 {
		tEntity := Entity.EmPlayer.GetEntityByID(msgResponse.MasterEntityID)
		if tEntity == nil {
			return
		}
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		if tEntityPlayer.GetClubID() > 0 {
			return
		}
		tEntityPlayer.UpdateClubTags()
		//保存角色
		emClub := Entity.EmClub.GetEntityByID(msgResponse.ClubID)
		club := emClub.(*entity.Club)
		tEntityPlayer.ClubId = msgResponse.ClubID
		tEntityPlayer.ClubName = club.ClubName
		tEntityPlayer.ClubBadge = club.ClubBadge
		tEntityPlayer.ClubRate = club.ClubRate
		c.ClubTaskInit(msgResponse.MasterEntityID)
		tEntityPlayer.SyncEntity(1)
		//推送角色同步给前端
		Player.SyncClubToPlayer(msgResponse.MasterEntityID)
		ConditionalMr.SyncConditional(msgResponse.MasterEntityID, []consts.ConditionData{{consts.JoinOrCreateClub, 1, false}})
		club.ShopList = c.ClubShopList
		club.SyncEntity(1)
		c.SyncClubTaskList(tEntityPlayer)
		c.toEntityEmailClub(true, club.ClubName, msgResponse.MasterEntityID)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubCreateResponse, msgResponse, []uint32{msgResponse.MasterEntityID})
}

func (c *_Club) clubShopCfg() {
	list := Table.GetClubShopCfgMap()
	if len(list) == 0 {
		log.Error("-->ClubShopCfg is nil:")
		return
	}

	c.ClubShopList = nil
	for _, v := range list {
		shopItem := new(entity.ClubShopItem)
		shopItem.ItemID = v.TableID
		shopItem.TableID = v.ItemTableID
		shopItem.MaxBuyNum = v.MaxBuyNum
		shopItem.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		shopItem.Sort = int(v.Sort)
		shopItem.Unlock = v.Unlock
		c.ClubShopList = append(c.ClubShopList, *shopItem)
	}

	//升序
	sort.Slice(c.ClubShopList, func(i, j int) bool {
		return c.ClubShopList[i].Sort < c.ClubShopList[j].Sort
	})

	log.Info("-->clubShopCfg-->success-->", c.ClubShopList)
}

func (c *_Club) OnClubUpdateRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.UpdateClubRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}
	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.GetClubID())
	club := emClub.(*entity.Club)
	if club == nil {
		return
	}

	msgResponse := &gmsg.UpdateClubResponse{}
	msgResponse.Code = uint32(1)

	//if *msgBody.ClubNotice != "" && utf8.RuneCountInString(*msgBody.ClubNotice) > 30 {
	//	log.Error("ClubNotice大于30")
	//	return
	//}
	//
	//if *msgBody.ClubName != "" && !tools.LimitChinese(*msgBody.ClubName) {
	//	log.Error("ClubName不符合要求")
	//	return
	//}

	// 管理人员操作
	if club.MemberPosition(msgBody.EntityID) >= consts.Second_Master {
		log.Info("-->OnClubUpdateRequest-->begin-->")
		if *msgBody.ClubNotice != "" {
			club.ClubNotice = *msgBody.ClubNotice
			club.JoinLevel = *msgBody.JoinLevel
			if *msgBody.IsOpen {
				club.IsOpen = true
			} else {
				club.IsOpen = false
			}
		} else if *msgBody.ClubName != "" {
			club.ClubName = *msgBody.ClubName
		}
		club.SyncEntity(1)
		msgResponse.Code = uint32(0)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UpdateClubResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 审核成员加入，游戏->db服
func (c *_Club) ClubRatifyJoinRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubRatifyJoinRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.ClubRatifyJoinResponse{}
	msgResponse.Code = uint32(1)

	emClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := emClub.(*entity.Club)
	//权限不足
	if club == nil || club.MemberPosition(msgBody.EntityID) < consts.Second_Master {
		log.Info("//权限不足")
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRatifyJoinResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	// 超过人数上限
	if len(club.Members) >= int(club.MaxNum) || int(club.MaxNum) < len(club.Members)+len(msgBody.List) {
		log.Info("//超过人数上限")
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRatifyJoinResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	c.ratifyJoinFunc(club, msgBody.List)

	msgResponse.Code = uint32(0)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRatifyJoinResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 加入俱乐部逻辑,先判断是否离线，在线直接在gameServer处理，离线的去DB处理
func (c *_Club) ratifyJoinFunc(club *entity.Club, list []*gmsg.JoinEntityIDList) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	log.Info("-->ratifyJoinFunc-->", list)
	for _, vl := range list {
		ratifyJoinBody := &gmsg.ClubRatifyJoinToDB{}
		ratifyJoinBody.ClubID = club.ClubID
		ratifyJoinBody.EntityID = vl.EntityID
		ratifyJoinBody.AddEntityID = vl.EntityID
		// 俱乐部成员有位子，并且加入的 ，否则直接拒绝
		if len(club.Members) < int(club.MaxNum) && vl.IsJoin {
			tEntity := Entity.EmPlayer.GetEntityByID(vl.EntityID)
			if tEntity == nil {
				//通知DB处理
				ratifyJoinBody.IsJoin = vl.IsJoin
				ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubRatifyJoinToDB), ratifyJoinBody, network.ServerType_DB)
				continue
			}
			tEntityPlayer := tEntity.(*entity.EntityPlayer)
			if tEntityPlayer.GetClubID() > 0 {
				continue
			}
			club.AddMembers(vl.EntityID, consts.General)
			tEntityPlayer.UpdateClubTags()
			//清除用户申请数据
			c.clearEntityPlayerClubReq(tEntityPlayer)
			//修改用户
			tEntityPlayer.JoinClub(club.ClubID, club.ClubBadge, club.ClubRate, club.ClubName)
			c.ClubTaskInit(vl.EntityID)
			tEntityPlayer.SyncEntity(1)
			//推送俱乐部给客户端
			Player.SyncClubToPlayer(vl.EntityID)
			c.SyncClubTaskList(tEntityPlayer)
			club.SyncEntity(0)
			c.toEntityEmailClub(true, club.ClubName, vl.EntityID)
			ChatMgr.InnerSendClubMsg(uint32(gmsg.MsgType_MtSystem), 0, club.ClubID, consts.PRIVATELY_BUY_NOTICE)
			ConditionalMr.SyncConditional(vl.EntityID, []consts.ConditionData{{consts.JoinOrCreateClub, 1, false}})
		} else {
			tEntity := Entity.EmPlayer.GetEntityByID(vl.EntityID)
			if tEntity == nil {
				//通知DB处理
				ratifyJoinBody.IsJoin = vl.IsJoin
				ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubRatifyJoinToDB), ratifyJoinBody, network.ServerType_DB)
				continue
			}
			tEntityPlayer := tEntity.(*entity.EntityPlayer)
			tEntityPlayer.RemoveClubReq(club.ClubID)
			tEntityPlayer.SyncEntity(1)
		}
	}
	// 审核完了，清空申请列表
	club.ClearReqList()

	club.SyncEntity(1)

	log.Info("-->ratifyJoinFunc-->end")

	return nil
}

// 清除用户所有的俱乐部申请请求
func (c *_Club) clearEntityPlayerClubReq(tEntityPlayer *entity.EntityPlayer) {
	for _, val := range tEntityPlayer.ReqJoinClub {
		emClub := Entity.EmClub.GetEntityByID(val)
		if emClub == nil {
			continue
		}
		club := emClub.(*entity.Club)
		club.RemoveOneReqList(tEntityPlayer.EntityID)
		club.SyncEntity(1)
	}
}

// 审核成员,db服->游戏服
func (c *_Club) ClubRatifyJoinFromDbRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgResponse := &gmsg.ClubRatifyJoinToGame{}
	if err := msgEV.Unmarshal(msgResponse); err != nil {
		return
	}
	emClub := Entity.EmClub.GetEntityByID(msgResponse.ClubID)
	club := emClub.(*entity.Club)

	if msgResponse.IsJoin {
		c.toEntityEmailClub(true, club.ClubName, msgResponse.AddEntityID)
		log.Info("-->ClubRatifyJoinFromDbRequest--begin-->", msgResponse)
		club.AddMembers(msgResponse.AddEntityID, consts.General)
		club.SyncEntity(1)
		ConditionalMr.SyncConditional(msgResponse.AddEntityID, []consts.ConditionData{{consts.JoinOrCreateClub, 1, false}})
	}

	log.Info("-->ClubRatifyJoinFromDbRequest->end-->", msgResponse)
}

func (c *_Club) toEntityEmailClub(event bool, clubName string, toEntityID uint32) {
	Tittle, Content := "", ""
	if event {
		Tittle = tools.StringReplace(Table.GetConstTextFromID(7, DefaultText), "s", clubName)
		Content = tools.StringReplace(Table.GetConstTextFromID(8, DefaultText), "s", clubName)
	} else {
		Tittle = tools.StringReplace(Table.GetConstTextFromID(9, DefaultText), "s", clubName)
		Content = tools.StringReplace(Table.GetConstTextFromID(10, DefaultText), "s", clubName)
	}

	email := new(gmsg.Email)
	email.EmailID = Player.GetMaxUuid(toEntityID)
	email.Date = tools.GetTimeByTimeStamp(time.Now().Unix())
	email.StateReward = false
	email.Tittle = Tittle
	email.Content = Content

	Email.AddEmail(toEntityID, email)
}

func (c *_Club) OnClubExitRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubExitRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	msgResponse := &gmsg.ClubExitResponse{}
	msgResponse.Code = uint32(0)

	if tEntityPlayer.GetClubID() == 0 {
		log.Info("角色没有俱乐部,", msgBody.EntityID)
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.GetClubID())
	club := emClub.(*entity.Club)
	if club == nil || club.IsMasterEntityID(msgBody.EntityID) {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubExitResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	tEntityPlayer.ResetClubID()
	tEntityPlayer.SetExitClubUnixSec()
	tEntityPlayer.ClubAttribute.ClubProgressRewardList = Table.GetClubProgressRewardList()
	tEntityPlayer.SyncEntity(1)

	club.ReMoveMember(msgBody.EntityID)
	club.SyncEntity(1)
	c.toEntityEmailClub(false, club.ClubName, msgBody.EntityID)
	// 推送俱乐部信息给前端
	Player.SyncClubToPlayer(msgBody.EntityID)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubExitResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 查询俱乐部主页
func (c *_Club) ClubHomePageRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubHomePageRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.ClubHomePageResponse{}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.ClubId == 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubHomePageResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	if club == nil {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubHomePageResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	membersArgs := club.GetMembers()
	members := make([]*gmsg.MembersInfo, 0)
	sdMasterNum := uint32(0)
	for _, vl := range membersArgs {
		data := new(gmsg.MembersInfo)
		stack.SimpleCopyProperties(data, vl)
		pEntity := Entity.EmPlayer.GetEntityByID(vl.EntityID)
		if pEntity == nil {
			data.Online = uint32(carbon.Now().Timestamp()-tools.GetUnixFromStr(Player.GetGamePlayerCurrentLoginTime(vl.EntityID))) / uint32(60)
		} else {
			data.Online = 0
		}

		members = append(members, data)
		sdMasterNum++
	}

	clubHomePage := new(gmsg.ClubHome)
	stack.SimpleCopyProperties(clubHomePage, club)
	clubHomePage.SecondMasterNum = sdMasterNum
	msgResponse.Code = uint32(0)
	msgResponse.List = members
	msgResponse.Data = clubHomePage
	msgResponse.IsDailySignIn = tEntityPlayer.IsClubDailySignIn()
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubHomePageResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 加入俱乐部请求
func (c *_Club) OnClubReqJoinRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubReqJoinRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	emClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := emClub.(*entity.Club)
	if club == nil {
		log.Error("俱乐部为空。")
		return
	}
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	if tEntityPlayer.GetClubID() > 0 {
		return
	}

	msgResponse := &gmsg.ClubReqJoinResponse{}
	msgResponse.Code = uint32(0)
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.ClubID = 0

	//判断用户离开俱乐部时间
	if !tEntityPlayer.IsGtExitClubUnixSec() {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubReqJoinResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	//达到人数不能申请
	if club.MaxNum <= club.Num {
		msgResponse.Code = uint32(3)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubReqJoinResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	// 如果开放，自动加入，不用放在申请列表了
	if club.IsOpen && club.Num < club.MaxNum {
		joinList := make([]*gmsg.JoinEntityIDList, 0)
		JoinEntity := new(gmsg.JoinEntityIDList)
		JoinEntity.EntityID = msgBody.EntityID
		JoinEntity.IsJoin = true
		joinList = append(joinList, JoinEntity)
		c.ratifyJoinFunc(club, joinList)
		msgResponse.ClubID = msgBody.ClubID

		ChatMgr.InnerSendClubMsg(uint32(gmsg.MsgType_MtSystem), 0, club.ClubID, consts.PRIVATELY_BUY_NOTICE)

		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubReqJoinResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	err := c.clubReqJoin(club, tEntityPlayer)
	if err != nil {
		log.Error(err)
		msgResponse.Code = uint32(1)
	}
	log.Info("----->OnClubReqJoinRequest-->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubReqJoinResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) clubReqJoin(club *entity.Club, tEntityPlayer *entity.EntityPlayer) error {
	if tEntityPlayer.IsIReqJoinClub(club.ClubID) {
		return errors.New("重复俱乐部id")
	}

	if tEntityPlayer.PlayerLv < club.JoinLevel {
		return errors.New("用户达不到俱乐部等级。")
	}

	if len(tEntityPlayer.ReqJoinClub) >= consts.MaxReqClub {
		return errors.New("申请俱乐部上限。")
	}

	if club.IsInReqList(tEntityPlayer.EntityID) {
		return errors.New("在申请表中。")
	}

	club.AddReqList(tEntityPlayer.EntityID)
	tEntityPlayer.ReqJoinClub = append(tEntityPlayer.ReqJoinClub, club.ClubID)
	tEntityPlayer.SyncEntity(1)
	club.SyncEntity(1)

	return nil
}

func (c *_Club) OnClubCancelJoinRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubCancelJoinRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	emClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := emClub.(*entity.Club)
	if club == nil {
		log.Error("俱乐部为空。")
		return
	}
	msgResponse := &gmsg.ClubCancelJoinResponse{}
	err := c.clubCancelJoin(club, msgBody.EntityID)
	if err != nil {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubCancelJoinResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	msgResponse.Code = uint32(0)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubCancelJoinResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) clubCancelJoin(club *entity.Club, EntityID uint32) error {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return errors.New("无用户数据。")
	}
	if tEntityPlayer.GetClubID() > 0 {
		log.Error("用户已有俱乐部。")
		return errors.New("用户已有俱乐部。")
	}
	club.RemoveReqList([]uint32{EntityID})
	tEntityPlayer.RemoveClubReq(club.ClubID)
	tEntityPlayer.SyncEntity(1)
	club.SyncEntity(1)
	return nil
}

// 部长任命与取消，二合一
func (c *_Club) OnCommissionSecondMasterRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.CommissionSecondMasterRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	emClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := emClub.(*entity.Club)
	if club == nil {
		log.Error("俱乐部为空。", msgBody.ClubID)
		return
	}

	msgResponse := &gmsg.CommissionSecondMasterResponse{}
	//不能任命自己
	if msgBody.SecondMasterID == msgBody.EntityID {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_CommissionSecondMasterResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if !club.IsMasterEntityID(msgBody.EntityID) || (msgBody.Position == consts.Second_Master && Table.GetCluMasterNum(club.ClubLV) <= club.TotalPosition(consts.Second_Master)) {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_CommissionSecondMasterResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	c.commissionSecondMaster(club, msgBody.SecondMasterID, msgBody.Position)
	msgResponse.Code = uint32(0)
	msgResponse.EntityID = msgBody.EntityID
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_CommissionSecondMasterResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) commissionSecondMaster(club *entity.Club, secondMasterID, position uint32) {
	entityMap := make(map[uint32]uint32, 0)
	entityMap[secondMasterID] = position
	club.SetMemberPosition(entityMap)
	for key, vl := range club.Members {
		if vl.EntityID == secondMasterID {
			a := club.Members[key]
			a.Position = position
			club.Members[key] = a
		}
	}
	club.SyncEntity(1)
}

// 转让部长
func (c *_Club) OnTransferMasterRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.TransferMasterRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer == nil {
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.GetClubID())
	club := emClub.(*entity.Club)
	if club == nil {
		log.Error("俱乐部为空。")
		return
	}

	msgResponse := &gmsg.TransferMasterResponse{}
	if msgBody.NewMasterEntityID == msgBody.EntityID || !club.IsMasterEntityID(msgBody.EntityID) {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_TransferMasterResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	c.transferMaster(club, msgBody.EntityID, msgBody.NewMasterEntityID)
	msgResponse.Code = uint32(0)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_TransferMasterResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) transferMaster(club *entity.Club, entityID, newEntityID uint32) {
	club.SetNewMaster(newEntityID)
	//设置部长职位和旧部长职位
	entityMap := make(map[uint32]uint32, 0)
	entityMap[newEntityID] = consts.Master
	entityMap[entityID] = consts.General
	club.SetMemberPosition(entityMap)
	club.SyncEntity(1)
}

func (c *_Club) OnClubRatifyListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubRatifyListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	emClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := emClub.(*entity.Club)
	if club == nil {
		log.Error("俱乐部为空。", msgBody.ClubID)
		return
	}
	msgResponse := &gmsg.ClubRatifyListResponse{}
	msgResponse.Code = uint32(1)
	if club.MemberPosition(msgBody.EntityID) >= consts.Second_Master {
		msgResponse.Code = uint32(0)
		msgResponse.EntityID = msgBody.EntityID
		msgResponse.List = club.ReqList
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRatifyListResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) OnClubDelMembersRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubDelMembersRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	emClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	if emClub == nil {
		return
	}
	club := emClub.(*entity.Club)

	msgResponse := &gmsg.ClubDelMemberResponse{}
	msgResponse.Code = uint32(1)
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.DelEntityID = msgBody.DelEntityID
	if club.MemberPosition(msgBody.EntityID) < consts.Second_Master || msgResponse.EntityID == msgResponse.DelEntityID {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDelMembersResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	club.ReMoveMember(msgBody.DelEntityID)
	club.SyncEntity(1)
	c.toEntityEmailClub(false, club.ClubName, msgBody.DelEntityID)
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.DelEntityID)
	//用户离线，通知DB处理
	if tEntity == nil {
		delMember := &gmsg.ClubDelMembersToDB{}
		delMember.EntityID = msgBody.EntityID
		delMember.DelEntityID = msgBody.DelEntityID
		delMember.ClubID = msgBody.ClubID
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubDelMembersToDB), delMember, network.ServerType_DB)
		return
	}
	msgResponse.Code = uint32(0)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.ResetClubID()
	tEntityPlayer.SetExitClubUnixSec()
	tEntityPlayer.SyncEntity(1)

	//推送角色同步给前端
	Player.SyncClubToPlayer(msgBody.DelEntityID)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDelMembersResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) OnClubDelMembersResponse(msgEV *network.MsgBodyEvent) {
	msgResponse := &gmsg.ClubDelMemberResponse{}
	if err := msgEV.Unmarshal(msgResponse); err != nil {
		return
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDelMembersResponse, msgResponse, []uint32{msgResponse.EntityID})
}

// OnClubShopListRequest 获取俱乐部商店列表
func (c *_Club) OnClubShopListRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ClubShopListRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--ClubMgr--OnClubShopListRequest--msgEV.Unmarshal(req) err:", err)
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	resp := &gmsg.ClubShopListResponse{}
	resp.EntityID = req.EntityID
	resp.Code = uint32(1)

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.GetClubID())
	if emClub == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopListResponse, resp, []uint32{req.EntityID})
		return
	}

	club := emClub.(*entity.Club)

	resp.ClubShopList = make([]*gmsg.ClubShopItem, 0)
	for _, val := range club.ShopList {
		item := new(gmsg.ClubShopItem)
		item.ItemID = val.ItemID
		item.TableID = val.TableID
		item.MaxBuyNum = val.MaxBuyNum
		item.BuyNum = 0
		buyItem := tEntityPlayer.GetClubShopItemByTableID(val.TableID)
		if buyItem != nil {
			item.BuyNum = buyItem.BuyNum
		}
		item.IsUnlock = false
		if club.ClubLV >= val.Unlock {
			item.IsUnlock = true
		}
		resp.ClubShopList = append(resp.ClubShopList, item)
	}
	resp.Code = uint32(0)

	//广播初始化消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopListResponse, resp, []uint32{req.EntityID})
	return
}

// 查询道具
func (c *_Club) GetItemByItemID(tEntityPlayer *entity.EntityPlayer, itemID uint32) (item *entity.ShopItem, index int, err error) {
	for key, itemData := range tEntityPlayer.ClubShopBuyList {
		if itemData.ItemID == itemID {
			return &itemData, key, nil
		}
	}
	return nil, -1, errors.New("查找失败")
}

// 俱乐部商店购买物品
func (c *_Club) OnClubShopBuyRequest(msgEV *network.MsgBodyEvent) {
	req := &gmsg.ClubShopBuyRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(req.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.GetClubID())

	//购买返回消息
	resp := &gmsg.ClubShopBuyResponse{}
	resp.Code = uint32(1)
	//itemID异常或者俱乐部不存在
	if req.ItemID < 0 || emClub == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
		return
	}

	club := emClub.(*entity.Club)

	clubShopItem := club.GetClubShopItem(req.ItemID)
	//校验商品是否存在并已解锁
	if clubShopItem == nil || club.ClubLV < clubShopItem.Unlock {
		resp.Code = uint32(3)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
		return
	}

	buyNum := uint32(0)
	buyItem := tEntityPlayer.GetClubShopItemByTableID(clubShopItem.TableID)
	if buyItem != nil {
		buyNum = buyItem.BuyNum
	}
	if buyNum+req.Num > clubShopItem.MaxBuyNum {
		resp.Code = uint32(4)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
		return
	}

	rewardEntity := new(entity.RewardEntity)
	clubShopCfg := Table.GetClubShopCfgById(req.ItemID)
	if clubShopCfg == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
		return
	}

	if c.getGoldPriceType(clubShopCfg.PriceType, tEntityPlayer) < req.Num*clubShopCfg.Price {
		resp.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
		return
	}
	rewardEntity.ItemTableId = clubShopCfg.ItemTableID
	rewardEntity.Num = req.Num
	rewardEntity.ExpireTimeId = 0

	resParam := GetResParam(consts.SYSTEM_ID_CLUB_SHOP, consts.Buy)
	//发放物品
	addItemErr, _, _ := Backpack.BackpackAddOneItemAndSave(req.EntityID, *rewardEntity, *resParam)
	if addItemErr != nil {
		log.Error(addItemErr)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
		return
	}
	tEntityPlayer.BuyClubShopItemByTableID(req.ItemID, clubShopItem.TableID, req.Num)
	tEntityPlayer.SyncEntity(0)

	//TODO:如果允许创建订单失败丢失日志则直接发一条消息即可，反之
	//创建订单
	//ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Login_EnterGameRequest, msgBody, network.ServerType_DB)

	resp.Code = 0
	resp.ItemID = req.ItemID
	msgRewardEntity := new(gmsg.RewardInfo)
	stack.SimpleCopyProperties(msgRewardEntity, rewardEntity)
	resp.RewardInfo = msgRewardEntity

	Player.UpdatePlayerPropertyItem(req.EntityID, clubShopCfg.PriceType, int32(-req.Num*clubShopCfg.Price), *resParam)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubShopBuyResponse, resp, []uint32{req.EntityID})
}

// 判断货币是否足够
func (c *_Club) getGoldPriceType(pType uint32, player *entity.EntityPlayer) (gold uint32) {
	switch pType {
	case consts.Gold:
		gold = player.NumGold
	case consts.Diamond:
		gold = player.NumStone
	case consts.ClubGold:
		gold = player.ClubNumGold
	}
	return
}

// 查询俱乐部红包列表
func (c *_Club) OnClubRedEnvelopeListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubRedEnvelopeListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.ClubRedEnvelopeListResponse{}
	msgResponse.EntityID = msgBody.EntityID

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.ClubId == 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	if club == nil {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	msgResponse.Code = 0
	msgList := make([]*gmsg.ClubRedEnvelopeItem, 0)
	// 大于24小时不显示
	redEnvelopeTamp := time.Now().Unix() - 86400
	for _, vl := range club.RedEnvelopeList {
		if vl.SendTime < redEnvelopeTamp {
			continue
		}
		data := new(gmsg.ClubRedEnvelopeItem)
		stack.SimpleCopyProperties(data, vl)
		data.RedEnvelopeID = vl.RedEnvelopeID.Hex()
		data.ClubRedEnvelopeRecordList = make([]*gmsg.ClubRedEnvelopeRecordItem, 0)
		for _, val := range vl.ClubRedEnvelopeRecordList {
			recordItem := new(gmsg.ClubRedEnvelopeRecordItem)
			stack.SimpleCopyProperties(recordItem, val)
			data.ClubRedEnvelopeRecordList = append(data.ClubRedEnvelopeRecordList, recordItem)
		}
		msgList = append(msgList, data)
	}
	msgResponse.ClubRedEnvelopeList = msgList
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 查询俱乐部红包列表
func (c *_Club) ClubRedEnvelopeListRequest(tEntityPlayer *entity.EntityPlayer) []*gmsg.ClubRedEnvelopeItem {
	msgList := make([]*gmsg.ClubRedEnvelopeItem, 0)
	if tEntityPlayer.ClubId == 0 {
		return msgList
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	if emClub == nil {
		return msgList
	}
	club := emClub.(*entity.Club)
	if club == nil {
		return msgList
	}

	// 大于24小时不显示
	redEnvelopeTamp := time.Now().Unix() - 86400
	for _, vl := range club.RedEnvelopeList {
		if vl.SendTime < redEnvelopeTamp {
			continue
		}
		data := new(gmsg.ClubRedEnvelopeItem)
		stack.SimpleCopyProperties(data, vl)
		data.RedEnvelopeID = vl.RedEnvelopeID.Hex()
		data.ClubRedEnvelopeRecordList = make([]*gmsg.ClubRedEnvelopeRecordItem, 0)
		for _, val := range vl.ClubRedEnvelopeRecordList {
			recordItem := new(gmsg.ClubRedEnvelopeRecordItem)
			stack.SimpleCopyProperties(recordItem, val)
			data.ClubRedEnvelopeRecordList = append(data.ClubRedEnvelopeRecordList, recordItem)
		}
		msgList = append(msgList, data)
	}
	return msgList
}

// 俱乐部发送红包
func (c *_Club) OnClubSendRedEnvelopeRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubSendRedEnvelopeRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.ClubSendRedEnvelopeResponse{}
	msgResponse.EntityID = msgBody.EntityID

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.ClubId == 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSendRedEnvelopeResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	if club == nil || msgBody.TotalSendNum > msgBody.SendCoinNum {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSendRedEnvelopeResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if msgBody.TotalSendNum > uint32(len(club.GetMembers())) {
		msgResponse.Code = uint32(3)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSendRedEnvelopeResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	// 扣减加上手续费
	redEnvelopeCommission := msgBody.SendCoinNum / consts.RedEnvelopeCommission
	sendRedEnvelopeCost := int32(msgBody.SendCoinNum + redEnvelopeCommission)
	if tEntityPlayer.NumGold < uint32(sendRedEnvelopeCost) {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSendRedEnvelopeResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	msgResponse.Code = 0

	redEnvelope := new(entity.RedEnvelope)
	redEnvelope.RedEnvelopeID = bson.NewObjectId()
	redEnvelope.ClubRedEnvelopeRecordList = make([]entity.RedEnvelopeRecord, 0)
	redEnvelope.SendTime = time.Now().Unix()
	redEnvelope.SendCoinNum = msgBody.SendCoinNum
	redEnvelope.TotalSendNum = msgBody.TotalSendNum
	redEnvelope.SendEnvelopeEntityID = msgBody.EntityID
	redEnvelope.SendEnvelopeEntityName = tEntityPlayer.PlayerName
	redEnvelope.SendEnvelopeEntityAvatarID = tEntityPlayer.PlayerIcon
	redEnvelope.SendEnvelopeEntityIconFrameID = tEntityPlayer.IconFrame
	redEnvelope.BlessWorld = msgBody.BlessWorld
	club.RedEnvelopeList = append(club.RedEnvelopeList, *redEnvelope)
	club.SyncEntity(1)
	c.BuildNewRedPack(redEnvelope)

	data := new(gmsg.ClubRedEnvelopeItem)
	stack.SimpleCopyProperties(data, redEnvelope)

	resParam := GetResParam(consts.SYSTEM_ID_RED_Envelope, consts.SendRedEnvelope)
	// 扣减金币
	Player.UpdatePlayerPropertyItem(msgBody.EntityID, consts.Gold, -sendRedEnvelopeCost, *resParam)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSendRedEnvelopeResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 同步俱乐部红包
func (c *_Club) ClubRedEnvelopeSync(entityID uint32, data *gmsg.ClubRedEnvelopeItem) {
	msgResponse := &gmsg.ClubRedEnvelopeSync{}
	msgResponse.EntityID = entityID
	msgResponse.ClubRedEnvelopeItem = data
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeSync, msgResponse, []uint32{entityID})
}

// 创建红包
func (c *_Club) BuildNewRedPack(redEnvelope *entity.RedEnvelope) {
	c.RedPack[redEnvelope.RedEnvelopeID.Hex()] = NewRedPack(redEnvelope.RedEnvelopeID.Hex(), int(redEnvelope.TotalSendNum), int(redEnvelope.SendCoinNum), redEnvelope.SendTime)
	log.Info("-->newredpack--->:", c.RedPack[redEnvelope.RedEnvelopeID.Hex()])
}

// 获取红包
func (c *_Club) GetRedPack(redEnvelopeID string) *RedPack {
	if len(c.RedPack) == 0 {
		return nil
	}

	if rd, ok := c.RedPack[redEnvelopeID]; ok {
		return rd
	}
	return nil
}

// 俱乐部打开红包
func (c *_Club) OnClubRedEnvelopeOpenRequest(msgEV *network.MsgBodyEvent) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	msgBody := &gmsg.ClubRedEnvelopeOpenRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("OnClubRedEnvelopeOpenRequest", msgBody)
	msgResponse := &gmsg.ClubRedEnvelopeOpenResponse{}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.ClubId == 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	if club == nil {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	redEnvelope := club.GetRedEnvelopeFromRedEnvelopeID(msgBody.RedEnvelopeID)
	if redEnvelope == nil || redEnvelope.SendTime < time.Now().Unix()-86400 {
		msgResponse.Code = uint32(3)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	hadOpen := club.CheckHadOpenRedEnvelopeByRedEnvelope(redEnvelope, msgBody.EntityID)
	if hadOpen {
		msgResponse.Code = uint32(4)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	redpack := c.GetRedPack(msgBody.RedEnvelopeID)
	amount, NumDelivered, AmountDelivered := redpack.OpenRedPack()
	if redpack == nil || amount == 0 {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	redEnvelopeRecord := new(entity.RedEnvelopeRecord)
	redEnvelopeRecord.EntityID = msgBody.EntityID
	redEnvelopeRecord.EntityName = tEntityPlayer.PlayerName
	redEnvelopeRecord.GetCoinNum = uint32(amount)
	redEnvelopeRecord.GetTime = time.Now().Unix()
	redEnvelope.ClubRedEnvelopeRecordList = append(redEnvelope.ClubRedEnvelopeRecordList, *redEnvelopeRecord)
	redEnvelope.NumDelivered = uint32(NumDelivered)
	redEnvelope.AmountDelivered = uint32(AmountDelivered)
	club.SaveClubRedEnvelope(redEnvelope)
	club.SyncEntity(1)

	msgResponse.Code = 0
	msgResponse.RedEnvelopeID = msgBody.RedEnvelopeID
	msgResponse.CoinNum = redEnvelopeRecord.GetCoinNum

	data := new(gmsg.ClubRedEnvelopeItem)
	stack.SimpleCopyProperties(data, redEnvelope)
	data.ClubRedEnvelopeRecordList = make([]*gmsg.ClubRedEnvelopeRecordItem, 0)
	for _, val := range redEnvelope.ClubRedEnvelopeRecordList {
		recordItem := new(gmsg.ClubRedEnvelopeRecordItem)
		stack.SimpleCopyProperties(recordItem, val)
		data.ClubRedEnvelopeRecordList = append(data.ClubRedEnvelopeRecordList, recordItem)
	}
	msgResponse.ClubRedEnvelopeItem = data

	resParam := GetResParam(consts.SYSTEM_ID_RED_Envelope, consts.RedEnvelopeOpen)
	Player.UpdatePlayerPropertyItem(msgBody.EntityID, consts.Gold, int32(msgResponse.CoinNum), *resParam)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeOpenResponse, msgResponse, []uint32{msgBody.EntityID})
	c.ClubRedEnvelopeSync(msgBody.EntityID, data)
}

// 俱乐部红包记录
func (c *_Club) OnClubRedEnvelopeRecordListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubRedEnvelopeRecordListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("-->OnClubRedEnvelopeRecordListRequest-->begin-->", msgBody)
	msgResponse := &gmsg.ClubRedEnvelopeRecordListResponse{}
	msgResponse.EntityID = msgBody.EntityID

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.ClubId == 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeRecordListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	if club == nil {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeRecordListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	msgResponse.Code = 0
	list, err := club.GetRedEnvelopeRecordFromRedEnvelopeID(msgBody.RedEnvelopeID)
	fmt.Println("GetRedEnvelopeRecordFromRedEnvelopeID", list)
	if err != nil {
		msgResponse.Code = uint32(3)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeRecordListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}
	msgResponse.ClubRedEnvelopeRecordList = make([]*gmsg.ClubRedEnvelopeRecordItem, 0)
	for _, vl := range list {
		data := new(gmsg.ClubRedEnvelopeRecordItem)
		stack.SimpleCopyProperties(data, vl)
		msgResponse.ClubRedEnvelopeRecordList = append(msgResponse.ClubRedEnvelopeRecordList, data)
	}
	log.Info("-->OnClubRedEnvelopeRecordListRequest-->begin-->", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRedEnvelopeRecordListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 俱乐部打卡请求
func (c *_Club) OnClubDailySignInRequest(msgEV *network.MsgBodyEvent) {

	msgBody := &gmsg.ClubDailySignInRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.ClubDailySignInResponse{}
	msgResponse.Code = uint32(1)

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	log.Info("-->OnClubDailySignInRequest-->begin-->", msgBody)
	if tEntityPlayer.GetClubID() == 0 || club == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDailySignInResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if tEntityPlayer.IsClubDailySignIn() {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDailySignInResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if tEntityPlayer.IsClubReFreshUnix() {
		score := Table.GetClubTaskScoreFromCondID(consts.DailyTask)
		if score == 0 {
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDailySignInResponse, msgResponse, []uint32{msgBody.EntityID})
			return
		}
		tEntityPlayer.ClubDailySignIn()
		resParam := GetResParam(consts.SYSTEM_ID_CLUB_TASk, consts.ClubDailySign)
		c.ToUpdateEmPlayerClubTask(msgBody.EntityID, score, tEntityPlayer.GetClubID(), consts.DailyTask, 1, *resParam)

		msgResponse.Code = uint32(0)
	}

	log.Info("-->OnClubDailySignInRequest-->end-->", msgResponse)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubDailySignInResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) ToUpdateEmPlayerClubTask(EntityID, score, clubID, condID, progress uint32, resParam entity.ResParam) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	c.updateEmPlayerClubTask(EntityID, score, clubID, condID, progress, resParam)
	tEntityPlayer.SyncEntity(1)
	emClub := Entity.EmClub.GetEntityByID(clubID)
	if emClub == nil {
		return
	}
	club := emClub.(*entity.Club)
	club.SyncEntity(1)
}

func (c *_Club) updateEmPlayerClubTask(EntityID, score, clubID, condID, progress uint32, resParam entity.ResParam) {
	if clubID == 0 {
		return
	}
	//更新个人俱乐部任务
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if condID == consts.DailyTask {
		score = progress
	}
	tEntityPlayer.UpdateClubTaskFromConditionID(condID, score)
	tEntityPlayer.AddClubActiveValue(score)

	Player.UpdatePlayerPropertyItem(EntityID, consts.ClubGold, int32(score), resParam)

	if condID == 0 {
		return
	}
	//更新俱乐部
	emClub := Entity.EmClub.GetEntityByID(clubID)
	if emClub == nil {
		log.Error(fmt.Sprintf("俱乐部为空。", clubID))
		return
	}
	club := emClub.(*entity.Club)
	club.AddClubActiveValueNumExp(score)
	cfg := Table.GetClubCfg(club.ClubLV)
	if cfg == nil {
		log.Error(errors.New("俱乐部对应等级为空。"))
		return
	}
	// 达到最大等级不用升级
	if cfg.Exp == 0 {
		return
	}
	if club.NumExp >= cfg.Exp {
		nextCfg := Table.GetClubCfg(club.ClubLV + 1)
		club.UpgradeClub(nextCfg.Num)
	}
	club.UpdateMemberActive(EntityID, score)

	//推送俱乐部任务更新
	c.SyncClubTaskList(tEntityPlayer)
}

// 俱乐部任务表
func (c *_Club) OnClubTaskListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubTaskListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnClubTaskListRequest-->", msgBody)
	msgResponse := &gmsg.ClubTaskListResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.ClubActiveValue = 0
	msgResponse.MyActiveValue = 0
	msgResponse.ClubProgressRewardList = nil
	msgResponse.ClubTaskProgressList = nil
	msgResponse.ClubTaskList = nil

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)

	if tEntityPlayer.GetClubID() == 0 || club == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubTaskListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	msgResponse.ClubActiveValue = club.GetClubActiveValue()
	msgResponse.MyActiveValue = tEntityPlayer.GetPlayerClubActiveValue()

	msgResponse.ClubProgressRewardList = make([]*gmsg.ClubProgressReward, 0)
	for _, vl := range tEntityPlayer.ClubAttribute.ClubProgressRewardList {
		clubProgressReward := new(gmsg.ClubProgressReward)
		stack.SimpleCopyProperties(clubProgressReward, vl)
		msgResponse.ClubProgressRewardList = append(msgResponse.ClubProgressRewardList, clubProgressReward)
	}

	msgResponse.ClubTaskProgressList = make([]*gmsg.ClubTaskProgress, 0)
	for _, vl := range tEntityPlayer.ClubAttribute.ClubTaskProgressList {
		clubTaskProgress := new(gmsg.ClubTaskProgress)
		stack.SimpleCopyProperties(clubTaskProgress, vl)
		task := Table.GetClubTaskProgressCfg(vl.ProgressID)
		if task == nil || len(task.Rewards) < 1 {
			log.Error("nil task:", vl.ProgressID)
			continue
		}
		clubTaskProgress.ItemTableId = Table.GetClubTaskProgressCfg(vl.ProgressID).Rewards[0]
		clubTaskProgress.Num = Table.GetClubTaskProgressCfg(vl.ProgressID).Rewards[1]
		msgResponse.ClubTaskProgressList = append(msgResponse.ClubTaskProgressList, clubTaskProgress)
	}

	msgResponse.ClubTaskList = make([]*gmsg.ClubWeekTask, 0)
	for _, vl := range tEntityPlayer.ClubAttribute.ClubTaskList {
		clubWeekTask := new(gmsg.ClubWeekTask)
		stack.SimpleCopyProperties(clubWeekTask, vl)
		clubWeekTask.ClubDailyTaskList = make([]*gmsg.ClubDailyTask, 0)
		if len(vl.ClubDailyTaskList) > 0 {
			for _, vd := range vl.ClubDailyTaskList {
				clubDailyTask := new(gmsg.ClubDailyTask)
				stack.SimpleCopyProperties(clubDailyTask, vd)
				clubWeekTask.ClubDailyTaskList = append(clubWeekTask.ClubDailyTaskList, clubDailyTask)
			}
		}
		msgResponse.ClubTaskList = append(msgResponse.ClubTaskList, clubWeekTask)
	}

	log.Info("-->OnClubTaskListRequest-->", msgResponse)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubTaskListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 俱乐部任务表
func (c *_Club) GetClubTaskList(entityID uint32) *gmsg.ClubTaskListResponse {
	msgResponse := &gmsg.ClubTaskListResponse{}
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return nil
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	msgResponse.EntityID = entityID

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	if emClub == nil {
		return msgResponse
	}
	club := emClub.(*entity.Club)

	msgResponse.ClubActiveValue = club.GetClubActiveValue()
	msgResponse.MyActiveValue = tEntityPlayer.GetPlayerClubActiveValue()

	msgResponse.ClubProgressRewardList = make([]*gmsg.ClubProgressReward, 0)
	for _, vl := range tEntityPlayer.ClubAttribute.ClubProgressRewardList {
		clubProgressReward := new(gmsg.ClubProgressReward)
		stack.SimpleCopyProperties(clubProgressReward, vl)
		msgResponse.ClubProgressRewardList = append(msgResponse.ClubProgressRewardList, clubProgressReward)
	}

	msgResponse.ClubTaskProgressList = make([]*gmsg.ClubTaskProgress, 0)
	for _, vl := range tEntityPlayer.ClubAttribute.ClubTaskProgressList {
		clubTaskProgress := new(gmsg.ClubTaskProgress)
		stack.SimpleCopyProperties(clubTaskProgress, vl)
		clubTaskProgress.ItemTableId = Table.GetClubTaskProgressCfg(vl.ProgressID).Rewards[0]
		clubTaskProgress.Num = Table.GetClubTaskProgressCfg(vl.ProgressID).Rewards[1]
		msgResponse.ClubTaskProgressList = append(msgResponse.ClubTaskProgressList, clubTaskProgress)
	}

	msgResponse.ClubTaskList = make([]*gmsg.ClubWeekTask, 0)
	for _, vl := range tEntityPlayer.ClubAttribute.ClubTaskList {
		clubWeekTask := new(gmsg.ClubWeekTask)
		stack.SimpleCopyProperties(clubWeekTask, vl)
		clubWeekTask.ClubDailyTaskList = make([]*gmsg.ClubDailyTask, 0)
		if len(vl.ClubDailyTaskList) > 0 {
			for _, vd := range vl.ClubDailyTaskList {
				clubDailyTask := new(gmsg.ClubDailyTask)
				stack.SimpleCopyProperties(clubDailyTask, vd)
				clubWeekTask.ClubDailyTaskList = append(clubWeekTask.ClubDailyTaskList, clubDailyTask)
			}
		}
		msgResponse.ClubTaskList = append(msgResponse.ClubTaskList, clubWeekTask)
	}

	//log.Info("-->SyncClubTaskList-->", msgResponse)
	return msgResponse
}

func (c *_Club) SyncClubTaskList(tEntityPlayer *entity.EntityPlayer) {
	resList := c.GetClubTaskList(tEntityPlayer.EntityID)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubTaskListSync, resList, []uint32{tEntityPlayer.EntityID})
}

// 俱乐部赞助资金
func (c *_Club) OnClubSupportRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubSupportRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.ClubSupportResponse{}
	msgResponse.Code = uint32(0)

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	if !tEntityPlayer.IsClubReFreshUnix() {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSupportResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	if tEntityPlayer.GetNumStone() < consts.DeductStone {
		msgResponse.Code = uint32(2)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSupportResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	score := Table.GetClubTaskScoreFromCondID(consts.SupportFunds)
	if score == 0 {
		msgResponse.Code = uint32(1)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSupportResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	resParam := GetResParam(consts.SYSTEM_ID_CLUB_TASk, consts.ClubSupportFunds)
	// 扣减钻石
	Player.UpdatePlayerPropertyItem(msgBody.EntityID, consts.Diamond, int32(-consts.DeductStone), *resParam)
	//更新赞助任务
	c.ToUpdateEmPlayerClubTask(msgBody.EntityID, score*20, tEntityPlayer.GetClubID(), consts.SupportFunds, 1, *resParam)
	//更新消费任务
	c.UpdateConsumeTask(msgBody.EntityID, consts.DeductStone, tEntityPlayer.ClubId, *resParam)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubSupportResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 领取俱乐部活跃奖励，退出俱乐部会重置这个数据
func (c *_Club) OnClubClaimTaskProgressRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubClaimTaskProgressRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.ClubClaimTaskProgressResponse{}
	msgResponse.Code = uint32(1)
	msgResponse.ClubProgressID = msgBody.ClubProgressID
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)

	// 俱乐部为空或者任务过期不能领取
	if tEntityPlayer.GetClubID() == 0 || club == nil || !tEntityPlayer.IsClubReFreshUnix() {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubClaimTaskProgressResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	cfg := Table.GetClubProgressRewardCfgMap(msgBody.ClubProgressID)
	if cfg == nil {
		return
	}
	if club.ClubActiveValue >= cfg.Progress && !tEntityPlayer.IsClaimClubProgressReward(msgBody.ClubProgressID) {
		tEntityPlayer.ClaimClubProgressReward(msgBody.ClubProgressID)
		tEntityPlayer.SyncEntity(1)
		rewardEntity := new(entity.RewardEntity)
		rewardEntity.ItemTableId = cfg.Rewards[0]
		rewardEntity.Num = cfg.Rewards[1]
		rewardEntity.ExpireTimeId = 0

		resParam := GetResParam(consts.SYSTEM_ID_CLUB_TASk, consts.Reward)
		//发放奖励
		err, _ := Backpack.BackpackAddItemListAndSave(msgBody.EntityID, []entity.RewardEntity{*rewardEntity}, *resParam)
		if err != nil {
			log.Error(err)
		}
		log.Info("-->OnClubClaimTaskProgressRequest-->success!", "-->EntityID", msgBody.EntityID)
		msgResponse.Code = uint32(0)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubClaimTaskProgressResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 领取我的活跃值奖励
func (c *_Club) OnClaimMyClubTaskProgressRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClaimMyClubTaskProgressRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.ClaimMyClubTaskProgressResponse{}
	msgResponse.Code = uint32(1)
	msgResponse.ProgressID = msgBody.ProgressID
	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	log.Info("-->OnClaimMyClubTaskProgressRequest-->begin-->", msgBody)
	cfg := Table.GetClubTaskProgressCfg(msgBody.ProgressID)
	if cfg == nil {
		return
	}

	if tEntityPlayer.GetPlayerClubActiveValue() >= cfg.Progress && !tEntityPlayer.IsClaimMyClubTaskProgressReward(msgBody.ProgressID) && tEntityPlayer.IsClubReFreshUnix() {
		tEntityPlayer.ClaimMyClubTaskProgressReward(msgBody.ProgressID)
		tEntityPlayer.SyncEntity(1)
		rewardEntity := new(entity.RewardEntity)
		rewardEntity.ItemTableId = cfg.Rewards[0]
		rewardEntity.Num = cfg.Rewards[1]
		rewardEntity.ExpireTimeId = 0

		resParam := GetResParam(consts.SYSTEM_ID_CLUB_TASk, consts.Reward)
		//发放奖励
		err, _ := Backpack.BackpackAddItemListAndSave(msgBody.EntityID, []entity.RewardEntity{*rewardEntity}, *resParam)
		if err != nil {
			log.Error(err)
		}
		log.Info("-->OnClaimMyClubTaskProgressRequest-->success!", "-->EntityID", msgBody.EntityID)
		msgResponse.Code = uint32(0)
	}
	log.Info("-->OnClaimMyClubTaskProgressRequest-->end-->", msgResponse)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClaimMyClubTaskProgressResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 俱乐部评级列表请求
func (c *_Club) OnClubRateListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubRateListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)

	msgResponse := &gmsg.ClubRateListResponse{}
	if tEntityPlayer.GetClubID() == 0 || club == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRateListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	msgResponse.MyClubRate = club.ClubRate
	msgResponse.MyClubScore = club.ClubScore
	msgResponse.ClubRateList = make([]*gmsg.ClubRateInfo, 0)
	num := uint32(1)
	rate := club.ClubRate
	if club.ClubRate == consts.ClubRateSPlus {
		rate = consts.ClubRateS
	}
	if list, ok := c.RankList[rate]; ok {
		for rank, vl := range list {
			a := new(gmsg.ClubRateInfo)
			a.ClubID = vl.ClubID
			a.ClubBadge = vl.ClubBadge
			a.MasterEntityID = vl.MasterEntityID
			a.ClubRate = vl.ClubRate
			a.ClubScore = vl.ClubScore
			a.ClubName = vl.ClubName
			a.RateRank = num
			a.RankTags = c.getRankTags(rate, vl.ClubScore, rank)
			if vl.ClubID == tEntityPlayer.GetClubID() {
				msgResponse.MyClubRank = num
			}
			msgResponse.ClubRateList = append(msgResponse.ClubRateList, a)
			num++
		}
	}

	if len(msgResponse.ClubRateList) > consts.MaxClubRateNum {
		msgResponse.ClubRateList = msgResponse.ClubRateList[0:consts.MaxClubRateNum]
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubRateListResponse, msgResponse, []uint32{msgBody.EntityID})
}

func (c *_Club) getRankTags(clubRate, clubScore uint32, rank int) uint32 {
	tags := uint32(0)
	if clubRate < consts.ClubRateA {
		resRate := c.statisticsBeforeClubRateB(clubRate, clubScore)
		tags = c.getTags(clubRate, resRate)
	} else if clubRate == consts.ClubRateA {
		resRate := c.statisticsClubRateA(clubRate, clubScore, rank)
		tags = c.getTags(clubRate, resRate)
	} else if clubRate >= consts.ClubRateS {
		resRate := c.statisticsClubRateS(clubRate, clubScore, rank)
		tags = c.getTags(clubRate, resRate)
	}

	return tags
}

func (c *_Club) getTags(clubRate, resRate uint32) uint32 {
	tagsAsc, tagsDesc := consts.ClubUpgradeRate, consts.ClubReduceRate
	tags := consts.ClubKeepRate
	if resRate > clubRate {
		tags = tagsAsc
	} else if resRate < clubRate {
		tags = tagsDesc
	}
	return tags
}

// 殿堂俱乐部评级列表请求
func (c *_Club) OnPalaceClubRateListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PalaceClubRateListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.PalaceClubRateListResponse{}
	msgResponse.ClubRateList = make([]*gmsg.ClubRateInfo, 0)
	num := uint32(1)
	for _, vl := range c.PalaceRankList {
		a := new(gmsg.ClubRateInfo)
		a.ClubID = vl.ClubID
		a.ClubBadge = vl.ClubBadge
		a.MasterEntityID = vl.MasterEntityID
		a.ClubRate = vl.ClubRate
		a.ClubScore = vl.TotalScore
		a.ClubName = vl.ClubName
		a.RateRank = num
		msgResponse.ClubRateList = append(msgResponse.ClubRateList, a)
		num++
		if num > consts.MaxPalaceClubNum {
			break
		}
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_PalaceClubRateListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 俱乐部盈利排行榜
func (c *_Club) OnClubProfitGoldListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubProfitGoldListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	emClub := Entity.EmClub.GetEntityByID(tEntityPlayer.ClubId)
	club := emClub.(*entity.Club)
	msgResponse := &gmsg.ClubProfitGoldListResponse{}
	msgResponse.ClubList = make([]*gmsg.ClubProfitGold, 0)
	if tEntityPlayer.GetClubID() == 0 || club == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubProfitGoldListResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	num := uint32(1)
	myClub := new(gmsg.ClubProfitGold)
	for _, vl := range c.ProfitGoldList {
		a := new(gmsg.ClubProfitGold)
		stack.SimpleCopyProperties(a, vl)
		a.Rank = num
		if tEntityPlayer.GetClubID() == vl.ClubID {
			stack.SimpleCopyProperties(myClub, a)
		}
		msgResponse.ClubList = append(msgResponse.ClubList, a)
		num++
	}

	if len(msgResponse.ClubList) > consts.MaxProfitNum {
		msgResponse.ClubList = msgResponse.ClubList[0:consts.MaxProfitNum]
	}
	msgResponse.ClubList = append(msgResponse.ClubList, myClub)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_ClubProfitGoldListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 更新消息充值任务
func (c *_Club) UpdateConsumeTask(entityID, num, clubID uint32, resParam entity.ResParam) {
	if clubID == 0 {
		return
	}
	c.ToUpdateEmPlayerClubTask(entityID, num, clubID, consts.ConsumeTask, num, resParam)
}

func (c *_Club) ClubTaskInit(EntityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	tEntityPlayer.ClubAttribute.ClubProgressRewardList = Table.GetClubProgressRewardList()

	if tEntityPlayer.ClubAttribute.ClubReFreshUnix < tools.GetThisWeekSaturday() {
		tEntityPlayer.ClubAttribute.ClubTaskProgressList = Table.GetClubTaskProgressList()
		tEntityPlayer.ClubAttribute.ClubTaskList = Table.GetClubTaskList()
		tEntityPlayer.ClubAttribute.ClubActiveValue = 0
	}

	tEntityPlayer.UpdateClubReFreshUnix()
}

// 对战结算
func (c *_Club) UpdateClubFromGameResult(data UpdateClubData, resParam entity.ResParam) {
	ClubMutex.Lock()
	defer ClubMutex.Unlock()
	tEntity := Entity.EmPlayer.GetEntityByID(data.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	if tEntityPlayer.GetClubID() > 0 {
		c.updateBattleClubProgressReward(data.GameType, data.RoomType, data.Result, data.EntityID, tEntityPlayer.GetClubID(), resParam)
		c.updateBattleClubScore(data.GameType, data.RoomType, data.Result, tEntityPlayer.GetClubID(), data.Gold, data.SettlementType)
	}
	//更新宝箱数据
	BoxMr.SettlementAddBox(data, resParam)
}

func (c *_Club) updateBattleClubScore(gameType, roomType, result, clubID, gold, SettlementType uint32) {
	key := c.getBattleClubScoreKey(gameType, roomType)
	scfg := Table.GetBattleClubScoreRewardCfgMap(key)
	if scfg == nil {
		return
	}
	emClub := Entity.EmClub.GetEntityByID(clubID)
	if emClub == nil {
		return
	}
	club := emClub.(*entity.Club)
	score := uint32(0)
	if result == consts.RESULT_VICTORY {
		score = scfg.WinScore
	} else if result == consts.RESULT_TRANSPORT {
		score = scfg.LoseScore
		gold = 0
	}

	if SettlementType == consts.SETTLEMENT_TYPE_SURRENDER {
		gold = 0
		score = 0
	}

	club.AddProfitGold(gold)
	club.AddClubScore(score)
	club.SyncEntity(1)
}

func (c *_Club) updateBattleClubProgressReward(gameType, roomType, result, EntityID, clubID uint32, resParam entity.ResParam) {
	key := c.getBattleClubScoreKey(gameType, roomType)
	acfg := Table.GetBattleClubProgressRewardCfgMap(key)
	if acfg == nil {
		log.Error("GetBattleClubProgressRewardCfgMap is err")
		return
	}

	score := uint32(0)
	if result == consts.RESULT_VICTORY {
		score = acfg.WinReward
	} else if result == consts.RESULT_TRANSPORT {
		score = acfg.LoseReward
	}

	c.updateEmPlayerClubTask(EntityID, score, clubID, consts.BattleTask, 1, resParam)
}

func (c *_Club) getBattleClubScoreKey(gameType, roomType uint32) string {
	return fmt.Sprintf("%d_%d", gameType, roomType)
}

func (c *_Club) sendRewardEmail(club *entity.Club, index int) {
	rateCfg := Table.GetClubRateRewardCfg(club.ClubRate)
	if rateCfg == nil {
		return
	}
	itemRewardEntity := new(gmsg.RewardInfo)
	if club.ClubRate >= consts.ClubRateA && len(rateCfg.ItemReward) > 0 {
		itemRewardEntity.ItemTableId = rateCfg.ItemReward[0][0]
		itemRewardEntity.Num = rateCfg.ItemReward[0][1]
		itemRewardEntity.ExpireTimeId = rateCfg.ItemReward[0][2]
	}
	for _, vl := range club.GetMembers() {
		Tittle := Table.GetConstTextFromID(11, DefaultText)
		Content := tools.StringReplace(Table.GetConstTextFromID(12, DefaultText), "s", rateCfg.LevelSymbol)
		email := new(gmsg.Email)
		email.EmailID = Player.GetMaxUuid(vl.EntityID)
		email.Date = tools.GetTimeByTimeStamp(time.Now().Unix())
		email.StateReward = false
		email.Tittle = Tittle
		email.Content = Content
		email.IsRewardEmail = true
		email.RewardList = make([]*gmsg.RewardInfo, 0)

		emailRewardEntity := new(gmsg.RewardInfo)
		emailRewardEntity.ItemTableId = rateCfg.Reward[0]
		emailRewardEntity.Num = rateCfg.Reward[1]
		email.RewardList = append(email.RewardList, emailRewardEntity)
		if itemRewardEntity.ItemTableId > 0 {
			email.RewardList = append(email.RewardList, itemRewardEntity)
		}
		Email.AddEmail(vl.EntityID, email)
		//更新称号
		if club.ClubRate == consts.ClubRateA {
			if index < consts.ClubItemRewardRank {
				ConditionalMr.SyncConditional(vl.EntityID, []consts.ConditionData{{consts.ATop10Club, 1, false}})
			}
		} else if club.ClubRate == consts.ClubRateS {
			ConditionalMr.SyncConditional(vl.EntityID, []consts.ConditionData{{consts.SClub, 1, false}})
		} else if club.ClubRate == consts.ClubRateSPlus {
			ConditionalMr.SyncConditional(vl.EntityID, []consts.ConditionData{{consts.SMaxClub, 1, false}})
		}
	}
}

// 发送奖励
func (c *_Club) ClubSendItemRewardEmail() {
	for rate := consts.ClubRateE; rate <= consts.ClubRateS; rate++ {
		list, ok := c.RankList[rate]
		if !ok {
			continue
		}

		for index, val := range list {
			emClub := Entity.EmClub.GetEntityByID(val.ClubID)
			if emClub == nil {
				continue
			}
			club := emClub.(*entity.Club)
			c.sendRewardEmail(club, index)
		}
		log.Info("-->ClubSendItemRewardEmail-->end-->", rate)
	}
}
