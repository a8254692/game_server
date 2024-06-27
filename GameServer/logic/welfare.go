package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/table"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/tools"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"
)

/***
 *@disc:福利模块
 *@author: lsj
 *@date: 2023/10/17
 */

type _Welfare struct {
	lock               sync.RWMutex
	ReFreshHour        []uint32
	ReFreshDeduct      []uint32
	FreeShopNum        int
	FreeReFreshTimes   uint32
	NextRefreshHour    uint32
	FreeShopList       []ShopProductKey
	FreeShopCfg        map[string]*table.FreeStoreCfg
	FreeShopPlayerList map[string][]*gmsg.FreeShopProduct
	FreeShopRand       map[uint32][]RandRewardInfo
}

type ShopProductKey struct {
	ShopKey uint32
}

var WelfareMr _Welfare

func (c *_Welfare) Init() {
	c.FreeShopNum = 10
	c.FreeShopPlayerList = make(map[string][]*gmsg.FreeShopProduct, 0)
	c.setReFreshHour()
	c.getFreeShopRandList()
	c.refreshFreeShopList()
	c.refreshTick()
	event.OnNet(gmsg.MsgTile_Welfare_SignInRequest, reflect.ValueOf(c.OnWelfareSignInRequest))
	event.OnNet(gmsg.MsgTile_Welfare_SignInListRequest, reflect.ValueOf(c.OnWelfareSignInListRequest))
	event.OnNet(gmsg.MsgTile_Welfare_FreeShopListRequest, reflect.ValueOf(c.OnFreeSHopListRequest))
	event.OnNet(gmsg.MsgTile_Welfare_FreeShopBuyRequest, reflect.ValueOf(c.OnFreeSHopBuyRequest))
	event.OnNet(gmsg.MsgTile_Welfare_RefreshFreeShopRequest, reflect.ValueOf(c.OnRefreshFreeShopRequest))
	//timer.AddTimer(c, "GetFreeShopPlayerList", 30000, true)
}

//func (c *_Welfare) GetFreeShopPlayerList() {
//	log.Info("GetFreeShopPlayerList:", c.FreeShopPlayerList)
//}

func (c *_Welfare) setReFreshHour() {
	config, ok := Table.GetConstMap()["13"]
	if !ok || config.Paramater1 == uint32(0) || len(config.Paramater10) == 0 || len(config.Paramater20) == 0 {
		log.Error("Welfare setReFreshHour is err!")
		return
	}
	if len(config.Paramater10) > 1 {
		c.ReFreshDeduct = config.Paramater10
	}
	if len(config.Paramater20) > 0 {
		c.ReFreshHour = config.Paramater20[0]
	}
	c.FreeReFreshTimes = config.Paramater1
	c.FreeShopCfg = Table.GetAllFreeStoreCfg()
}

func (c *_Welfare) getFreeShopCfg(shopKey uint32) *table.FreeStoreCfg {
	target, ok := c.FreeShopCfg[strconv.Itoa(int(shopKey))]
	if !ok {
		return nil
	}
	return target
}

func (c *_Welfare) resetFreeShopPlayerList() {
	c.FreeShopPlayerList = nil
}

func (c *_Welfare) refreshFreeShopList() {
	c.FreeShopList = make([]ShopProductKey, 0)
	c.FreeShopList = c.getNewFreeShopList()
	c.resetFreeShopPlayerList()
	log.Info("c.FreeShopList", c.FreeShopList)
}

func (c *_Welfare) getFreeShopRandList() {
	c.FreeShopRand = make(map[uint32][]RandRewardInfo, 0)
	for _, val := range Table.FreeStoreCfg {
		randRewardInfo := new(RandRewardInfo)
		randRewardInfo.itemTableId = int(val.TableID)
		if len(val.Product) > 1 {
			randRewardInfo.weight = val.Product[1]
		}
		c.FreeShopRand[val.RandomType] = append(c.FreeShopRand[val.RandomType], *randRewardInfo)
	}

	log.Info("c.FreeShopRand", c.FreeShopRand)
}

func (c *_Welfare) getNewFreeShopList() []ShopProductKey {
	shopList := make([]ShopProductKey, 0)
	r2, num := c.FreeShopNum-3, 0
	for randomType, randList := range c.FreeShopRand {
		if randomType == conf.RandomTYpe_2 {
			haveList := make(map[int]uint8, 0)
			for {
				list := make([]RandRewardInfo, 0)
				for _, v := range randList {
					if _, ok := haveList[v.itemTableId]; !ok {
						list = append(list, v)
					}
				}

				reward := randomDraw(list)
				haveList[reward.itemTableId] = 1
				shopProduct := new(ShopProductKey)
				shopProduct.ShopKey = uint32(reward.itemTableId)
				shopList = append(shopList, *shopProduct)
				num++
				if num == r2 {
					break
				}
			}
		} else {
			shopProduct := new(ShopProductKey)
			reward := randomDraw(randList)
			shopProduct.ShopKey = uint32(reward.itemTableId)
			shopList = append(shopList, *shopProduct)
		}
	}
	sort.Slice(shopList, func(i, j int) bool {
		return shopList[i].ShopKey < shopList[j].ShopKey
	})
	log.Info("shopList", shopList)
	return shopList
}

func (c *_Welfare) getFreeShopListProperties(shopList []ShopProductKey) (list []*gmsg.FreeShopProduct) {
	for _, val := range shopList {
		freeShopProduct := new(gmsg.FreeShopProduct)
		freeShopProduct.ShopKey = val.ShopKey
		list = append(list, freeShopProduct)
	}
	return
}

// 按时刷新商店
func (c *_Welfare) refreshTick() {
	c.lock.Lock()
	defer c.lock.Unlock()

	leftSecond := time.Duration(c.getRefreshHourUnix()) * time.Second
	time.AfterFunc(leftSecond, func() {
		nowHour := uint32(time.Now().Hour())
		if c.NextRefreshHour == nowHour {
			go c.refreshFreeShopList()
		}
		c.refreshTick()
	})
}

// 领取签到奖励
func (c *_Welfare) OnWelfareSignInRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()
	msgBody := &gmsg.SignInRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.SignInResponse{}
	msgResponse.EntityID = msgBody.EntityID
	msgResponse.Code = 1
	msgResponse.SignType = msgBody.SignType
	if msgBody.SignType > 1 || msgBody.SignType < 0 {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_SignInResponse, msgResponse, []uint32{msgBody.EntityID})
		return
	}

	isSign, _ := tEntityPlayer.IsDailySignIn()
	if !isSign {
		tEntityPlayer.DailySignIn(msgBody.SignType)
		continueDay, days := c.GetPlayerDailySignInDays(tEntityPlayer)
		msgResponse.Code = 0
		msgResponse.SignInContinueDays = continueDay
		// 必须先领取奖励，再重置
		c.signInRewardFromCfg(msgBody.EntityID, continueDay, msgBody.SignType)
		// 达到连续签到7天，重置
		if continueDay == conf.ContinueDays {
			tEntityPlayer.ResetDailySignInElement(days)
		}
		//同步签到到任务
		ConditionalMr.SyncConditional(msgBody.EntityID, []conf.ConditionData{{conf.SignInDayTimes, 1, false}})
		//更新成就
		go c.GetPlayerSummaryDailySignInDays(tEntityPlayer)
	}

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_SignInResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 根据配置奖励
func (c *_Welfare) signInRewardFromCfg(EntityID, days, signType uint32) {
	key := c.getRewardFromDayAndType(days, signType)
	dailySignInCfg := Table.GetDailySigInCfgMap(key)

	if dailySignInCfg == nil {
		return
	}

	rewardList := make([]entity.RewardEntity, 0)
	for _, value := range dailySignInCfg.Rewards {
		rewardEntity := new(entity.RewardEntity)
		rewardEntity.ItemTableId = value[0]
		rewardEntity.Num = value[1]
		rewardEntity.ExpireTimeId = 0
		rewardList = append(rewardList, *rewardEntity)
	}

	resParam := GetResParam(conf.SYSTEM_ID_WELFARE, conf.Reward)
	//发放奖励
	err, _ := Backpack.BackpackAddItemListAndSave(EntityID, rewardList, *resParam)
	if err != nil {
		log.Error(err)
	}
	log.Info("-->signInRewardFromCfg-->success!", "-->EntityID", EntityID, "-->days-->", days, "-->signType-->", signType)
}

// 组合key
func (c *_Welfare) getRewardFromDayAndType(days, signType uint32) string {
	return fmt.Sprintf("%d_%d", days, signType)
}

// 签到列表
func (c *_Welfare) OnWelfareSignInListRequest(msgEV *network.MsgBodyEvent) {
	msgBody := &gmsg.SignInListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.SignInListResponse{}
	msgResponse.SignInReward = 0
	isSign, signType := tEntityPlayer.IsDailySignIn()
	msgResponse.SignType = conf.SignTypeAd
	if isSign {
		msgResponse.SignInReward = 1
		msgResponse.SignType = signType
	}
	continueDay, _ := c.GetPlayerDailySignInDays(tEntityPlayer)
	msgResponse.SignInContinueDays = continueDay

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_SignInListResponse, msgResponse, []uint32{msgBody.EntityID})
}

// 获取最近7天的签到数据
func (c *_Welfare) GetPlayerDailySignInDays(tEntityPlayer *entity.EntityPlayer) (num uint32, sevenDays []string) {
	thisM, _, _ := tools.GetNowTimeMonthAndUnix()

	// 获取7天前的时间,包含当天
	bfSevenDays := tools.GetBeforeNDayString(conf.ContinueDays)
	isTodaySignIn, _ := tEntityPlayer.IsDailySignIn()

	for key, value := range bfSevenDays {
		// 当天未签到，直接跳过不用去查询
		if key == 0 && !isTodaySignIn {
			continue
		}
		month, _ := strconv.Atoi(value[5:7])
		day, _ := strconv.Atoi(value[8:10])
		if month == thisM {
			this := tEntityPlayer.GetMonthSignInDays(month)
			if this == nil {
				return
			}
			if this.Test(uint(day)) {
				num++
			} else {
				break
			}
		} else {
			last := tEntityPlayer.GetMonthSignInDays(tools.FormatToNMonthInt(-1))
			if last == nil {
				return
			}
			if last.Test(uint(day)) {
				num++
			} else {
				break
			}
		}
	}

	return
}

// 获取汇总签到的数据
func (c *_Welfare) GetPlayerSummaryDailySignInDays(tEntityPlayer *entity.EntityPlayer) {
	continueNDay := uint32(0)
	// 获取90天前的时间,包含当天
	bfNDays := tools.GetBeforeNDayString(conf.ContinueNDays90)

	for _, value := range bfNDays {
		month, _ := strconv.Atoi(value[5:7])
		day, _ := strconv.Atoi(value[8:10])

		summaryDailyLog := tEntityPlayer.GetMonthSummarySignInDays(month)
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
	ConditionalMr.SyncConditional(tEntityPlayer.EntityID, []conf.ConditionData{{conf.SignDaysConnecting, continueNDay, true}})
}

// 同步签到到客户端
func (c *_Welfare) LoginPlayerSinInListSync(EntityID uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	msgResponse := &gmsg.SignInListResponse{}
	msgResponse.SignInReward = 0
	isSign, signType := tEntityPlayer.IsDailySignIn()
	msgResponse.SignType = conf.SignTypeAd
	if isSign {
		msgResponse.SignInReward = 1
		msgResponse.SignType = signType
	}
	continueDay, _ := c.GetPlayerDailySignInDays(tEntityPlayer)
	msgResponse.SignInContinueDays = continueDay

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Login_PlayerSignInListSync, msgResponse, []uint32{EntityID})
}

// 免费商店列表请求
func (c *_Welfare) OnFreeSHopListRequest(msgEV *network.MsgBodyEvent) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	msgBody := &gmsg.FreeShopListRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	resResponse := &gmsg.FreeShopListResponse{
		NextRefreshHour: c.NextRefreshHour,
	}
	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--_OnFreeSHopListRequest--GetEntityPlayerById--err--", err)
		return
	}

	resResponse.List = make([]*gmsg.FreeShopProduct, 0)
	entityKey := c.queryFreeHourKey(msgBody.EntityID)
	if list, ok := c.FreeShopPlayerList[entityKey]; ok {
		resResponse.List = list
	} else {
		resResponse.List = c.getFreeShopListProperties(c.FreeShopList)
	}

	resResponse.EntityKey = entityKey
	resResponse.ReFreshTimes = c.getRefreshAdTimes(tEntityPlayer)

	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopListResponse, resResponse, []uint32{msgBody.EntityID})
}

// 免费商店购买请求
func (c *_Welfare) OnFreeSHopBuyRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()

	msgBody := &gmsg.FreeShopBuyRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--OnFreeSHopBuyRequest--GetEntityPlayerById--err--", err)
		return
	}
	resResponse := &gmsg.FreeShopBuyResponse{
		Code:    uint32(3),
		Product: &gmsg.FreeShopProduct{ShopKey: msgBody.ShopKey},
	}

	log.Info("--OnFreeSHopBuyRequest-->", msgBody)

	entityKey := c.queryFreeHourKey(msgBody.EntityID)
	if msgBody.EntityKey != entityKey {
		resResponse.Code = uint32(4)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	productCfg := c.getFreeShopCfg(msgBody.ShopKey)
	if productCfg == nil || msgBody.ShopKey == uint32(0) {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}
	disCountPrice := productCfg.Price * (uint32(10) - productCfg.Discount) / 10
	if productCfg.BuyType == conf.Gold {
		if tEntityPlayer.NumGold < disCountPrice {
			resResponse.Code = uint32(1)
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
			return
		}
	} else if productCfg.BuyType == conf.Diamond {
		if tEntityPlayer.NumStone < disCountPrice {
			resResponse.Code = uint32(2)
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
			return
		}
	}

	err = c.sendReward(productCfg.Product, productCfg.BuyType, disCountPrice, msgBody.EntityID)
	if err != nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	res := c.updateFreeShop(msgBody.ShopKey, entityKey)
	if res == nil {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	resResponse.Code = uint32(0)
	resResponse.Product.BuyStatus = uint32(1)
	log.Info("-->OnFreeSHopBuyRequest-->", resResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_FreeShopBuyResponse, resResponse, []uint32{msgBody.EntityID})
}

func (c *_Welfare) updateFreeShop(shopKey uint32, entityKey string) (res *gmsg.FreeShopProduct) {
	if len(c.FreeShopPlayerList) == 0 {
		c.FreeShopPlayerList = make(map[string][]*gmsg.FreeShopProduct, 0)
	}
	_, ok := c.FreeShopPlayerList[entityKey]
	if !ok {
		c.FreeShopPlayerList[entityKey] = c.getFreeShopListProperties(c.FreeShopList)
	}

	for index, val := range c.FreeShopPlayerList[entityKey] {
		if val.ShopKey == shopKey {
			value := val
			value.BuyStatus = uint32(1)
			c.FreeShopPlayerList[entityKey][index] = value
			res = value
			break
		}
	}
	return
}

// 发放物品
func (c *_Welfare) sendReward(reReward []uint32, BuyType, disCountPrice uint32, entityID uint32) error {
	if len(reReward) < 3 {
		return errors.New("配置异常。")
	}
	resParam := GetResParam(conf.SYSTEM_ID_WELFARE, conf.FreeShopBuy)
	rewardEntity := new(entity.RewardEntity)
	rewardEntity.ItemTableId = reReward[0]
	if len(reReward) == 3 {
		rewardEntity.ItemTableId = reReward[0]
		rewardEntity.Num = reReward[2]
		rewardEntity.ExpireTimeId = 0
	} else if len(reReward) == 4 {
		rewardEntity.ItemTableId = reReward[0]
		rewardEntity.Num = reReward[2]
		rewardEntity.ExpireTimeId = reReward[3]
	} else {
		return errors.New("配置异常。")
	}

	err, _ := Backpack.BackpackAddItemListAndSave(entityID, []entity.RewardEntity{*rewardEntity}, *resParam)
	if err != nil {
		return err
	}
	if disCountPrice > uint32(0) && BuyType > conf.FreeShopBuyAd {
		Player.UpdatePlayerPropertyItem(entityID, BuyType, int32(-disCountPrice), *resParam)
	}
	return nil
}

// 免费商店刷新请求
func (c *_Welfare) OnRefreshFreeShopRequest(msgEV *network.MsgBodyEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()

	msgBody := &gmsg.RefreshFreeShopRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	tEntityPlayer, err := GetEntityPlayerById(msgBody.EntityID)
	if err != nil {
		log.Waring("-->logic--_OnFreeSHopListRequest--GetEntityPlayerById--err--", err)
		return
	}

	resResponse := &gmsg.RefreshFreeShopResponse{
		Code: uint32(1),
	}

	if len(c.ReFreshDeduct) < 2 || c.FreeReFreshTimes == uint32(0) {
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_RefreshFreeShopResponse, resResponse, []uint32{msgBody.EntityID})
		return
	}

	if msgBody.RefreshType == conf.RefreshType {
		if tEntityPlayer.GetNumStone() < c.ReFreshDeduct[1] {
			resResponse.Code = uint32(3)
			ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_RefreshFreeShopResponse, resResponse, []uint32{msgBody.EntityID})
			return
		}

		resParam := GetResParam(conf.SYSTEM_ID_WELFARE, conf.FreeShopRefresh)
		Player.UpdatePlayerPropertyItem(msgBody.EntityID, c.ReFreshDeduct[0], int32(-c.ReFreshDeduct[1]), *resParam)
		resResponse.List = c.refreshPlayerShopList(tEntityPlayer, false)
	}

	if msgBody.RefreshType == conf.RefreshTypeAd {
		if tEntityPlayer.FreeShopRefresh.LastRefreshStamp >= tools.GetTodayBeginTime() {
			if tEntityPlayer.FreeShopRefresh.RefreshAdTimes == uint32(0) {
				resResponse.Code = uint32(2)
				ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_RefreshFreeShopResponse, resResponse, []uint32{msgBody.EntityID})
				return
			}
			resResponse.List = c.refreshPlayerShopList(tEntityPlayer, true)
		} else {
			tEntityPlayer.FreeShopRefresh.RefreshAdTimes = c.FreeReFreshTimes
			resResponse.List = c.refreshPlayerShopList(tEntityPlayer, true)
		}
	}
	resResponse.Code = uint32(0)
	resResponse.ReFreshTimes = c.getRefreshAdTimes(tEntityPlayer)
	log.Info("-->OnRefreshFreeShopRequest-->", resResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Welfare_RefreshFreeShopResponse, resResponse, []uint32{msgBody.EntityID})
}

func (c *_Welfare) getRefreshAdTimes(tEntityPlayer *entity.EntityPlayer) uint32 {
	if tEntityPlayer.FreeShopRefresh.LastRefreshStamp >= tools.GetTodayBeginTime() {
		return tEntityPlayer.FreeShopRefresh.RefreshAdTimes
	}
	return c.FreeReFreshTimes
}

// 刷新用户免费商店
func (c *_Welfare) refreshPlayerShopList(tEntityPlayer *entity.EntityPlayer, isRefreshUpdate bool) (list []*gmsg.FreeShopProduct) {
	key := c.queryFreeHourKey(tEntityPlayer.EntityID)
	newShopList := c.getNewFreeShopList()
	if len(c.FreeShopPlayerList) == 0 {
		c.FreeShopPlayerList = make(map[string][]*gmsg.FreeShopProduct, 0)
	}
	c.FreeShopPlayerList[key] = c.getFreeShopListProperties(newShopList)
	if isRefreshUpdate {
		tEntityPlayer.RefreshFreeShop()
		tEntityPlayer.SyncEntity(1)
	}

	return c.FreeShopPlayerList[key]
}

func (c *_Welfare) queryFreeHourKey(entityID uint32) (key string) {
	if len(c.ReFreshHour) < 2 {
		return key
	}
	nowHour := uint32(time.Now().Hour())
	hour := uint32(0)
	for i := 0; i < len(c.ReFreshHour); i++ {
		if c.ReFreshHour[0] <= nowHour && nowHour < c.ReFreshHour[len(c.ReFreshHour)-1] {
			if c.ReFreshHour[i] <= nowHour && nowHour < c.ReFreshHour[i+1] {
				hour = c.ReFreshHour[i]
				break
			}
		} else {
			hour = c.ReFreshHour[len(c.ReFreshHour)-1]
			break
		}
	}

	if hour < uint32(10) {
		key = fmt.Sprintf("%d0%d", entityID, hour)
	} else {
		key = fmt.Sprintf("%d%d", entityID, hour)
	}

	return
}

func (c *_Welfare) getRefreshHourUnix() int {
	now := time.Now()
	nowHour := uint32(now.Hour())
	hour := uint32(0)
	for i := 0; i < len(c.ReFreshHour); i++ {
		if c.ReFreshHour[0] <= nowHour && nowHour < c.ReFreshHour[len(c.ReFreshHour)-1] {
			if c.ReFreshHour[i] <= nowHour && nowHour < c.ReFreshHour[i+1] {
				hour = c.ReFreshHour[i+1]
				break
			}
		} else {
			hour = c.ReFreshHour[0]
			break
		}
	}
	c.NextRefreshHour = hour
	log.Info("-->NextRefreshHour->", hour)
	return int(tools.Tool_GetTimeGap(int(hour), 0))
}
