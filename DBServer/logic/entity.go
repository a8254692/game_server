package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/jwt"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/timer"
	"BilliardServer/Util/tools"
	"encoding/binary"
	"errors"
	"google.golang.org/protobuf/proto"
	"gopkg.in/mgo.v2"
	"reflect"
	"time"
)

// 登录部件
type _Entity struct {
	UnitName       string
	EmEntityAcc    *entity.Entity_Manager
	EmEntityPlayer *entity.Entity_Manager
	EmClub         *entity.Entity_Manager
}

// 登录部件
var Entity _Entity

func (c *_Entity) Init() {
	c.EmEntityAcc = new(entity.Entity_Manager)
	c.EmEntityAcc.Init("acc")
	c.EmEntityPlayer = new(entity.Entity_Manager)
	c.EmEntityPlayer.Init("player")
	c.EmClub = new(entity.Entity_Manager)
	c.EmClub.Init("club")

	timer.AddTimer(c, "SaveEntityAll", 1000*10, true)

	event.On(entity.UnitSyncentity, reflect.ValueOf(c.OnSyncEntityToGame))
	event.OnNet(gmsg.MsgTile_Sys_SyncEntity, reflect.ValueOf(c.OnSyncEntityFormGame))
	event.OnNet(gmsg.MsgTile_Login_EnterGameRequest, reflect.ValueOf(c.OnEnterGameRequest))
	event.OnNet(gmsg.MsgTile_Login_PlayerCreateRequest, reflect.ValueOf(c.OnPlayerCreateRequest))
	event.OnNet(gmsg.MsgTile_Player_QueryEntityPlayerByIDRequest, reflect.ValueOf(c.OnQueryEntityPlayerByIDFromDB))
}

// 清空数据集
func (c *_Entity) ClearEntityAll() {
	//DBConnect.RemoveAllData("acc")
	//DBConnect.RemoveAllData("player")
	//DBConnect.RemoveAllData("entity")
	//DBConnect.RemoveAllData("club")
}

// 保存所有数据集
func (c *_Entity) SaveEntityAll() {
	countAcc := 0
	countPlayer := 0
	countClub := 0
	for _, ValueAcc := range c.EmEntityAcc.EntityMap {
		tEntityAcc := ValueAcc.(*entity.EntityAcc)
		if tEntityAcc.FlagChange {
			tEntityAcc.SaveEntity(DBConnect)
			tEntityAcc.FlagChange = false
			countAcc++
		}
	}
	for _, ValuePlayer := range c.EmEntityPlayer.EntityMap {
		tEntityPlayer := ValuePlayer.(*entity.EntityPlayer)
		if tEntityPlayer.FlagChange {
			tEntityPlayer.SaveEntity(DBConnect)
			tEntityPlayer.FlagChange = false
			countPlayer++
		}
	}
	for _, ValueClub := range c.EmClub.EntityMap {
		tEntityClub := ValueClub.(*entity.Club)
		if tEntityClub.FlagChange {
			tEntityClub.SaveEntity(DBConnect)
			tEntityClub.FlagChange = false
			countClub++
		}
	}
	log.Info("-->SaveEntityAll countAcc:", countAcc, " countPlayer:", countPlayer, " countClub:", countClub)
	return
}

// 初始化数据集管理器
func (c *_Entity) LoadEntityAll() {
	//加载所有的EntityAcc数据并填充进数据集管理器中
	tEntityAccArgs := make([]entity.EntityAcc, 0)
	tEntityAccType := entity.UnitAcc
	errEntityAcc := DBConnect.GetAll(tEntityAccType, nil, nil, &tEntityAccArgs)
	if errEntityAcc == nil {
		for i := 0; i < len(tEntityAccArgs); i++ {
			tEntityAccArgs[i].SetDBConnect(tEntityAccType)
			c.EmEntityAcc.AddEntity(&tEntityAccArgs[i])
		}
		log.Info("-->Load All ", tEntityAccType, " Entity Length:", len(tEntityAccArgs))
	} else {
		log.Info("-->Load All ", tEntityAccType, " Entity Error:", errEntityAcc)
	}

	//加载所有的EntityAcc数据并填充进数据集管理器中
	tEntityPlayerArgs := make([]entity.EntityPlayer, 0)
	tEntityPlayerType := entity.UnitPlayer
	errEntityPlayer := DBConnect.GetAll(tEntityPlayerType, nil, nil, &tEntityPlayerArgs)
	if errEntityPlayer == nil {
		for i := 0; i < len(tEntityPlayerArgs); i++ {
			tEntityPlayerArgs[i].SetDBConnect(tEntityPlayerType)
			c.EmEntityPlayer.AddEntity(&tEntityPlayerArgs[i])
		}
		log.Info("-->Load All ", tEntityPlayerType, " Entity Length:", len(tEntityPlayerArgs))
	} else {
		log.Info("-->Load All ", tEntityPlayerType, " Entity Error:", errEntityPlayer)
	}

	//加载所有的Club数据并填充进数据集管理器中
	tEntityClubArgs := make([]entity.Club, 0)
	tEntityClubType := entity.UnitClub
	errEntityClub := DBConnect.GetAll(tEntityClubType, nil, nil, &tEntityClubArgs)
	if errEntityClub == nil {
		for i := 0; i < len(tEntityClubArgs); i++ {
			tEntityClubArgs[i].SetDBConnect(tEntityClubType)
			c.EmClub.AddEntity(&tEntityClubArgs[i])
		}
		log.Info("-->Load All ", tEntityClubType, " Entity Length:", len(tEntityClubArgs))
	} else {
		log.Info("-->Load All ", tEntityClubType, " Entity Error:", errEntityClub)
	}
}

// 同步实体数据，DB服->游戏服
func (c *_Entity) OnSyncEntityToGame(ev *entity.EntityEvent) {
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
		log.Info("---->OnSyncEntityToGame-->", ev.Entity.(*entity.Club).ClubID)
		buf, _ := stack.StructToBytes_Gob(ev.Entity.(*entity.Club))
		tBuff.WriteBytes(buf)
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile_Sys_SyncEntity, tBuff.GetBytes(), network.ServerType_Game)
}

// 同步实体数据，游戏服->DB服
func (c *_Entity) OnSyncEntityFormGame(msgEV *network.MsgBodyEvent) {
	typeSave := binary.LittleEndian.Uint32(msgEV.MsgBody[0:])
	typeEntity := binary.LittleEndian.Uint32(msgEV.MsgBody[4:])
	if typeEntity == entity.EntityTypeAcc {
		var tEntityAcc entity.EntityAcc
		stack.BytesToStruct_Gob(msgEV.MsgBody[12:], &tEntityAcc)
		c.EmEntityAcc.AddEntity(&tEntityAcc)
		tEntityAcc.FlagChange = true

		var modeTypeSave bool
		if typeSave == 1 {
			if !tools.IsModeProd(Mode) {
				modeTypeSave = true
			}
		}

		if modeTypeSave || typeSave == 2 {
			tEntityAcc.SetDBConnect(entity.UnitAcc)
			tEntityAcc.SaveEntity(DBConnect)
			tEntityAcc.FlagChange = false
		}
	} else if typeEntity == entity.EntityTypePlayer {
		var tEntityPlayer entity.EntityPlayer
		stack.BytesToStruct_Gob(msgEV.MsgBody[12:], &tEntityPlayer)
		c.EmEntityPlayer.AddEntity(&tEntityPlayer)
		tEntityPlayer.FlagChange = true

		var modeTypeSave bool
		if typeSave == 1 {
			if !tools.IsModeProd(Mode) {
				modeTypeSave = true
			}
		}

		if modeTypeSave || typeSave == 2 {
			tEntityPlayer.SetDBConnect(entity.UnitPlayer)
			tEntityPlayer.SaveEntity(DBConnect)
			tEntityPlayer.FlagChange = false
		}
	} else if typeEntity == entity.EntityTypeClub {
		var tEntityClub entity.Club
		stack.BytesToStruct_Gob(msgEV.MsgBody[12:], &tEntityClub)
		c.EmClub.AddEntity(&tEntityClub)
		tEntityClub.FlagChange = true

		var modeTypeSave bool
		if typeSave == 1 {
			if !tools.IsModeProd(Mode) {
				modeTypeSave = true
			}
		}

		if modeTypeSave || typeSave == 2 {
			tEntityClub.SetDBConnect(entity.UnitClub)
			tEntityClub.SaveEntity(DBConnect)
			tEntityClub.FlagChange = false
		}
	}
}

func (c *_Entity) checkEntityToken(token string, entityID uint32) bool {
	var resp bool

	if entityID <= 0 {
		return resp
	}

	if token == "" {
		return resp
	}

	claims, err := jwt.ParseToken(token)
	if err != nil {
		return resp
	}
	if claims == nil {
		return resp
	}

	if claims.EntityId <= 0 {
		return resp
	}

	if claims.EntityId != entityID {
		return resp
	}

	resp = true

	return resp
}

func (c *_Entity) OnEnterGameRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.EnterGameRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}

	msgResponse := &gmsg.EnterGameResponse{}

	checkToken := c.checkEntityToken(msgBody.Token, msgBody.EntityId)

	if !checkToken {
		msgResponse.Result = 1 //进入游戏失败
		msgResponse.Code = 1   //无此帐号
	} else {
		tEntityPlayer := new(entity.EntityPlayer)
		tEntityPlayer.SetDBConnect(entity.UnitPlayer)
		yes, err := tEntityPlayer.InitFormDB(msgBody.EntityId, DBConnect)
		if err != nil && !errors.Is(err, mgo.ErrNotFound) {
		}

		var isSetPlayer bool
		if yes && tEntityPlayer != nil && tEntityPlayer.EntityID > 0 {
			isSetPlayer = true
		}

		var isSuccessLogin bool
		tEntity := c.EmEntityAcc.GetEntityByID(msgBody.EntityId)
		if tEntity == nil {
			tEntityAcc := new(entity.EntityAcc)
			tEntityAcc.SetDBConnect(entity.UnitAcc)

			ok, err := tEntityAcc.InitFormDB(msgBody.EntityId, DBConnect)
			if err != nil && !errors.Is(err, mgo.ErrNotFound) {
				return
			}

			if !ok {
				msgResponse.Result = 1 //进入游戏失败
				msgResponse.Code = 1   //无此帐号
				msgResponse.EntityId = msgBody.EntityId
			} else {
				c.EmEntityAcc.AddEntity(tEntityAcc) //添加进实体管理器
				c.DoMainAccSync_Bytes(tEntityAcc)

				if isSetPlayer {
					isSuccessLogin = true
					msgResponse.Result = 0 //进入游戏成功
					msgResponse.Code = 0   //已登录
					msgResponse.EntityId = tEntityAcc.EntityID
				} else {
					msgResponse.Result = 1 //进入游戏失败
					msgResponse.Code = 2   //末创建角色
					msgResponse.EntityId = msgBody.EntityId
				}
			}
		} else {
			if isSetPlayer {
				isSuccessLogin = true
				msgResponse.Result = 0 //进入游戏成功
				msgResponse.Code = 0   //已登录
				msgResponse.EntityId = msgBody.EntityId
			} else {
				msgResponse.Result = 1 //进入游戏失败，无此帐号
				msgResponse.Code = 2   //末创建角色
				msgResponse.EntityId = msgBody.EntityId
			}
		}

		if isSuccessLogin {
			c.EmEntityPlayer.AddEntity(tEntityPlayer) //添加进实体管理器
			c.DoPlayerBaseSync_Bytes(tEntityPlayer)
			c.DoMainPlayerSync_Bytes(tEntityPlayer)

			//如果角色有俱乐部
			//if tEntityPlayer.ClubId > 0 {
			//	tEntityClub := new(entity.Club)
			//	tEntityClub.SetDBConnect(entity.UnitClub)
			//	yes1, errs := tEntityClub.InitFormDB(tEntityPlayer.ClubId, DBConnect)
			//	if !yes1 || errs != nil {
			//		return
			//	}
			//	c.EmClub.AddEntity(tEntityClub) //添加进实体管理器
			//	c.DoPlayerClubSync_Bytes(tEntityClub)
			//}
		}
	}

	ConnectManager.SendMsgBodyPB(gmsg.MsgTile_Login_EnterGameResponse, msgResponse)
	return
}

func (c *_Entity) OnPlayerCreateRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.PlayerCreateRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}
	tEntity := c.EmEntityAcc.GetEntityByID(msgBody.EntityId)
	if tEntity == nil {
		return
	}
	tEntityAcc := tEntity.(*entity.EntityAcc)
	tEntityAcc.SetDBConnect(entity.UnitAcc)
	tEntityPlayer := new(entity.EntityPlayer)
	tEntityPlayer.InitByFirst(entity.UnitPlayer, msgBody.EntityId)
	tEntityPlayer.PlayerName = msgBody.Name
	tEntityPlayer.Sex = *msgBody.Sex
	if tEntityPlayer.Sex == consts.USER_WOMEN {
		tEntityPlayer.PlayerDress = Table.DefaultDress[*msgBody.Sex]
	}
	tEntityPlayer.InsertEntity(DBConnect)

	tUnitPlayerBase := new(entity.UnitPlayerBase)
	stack.SimpleCopyProperties(tUnitPlayerBase, tEntityPlayer)
	tEntityAcc.ListPlayer = append(tEntityAcc.ListPlayer, *tUnitPlayerBase)
	tEntityAcc.SaveEntity(DBConnect)

	Entity.EmEntityPlayer.AddEntity(tEntityPlayer)
	InitPlayerMr.InitPlayerData(msgBody.EntityId)
	//处理有角色的消息，一定要在返回PlayerCreateResponse之前发送
	c.DoMainAccSync_Bytes(tEntityAcc)
	c.DoMainPlayerSync_Bytes(tEntityPlayer)

	msgResponse := &gmsg.PlayerCreateResponse{}
	msgResponse.Result = 0 //创建成功
	msgResponse.Code = 0   //无错误
	msgResponse.EntityId = tEntityAcc.EntityID
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Login_PlayerCreateResponse, msgResponse, network.ServerType_Game)

}
func (c *_Entity) DoMainAccSync_Bytes(tEntityAcc *entity.EntityAcc) {
	buf, _ := stack.StructToBytes_Gob(tEntityAcc)
	if len(buf) < 1 {
		return
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile_Login_MainAccSync, buf, network.ServerType_Game)
}
func (c *_Entity) DoMainPlayerSync_Bytes(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer = InitPlayerMr.checkDefaultItem(tEntityPlayer.EntityID)
	buf, _ := stack.StructToBytes_Gob(tEntityPlayer)
	if len(buf) < 1 {
		return
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile_Login_MainPlayerSync, buf, network.ServerType_Game)

	//调用进入游戏加载用户统记数据至GAME_SERVER的方法
	Statistics.SyncUserDataStatisticsToGame(tEntityPlayer.EntityID)
	Chat.SyncChatFriendsListToGame(tEntityPlayer.EntityID)
}

func (c *_Entity) DoPlayerClubSync_Bytes(tEntityClub *entity.Club) {
	buf, _ := stack.StructToBytes_Gob(tEntityClub)
	if len(buf) < 1 {
		return
	}
	ConnectManager.SendMsgToOtherServer(gmsg.MsgTile_Login_PlayerClubSync, buf, network.ServerType_Game)
}

func (c *_Entity) DoPlayerBaseSync_Bytes(tEntityPlayer *entity.EntityPlayer) {
	msgResponse := &gmsg.QueryEntityPlayerByIDResponse{}
	msgResponse.Player = make([]*gmsg.PlayerBase, 0)
	for _, val := range tEntityPlayer.MyFriends {
		emPlayer := c.EmEntityPlayer.GetEntityByID(val.EntityID)
		if emPlayer == nil {
			continue
		}
		player := emPlayer.(*entity.EntityPlayer)
		playerBase := &gmsg.PlayerBase{}
		stack.SimpleCopyProperties(playerBase, player)
		msgResponse.Player = append(msgResponse.Player, playerBase)
	}
	for _, val := range tEntityPlayer.FansList.List {
		emPlayer := c.EmEntityPlayer.GetEntityByID(val.EntityID)
		if emPlayer == nil {
			continue
		}
		player := emPlayer.(*entity.EntityPlayer)
		playerBase := &gmsg.PlayerBase{}
		stack.SimpleCopyProperties(playerBase, player)
		msgResponse.Player = append(msgResponse.Player, playerBase)
	}

	playerBase := &gmsg.PlayerBase{}
	stack.SimpleCopyProperties(playerBase, tEntityPlayer)
	playerBase.CurrentLoginTime = tools.GetTimeByTimeStamp(time.Now().Unix())
	msgResponse.Player = append(msgResponse.Player, playerBase)
	msgResponse.Code = *proto.Uint32(0)
	if len(msgResponse.Player) > 0 {
		log.Info("-->DoPlayerBaseSync_Bytes-->end-->", msgResponse)
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Player_QueryEntityPlayerByIDResponse, msgResponse, network.ServerType_Game)
	}
	return
}

func (c *_Entity) OnQueryEntityPlayerByIDFromDB(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.QueryEntityPlayerByIDRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}
	log.Info("-->OnQueryEntityPlayerByIDFromDB-->begin-->", msgBody)
	tEntity := c.EmEntityPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}

	msgResponse := &gmsg.QueryEntityPlayerByIDResponse{}
	msgResponse.Player = make([]*gmsg.PlayerBase, 0)
	for _, val := range msgBody.QueryEntityID {
		emPlayer := c.EmEntityPlayer.GetEntityByID(val)
		if emPlayer == nil {
			continue
		}
		player := emPlayer.(*entity.EntityPlayer)
		playerBase := &gmsg.PlayerBase{}
		stack.SimpleCopyProperties(playerBase, player)
		msgResponse.Player = append(msgResponse.Player, playerBase)
	}

	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = *proto.Uint32(0)
	log.Info("-->OnQueryEntityPlayerByIDFromDB-->end-->", msgResponse)
	ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Player_QueryEntityPlayerByIDResponse, msgResponse, network.ServerType_Game)
}

func (c *_Entity) SendPlayerBaseSync(tEntityPlayer *entity.EntityPlayer) {
	msgResponse := &gmsg.QueryEntityPlayerByIDResponse{}
	msgResponse.Player = make([]*gmsg.PlayerBase, 0)

	playerBase := &gmsg.PlayerBase{}
	stack.SimpleCopyProperties(playerBase, tEntityPlayer)

	msgResponse.Player = append(msgResponse.Player, playerBase)
	msgResponse.Code = *proto.Uint32(0)
	if len(msgResponse.Player) > 0 {
		ConnectManager.SendMsgPbToOtherServer(gmsg.MsgTile_Player_QueryEntityPlayerByIDResponse, msgResponse, network.ServerType_Game)
	}
	return
}
