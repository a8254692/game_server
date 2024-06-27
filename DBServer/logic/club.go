package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/DBServer/initialize/consts"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"
)

type _Club struct {
	lock      sync.RWMutex
	ClubNow   int //当前自增长ID
	ClubCount int //当前总数
}

var Club _Club

var RangeNum = 100

func (c *_Club) Init() {
	c.ClubNow = 10000
	//c.testCreateClub()
	event.OnNet(gmsg.MsgTile_Hall_ClubListRequest, reflect.ValueOf(c.OnClubListDBRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubCreateRequest, reflect.ValueOf(c.OnClubCreateDBRequest))
	event.OnNet(gmsg.MsgTile_Hall_ClubRatifyJoinRequest, reflect.ValueOf(c.OnClubRatifyJoinDBRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubDelMembersToDB), reflect.ValueOf(c.OnClubDelMembersDBRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_IN_ClubTOP10_DB_Data_Request), reflect.ValueOf(c.OnClubTop10DBRequest))

	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubRatifyJoinToDB), reflect.ValueOf(c.OnClubRatifyJoinDBRequest))
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncEntityClub), reflect.ValueOf(c.GetGameSyncClubRequest))

	//内部测试协议
	event.OnNet(gmsg.MsgTile(gmsg.InternalMsgTile_In_BatchCreateClubRequest), reflect.ValueOf(c.OnBatchReqClubRequest))

	timer.AddTimer(c, "OnMongoDBInItComplete", 200, false)
}

// 接收game请求同步club数据
func (c *_Club) GetGameSyncClubRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SyncEntityClubNoticeDB{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	c.getEntityClubArgs()
}

func (c *_Club) OnMongoDBInItComplete() {
	count, _ := DBConnect.GetTableCount(consts.COLLECTION_CLUB)
	c.ClubCount = count
	c.ClubNow = c.ClubNow + count
	log.Info("-->Club Init Complete, ClubNow:", c.ClubNow)
}

// 分批推送俱乐部数据
func (c *_Club) getEntityClubArgs() {
	count := math.Ceil(float64(Entity.EmClub.EntityCount) / float64(RangeNum))
	args := make(map[uint32]uint32, 0)
	for i := 1; i <= int(count); i++ {
		tEntityClubArgs, num := make([]entity.Club, 0), 0
		for _, ValueClub := range Entity.EmClub.EntityMap {
			tEntityClub := ValueClub.(*entity.Club)
			if _, ok := args[tEntityClub.ClubID]; ok {
				continue
			}
			tEntityClubArgs = append(tEntityClubArgs, *tEntityClub)
			args[tEntityClub.ClubID] = 1
			num++
			if num == RangeNum {
				break
			}
		}

		buf, _ := stack.StructToBytes_Gob(tEntityClubArgs)
		if len(buf) < 1 {
			continue
		}
		c.syncEntityClubNoticeDB(buf)
	}
}

func (c *_Club) syncEntityClubNoticeDB(buf []byte) {
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_SyncEntityClubDBToGame), buf, network.ServerType_Game)
}

func (c *_Club) OnClubListDBRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	clubList, count := c.getClubList(*msgBody.ClubID, *msgBody.PlayerLV, msgBody.PageNum, msgBody.PageSize, *msgBody.IsJoinLevel)

	msgResponse := &gmsg.ClubListResponse{}
	msgResponse.Code = 0
	msgResponse.List = clubList
	msgResponse.Total = uint32(count)
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.PageSize = msgBody.PageSize
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_ClubListResponse, msgResponse, network.ServerType_Game)
}

func (c *_Club) getClubList(clubID, playerLV, pageNum, pageSize uint32, isJoinLevel bool) ([]*gmsg.ClubInfo, int) {
	clubList := make([]entity.Club, 0)

	var query, selector bson.M

	if clubID > 0 && !isJoinLevel {
		query = bson.M{"ClubID": clubID}
	} else if clubID == 0 && isJoinLevel {
		selector = bson.M{"$lt": playerLV}
		query = bson.M{"JoinLevel": selector}
	} else if clubID > 0 && isJoinLevel {
		selector = bson.M{"$lt": playerLV}
		query = bson.M{"ClubID": clubID, "JoinLevel": selector}
	}

	err := DBConnect.GetDataLimitAndPage(consts.COLLECTION_CLUB, query, int(pageNum), int((pageSize-1)*pageNum), &clubList, "-ClubLV")
	if err != nil {
		log.Error(err)
		return nil, 0
	}

	list := make([]*gmsg.ClubInfo, 0)
	for _, vl := range clubList {
		clubInfo := new(gmsg.ClubInfo)
		player := new(entity.EntityPlayer)
		player.SetDBConnect(entity.UnitPlayer)
		yes, errs := player.InitFormDB(vl.MasterEntityID, DBConnect)
		if !yes {
			log.Error(errs)
			return nil, 0
		}
		stack.SimpleCopyProperties(clubInfo, &vl)
		clubInfo.MasterPlayerIcon = player.PlayerIcon
		clubInfo.MasterPlayerName = player.PlayerName
		list = append(list, clubInfo)
	}

	count, errs := DBConnect.GetDataCountTotal(consts.COLLECTION_CLUB)
	if errs != nil {
		log.Error(errs)
		return nil, 0
	}
	return list, count
}

func (c *_Club) OnClubTop10DBRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.ClubTop10DBRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	clubArgs := make([]entity.Club, 0)
	err := DBConnect.GetLimitDataAndSort(consts.COLLECTION_CLUB, 10, nil, &clubArgs, "-ClubScore")
	if err != nil {
		log.Error("-->OnClubTop10DBRequest--err--", err)
		return
	}
	list := make([]*gmsg.ClubTop10DBInfo, 0)
	for _, vl := range clubArgs {
		player := new(entity.EntityPlayer)
		player.SetDBConnect(entity.UnitPlayer)
		yes, errs := player.InitFormDB(vl.MasterEntityID, DBConnect)
		if !yes {
			log.Error(errs)
			break
		}
		club := new(gmsg.ClubTop10DBInfo)
		stack.SimpleCopyProperties(club, &vl)
		club.MasterPlayerName = player.PlayerName
		club.MasterPlayerIcon = player.PlayerIcon
		list = append(list, club)
	}

	msgResponse := &gmsg.ClubTop10DBResponse{}
	msgResponse.TimeStamp = msgBody.TimeStamp
	msgResponse.List = list
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_ClubTOP10_DB_Data_Response), msgResponse, network.ServerType_Game)
}

func (c *_Club) OnClubCreateDBRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	msgBody := &gmsg.ClubCreateRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	msgResponse := &gmsg.ClubCreateResponse{}
	eEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	tEntityPlayer := eEntity.(*entity.EntityPlayer)
	if tEntityPlayer.ClubId > 0 {
		msgResponse.Code = 1
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_ClubCreateResponse, msgResponse, network.ServerType_Game)
		return
	}

	club := new(entity.Club)
	club.InitByFirst(consts.COLLECTION_CLUB, msgBody.EntityID)
	club.ClubID = uint32(c.getCreateClubID())
	club.ClubNotice = msgBody.ClubNotice
	club.ClubName = msgBody.ClubName
	club.ClubBadge = msgBody.ClubBadge
	club.AddMembers(msgBody.EntityID, conf.Master)
	club.SaveEntity(DBConnect)
	Entity.EmClub.AddEntity(club)
	// 推送到游戏服同步entity club
	Entity.DoPlayerClubSync_Bytes(club)
	msgResponse.Code = 0
	msgResponse.MasterEntityID = msgBody.EntityID
	msgResponse.ClubID = club.ClubID

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_ClubCreateResponse, msgResponse, network.ServerType_Game)
}

func (c *_Club) getCreateClubID() int {
	clubID32 := int32(c.ClubNow)
	clubID := tools.GetEntityID(&clubID32)
	c.ClubNow = int(clubID)
	return c.ClubNow
}

// 审核成员 游戏服->db服
func (c *_Club) OnClubRatifyJoinDBRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	msgBody := &gmsg.ClubRatifyJoinToDB{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("-->OnClubRatifyJoinDBRequest-->begin-->", msgBody)
	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.AddEntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	tEntityClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := tEntityClub.(*entity.Club)
	if club == nil {
		return
	}

	msgResponse := &gmsg.ClubRatifyJoinToGame{}
	msgResponse.IsJoin = false

	if tEntityPlayer.GetClubID() == 0 {
		msgResponse.IsJoin = true
		//加入并清空申请列表
		tEntityPlayer.JoinClub(msgBody.ClubID, club.ClubBadge, club.ClubRate, club.ClubName)
		tEntityPlayer.UpdateClubTags()
		// 更新俱乐部任务
		c.ClubTaskInit(msgBody.EntityID)
		TaskDBManger.updateConditional(tEntityPlayer.EntityID, []conf.ConditionData{{conf.JoinOrCreateClub, 1, false}})
	} else {
		tEntityPlayer.RemoveClubReq(msgBody.ClubID)
	}
	tEntityPlayer.FlagChang()

	msgResponse.EntityID = msgBody.EntityID
	msgResponse.AddEntityID = msgBody.AddEntityID
	msgResponse.ClubID = msgBody.ClubID

	log.Info("-->OnClubRatifyJoinDBRequest-->end-->", msgResponse)
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_ClubRatifyJoinToGame), msgResponse, network.ServerType_Game)
}

// 删除俱乐部成员，在DB处理离线用户
func (c *_Club) OnClubDelMembersDBRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	msgBody := &gmsg.ClubDelMembersToDB{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntityClub := Entity.EmClub.GetEntityByID(msgBody.ClubID)
	club := tEntityClub.(*entity.Club)
	if club == nil {
		return
	}

	tEntity := Entity.EmEntityPlayer.GetEntityByID(msgBody.DelEntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.ResetClubID()
	tEntityPlayer.SetExitClubUnixSec()
	tEntityPlayer.ClubAttribute.ClubProgressRewardList = c.getClubProgressRewardCfg()
	tEntityPlayer.FlagChang()

	msgResponse := &gmsg.ClubDelMemberResponse{}
	msgResponse.Code = uint32(0)
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.DelEntityID = msgBody.DelEntityID

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Hall_ClubDelMembersResponse, msgResponse, network.ServerType_Game)
}

func (c *_Club) getClubProgressRewardCfg() []entity.ClubProgressReward {
	clubProgressRewardList := make([]entity.ClubProgressReward, 0)
	for _, vl := range Table.GetClubProgressRewardCfg() {
		progressReward := new(entity.ClubProgressReward)
		progressReward.ProgressID = vl.TableID
		progressReward.StateReward = 0
		progressReward.Progress = vl.Progress
		progressReward.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		clubProgressRewardList = append(clubProgressRewardList, *progressReward)
	}
	sort.Slice(clubProgressRewardList, func(i, j int) bool {
		return clubProgressRewardList[i].ProgressID < clubProgressRewardList[j].ProgressID
	})
	return clubProgressRewardList
}

func (c *_Club) ClubTaskInit(EntityID uint32) {
	tEntity := Entity.EmEntityPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.UpdateClubReFreshUnix()
	if len(tEntityPlayer.ClubAttribute.ClubProgressRewardList) > 0 {
		return
	}
	tEntityPlayer.ClubAttribute.ClubProgressRewardList = Table.GetClubProgressRewardList()
	tEntityPlayer.ClubAttribute.ClubTaskProgressList = Table.GetClubTaskProgressList()
	tEntityPlayer.ClubAttribute.ClubTaskList = Table.GetClubTaskList()
}

func (c *_Club) OnBatchReqClubRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	msgBody := &gmsg.BatchCreateClubRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	msgResponse := &gmsg.BatchCreateClubResponse{}
	msgResponse.ClubID = make([]uint32, 0)
	for i := uint32(1); i <= msgBody.RegNum; i++ {
		us := fmt.Sprintf("e%d", c.getRandomID(10000, 99999))
		enid, username := RegRobotMr.regAccForName(us)
		if enid == 0 {
			continue
		}
		RegRobotMr.player(enid, username, 10000)
		tEntity := Entity.EmEntityPlayer.GetEntityByID(enid)

		club := new(entity.Club)
		club.InitByFirst(consts.COLLECTION_CLUB, enid)
		club.ClubID = uint32(c.getCreateClubID())
		club.ClubNotice = ""
		club.ClubName = username
		club.ClubBadge = 1
		club.AddMembers(enid, conf.Master)
		club.SaveEntity(DBConnect)
		tEntityPlayer := tEntity.(*entity.EntityPlayer)
		tEntityPlayer.ClubId = club.ClubID
		tEntityPlayer.ClubName = club.ClubName
		tEntityPlayer.ClubBadge = club.ClubBadge
		c.ClubTaskInit(enid)
		tEntityPlayer.SaveEntity(DBConnect)
		Entity.EmClub.AddEntity(club)
		// 推送到游戏服同步entity club
		Entity.DoPlayerClubSync_Bytes(club)
		msgResponse.ClubID = append(msgResponse.ClubID, club.ClubID)
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_In_BatchCreateClubResponse), msgResponse, network.ServerType_Game)
}

// 生成不重复的房间ID
func (c *_Club) getRandomID(start int, end int) uint32 {
	//范围检查
	if end < start {
		return 0
	}
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//生成随机数
	num := uint32(r.Intn((end - start)) + start)
	return num
}
