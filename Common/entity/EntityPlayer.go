package entity

import (
	conf "BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"errors"
	"github.com/bits-and-blooms/bitset"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

// 角色对象
type EntityPlayer struct {
	CollectionName          string                   `bson:"-"`                       //数据集名称
	FlagChange              bool                     `bson:"-"`                       //是否被修改
	FlagKick                bool                     `bson:"-"`                       //被T标记
	ObjID                   bson.ObjectId            `bson:"_id,omitempty"`           //唯一ID
	EntityID                uint32                   `bson:"EntityID"`                //实体ID
	PlayerID                uint32                   `bson:"PlayerID"`                //实体ID
	AccUnique               string                   `bson:"AccUnique"`               //帐号
	TimeCreate              string                   `bson:"TimeCreate"`              //创建时间
	TimeUpdate              string                   `bson:"TimeUpdate"`              //更新时间
	TimeExit                string                   `bson:"TimeExit"`                //退出时间
	CurrentLoginTime        string                   `bson:"CurrentLoginTime"`        // 当前登录时间
	LastLoginTime           string                   `bson:"LastLoginTime"`           // 上一次登录时间
	TimeTotal               uint32                   `bson:"TimeTotal"`               //总在线时长(秒)
	PlayerLv                uint32                   `bson:"PlayerLv"`                //角色等级
	PlayerName              string                   `bson:"PlayerName"`              //角色名称
	PlayerIcon              uint32                   `bson:"PlayerIcon"`              //角色头像
	NumExp                  uint32                   `bson:"NumExp"`                  //经验数量
	NumGold                 uint32                   `bson:"NumGold"`                 //金币数量
	NumStone                uint32                   `bson:"NumStone"`                //钻石数量
	VipLv                   uint32                   `bson:"VipLv"`                   //Vip等级
	VipExp                  uint32                   `bson:"VipExp"`                  //Vip等级
	BagMax                  uint32                   `bson:"BagMax"`                  //背包最大值
	BagNow                  uint32                   `bson:"BagNow"`                  //背包物品当前数量
	BagList                 []Item                   `bson:"BagList"`                 //背包物品列表
	TaskList                []Task                   `bson:"TaskList"`                //任务列表
	TaskResetDate           TaskResetDate            `bson:"TaskResetDate"`           //记录重置日期
	RedTipsList             []RedTips                `bson:"RedTipsList"`             //红点提示列表
	EmailList               []Email                  `bson:"EmailList"`               //邮件列表
	Sex                     uint32                   `bson:"Sex"`                     //用户性别
	RoomId                  uint32                   `bson:"RoomId"`                  //房间ID
	BehaviorStatus          uint8                    `bson:"BehaviorStatus"`          //行为状态 1大厅中 2匹配中 3房间中 4对战中
	IsRobot                 bool                     `bson:"Robot"`                   //是否机器人
	Online                  bool                     `bson:"Online"`                  //是否在线
	GiftsList               []GiveGift               `bson:"GiftsList"`               //赠送礼物列表
	ReceivingGifts          []RecGift                `bson:"ReceivingGifts"`          //接收到的礼物列表
	OpenGifts               bool                     `bson:"OpenGifts"`               //是否公开,默认不公开
	CharmNum                uint32                   `bson:"CharmNum"`                //魅力值
	PeakRankLv              uint32                   `bson:"PeakRankLv"`              //天梯/排位等级
	PeakRankExp             uint32                   `bson:"PeakRankExp"`             //天梯/排位赛星数
	FansNum                 uint32                   `bson:"FansNum"`                 //粉丝数
	IconFrame               uint32                   `bson:"IconFrame"`               //装扮头像框
	PeakRankHist            []PeakRankHist           `bson:"PeakRankHist"`            //赛季历史记录
	ClubId                  uint32                   `bson:"ClubId"`                  //俱乐部id
	ClubName                string                   `bson:"ClubName"`                //俱乐部名称
	ClubRate                uint32                   `bson:"ClubRate"`                // 俱乐部评级
	ClubBadge               uint32                   `bson:"ClubBadge"`               //俱乐部徽章
	PlayerSign              string                   `bson:"PlayerSign"`              //个性签名
	CueTableId              uint32                   `bson:"CueTableId"`              //使用球杆id
	CollectId               uint32                   `bson:"CollectId"`               //称号id
	Badge                   uint32                   `bson:"Badge"`                   //徽章
	PopularityValue         uint32                   `bson:"PopularityValue"`         //人气值
	BattingEffect           uint32                   `bson:"BattingEffect"`           //特效，击球特效
	GoalInEffect            uint32                   `bson:"GoalInEffect"`            //特效，入球效果
	CueBall                 uint32                   `bson:"CueBall"`                 //特效，主球
	TableCloth              uint32                   `bson:"TableCloth"`              //特效，桌布
	PlayerDress             uint32                   `bson:"PlayerDress"`             //人物着装
	PlayerBGImg             uint32                   `bson:"PlayerBGImg"`             //人物背景
	ClothingIcon            uint32                   `bson:"ClothingIcon"`            //装扮头像 这个字段先不用了。以后删除
	ClothingCountDown       uint32                   `bson:"ClothingCountDown"`       //装扮倒计时
	ClothingBubble          uint32                   `bson:"ClothingBubble"`          //装扮气泡
	ReqJoinClub             []uint32                 `bson:"ReqJoinClub"`             //申请加入俱乐部列表
	MyFriends               []Friend                 `bson:"MyFriends"`               // 我的关注
	FansList                FansAttribute            `bson:"FansList"`                // 粉丝列表
	GiveGoldList            []GiveGold               `bson:"GiveGoldList"`            // 赠送金币列表
	GiveGoldData            GiveGoldDate             `bson:"GiveGoldData"`            // 赠送金币,bitmap结构
	VipLvReward             []uint32                 `bson:"VipLvReward"`             // Vip等级礼拜购买记录
	SignInRewardList        []SignInReward           `bson:"SignInRewardList"`        // 签到记录
	DayProgressValue        uint32                   `bson:"DayProgressValue"`        //今日任务活跃值 每天0点重置
	WeekProgressValue       uint32                   `bson:"WeekProgressValue"`       //每周任务活跃值 每周1的0点重置
	AchievementLV           uint32                   `bson:"AchievementLV"`           //人物成就等级
	AchievementScore        uint32                   `bson:"AchievementScore"`        //成就积分
	NextRewardAchievementLV uint32                   `bson:"NextRewardAchievementLV"` //下一等级待领取的人物成就等级 ，默认1
	ClubShopBuyList         []ShopItem               `bson:"ClubShopBuyList"`         //俱乐部商品购买列表
	ClubTags                bool                     `bson:"ClubTags"`                // 是否加入或者创建俱乐部，只要有就修改成true，就不会计算到每日任务中
	DayProgressReward       []ProgressList           `bson:"DayProgressReward"`       // 每日活跃领取奖励表
	WeekProgressReward      []ProgressList           `bson:"WeekProgressReward"`      // 每周活跃领取奖励表
	CollectList             []Collect                `bson:"CollectList"`             //称号集合
	AchievementLVRewardList []AchievementLVReward    `bson:"AchievementLVRewardList"` //成就奖励列表
	AchievementList         []Achievement            `bson:"AchievementList"`         // 成就列表
	DailySignInList         []DailySignInElement     `bson:"DailySignInList"`         //每日签到
	ClubAttribute           ClubAttribute            `bson:"ClubAttribute"`           //俱乐部相关
	ClubNumGold             uint32                   `bson:"ClubNumGold"`             //俱乐部币，不可重置（有俱乐部的时候才会显示）
	ExchangeGold            uint32                   `bson:"ExchangeGold"`            //兑换卷
	BoxList                 []Box                    `bson:"BoxList"`                 //宝箱列表
	State                   uint32                   `bson:"State"`                   //当前状态(与acc表状态同步)
	MaxUuid                 uint32                   `bson:"MaxUuid"`                 //递增ID
	CueHandBook             []ElemBook               `bson:"CueHandBook"`             //球杆图鉴
	ShopScore               uint32                   `bson:"ShopScore"`               //商城积分
	ProgressActivityList    []ProgressActivityStatus `bson:"ProgressActivityList"`    //进度活动的进度及领奖列表
	DayBattleNum            DayBattleNum             `bson:"DayBattleNum"`            //转盘活动日对局数统记
	DayReceiveStatusNumList []DayReceiveStatusNum    `bson:"DayReceiveStatusNumList"` //转盘活动日领取次数
	ReceivePayLotteryList   []PayLotteryStatus       `bson:"ReceivePayLotteryList"`   //付费抽奖状态列表
	KingRodeActivityList    []KingRodeProgress       `bson:"KingRodeActivityList"`    //王者之路进度表
	FreeShopRefresh         FreeShopRefresh          `bson:"FreeShopRefresh"`         //免费商店刷新参数
	LoginRewardList         []LoginReward            `bson:"LoginRewardList"`         //登录奖励列表
	PointsShopBuyList       []PointsShopBuy          `bson:"PointsShopBuyList"`       //积分商城商品购买记录
	FirstRecharge           []Recharge               `bson:"FirstRecharge"`           //首充列表
}

// 初始化 第一次
func (this *EntityPlayer) InitByFirst(collectionName string, tEntityID uint32) {
	this.CollectionName = collectionName
	this.State = 0
	this.FlagChange = false
	this.FlagKick = false
	this.ObjID = bson.NewObjectId()
	this.EntityID = tEntityID
	this.PlayerID = tEntityID
	this.AccUnique = ""
	this.TimeCreate = tools.GetTimeByTimeStamp(time.Now().Unix())
	this.TimeUpdate = this.TimeCreate
	this.TimeExit = this.TimeCreate
	this.CurrentLoginTime = this.TimeCreate
	this.LastLoginTime = this.TimeCreate
	this.TimeTotal = 0
	this.PlayerLv = 1
	this.PlayerName = ""
	this.PlayerIcon = 50100001
	this.NumExp = 0
	this.NumGold = 10000
	this.NumStone = 10000
	this.BagMax = 200
	this.BagNow = 0
	this.VipLv = 0
	this.VipExp = 0
	this.Sex = 1
	this.RoomId = 0
	this.BehaviorStatus = 0
	this.IsRobot = false
	this.BagList = make([]Item, 0)
	this.TaskList = make([]Task, 0)
	this.RedTipsList = make([]RedTips, 0)
	this.GiftsList = make([]GiveGift, 0)
	this.ReceivingGifts = make([]RecGift, 0)
	this.OpenGifts = false
	this.CharmNum = 0
	this.PeakRankLv = 1
	this.PeakRankExp = 0
	this.FansNum = 0
	this.PeakRankHist = make([]PeakRankHist, 0)
	this.IconFrame = 50200001
	this.ClubId = 0
	this.ClubName = ""
	this.ClubRate = 0
	this.PlayerSign = ""
	this.CueTableId = 10100001
	this.CollectId = 0
	this.Badge = 0
	this.PopularityValue = 0
	this.BattingEffect = 30200001
	this.GoalInEffect = 30300001
	this.CueBall = 30400001
	this.TableCloth = 30100001
	this.PlayerDress = 20000001
	this.PlayerBGImg = 0
	this.ClothingIcon = 0
	this.ClothingCountDown = 50400001
	this.ClothingBubble = 50300001
	this.ReqJoinClub = make([]uint32, 0)
	this.MyFriends = make([]Friend, 0)
	this.GiveGoldList = make([]GiveGold, 0)
	this.DayProgressValue = 0
	this.WeekProgressValue = 0
	this.ClubTags = false
	this.DayProgressReward = make([]ProgressList, 0)
	this.WeekProgressReward = make([]ProgressList, 0)
	this.AchievementLV = 0
	this.AchievementScore = 0
	this.AchievementLVRewardList = make([]AchievementLVReward, 0)
	this.CollectList = make([]Collect, 0)
	this.NextRewardAchievementLV = 1
	this.AchievementList = make([]Achievement, 0)
	this.SignInRewardList = make([]SignInReward, 0)
	this.DailySignInList = make([]DailySignInElement, 0)
	this.ClubNumGold = 0
	this.BoxList = make([]Box, 0)
	this.ExchangeGold = 10000
	this.ClubBadge = 0
	this.MaxUuid = 10000 //不可修改
	this.ShopScore = 10000
	this.CueHandBook = make([]ElemBook, 0)
	this.ProgressActivityList = make([]ProgressActivityStatus, 0)
	this.ReceivePayLotteryList = make([]PayLotteryStatus, 0)
	this.KingRodeActivityList = make([]KingRodeProgress, 0)
	this.LoginRewardList = make([]LoginReward, 0)
	this.FirstRecharge = make([]Recharge, 0)
}

// 获取ObjID
func (this *EntityPlayer) GetObjID() string {
	return this.ObjID.String()
}

// 获取ObjID
func (this *EntityPlayer) GetEntityID() uint32 {
	return this.EntityID
}

// 设置DBConnect
func (this *EntityPlayer) SetDBConnect(collectionName string) {
	this.CollectionName = collectionName
}

// 初始化 by数据结构
func (this *EntityPlayer) InitByData(playerData interface{}) {
	stack.SimpleCopyProperties(this, playerData)
}

// 初始化 by数据库
func (this *EntityPlayer) InitFormDB(tEntityID uint32, tDBConnect *mongodb.DBConnect) (bool, error) {
	if tDBConnect == nil {
		return false, errors.New("tDBConnect == nil")
	}
	err := tDBConnect.GetData(this.CollectionName, "EntityID", tEntityID, this)
	if err != nil {
		return false, err
	}

	return true, err
}

// 插入数据库
func (this *EntityPlayer) InsertEntity(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(this.CollectionName, this)
}

// 保存致数据库
func (this *EntityPlayer) SaveEntity(tDBConnect *mongodb.DBConnect) {
	if tDBConnect == nil {
		return
	}
	tDBConnect.SaveData(this.CollectionName, "_id", this.ObjID, this)
}

// 清理实体
func (this *EntityPlayer) ClearEntity() {
	this.CollectionName = ""
}

// 同步实体
// typeSave: 0定时同步 1根据环境默认 2立即同步
func (this *EntityPlayer) SyncEntity(typeSave uint32) {
	evEntity := new(EntityEvent)
	evEntity.TypeSave = typeSave
	evEntity.TypeEntity = EntityTypePlayer
	evEntity.Entity = this
	event.Emit(UnitSyncentity, evEntity)
}

func (this *EntityPlayer) FlagChang() {
	this.FlagChange = true
}

func (this *EntityPlayer) GetFindIn(tDBConnect *mongodb.DBConnect, query string, slice []uint32, i interface{}) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.GetFindIn(this.CollectionName, query, slice, i)
}

func (this *EntityPlayer) SetExitTime(t string) {
	this.TimeExit = t
}

// 设置用户行为状态
func (this *EntityPlayer) SetBehaviorStatus(s uint8) {
	this.BehaviorStatus = s
}

// 获取用户行为状态
func (this *EntityPlayer) GetBehaviorStatus() uint8 {
	return this.BehaviorStatus
}
func (this *EntityPlayer) SetOnline(o bool) {
	this.Online = o
}

func (this *EntityPlayer) ResetOnline() {
	this.Online = false
}

// 重置角色房间id
func (this *EntityPlayer) ResetRoomId() {
	this.RoomId = 0
}

// 角色房间id
func (this *EntityPlayer) SetRoomId(roomId uint32) {
	this.RoomId = roomId
}

func (this *EntityPlayer) GetClubID() uint32 {
	return this.ClubId
}

func (this *EntityPlayer) ResetClubID() {
	this.ClubId = 0
	this.ClubName = ""
	this.ClubBadge = 0
	this.ClubRate = 0
}

func (this *EntityPlayer) JoinClub(clubID, ClubBadge, clubRate uint32, clubName string) {
	this.ClubId = clubID
	this.ClubName = clubName
	this.ClubBadge = ClubBadge
	this.ClubRate = clubRate
	this.ReqJoinClub = nil
}

func (this *EntityPlayer) IsIReqJoinClub(clubID uint32) bool {
	for _, vl := range this.ReqJoinClub {
		if vl == clubID {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) RemoveClubReq(clubID uint32) {
	for index, val := range this.ReqJoinClub {
		if val == clubID {
			this.ReqJoinClub = append(this.ReqJoinClub[:index], this.ReqJoinClub[(index+1):]...)
		}
	}
}

func (this *EntityPlayer) GetItemFromTableID(tableID uint32) (*Item, int) {
	for key, vl := range this.BagList {
		if vl.TableID == tableID {
			return &vl, key
		}
	}
	return nil, 0
}

func (this *EntityPlayer) AddMyFriends(entityID uint32) *Friend {
	f := Friend{EntityID: entityID, AddTime: uint64(time.Now().Unix())}
	if this.IsGiveGold(entityID) > 0 {
		f.GiveGoldSec = this.IsGiveGold(entityID)
		f.Gold = 100
	}
	this.MyFriends = append(this.MyFriends, f)
	return &f
}

func (this *EntityPlayer) IsInMyFriends(entityID uint32) bool {
	for _, vl := range this.MyFriends {
		if vl.EntityID == entityID {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) GetMyFriends() []Friend {
	return this.MyFriends
}

func (this *EntityPlayer) AddFansList(entityID uint32) {
	if len(this.FansList.List) == 0 {
		this.InitFansList()
	}
	f := Fans{EntityID: entityID, AddTime: uint64(time.Now().Unix())}
	this.FansList.List = append(this.FansList.List, f)
	this.FansNum = uint32(len(this.FansList.List))
}

func (this *EntityPlayer) InitFansList() {
	this.FansList.List = make([]Fans, 0)
}

func (this *EntityPlayer) IsInFansList(entityID uint32) bool {
	for _, vl := range this.FansList.List {
		if vl.EntityID == entityID {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) UpdateFansUnixSec() {
	this.FansList.FansUnixSec = time.Now().Unix()
}

func (this *EntityPlayer) IsHaveFriend(entityID uint32) bool {
	return this.IsInMyFriends(entityID) && this.IsInFansList(entityID)
}

func (this *EntityPlayer) CancelMyFriends(entityID uint32) Friend {
	for index, val := range this.MyFriends {
		if val.EntityID == entityID {
			this.MyFriends = append(this.MyFriends[:index], this.MyFriends[(index+1):]...)
			return val
		}
	}
	return Friend{}
}

func (this *EntityPlayer) DelFans(entityID uint32) {
	for index, val := range this.FansList.List {
		if val.EntityID == entityID {
			this.FansList.List = append(this.FansList.List[:index], this.FansList.List[(index+1):]...)
			break
		}
	}
	this.FansNum = uint32(len(this.FansList.List))
}

func (this *EntityPlayer) IsGiveGold(entityID uint32) int64 {
	for _, vl := range this.GiveGoldList {
		if vl.EntityID == entityID {
			if vl.GiveGoldSec >= tools.GetTodayBeginTime() {
				return vl.GiveGoldSec
			}
		}
	}
	return int64(0)
}

func (this *EntityPlayer) GetMyFriendFromID(entityID uint32) (*Friend, int) {
	for index, vl := range this.MyFriends {
		if vl.EntityID == entityID {
			return &vl, index

		}
	}
	return nil, -1
}

func (this *EntityPlayer) AddGoldToFriend(entityID, gold uint32) *Friend {
	friend, index := this.GetMyFriendFromID(entityID)
	if friend != nil && index > -1 {
		f := this.MyFriends[index]
		f.GiveGoldSec = time.Now().Unix()
		f.Gold = gold
		this.MyFriends[index] = f
		this.AddGiveGoldList(entityID, gold)
		return &f
	}
	return nil
}

func (this *EntityPlayer) GiveGoldNum() int {
	return len(this.GiveGoldList)
}

func (this *EntityPlayer) ResetGiveGoldList() {
	this.GiveGoldList = nil
}

func (this *EntityPlayer) AddGiveGoldList(entityID, gold uint32) {
	this.GiveGoldList = append(this.GiveGoldList, GiveGold{entityID, gold, time.Now().Unix()})
}

func (this *EntityPlayer) IsInGiveGoldList(entityID uint32) bool {
	for _, vl := range this.GiveGoldList {
		if vl.EntityID == entityID {
			return true
		}
	}
	return false
}

// 根据俱乐部商店配置表ID获取数据
func (this *EntityPlayer) GetClubShopItemByTableID(tableID uint32) *ShopItem {
	for _, vl := range this.ClubShopBuyList {
		if vl.TableID == tableID {
			return &vl
		}
	}
	return nil
}

func (this *EntityPlayer) BuyClubShopItemByTableID(itemID, tableID, num uint32) {
	var isBuy bool
	for key, vl := range this.ClubShopBuyList {
		if vl.TableID == tableID {
			isBuy = true
			item := this.ClubShopBuyList[key]
			item.BuyNum += num
			item.BuyTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.ClubShopBuyList[key] = item
		}
	}

	if !isBuy {
		item := new(ShopItem)
		item.TableID = tableID
		item.ItemID = itemID
		item.BuyNum += num
		item.BuyTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		this.ClubShopBuyList = append(this.ClubShopBuyList, *item)
	}
}

// 重置俱乐部购买记录
func (this *EntityPlayer) ResetEntityClubShop() {
	this.ClubShopBuyList = nil
}

func (this *EntityPlayer) IsCanClaimRewardTaskDayProgress() bool {
	if len(this.DayProgressReward) == 0 {
		return false
	}

	if this.DayProgressReward[0].DateStamp < tools.GetTodayBeginTime() {
		return false
	}
	return true
}

func (this *EntityPlayer) IsCanClaimRewardTaskWeekProgress() bool {
	if len(this.WeekProgressReward) == 0 {
		return false
	}

	if this.WeekProgressReward[0].DateStamp < tools.GetThisWeekFirstDate() {
		return false
	}
	return true
}

func (this *EntityPlayer) TaskDayProgressToValue(taskProgressValue uint32) bool {
	return this.DayProgressValue >= taskProgressValue
}

func (this *EntityPlayer) TaskWeekProgressToValue(taskProgressValue uint32) bool {
	return this.WeekProgressValue >= taskProgressValue
}

func (this *EntityPlayer) IsInDayProgressRewardList(progressId uint32) bool {
	if len(this.DayProgressReward) == 0 {
		return false
	}
	for _, vl := range this.DayProgressReward[0].ProgressRewardList {
		if progressId == vl.ProgressID && vl.StateReward == 1 {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) IsInWeekProgressRewardList(progressId uint32) bool {
	if len(this.WeekProgressReward) == 0 {
		return false
	}
	for _, vl := range this.WeekProgressReward[0].ProgressRewardList {
		if progressId == vl.ProgressID && vl.StateReward == 1 {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) TaskDayProgressClaimReward(progressID uint32) {
	for k, vl := range this.DayProgressReward[0].ProgressRewardList {
		if vl.ProgressID == progressID && vl.StateReward == 0 {
			progressReward := vl
			progressReward.StateReward = 1
			progressReward.RewardTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.DayProgressReward[0].ProgressRewardList[k] = progressReward
			break
		}
	}
}

func (this *EntityPlayer) TaskWeekProgressClaimReward(progressID uint32) {
	for k, vl := range this.WeekProgressReward[0].ProgressRewardList {
		if vl.ProgressID == progressID && vl.StateReward == 0 {
			progressReward := vl
			progressReward.StateReward = 1
			progressReward.RewardTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.WeekProgressReward[0].ProgressRewardList[k] = progressReward
			break
		}
	}
}

func (this *EntityPlayer) IsInTaskList(taskID uint32) (*Task, int) {
	for index, vl := range this.TaskList {
		if vl.TaskId == taskID {
			return &vl, index
		}
	}
	return nil, -1
}

func (this *EntityPlayer) IsCanClaimTaskListReward() bool {
	if len(this.TaskList) == 0 {
		return false
	}
	if this.TaskList[0].Timestamp < tools.GetTodayBeginTime() {
		return false
	}
	return true
}

func (this *EntityPlayer) ClaimTaskListReward(index int) {
	this.TaskList[index].StateReward = 1
	this.TaskList[index].Timestamp = time.Now().Unix()
}

func (this *EntityPlayer) AddTaskProgressValue(value uint32) (uint32, uint32) {
	this.DayProgressValue = this.DayProgressValue + value
	this.WeekProgressValue = this.WeekProgressValue + value
	return this.DayProgressValue, this.WeekProgressValue

}

func (this *EntityPlayer) GetCollect(collectID uint32) *Collect {
	for _, vl := range this.CollectList {
		if vl.CollectID == collectID {
			return &vl
		}
	}
	return nil
}

func (this *EntityPlayer) CollectApply(collectID, oldCollectID uint32) *Collect {
	collectResult := new(Collect)
	for key, vl := range this.CollectList {
		if vl.CollectID == collectID && vl.State == 2 {
			collect := vl
			collect.Apply = 1
			collect.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.CollectList[key] = collect
			stack.SimpleCopyProperties(collectResult, collect)
		} else if vl.CollectID == oldCollectID {
			collect := vl
			collect.Apply = 0
			collect.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.CollectList[key] = collect
		}
	}

	return collectResult
}

func (this *EntityPlayer) CollectActivate(collectID uint32) *Collect {
	collectResult := new(Collect)
	for key, vl := range this.CollectList {
		if vl.CollectID == collectID && vl.State == 1 {
			collect := vl
			collect.State = 2
			this.CollectList[key] = collect
			stack.SimpleCopyProperties(collectResult, collect)
		}
	}
	return collectResult
}

func (this *EntityPlayer) IsCanRewardAchievementLV() bool {
	return this.AchievementLV >= this.NextRewardAchievementLV
}

func (this *EntityPlayer) IsInAchievementLVRewardList(achievementLVID uint32) bool {
	return this.AchievementLVRewardList[achievementLVID-1].StateReward == 1
}

func (this *EntityPlayer) AchievementLVClaimReward(achievementLVID uint32) {
	reward := this.AchievementLVRewardList[achievementLVID-1]
	reward.StateReward = 1
	reward.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
	this.AchievementLVRewardList[achievementLVID-1] = reward
}

func (this *EntityPlayer) SaveNextRewardAchievementLV(achievementLVID uint32) {
	this.NextRewardAchievementLV = achievementLVID + 1
}

func (this *EntityPlayer) GetChildAchievementList(achievementID uint32) []ChildAchievement {
	child := make([]ChildAchievement, 0)
	for _, vl := range this.AchievementList {
		if vl.AchievementID == achievementID {
			for _, v := range vl.ChildList {
				child = append(child, v)
			}
		}
	}
	return child
}

func (this *EntityPlayer) UpdateTaskFromConditionID(condID, progress uint32) (resTask *Task) {
	for key, task := range this.TaskList {
		if task.ConditionId == condID {
			if task.State == 0 {
				t := this.TaskList[key]
				t.CompleteProgress = t.CompleteProgress + progress
				if t.CompleteProgress >= task.TaskProgress {
					t.State = 1
				}
				t.Timestamp = time.Now().Unix()
				this.TaskList[key] = t
				resTask = &t
			}
		}
	}
	return
}

func (this *EntityPlayer) UpdateCollectFromConditionID(condID, progress uint32, isTotal bool) (resCollect []*Collect) {
	for key, collect := range this.CollectList {
		if collect.ConditionID == condID {
			if collect.State == 0 {
				t := this.CollectList[key]
				if isTotal {
					if progress <= t.CompleteProgress {
						continue
					}
					t.CompleteProgress = progress
				} else {
					t.CompleteProgress += progress
				}
				if t.CompleteProgress >= collect.TaskProgress {
					t.State = 1
				}
				t.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
				this.CollectList[key] = t
				resCollect = append(resCollect, &t)
			}
		}
	}
	return
}

func (this *EntityPlayer) UpdateAchievementFromConditionID(condID, progress uint32, isTotal bool) (childID []uint32) {
	for key, achi := range this.AchievementList {
		if len(achi.ChildList) == 0 {
			continue
		}
		if achi.ChildList[0].ConditionID == condID {
			if condID != conf.XYCue {
				for k, v := range achi.ChildList {
					if v.State == 1 {
						continue
					}
					childAchievement := v
					if isTotal {
						if progress <= childAchievement.CompleteProgress {
							continue
						}
						childAchievement.CompleteProgress = progress
					} else {
						childAchievement.CompleteProgress += progress
					}
					if childAchievement.CompleteProgress >= v.TaskProgress {
						childAchievement.State = 1
						childID = append(childID, v.ChildID)
					}
					childAchievement.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
					this.AchievementList[key].ChildList[k] = childAchievement
				}
			} else {
				childID = this.UpdateAchievementFromChild(key, achi.ChildList)
			}
		}
	}
	return
}

func (this *EntityPlayer) UpdateAchievementFromChild(index int, child []ChildAchievement) (childIds []uint32) {
	progress := uint32(0)
	resCueQuality := this.SumCueInfoQuality()
	if resCueQuality.QualityS == 0 && resCueQuality.QualitySs == 0 && resCueQuality.QualitySss == 0 {
		return
	}
	for k, v := range child {
		if v.State == 1 {
			continue
		}
		if v.ChildID == conf.XYCueS {
			progress = uint32(resCueQuality.QualityS)
		} else if v.ChildID == conf.XYCueSs {
			progress = uint32(resCueQuality.QualitySs)
		} else if v.ChildID == conf.XYCueSss {
			progress = uint32(resCueQuality.QualitySss)
		}
		childAchievement := v
		if progress <= childAchievement.CompleteProgress {
			continue
		}
		childAchievement.CompleteProgress = progress
		if childAchievement.CompleteProgress >= v.TaskProgress {
			childAchievement.State = 1
			childIds = append(childIds, v.ChildID)
		}
		childAchievement.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		this.AchievementList[index].ChildList[k] = childAchievement
		log.Info("-->UpdateAchievementFromChild-->ChildID-->", v.ChildID, "--childIds-->", childIds)
	}
	return
}

func (this *EntityPlayer) UpdatePlayerAchievement(score uint32) {
	this.AchievementScore = this.AchievementScore + score
}

func (this *EntityPlayer) UpgradeAchievementLV() {
	this.AchievementLV = this.AchievementLV + 1
}

func (this *EntityPlayer) UpdateClubTags() {
	this.ClubTags = true
}

func (this *EntityPlayer) DailySignIn(signType uint32) {
	month, day, now := tools.GetNowTimeMonthAndUnix()
	if !this.IsThisMonthSignIn(month) {
		this.AddDailySignInElement(month, now)
	}
	for k, vl := range this.DailySignInList {
		if vl.MonthKey == month && vl.LastSignInUnixSec < tools.GetTodayBeginTime() {
			sign := vl
			var bitDay, bitSum bitset.BitSet
			bitDay = *bitset.From(sign.SignLog)
			bitSum = *bitset.From(sign.SummarySignLog)
			if !bitDay.Test(uint(day)) {
				bitDay.Set(uint(day))
				bitSum.Set(uint(day))
				sign.SignLog = bitDay.Bytes()
				sign.SummarySignLog = bitSum.Bytes()
				sign.LastSignInUnixSec = time.Now().Unix()
				sign.SignType = signType
				this.DailySignInList[k] = sign
			}
		}
	}

}

func (this *EntityPlayer) IsDailySignIn() (bool, uint32) {
	month, _, _ := tools.GetNowTimeMonthAndUnix()

	for _, vl := range this.DailySignInList {
		if vl.MonthKey == month && vl.LastSignInUnixSec >= tools.GetTodayBeginTime() {
			return true, vl.SignType
		}
	}
	return false, 0
}

func (this *EntityPlayer) AddDailySignInElement(month int, now int64) {
	signLog := make([]uint64, 0)
	signLog = append(signLog, 0)
	sign := DailySignInElement{
		ObjID:              bson.NewObjectId(),
		MonthKey:           month,
		SignLog:            signLog,
		SummarySignLog:     signLog,
		LastSignInUnixSec:  0,
		FirstSignInUnixSec: now,
	}
	this.DailySignInList = append(this.DailySignInList, sign)
}

func (this *EntityPlayer) IsThisMonthSignIn(month int) bool {
	for _, vl := range this.DailySignInList {
		if vl.MonthKey == month {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) GetMonthSignInDays(month int) *bitset.BitSet {
	for _, vl := range this.DailySignInList {
		if vl.MonthKey == month {
			var bit bitset.BitSet
			bit = *bitset.From(vl.SignLog)
			return &bit
		}
	}
	return nil
}

func (this *EntityPlayer) GetMonthSummarySignInDays(month int) *bitset.BitSet {
	for _, vl := range this.DailySignInList {
		if vl.MonthKey == month {
			var bit bitset.BitSet
			bit = *bitset.From(vl.SummarySignLog)
			return &bit
		}
	}
	return nil
}

func (this *EntityPlayer) ResetDailySignInElement(days []string) {
	for _, value := range days {
		month, _ := strconv.Atoi(value[5:7])
		day, _ := strconv.Atoi(value[8:10])
		for k, vl := range this.DailySignInList {
			if vl.MonthKey == month {
				sign := vl
				var bit bitset.BitSet
				bit = *bitset.From(sign.SignLog)
				bit.Clear(uint(day))
				sign.SignLog = bit.Bytes()
				this.DailySignInList[k] = sign
			}
		}
	}
}

func (this *EntityPlayer) AddGiveGoldDataDate() {
	month, day, now := tools.GetNowTimeMonthAndUnix()
	this.GiveGoldData.LastGiveGoldUnixSec = now
	if !this.IsThisMonthGiveGold(month) {
		this.AddGiveGoldElement(month)
	}
	for k, vl := range this.GiveGoldData.GiveElementList {
		if vl.MonthKey == month {
			log := vl
			var bitDay bitset.BitSet
			bitDay = *bitset.From(log.ElementLog)
			if !bitDay.Test(uint(day)) {
				bitDay.Set(uint(day))
				log.ElementLog = bitDay.Bytes()
				this.GiveGoldData.GiveElementList[k] = log
			}
		}
	}
}

func (this *EntityPlayer) IsThisMonthGiveGold(month int) bool {
	for _, vl := range this.GiveGoldData.GiveElementList {
		if vl.MonthKey == month {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) AddGiveGoldElement(month int) {
	sign := GiveElement{
		ObjID:      bson.NewObjectId(),
		MonthKey:   month,
		ElementLog: nil,
	}
	this.GiveGoldData.GiveElementList = append(this.GiveGoldData.GiveElementList, sign)
}

func (this *EntityPlayer) GetMonthGiveGoldDays(month int) *bitset.BitSet {
	for _, vl := range this.GiveGoldData.GiveElementList {
		if vl.MonthKey == month {
			var bit bitset.BitSet
			bit = *bitset.From(vl.ElementLog)
			return &bit
		}
	}
	return nil
}

func (this *EntityPlayer) IsDailyGiveGold() bool {
	return this.GiveGoldData.LastGiveGoldUnixSec >= tools.GetTodayBeginTime()
}

func (this *EntityPlayer) IsClubDailySignIn() bool {
	return this.ClubAttribute.DailySignInUnixSec >= tools.GetTodayBeginTime()
}

func (this *EntityPlayer) ClubDailySignIn() {
	this.ClubAttribute.DailySignInUnixSec = time.Now().Unix()
}

func (this *EntityPlayer) IsGtExitClubUnixSec() bool {
	return time.Now().Unix()-this.ClubAttribute.ExitClubUnixSec > 60*60
}

func (this *EntityPlayer) SetExitClubUnixSec() {
	this.ClubAttribute.ExitClubUnixSec = time.Now().Unix()
	this.ClubAttribute.DailySignInUnixSec = 0
	this.ClubAttribute.ClubProgressRewardList = nil
	for k, vl := range this.ClubAttribute.ClubTaskList {
		if len(vl.ClubDailyTaskList) > 0 {
			value := vl.ClubDailyTaskList[0]
			value.State = 0
			value.CompleteProgress = 0
			this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0] = value
		}
	}
}

func (this *EntityPlayer) UpdateClubReFreshUnix() {
	this.ClubAttribute.ClubReFreshUnix = time.Now().Unix()
}

func (this *EntityPlayer) UpdateClubTaskFromConditionID(condID, progress uint32) {
	for k, vl := range this.ClubAttribute.ClubTaskList {
		if vl.ConditionID == condID {
			if vl.State == 0 {
				if len(vl.ClubDailyTaskList) > 0 {
					dailyTask := this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0]
					dailyTask.CompleteProgress = dailyTask.CompleteProgress + 1
					dailyTask.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
					if dailyTask.CompleteProgress >= dailyTask.TaskProgress {
						dailyTask.State = 1
					}
					this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0] = dailyTask
				}
				if vl.TaskProgress > 0 {
					weekTask := vl
					weekTask.CompleteProgress = weekTask.CompleteProgress + progress
					weekTask.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
					if weekTask.CompleteProgress >= weekTask.TaskProgress {
						weekTask.State = 1
					}
					this.ClubAttribute.ClubTaskList[k] = weekTask
				}
			}
		}
	}
}

func (this *EntityPlayer) GetPlayerClubActiveValue() uint32 {
	return this.ClubAttribute.ClubActiveValue
}

func (this *EntityPlayer) AddClubActiveValue(value uint32) {
	this.ClubAttribute.ClubActiveValue += value
}

// 周6，俱乐部任务重置
func (this *EntityPlayer) ReSetClubAttribute() {
	this.UpdateClubReFreshUnix()
	this.ClubAttribute.ClubActiveValue = 0
	for k, vl := range this.ClubAttribute.ClubProgressRewardList {
		n := vl
		n.StateReward = 0
		n.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		this.ClubAttribute.ClubProgressRewardList[k] = n
	}
	for k, vl := range this.ClubAttribute.ClubTaskProgressList {
		n := vl
		n.StateReward = 0
		n.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		this.ClubAttribute.ClubTaskProgressList[k] = n
	}
	for k, vl := range this.ClubAttribute.ClubTaskList {
		if len(vl.ClubDailyTaskList) > 0 {
			dailyTask := this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0]
			dailyTask.CompleteProgress = 0
			dailyTask.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			dailyTask.State = 0
			this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0] = dailyTask
		}
		n := vl
		n.CompleteProgress = 0
		n.State = 0
		n.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		this.ClubAttribute.ClubTaskList[k] = n
	}
}

func (this *EntityPlayer) DailyReSetClubAttribute() {
	this.UpdateClubReFreshUnix()
	for k, vl := range this.ClubAttribute.ClubTaskList {
		if len(vl.ClubDailyTaskList) > 0 {
			dailyTask := this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0]
			dailyTask.CompleteProgress = 0
			dailyTask.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			dailyTask.State = 0
			this.ClubAttribute.ClubTaskList[k].ClubDailyTaskList[0] = dailyTask
		}
	}
}

func (this *EntityPlayer) GetNumStone() uint32 {
	return this.NumStone
}

func (this *EntityPlayer) IsClubReFreshUnix() bool {
	return this.ClubAttribute.ClubReFreshUnix >= tools.GetTodayBeginTime()
}

func (this *EntityPlayer) IsClaimClubProgressReward(clubProgressID uint32) bool {
	for _, vl := range this.ClubAttribute.ClubProgressRewardList {
		if vl.ProgressID == clubProgressID && vl.StateReward == 1 {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) ClaimClubProgressReward(clubProgressID uint32) {
	for k, vl := range this.ClubAttribute.ClubProgressRewardList {
		if vl.ProgressID == clubProgressID && vl.StateReward == 0 {
			reward := vl
			reward.StateReward = 1
			reward.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.ClubAttribute.ClubProgressRewardList[k] = reward
		}
	}
}

func (this *EntityPlayer) IsClaimMyClubTaskProgressReward(progressID uint32) bool {
	for _, vl := range this.ClubAttribute.ClubTaskProgressList {
		if vl.ProgressID == progressID && vl.StateReward == 1 {
			return true
		}
	}
	return false
}

func (this *EntityPlayer) ClaimMyClubTaskProgressReward(progressID uint32) {
	for k, vl := range this.ClubAttribute.ClubTaskProgressList {
		if vl.ProgressID == progressID && vl.StateReward == 0 {
			reward := vl
			reward.StateReward = 1
			reward.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.ClubAttribute.ClubTaskProgressList[k] = reward
		}
	}
}

func (this *EntityPlayer) GetBoxList() []Box {
	return this.BoxList
}

func (this *EntityPlayer) AddBoxInit(num int) {
	if len(this.BoxList) > 0 {
		return
	}
	for i := 1; i <= num; i++ {
		b := new(Box)
		b.ObjID = bson.NewObjectId()
		b.BoxNum = uint32(i)
		this.BoxList = append(this.BoxList, *b)
	}
}

func (this *EntityPlayer) GetEmptyBoxNum() int {
	n := 0
	for _, vl := range this.BoxList {
		if vl.BoxID == 0 {
			n++
		}
	}
	return n
}

func (this *EntityPlayer) GetBoxNum() int {
	n := 0
	for _, vl := range this.BoxList {
		if vl.BoxID > 0 {
			n++
		}
	}
	return n
}

func (this *EntityPlayer) AddBox(boxID, gameType, roomType uint32) {
	for k, vl := range this.BoxList {
		if vl.BoxID == 0 {
			b := vl
			b.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			b.BoxID = boxID
			b.RoomType = roomType
			b.GameType = gameType
			this.BoxList[k] = b
			break
		}
	}
}

func (this *EntityPlayer) BoxUnlock(ID string, boxID uint32, timeUnix int64) int64 {
	for k, vl := range this.BoxList {
		if vl.ObjID.Hex() == ID && vl.BoxID == boxID {
			b := vl
			b.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			b.UnlockTimeStamp = time.Now().Unix() + timeUnix
			this.BoxList[k] = b
			return this.BoxList[k].UnlockTimeStamp
		}
	}
	return 0
}

func (this *EntityPlayer) BoxFastReward(ID string, boxID, timeUnix uint32) *Box {
	for k, vl := range this.BoxList {
		if vl.ObjID.Hex() == ID && vl.BoxID == boxID {
			b := vl
			b.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			b.ReduceTime += timeUnix
			this.BoxList[k] = b
			return &b
		}
	}
	return nil
}

func (this *EntityPlayer) GetBoxCountDown() int64 {
	for _, vl := range this.BoxList {
		if vl.UnlockTimeStamp > 0 {
			return vl.UnlockTimeStamp - time.Now().Unix()
		}
	}
	return 0
}

func (this *EntityPlayer) GetBox(ID string) *Box {
	for _, vl := range this.BoxList {
		if vl.ObjID.Hex() == ID {
			return &vl
		}
	}
	return nil
}

func (this *EntityPlayer) BoxClaim(ID string, boxID uint32) *Box {
	for k, vl := range this.BoxList {
		if vl.ObjID.Hex() == ID && vl.BoxID == boxID {
			b := vl
			b.AddTime = ""
			b.UnlockTimeStamp = 0
			b.ReduceTime = 0
			b.BoxID = 0
			b.RoomType = 0
			b.GameType = 0
			this.BoxList[k] = b
			return &b
		}
	}
	return nil
}

func (this *EntityPlayer) SetPlayerState(status uint32) {
	this.State = status
	return
}

func (this *EntityPlayer) GetMaxUuid() uint32 {
	maxuuid := int32(this.MaxUuid)
	newMax := uint32(tools.GetEntityID(&maxuuid))
	this.MaxUuid = newMax
	return newMax
}

func (this *EntityPlayer) GetCueHandBook(cueID uint32) *ElemBook {
	for _, val := range this.CueHandBook {
		if val.CueID == cueID {
			return &val
		}
	}
	return nil
}

func (this *EntityPlayer) CueHandBookActivate(cueID uint32) *ElemBook {
	for key, val := range this.CueHandBook {
		if val.CueID == cueID && val.State == 1 {
			book := val
			book.State = 2
			book.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.CueHandBook[key] = book
			return &book
		}
	}
	return nil
}

func (this *EntityPlayer) SumCueInfoQuality() *CueQualityS {
	cueQualityS := new(CueQualityS)
	for _, value := range this.BagList {
		switch value.CueInfo.Quality {
		case conf.CueQuality4:
			cueQualityS.QualityS++
		case conf.CueQuality5:
			cueQualityS.QualitySs++
		case conf.CueQuality6:
			cueQualityS.QualitySss++
		}
	}
	return cueQualityS
}

func (this *EntityPlayer) GetGiftsList(entityID uint32) (*GiveGift, int) {
	for key, val := range this.GiftsList {
		if val.EntityID == entityID {
			return &val, key
		}
	}
	return nil, 0
}

func (this *EntityPlayer) GetReceivingGifts(entityID uint32) (*RecGift, int) {
	for key, val := range this.ReceivingGifts {
		if val.EntityID == entityID {
			return &val, key
		}
	}
	return nil, 0
}

func (this *EntityPlayer) RefreshFreeShop() {
	this.FreeShopRefresh.RefreshAdTimes -= uint32(1)
	this.FreeShopRefresh.LastRefreshStamp = time.Now().Unix()
}
