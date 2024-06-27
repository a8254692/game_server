package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/table"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"
)

/***
 *@disc:定时登录奖励
 *@author: lsj
 *@date: 2024/1/19
 */

type _LoginReward struct {
	RewardCfg map[uint32]LoginRewardCfg
}

type LoginRewardCfg struct {
	TimeKey   uint32
	StartHour int
	EndHour   int
	Reward    []entity.RewardEntity
}

var LoginReward _LoginReward

func (c *_LoginReward) Init() {
	c.setLoginRewardCfg()
	event.OnNet(gmsg.MsgTile_Login_RewardClaimRequest, reflect.ValueOf(c.OnLoginRewardClaimRequest))
}

// 配置表初始化
func (c *_LoginReward) setLoginRewardCfg() {
	rewardCfg1, ok1 := Table.GetConstMap()["14"]
	rewardCfg2, ok2 := Table.GetConstMap()["15"]
	if !ok1 || !ok2 {
		log.Error("LoginReward setLoginRewardCfg is err!")
		return
	}

	constCfg := make([]*table.ConstCfg, 0)
	constCfg = append(constCfg, rewardCfg1, rewardCfg2)
	c.RewardCfg = c.getLoginReward(constCfg)
	log.Info("-->setLoginRewardCfg-->end", c.RewardCfg)
}

func (c *_LoginReward) getLoginReward(constCfg []*table.ConstCfg) (list map[uint32]LoginRewardCfg) {
	list = make(map[uint32]LoginRewardCfg, 0)
	timeKey := uint32(1)
	for _, rewardCfg := range constCfg {
		cfg := new(LoginRewardCfg)
		if len(rewardCfg.Paramater10) > 1 && len(rewardCfg.Paramater20) > 1 {
			cfg.TimeKey = timeKey
			cfg.StartHour, cfg.EndHour = int(rewardCfg.Paramater10[0]), int(rewardCfg.Paramater10[1])
			cfg.Reward = make([]entity.RewardEntity, 0)
			for _, val := range rewardCfg.Paramater20 {
				cfg.Reward = append(cfg.Reward, entity.RewardEntity{ItemTableId: val[0], Num: val[1]})
			}
		}
		list[cfg.TimeKey] = *cfg
		timeKey++
	}
	return
}

func (c *_LoginReward) getPlayerLoginRewardList(entityID uint32) (list []*gmsg.LoginReward) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	for _, val := range tEntityPlayer.LoginRewardList {
		reward := new(gmsg.LoginReward)
		stack.SimpleCopyProperties(reward, &val)
		list = append(list, reward)
	}
	return
}

// 领取定时登录奖励
func (c *_LoginReward) OnLoginRewardClaimRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.LoginRewardClaimRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	resResponse := &gmsg.LoginRewardClaimResponse{
		Code: uint32(1),
	}

	rewardList, ok := c.RewardCfg[msgBody.TimeKey]
	if msgBody.TimeKey == 0 || !ok || len(rewardList.Reward) == 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_RewardClaimResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	now := time.Now()
	hour, minu := now.Hour(), now.Minute()
	hm := tools.GetHourMinuteInt(hour, minu)
	if msgBody.RewardType == conf.LoginStateReward_0 && (hm < rewardList.StartHour || hm > rewardList.EndHour) {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_RewardClaimResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--OnLoginRewardClaimRequest--GetEntityPlayerById--err--", err)
		return
	}

	resParam := GetResParam(conf.SYSTEM_ID_ACTIVITY, conf.LoginReward)
	err, _ = Backpack.BackpackAddItemListAndSave(msgBody.EntityID, rewardList.Reward, *resParam)
	if err != nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_RewardClaimResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	if err = c.loginRewardClaim(tEntityPlayer, msgBody.TimeKey); err != nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_RewardClaimResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	resResponse.Code = uint32(0)
	loginReward := &gmsg.LoginReward{TimeKey: msgBody.TimeKey, IsReward: true}
	resResponse.LoginReward = loginReward

	tEntityPlayer.SyncEntity(1)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_RewardClaimResponse, resResponse, []uint32{msgBody.EntityID})
}

func (c *_LoginReward) updateLoginReward(tEntityPlayer *entity.EntityPlayer) {
	tEntityPlayer.LoginRewardList = make([]entity.LoginReward, 0)
	for _, val := range c.RewardCfg {
		tEntityPlayer.LoginRewardList = append(tEntityPlayer.LoginRewardList,
			entity.LoginReward{TimeKey: val.TimeKey, IsReward: false})
	}
	sort.Slice(tEntityPlayer.LoginRewardList, func(i, j int) bool {
		return tEntityPlayer.LoginRewardList[i].TimeKey < tEntityPlayer.LoginRewardList[j].TimeKey
	})
}

func (c *_LoginReward) initLoginReward(tEntityPlayer *entity.EntityPlayer) {
	if len(tEntityPlayer.LoginRewardList) > 0 {
		return
	}
	c.updateLoginReward(tEntityPlayer)
}

func (c *_LoginReward) loginRewardClaim(tEntityPlayer *entity.EntityPlayer, timeKey uint32) error {
	for key, val := range tEntityPlayer.LoginRewardList {
		if val.IsReward == false && val.TimeKey == timeKey {
			v := val
			v.IsReward = true
			tEntityPlayer.LoginRewardList[key] = v
			return nil
		}
	}
	return errors.New(fmt.Sprintf("%s:数据异常。", timeKey))
}
