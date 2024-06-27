package table

import (
	"BilliardServer/Common/entity"
	"BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	ACHIEVEMENT_CFG = "AchievementCfg"
)

const (
	TextCn = "text_cn"
	TextEn = "text_en"
)

type Table struct {
	AchievementElementCfg          map[string]*AchievementElementCfg
	AchievementLvCfg               map[string]*AchievementLvCfg
	AchievementCfg                 map[string]*AchievementCfg
	ClothingCfg                    map[string]*ClothingCfg
	ClubCfg                        map[string]*ClubCfg
	ClubShopCfg                    map[string]*ClubShopCfg
	ClubTaskCfg                    map[string]*ClubTaskCfg
	ClubTaskProgressCfg            map[string]*ClubTaskProgressCfg
	ClubProgressRewardCfg          map[string]*ClubProgressRewardCfg
	ClubDailySignCfg               map[string]*ClubDailySignCfg
	ClubRateRewardCfg              map[string]*ClubRateRewardCfg
	BattleClubProgressRewardCfg    map[string]*BattleClubProgressRewardCfg
	BattleClubScoreRewardCfg       map[string]*BattleClubScoreRewardCfg
	CollectCfg                     map[string]*CollectCfg
	ConditionalCfg                 map[string]*ConditionalCfg
	ConstTextCfg                   map[string]*ConstTextCfg
	ConstCfg                       map[string]*ConstCfg
	CueCfg                         map[string]*CueCfg
	DressCfg                       map[string]*DressCfg
	EffectCfg                      map[string]*EffectCfg
	ItemCfg                        map[string]*ItemCfg
	DailySignCfg                   map[string]*DailySignCfg
	PlayerLevelCfg                 map[string]*PlayerLevelCfg
	RoomCfg                        map[string]*RoomCfg
	ShopCfg                        map[string]*ShopCfg
	TaskCfg                        map[string]*TaskCfg
	TaskProgressCfg                map[string]*TaskProgressCfg
	TollgateExpertAdvancedCfg      map[string]*TollgateExpertAdvancedCfg
	TollgateCfg                    map[string]*TollgateCfg
	DanCfg                         map[string]*DanCfg
	VipCfg                         map[string]*VipCfg
	TaskList                       []TaskData
	TaskMap                        map[uint32]TaskData
	TaskConditionMap               map[uint32]TaskData
	DailySigInCfgMap               map[string]*DailySignCfg
	BattleClubScoreRewardCfgMap    map[string]*BattleClubScoreRewardCfg
	BattleClubProgressRewardCfgMap map[string]*BattleClubProgressRewardCfg
	ClubTaskCfgMap                 map[string]*ClubTaskCfg
	BoxCfg                         map[string]*BoxCfg
	BoxCfgList                     []BoxCfg
	ItemTimeCfg                    map[string]*ItemTimeCfg
	EightBallRoomCfg               map[string]*EightBallRoomCfg
	EightBallRoomCfgMap            map[uint32]*EightBallRoomCfg
	EightBallRoomSlice             []*EightBallRoomCfg
	CueHandbookCfg                 map[string]*CueHandbookCfg
	SpecialShopCfg                 map[string]*SpecialShopCfg
	AchievementLvCfgData           []*AchievementLvCfg
	PropertyItemCfg                map[string]*PropertyItemCfg
	RandNameCfg                    map[string]*RandNameCfg
	RandPreNameCfg                 map[string]*RandPreNameCfg
	KingCfg                        map[string]*KingCfg
	FreeStoreCfg                   map[string]*FreeStoreCfg
	FirstRechargeCfg               map[string]*FirstRechargeCfg
	DefaultDress                   map[uint32]uint32
	DefaultPlayerIcon              map[uint32]uint32
}

func (c *Table) Init(mode string) {
	c.TaskList = make([]TaskData, 0)
	c.TaskMap = make(map[uint32]TaskData, 0)
	c.TaskConditionMap = make(map[uint32]TaskData, 0)
	c.DefaultDress, c.DefaultPlayerIcon = make(map[uint32]uint32, 0), make(map[uint32]uint32, 0)
	cfgPath := tools.GetModeTablePath(mode)
	if cfgPath == "" {
		log.Error("-->TableManager:cfgPath is empty!")
		return
	}

	filePtr, err := os.Open(cfgPath + "table.json")
	if err != nil {
		log.Error("-->TableManager:table load losed!", err, cfgPath)
		return
	}
	defer filePtr.Close()
	// 创建json解码器
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&c)
	c.SetTaskSlice()
	c.SetDailySigInCfg()
	c.SetBattleClubScoreRewardCfg()
	c.SetClubTaskCondID()
	c.SetBoxCfgList()
	c.SetEightBallRoomCfg()
	c.SetBattleClubProgressRewardCfg()
	c.AchievementLVInit()
	c.InitDefaultDress()
	c.InitPlayerIcon()

	if err != nil {
		log.Error("-->TableManager:table decode  losed", err)
		return
	}
	log.Info("-->TableManager:table load completed!")
	//发送table解析完成事件
	//event.Emit(model.Ek_LoadTableComplete, nil)
}

type CommonCfg struct {
	ItemType uint32
	SubType  uint32
	LifeTime uint32
}

func (c *Table) IsExistTable(tableID string) (cfg CommonCfg) {
	id, _ := strconv.Atoi(tableID)
	itemType, _ := tools.GetItemTypeByTableId(uint32(id))
	if itemType == 0 {
		return
	}
	if itemType == consts.Cue && c.GetCueCfg(tableID) != nil {
		cfg = CommonCfg{consts.Cue, c.GetCueCfg(tableID).TypeN, c.GetCueCfg(tableID).LifeTime}
	} else if itemType == consts.Dress && c.GetDressCfg(tableID) != nil {
		cfg = CommonCfg{consts.Dress, c.GetDressCfg(tableID).TypeN, c.GetDressCfg(tableID).LifeTime}
	} else if itemType == consts.Effect && c.GetCloneEffectCfg(tableID) != nil {
		cfg = CommonCfg{consts.Effect, c.GetCloneEffectCfg(tableID).TypeN, c.GetCloneEffectCfg(tableID).LifeTime}
	} else if itemType == consts.Item && c.GetCloneItemCfg(tableID) != nil {
		cfg = CommonCfg{consts.Item, c.GetCloneItemCfg(tableID).TypeN, c.GetCloneItemCfg(tableID).LifeTime}
	} else if itemType == consts.Clothing && c.GetCloneClothingCfg(tableID) != nil {
		cfg = CommonCfg{consts.Clothing, c.GetCloneClothingCfg(tableID).TypeN, c.GetCloneClothingCfg(tableID).LifeTime}
	}
	return cfg
}

func (c *Table) GetCueCfg(tableID string) *CueCfg {
	target := c.CueCfg[tableID]
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetCueCfgById(tableID uint32) *CueCfg {
	target := c.CueCfg[strconv.Itoa(int(tableID))]
	if target == nil {
		return nil
	}

	return target
}
func (c *Table) GetCueCfgMap() map[string]*CueCfg {
	return c.CueCfg
}

func (c *Table) GetConstMap() map[string]*ConstCfg {
	return c.ConstCfg
}

func (c *Table) GetCueCharmScore(tableID uint32) uint32 {
	target := c.CueCfg[strconv.Itoa(int(tableID))]
	if target == nil {
		return 0
	}
	return target.CharmScore
}

func (c *Table) GetCloneDressCfg(tableID string) *DressCfg {
	target := c.DressCfg[tableID]
	if target == nil {
		return nil
	}
	source := stack.DeepClone(target).(*DressCfg)
	return source
}

func (c *Table) GetCloneEffectCfg(tableID string) *EffectCfg {
	target := c.EffectCfg[tableID]
	if target == nil {
		return nil
	}
	source := stack.DeepClone(target).(*EffectCfg)
	return source
}

func (c *Table) GetEffectCfgById(tableID string) *EffectCfg {
	target := c.EffectCfg[tableID]
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetCloneItemCfg(tableID string) *ItemCfg {
	target := c.ItemCfg[tableID]
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetItemCfgById(tableID string) *ItemCfg {
	target := c.ItemCfg[tableID]
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetCloneClothingCfg(tableID string) *ClothingCfg {
	target := c.ClothingCfg[tableID]
	if target == nil {
		return nil
	}
	source := stack.DeepClone(target).(*ClothingCfg)
	return source
}

func (c *Table) GetClothingCfgById(tableID string) *ClothingCfg {
	target := c.ClothingCfg[tableID]
	if target == nil {
		return nil
	}
	return target
}
func (c *Table) GetAllShopCfg() map[string]*ShopCfg {
	target := c.ShopCfg
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetDressCfgMap() map[string]*DressCfg {
	return c.DressCfg
}

func (c *Table) GetDressCfg(tableID string) *DressCfg {
	target := c.DressCfg[tableID]
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetClothingCfgMap() map[string]*ClothingCfg {
	return c.ClothingCfg
}

func (c *Table) GetDanCfgMap() map[string]*DanCfg {
	return c.DanCfg
}

func (c *Table) GetPlayerLevelCfgMap() map[string]*PlayerLevelCfg {
	return c.PlayerLevelCfg
}

func (c *Table) GetClubLVNum(Level uint32) uint32 {
	club := c.ClubCfg[strconv.Itoa(int(Level))]
	if club == nil {
		return 0
	}
	return club.Num
}

func (c *Table) GetCluMasterNum(Level uint32) uint32 {
	club := c.ClubCfg[strconv.Itoa(int(Level))]
	if club == nil {
		return 0
	}
	return club.MasterNum
}

// 根据Id获取俱乐部商店信息
func (c *Table) GetClubShopCfgMap() map[string]*ClubShopCfg {
	return c.ClubShopCfg
}

// 根据Id获取俱乐部商店信息
func (c *Table) GetClubShopCfgById(tableID uint32) *ClubShopCfg {
	target := c.ClubShopCfg[strconv.Itoa(int(tableID))]
	if target == nil {
		return nil
	}
	return target
}

type TaskData struct {
	TaskID        uint32
	ConditionID   uint32
	Condition     uint32
	ConditionType uint32
}

func (c *Table) SetTaskSlice() {
	for _, v := range c.TaskCfg {
		var task TaskData
		task.TaskID = v.TableID
		if len(v.Condition) > 1 {
			task.Condition = v.Condition[1]
		} else {
			task.Condition = 0
		}
		task.ConditionID = v.Condition[0]
		if c.GetConditionalCfg(task.ConditionID) != nil {
			task.ConditionType = c.GetConditionalCfg(task.ConditionID).TypeN
		}
		c.TaskList = append(c.TaskList, task)
		c.TaskMap[task.TaskID] = task
		c.TaskConditionMap[task.ConditionID] = task
	}

	sort.Slice(c.TaskList, func(i, j int) bool {
		return c.TaskList[i].TaskID < c.TaskList[j].TaskID
	})

	log.Info("-->SetTaskSlice", c.TaskList, "-->SetTaskMap", c.TaskMap, "-->SetTaskConditionMap", c.TaskConditionMap)
}

func (c *Table) GetConditionalCfg(cKey uint32) *ConditionalCfg {
	target := c.ConditionalCfg[strconv.Itoa(int(cKey))]
	if target == nil {
		return nil
	}
	return target
}

func (c *Table) GetVipCfgMap() map[string]*VipCfg {
	return c.VipCfg
}

func (c *Table) GetTaskList() []TaskData {
	return c.TaskList
}

func (c *Table) GetTaskDataFromTaskID(taskID uint32) *TaskData {
	target, ok := c.TaskMap[taskID]
	if !ok {
		return nil
	}
	return &target
}

func (c *Table) GetTaskDataFromConditionID(conditionID uint32) *TaskData {
	target, ok := c.TaskConditionMap[conditionID]
	if !ok {
		return nil
	}
	return &target
}

func (c *Table) GetTaskProgressCfg(taskProgressKey uint32) *TaskProgressCfg {
	target, ok := c.TaskProgressCfg[strconv.Itoa(int(taskProgressKey))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetTaskCfg(taskKey uint32) *TaskCfg {
	target, ok := c.TaskCfg[strconv.Itoa(int(taskKey))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetConditionalCfgMap() map[string]*ConditionalCfg {
	return c.ConditionalCfg
}

func (c *Table) GetAchievementLvCfg(achievementLVid uint32) *AchievementLvCfg {
	target, ok := c.AchievementLvCfg[strconv.Itoa(int(achievementLVid))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetAchievementLVCfg() map[string]*AchievementLvCfg {
	return c.AchievementLvCfg
}

func (c *Table) GetAchievementCfg() map[string]*AchievementCfg {
	return c.AchievementCfg
}

func (c *Table) GetAchievementElementCfg(key uint32) *AchievementElementCfg {
	target, ok := c.AchievementElementCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetAchievementElementCfgScore(childId uint32) uint32 {
	target, ok := c.AchievementElementCfg[strconv.Itoa(int(childId))]
	if !ok {
		return 0
	}
	return target.Score
}

func (c *Table) GetCollectCfg() map[string]*CollectCfg {
	return c.CollectCfg
}

func (c *Table) GetTaskProgressData() map[string]*TaskProgressCfg {
	return c.TaskProgressCfg
}

func (c *Table) SetDailySigInCfg() {
	c.DailySigInCfgMap = make(map[string]*DailySignCfg, 0)
	for _, v := range c.DailySignCfg {
		c.DailySigInCfgMap[fmt.Sprintf("%d_%d", v.Day, v.TypeN)] = v
	}
	log.Info("-->SetDailySigInCfg-->", c.DailySigInCfgMap)
}

func (c *Table) GetDailySigInCfgMap(key string) *DailySignCfg {
	target, ok := c.DailySigInCfgMap[key]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetClubProgressRewardCfg() map[string]*ClubProgressRewardCfg {
	return c.ClubProgressRewardCfg
}

func (c *Table) GetClubProgressRewardCfgMap(key uint32) *ClubProgressRewardCfg {
	target, ok := c.ClubProgressRewardCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetClubTaskProgressCfg(key uint32) *ClubTaskProgressCfg {
	target, ok := c.ClubTaskProgressCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetClubTaskCfg() map[string]*ClubTaskCfg {
	return c.ClubTaskCfg
}

func (c *Table) GetClubTaskScoreFromCondID(key uint32) uint32 {
	target, ok := c.ClubTaskCfgMap[strconv.Itoa(int(key))]
	if !ok {
		return 0
	}
	return target.Score
}

func (c *Table) GetClubCfg(key uint32) *ClubCfg {
	target, ok := c.ClubCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) SetClubTaskCondID() {
	c.ClubTaskCfgMap = make(map[string]*ClubTaskCfg, 0)
	for _, v := range c.ClubTaskCfg {
		c.ClubTaskCfgMap[fmt.Sprintf("%d", v.WeekCondition[0])] = v
	}
	log.Info("-->SetClubTaskCondID-->", c.ClubTaskCfgMap)
}

func (c *Table) GetClubProgressRewardList() []entity.ClubProgressReward {
	clubProgressRewardList := make([]entity.ClubProgressReward, 0)
	for _, vl := range c.ClubProgressRewardCfg {
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

func (c *Table) GetClubTaskProgressList() []entity.ClubTaskProgress {
	clubTaskProgressList := make([]entity.ClubTaskProgress, 0)
	for _, vl := range c.ClubTaskProgressCfg {
		progressReward := new(entity.ClubTaskProgress)
		progressReward.ProgressID = vl.TableID
		progressReward.StateReward = 0
		progressReward.Progress = vl.Progress
		progressReward.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		clubTaskProgressList = append(clubTaskProgressList, *progressReward)
	}
	sort.Slice(clubTaskProgressList, func(i, j int) bool {
		return clubTaskProgressList[i].ProgressID < clubTaskProgressList[j].ProgressID
	})
	return clubTaskProgressList
}

func (c *Table) GetClubTaskList() []entity.ClubWeekTask {
	clubTaskList := make([]entity.ClubWeekTask, 0)
	for _, vl := range c.ClubTaskCfg {
		weekTask := new(entity.ClubWeekTask)
		if len(vl.WeekCondition) > 1 {
			weekTask.TaskID = vl.TableID
			weekTask.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			weekTask.CompleteProgress = 0
			weekTask.ConditionID = vl.WeekCondition[0]
			weekTask.TaskProgress = vl.WeekCondition[1]
			weekTask.State = 0
			weekTask.ClubDailyTaskList = make([]entity.ClubDailyTask, 0)
			if len(vl.DayCondition) > 1 && vl.DayCondition[1] > 0 {
				dailyTask := new(entity.ClubDailyTask)
				dailyTask.State = 0
				dailyTask.TaskProgress = vl.DayCondition[1]
				dailyTask.CompleteProgress = 0
				dailyTask.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
				weekTask.ClubDailyTaskList = append(weekTask.ClubDailyTaskList, *dailyTask)
			}
			clubTaskList = append(clubTaskList, *weekTask)
		}
	}
	sort.Slice(clubTaskList, func(i, j int) bool {
		return clubTaskList[i].TaskID < clubTaskList[j].TaskID
	})

	return clubTaskList
}

func (c *Table) SetBattleClubScoreRewardCfg() {
	c.BattleClubScoreRewardCfgMap = make(map[string]*BattleClubScoreRewardCfg, 0)
	for _, v := range c.BattleClubScoreRewardCfg {
		c.BattleClubScoreRewardCfgMap[fmt.Sprintf("%d_%d", v.TypeN, v.RoomType)] = v
	}
	log.Info("-->SetBattleClubScoreRewardCfg-->", c.BattleClubScoreRewardCfgMap)
}

func (c *Table) SetBattleClubProgressRewardCfg() {
	c.BattleClubProgressRewardCfgMap = make(map[string]*BattleClubProgressRewardCfg, 0)
	for _, v := range c.BattleClubProgressRewardCfg {
		c.BattleClubProgressRewardCfgMap[fmt.Sprintf("%d_%d", v.TypeN, v.RoomType)] = v
	}
	log.Info("-->SetBattleClubProgressRewardCfg-->", c.BattleClubProgressRewardCfgMap)
}

func (c *Table) GetBattleClubScoreRewardCfgMap(key string) *BattleClubScoreRewardCfg {
	target, ok := c.BattleClubScoreRewardCfgMap[key]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetBattleClubProgressRewardCfgMap(key string) *BattleClubProgressRewardCfg {
	target, ok := c.BattleClubProgressRewardCfgMap[key]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetClubRateRewardCfg(key uint32) *ClubRateRewardCfg {
	target, ok := c.ClubRateRewardCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetBoxCfg(key uint32) *BoxCfg {
	target, ok := c.BoxCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) SetBoxCfgList() {
	c.BoxCfgList = make([]BoxCfg, 0)
	for _, vl := range c.BoxCfg {
		c.BoxCfgList = append(c.BoxCfgList, *vl)
	}
	sort.Slice(c.BoxCfgList, func(i, j int) bool {
		return c.BoxCfgList[i].TableID < c.BoxCfgList[j].TableID
	})
}

func (c *Table) GetItemTimeCfgById(key uint32) *ItemTimeCfg {
	return c.ItemTimeCfg[strconv.Itoa(int(key))]
}

func (c *Table) GetConstTextFromID(id uint32, text string) string {
	target, ok := c.ConstTextCfg[strconv.Itoa(int(id))]
	if !ok {
		return ""
	}

	if text == TextCn {
		return target.TextCN
	} else if text == TextEn {
		return target.TextEN
	}

	return target.TextCN
}

func (c *Table) GetEightBallRoomCfg(id uint32) *EightBallRoomCfg {
	target, ok := c.EightBallRoomCfg[strconv.Itoa(int(id))]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) SetEightBallRoomCfg() {
	c.EightBallRoomCfgMap = make(map[uint32]*EightBallRoomCfg, 0)
	c.EightBallRoomSlice = make([]*EightBallRoomCfg, 0)
	for _, v := range c.EightBallRoomCfg {
		c.EightBallRoomCfgMap[v.Level] = v
		c.EightBallRoomSlice = append(c.EightBallRoomSlice, v)
	}
	sort.Slice(c.EightBallRoomSlice, func(i, j int) bool {
		return c.EightBallRoomSlice[i].TableID < c.EightBallRoomSlice[j].TableID
	})
}

func (c *Table) GetEightBallRoomCfgLevel(level uint32) *EightBallRoomCfg {
	target, ok := c.EightBallRoomCfgMap[level]
	if !ok {
		return nil
	}
	return target
}

func (c *Table) GetCueHandbookCfg() map[string]*CueHandbookCfg {
	return c.CueHandbookCfg
}

func (c *Table) GetAllSpecialShopCfg() map[string]*SpecialShopCfg {
	return c.SpecialShopCfg
}

func (c *Table) GetSpecialShopCfg(key uint32) *SpecialShopCfg {
	target, ok := c.SpecialShopCfg[strconv.Itoa(int(key))]
	if !ok {
		return nil
	}
	return target
}

// 初始化
func (c *Table) AchievementLVInit() {
	c.AchievementLvCfgData = make([]*AchievementLvCfg, 0)
	for _, vl := range c.AchievementLvCfg {
		c.AchievementLvCfgData = append(c.AchievementLvCfgData, vl)
	}

	sort.Slice(c.AchievementLvCfgData, func(i, j int) bool {
		return c.AchievementLvCfgData[i].TableID < c.AchievementLvCfgData[j].TableID
	})
}

// 成就升级
func (c *Table) IsUpgradeAchievementLV(AchievementScore, AchievementLV uint32) (isUp bool) {
	for _, val := range c.AchievementLvCfgData {
		if AchievementScore >= val.Score && AchievementLV < val.TableID+1 && val.Score > 0 {
			isUp = true
			break
		}
	}
	return isUp
}

func (c *Table) GetPlayerRandName() string {
	PreName, Name := "", ""
	for _, val := range c.RandPreNameCfg {
		PreName = val.EN
		break
	}
	for _, val := range c.RandNameCfg {
		Name = val.EN
		break
	}
	return PreName + " " + Name
}

func (c *Table) GetAllFreeStoreCfg() map[string]*FreeStoreCfg {
	target := c.FreeStoreCfg
	if target == nil {
		return nil
	}
	return target
}

// 设置默认服装
func (c *Table) InitDefaultDress() {
	config, ok := c.GetConstMap()["16"]
	if !ok || ok && len(config.Paramater20) <= 1 {
		log.Error("initDefaultDress is err!")
		return
	}
	for _, val := range config.Paramater20 {
		if val[0] == consts.USER_MAN {
			c.DefaultDress[consts.USER_MAN] = val[1]
		} else if val[0] == consts.USER_WOMEN {
			c.DefaultDress[consts.USER_WOMEN] = val[1]
		}
	}

	log.Info("-->DefaultDress->", c.DefaultDress)
}

// 设置默认头像
func (c *Table) InitPlayerIcon() {
	config, ok := c.GetConstMap()["17"]
	if !ok || ok && len(config.Paramater20) <= 1 {
		log.Error("InitPlayerIcon is err!")
		return
	}
	for _, val := range config.Paramater20 {
		if val[0] == consts.USER_MAN {
			c.DefaultPlayerIcon[consts.USER_MAN] = val[1]
		} else if val[0] == consts.USER_WOMEN {
			c.DefaultPlayerIcon[consts.USER_WOMEN] = val[1]
		}
	}

	log.Info("-->InitPlayerIcon->", c.DefaultPlayerIcon)
}
