package logic

import (
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/log"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"BilliardServer/Common/entity"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
)

type _Entity struct {
	EmAcc    *entity.Entity_Manager
	EmPlayer *entity.Entity_Manager
	EmClub   *entity.Entity_Manager
}

var Entity _Entity

func (s *_Entity) Init() {
	s.EmAcc = new(entity.Entity_Manager)
	s.EmAcc.Init("acc")
	s.EmPlayer = new(entity.Entity_Manager)
	s.EmPlayer.Init("player")
	s.EmClub = new(entity.Entity_Manager)
	s.EmClub.Init("club")

	//注册逻辑业务事件
	event.On(entity.UnitSyncentity, reflect.ValueOf(s.OnSyncEntityToDB))
	event.OnNet(gmsg.MsgTile_Sys_SyncEntity, reflect.ValueOf(s.OnSyncEntityFormDB))
	event.OnNet(gmsg.MsgTile_Login_EnterGameRequest, reflect.ValueOf(s.OnEnterGameRequest))
	event.OnNet(gmsg.MsgTile_Login_EnterGameResponse, reflect.ValueOf(s.OnEnterGameResponse))
	event.OnNet(gmsg.MsgTile_Login_PlayerCreateRequest, reflect.ValueOf(s.OnPlayerCreateRequest))
	event.OnNet(gmsg.MsgTile_Login_PlayerCreateResponse, reflect.ValueOf(s.OnPlayerCreateResponse))
	event.OnNet(gmsg.MsgTile_Login_MainAccSync, reflect.ValueOf(s.OnMainAccSync))
	event.OnNet(gmsg.MsgTile_Login_MainPlayerSync, reflect.ValueOf(s.OnMainPlayerSync))
	event.OnNet(gmsg.MsgTile_Sys_EntityOfflineToGameRequest, reflect.ValueOf(s.OnEntityOffline))

	//每10分钟同步在线人数
	timer.AddTimer(s, "TimingUpdateOlineUserNum", 600000, true)
}

func (s *_Entity) TimingUpdateOlineUserNum() {
	resp := &gmsg.InStatisticsUserOlineNumRequest{
		Num: uint32(s.EmPlayer.EntityCount),
	}

	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile(gmsg.InternalMsgTile_IN_Statistics_User_Oline_Num_Request), resp, network.ServerType_DB)
}

func (s *_Entity) OnEntityOffline(msgEV *network.MsgBodyEvent) {
	req := &gmsg.EntityOfflineToGameRequest{}
	err := msgEV.Unmarshal(req)
	if err != nil {
		log.Waring("-->logic--_Entity--OnEntityOffline--msgEV.Unmarshal(req) err:", err)
		return
	}

	if req.EntityID <= 0 {
		log.Waring("-->logic--_Entity--OnEntityOffline--req.EntityID <= 0")
		return
	}

	//TODO:离线后处理各种数据

	//最后持久化用户数据 踢出在线管理器
	player := s.EmPlayer.GetEntityByID(req.EntityID)
	if player != nil && player.GetEntityID() > 0 {
		tEntityPlayer := player.(*entity.EntityPlayer)

		//TODO：对战未结算先不踢出加入观察队列（待实现）
		if tEntityPlayer.GetBehaviorStatus() != consts.BEHAVIOR_STATUS_BATTLE {
			tEntityPlayer.ResetRoomId()
			tEntityPlayer.ResetOnline()
			tEntityPlayer.SetExitTime(tools.GetTimeByTimeStamp(time.Now().Unix()))
			tEntityPlayer.SyncEntity(1)
			s.EmPlayer.DelEntity(player)

			acc := s.EmAcc.GetEntityByID(req.EntityID)
			if acc != nil && acc.GetEntityID() > 0 {
				tEntityAcc := acc.(*entity.EntityAcc)
				tEntityAcc.SyncEntity(1)
				s.EmAcc.DelEntity(acc)
			}
		}
	}

	return
}

// 同步实体数据，游戏服->DB服
func (s *_Entity) OnSyncEntityToDB(ev *entity.EntityEvent) {
	tBuff := new(network.MyBuff)
	tBuff.WriteUint32(ev.TypeSave)
	tBuff.WriteUint32(ev.TypeEntity)
	if ev.TypeEntity == entity.EntityTypeAcc {
		buf, _ := stack.StructToBytes_Gob(ev.Entity.(*entity.EntityAcc))
		tBuff.WriteBytes(buf)
	} else if ev.TypeEntity == entity.EntityTypePlayer {
		buf, _ := stack.StructToBytes_Gob(ev.Entity.(*entity.EntityPlayer))
		tBuff.WriteBytes(buf)
	} else if ev.TypeEntity == entity.EntityTypeClub {
		buf, _ := stack.StructToBytes_Gob(ev.Entity.(*entity.Club))
		tBuff.WriteBytes(buf)
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile_Sys_SyncEntity, tBuff.GetBytes(), network.ServerType_DB)
}

// 同步实体数据，DB服->游戏服
func (s *_Entity) OnSyncEntityFormDB(msgEV *network.MsgBodyEvent) {
	//typeSave := binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
	typeEntity := binary.LittleEndian.Uint32(msgEV.MsgBody[4:])
	if typeEntity == entity.EntityTypeAcc {
		var tEntityAcc entity.EntityAcc
		stack.BytesToStruct_Gob(msgEV.MsgBody[12:], &tEntityAcc)
		s.EmAcc.AddEntity(&tEntityAcc)
	} else if typeEntity == entity.EntityTypePlayer {
		var tEntityPlayer entity.EntityPlayer
		stack.BytesToStruct_Gob(msgEV.MsgBody[12:], &tEntityPlayer)
		s.EmPlayer.AddEntity(&tEntityPlayer)
	} else if typeEntity == entity.EntityTypeClub {
		var tEntityClub entity.Club
		stack.BytesToStruct_Gob(msgEV.MsgBody[12:], &tEntityClub)
		s.EmClub.AddEntity(&tEntityClub)
	}
}

// 进入游戏，前端->游戏服->DB服
func (s *_Entity) OnEnterGameRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EnterGameRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Login_EnterGameRequest, msgBody, network.ServerType_DB)
}

// 进入游戏，DB服->游戏服->前端
func (s *_Entity) OnEnterGameResponse(msgEV *network.MsgBodyEvent) {
	msgResponse := &gmsg.EnterGameResponse{}
	err := msgEV.Unmarshal(msgResponse)
	if err != nil {
		return
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_EnterGameResponse, msgResponse, []uint32{msgResponse.EntityId})

	tEntity := Entity.EmPlayer.GetEntityByID(msgResponse.EntityId)
	if tEntity != nil {
		tEntityPlayer, ok := tEntity.(*entity.EntityPlayer)
		if ok {
			//登录发送跑马灯消息
			//TODO：后续去掉或改为vip等级限制
			SendMarqueeMsgSync(0, fmt.Sprintf(consts.LOGIN_GAME_MSG, tEntityPlayer.PlayerName))
		}
	}
	return
}

// 创建角色，前端->游戏服->DB服
func (s *_Entity) OnPlayerCreateRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PlayerCreateRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Login_PlayerCreateRequest, msgBody, network.ServerType_DB)
	return
}

// 进入游戏，DB服->游戏服->前端
func (s *_Entity) OnPlayerCreateResponse(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PlayerCreateResponse{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerCreateResponse, msgBody, []uint32{msgBody.EntityId})

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityId)
	if tEntity != nil {
		tEntityPlayer, ok := tEntity.(*entity.EntityPlayer)
		if ok {
			//登录发送跑马灯消息
			//TODO：后续去掉或改为vip等级限制
			SendMarqueeMsgSync(0, fmt.Sprintf(consts.LOGIN_GAME_MSG, tEntityPlayer.PlayerName))
		}
	}
	return
}

// 同步帐号信息，DB服->游戏服－>前端
func (s *_Entity) OnMainAccSync(msgEV *network.MsgBodyEvent) {
	var tEntityAcc entity.EntityAcc
	stack.BytesToStruct_Gob(msgEV.MsgBody, &tEntityAcc)
	Entity.EmAcc.AddEntity(&tEntityAcc)
	msgMainAccSync := &gmsg.MainAccSync{}
	msgMainAccSync.EntityId = tEntityAcc.EntityID
	msgMainAccSync.MainAcc = &gmsg.EntityAcc{}
	err := stack.StructCopySame_Json(msgMainAccSync.MainAcc, tEntityAcc)
	if err != nil {
		return
	}
	msgMainAccSync.MainAcc.Token = tEntityAcc.ObjID.Hex()
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_MainAccSync, msgMainAccSync, []uint32{tEntityAcc.EntityID})
}

// 同步角色信息，DB服->游戏服－>前端
func (s *_Entity) OnMainPlayerSync(msgEV *network.MsgBodyEvent) {
	var tEntityPlayer entity.EntityPlayer
	errs := stack.BytesToStruct_Gob(msgEV.MsgBody, &tEntityPlayer)
	if errs != nil {
		log.Info(errs)
	}
	Entity.EmPlayer.AddEntity(&tEntityPlayer)

	tempTime := tEntityPlayer.CurrentLoginTime
	tEntityPlayer.LastLoginTime, tEntityPlayer.CurrentLoginTime = tempTime, tools.GetTimeByTimeStamp(time.Now().Unix())
	LoginReward.initLoginReward(&tEntityPlayer)
	RechargeMr.playerFirstRechargeInit(&tEntityPlayer)
	//任务初始化
	Task.LoginTaskDailyReset(tEntityPlayer.EntityID)

	msgMainPlayerSync := &gmsg.MainPlayerSync{}
	msgMainPlayerSync.EntityId = tEntityPlayer.EntityID
	msgMainPlayerSync.MainPlayer = &gmsg.EntityPlayer{}

	//社交模块
	msgMainPlayerSync.MyFriends = SocialManager.getFriendsList(&tEntityPlayer)
	msgMainPlayerSync.MyFans, msgMainPlayerSync.AddFansNum = SocialManager.getFansList(&tEntityPlayer)

	dayProgressValue, weekProgressValue, dayProgressRewardList, weekProgressRewardList, taskList := Task.getPlayerTask(tEntityPlayer.EntityID)
	//任务模块
	msgMainPlayerSync.DayProgress = dayProgressValue
	msgMainPlayerSync.WeekProgress = weekProgressValue
	msgMainPlayerSync.TaskList = taskList
	msgMainPlayerSync.DayProgressList = dayProgressRewardList
	msgMainPlayerSync.WeekProgressList = weekProgressRewardList

	msgMainPlayerSync.CollectList = Collect.getCollectList(&tEntityPlayer)
	resAchievementLVid, isCanClaim := Achievement.getAchievementLVRewardID(&tEntityPlayer)
	msgMainPlayerSync.AchievementLvCanClaimReward = new(gmsg.AchievementLvCanClaimReward)
	msgMainPlayerSync.AchievementLvCanClaimReward.AchievementLvID = resAchievementLVid
	msgMainPlayerSync.AchievementLvCanClaimReward.IsCanClaim = isCanClaim
	//图鉴列表
	msgMainPlayerSync.CueHandBookList = CueHandBookMr.getCueHandBookList(&tEntityPlayer)
	//宝箱列表
	msgMainPlayerSync.BoxList = BoxMr.getBoxList(&tEntityPlayer)
	//Vip模块
	msgMainPlayerSync.VipInfoList = VipMgr.OnGetVipList(tEntityPlayer.EntityID)
	msgMainPlayerSync.VipDailyGetBoxStatus = VipMgr.isGetDailyBox(&tEntityPlayer)
	//邮件列表
	msgMainPlayerSync.EmailList = Email.getAllEmailByPlayer(&tEntityPlayer)
	//背包列表
	msgMainPlayerSync.Items = Backpack.GetAllItem(&tEntityPlayer)
	//俱乐部任务表
	msgMainPlayerSync.ClubTaskList = ClubManager.GetClubTaskList(tEntityPlayer.EntityID)
	//俱乐部红包列表
	msgMainPlayerSync.ClubRedEnvelopeList = ClubManager.ClubRedEnvelopeListRequest(&tEntityPlayer)
	//获取私聊消息红点
	msgMainPlayerSync.FriendMsgRedDotNum = ChatMgr.GetPrivateChatAllRedDotNum(tEntityPlayer.EntityID)
	//活动列表
	msgMainPlayerSync.ActivityList = Activity.GetLoginActivityListRequest(tEntityPlayer.EntityID)
	//天梯结算数据
	msgMainPlayerSync.PeakRankInfoSettle = Activity.GetPeakRankSettle()
	//登录公告列表
	msgMainPlayerSync.LoginNoticeList = LoginNotice.GetLoginNoticeListRequest(tEntityPlayer.EntityID)
	//定时登录奖励列表
	msgMainPlayerSync.RewardList = LoginReward.getPlayerLoginRewardList(tEntityPlayer.EntityID)
	//登录公告列表
	msgMainPlayerSync.PointsShopList = PointsShop.GetPointsShopListRequest(tEntityPlayer.EntityID)
	//首充列表
	msgMainPlayerSync.IsHaveRecharge = RechargeMr.getPlayerIsHaveRecharge(tEntityPlayer.EntityID)

	err := stack.StructCopySame_Json(msgMainPlayerSync.MainPlayer, tEntityPlayer)
	if err != nil {
		return
	}

	WelfareMr.LoginPlayerSinInListSync(tEntityPlayer.EntityID)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_MainPlayerSync, msgMainPlayerSync, []uint32{tEntityPlayer.EntityID})
	return
}
