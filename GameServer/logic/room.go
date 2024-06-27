package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Util/log"
	"BilliardServer/Util/timer"
	"fmt"
	"gitee.com/go-package/carbon/v2"
	"sync"
)

// 部件 房间对象
type UnitRoom struct {
	lock           sync.Mutex
	RoomID         uint32 //房间ID
	State          int    //房间状态 0 开启，1 关闭
	CountEntity    int
	MaxEntity      int
	Level          uint32 //房间类型
	Blind          uint64 //盲注
	TableFee       uint64 //台费
	WinExp         uint32 //胜利获得经验
	TransporterExp uint32 //失败获得经验
	CreateTime     string //最近一次日更时间
	UpdateTime     string //最近一次日更时间
	MapEntity      map[uint32]entity.Entity
	PlayNum        uint32 //对局次数
	ReplayConfirm  uint32 //重赛申请次数
}

// 是否还有空位
func (this *UnitRoom) YesForFree() bool {
	return this.CountEntity < this.MaxEntity
}

// 是否在此房间
func (this *UnitRoom) YesInRoom(tEntityID uint32) bool {
	yes := false
	for _, value := range this.MapEntity {
		if value.GetEntityID() == tEntityID {
			yes = true
			break
		}
	}
	return yes
}

// 房间所有玩家是否都已准备
func (this *UnitRoom) YesReadyAll() bool {
	yes := false
	var count int = 0
	for _, value := range this.MapEntity {
		tEntityMain := value.(*entity.EntityAcc)
		if tEntityMain.State == 1 {
			count++
		}
	}
	if count > 1 {
		yes = true
	}
	return yes
}

// 获取此房间所有EntityID
func (this *UnitRoom) GetAllEntityID() []uint32 {
	tArgs := make([]uint32, 0)
	for _, value := range this.MapEntity {
		tArgs = append(tArgs, value.GetEntityID())
	}
	return tArgs
}

func (this *UnitRoom) GetOtherPlayerId(entityId uint32) uint32 {
	for _, value := range this.MapEntity {
		if value.GetEntityID() != entityId {
			return value.GetEntityID()
		}
	}
	return 0
}

// 进入房间
func (this *UnitRoom) EnterRoom(tEntity entity.Entity) {
	this.MapEntity[tEntity.GetEntityID()] = tEntity
	this.CountEntity = len(this.MapEntity)
}

// 退出房间
func (this *UnitRoom) ExitRoom(tEntityID uint32) entity.Entity {
	this.lock.Lock()
	defer this.lock.Unlock()
	tEntity := this.MapEntity[tEntityID]
	delete(this.MapEntity, tEntityID)
	this.CountEntity = len(this.MapEntity)
	log.Info(fmt.Sprintf("UnitRoom:%d,%d>>>>退出房间成功.", this.RoomID, tEntityID))
	return tEntity
}

// 更新对局次数
func (this *UnitRoom) AddRoomPlayNum() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.PlayNum += 1
	this.UpdateTime = carbon.Now().ToDateTimeString()
}

func (this *UnitRoom) ResetReplayConfirm() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.ReplayConfirm = 0
}

// 开始倒计时
func (this *UnitRoom) StartCoutDown() {
	timer.AddTimer(this, "FirStart", 5*1000, false)
}

// 通知前端战斗开始
func (this *UnitRoom) FirStart() {
	// msgSync := &msg.RoomBattleStartSync{}
	// msgSync.RoomID = proto.Int32(int32(this.RoomID))
	// ConnectManager.SendMsgPbToGateBroadCast(msg.Room_RoomBattleStartSync, msgSync, this.GetAllEntityID())
}

// 杀死对手
func (this *UnitRoom) KillDead(sEntityID uint32, tEntityID uint32) {
	tEntityAcc := this.MapEntity[tEntityID].(*entity.EntityAcc)
	tEntityAcc.State = 1
	timer.AddTimer(this, "KillReback", 15*1000, false)
}

// 死亡复活
func (this *UnitRoom) KillReback() {
	tArgs := make([]uint32, 0)
	for _, value := range this.MapEntity {
		tEntityAcc := value.(*entity.EntityAcc)
		if tEntityAcc.State == 1 {
			tEntityAcc.State = 0
			tArgs = append(tArgs, value.GetEntityID())
		}
	}
	if len(tArgs) > 0 {
		// for i := 0; i < len(tArgs); i++ {
		// 	msgSync := &msg.RoomRebackSync{}
		// 	msgSync.RoomID = proto.Int32(int32(this.RoomID))
		// 	msgSync.EntityID = proto.Uint32(tArgs[i])
		// 	msgSync.PlayerName = proto.String("player" + strconv.FormatUint(uint64(tArgs[i]), 10))
		// 	msgSync.PlayerLv = proto.Int32(1)
		// 	tVec3 := new(msg.Vec3)
		// 	tVec3.PostX = proto.Int32(0)
		// 	tVec3.PostY = proto.Int32(0)
		// 	tVec3.PostZ = proto.Int32(0)
		// 	msgSync.SiteNow = tVec3
		// 	ConnectManager.SendMsgPbToGateBroadCast(msg.Battle_RoomRebackSync, msgSync, this.GetAllEntityID())
		// }
	}

}
