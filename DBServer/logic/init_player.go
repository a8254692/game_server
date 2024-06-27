package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/log"
	"BilliardServer/Util/tools"
	"fmt"
	"gitee.com/go-package/carbon/v2"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

/***
 *@disc:
 *@author: lsj
 *@date: 2023/11/24
 */

type _InitPlayer struct {
	AchievementList     []entity.Achievement
	AchievementLVReward []entity.AchievementLVReward
	CollectList         []entity.Collect
	CueHandBook         []entity.ElemBook
	DefaultItem         map[string]uint32
}

var InitPlayerMr _InitPlayer

func (c *_InitPlayer) Init() {
	c.DefaultItem = make(map[string]uint32, 0)
	c.initDefaultItem()
	c.PlayerAchievementListInit()
	c.PlayerAchievementLVInit()
	c.PlayerCollectListInit()
	c.PlayerCueHandBookListInit()
}

// 初始化默认道具
func (c *_InitPlayer) initDefaultItem() {
	config, ok := Table.GetConstMap()["18"]
	if !ok || ok && (len(config.Paramater20) == 0) {
		log.Error("initDefaultItem is err!")
		return
	}
	for _, val := range config.Paramater20 {
		if len(val) > 2 {
			c.DefaultItem[fmt.Sprintf("%d_%d", val[0], val[1])] = val[2]
		}
	}
	log.Info("-->initDefaultItem-->", c.DefaultItem)
}

// 初始化背包道具
func (c *_InitPlayer) InitPlayerData(EntityID uint32) {
	tEntity := Entity.EmEntityPlayer.GetEntityByID(EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	tEntityPlayer.AchievementList = c.AchievementList
	tEntityPlayer.AchievementLVRewardList = c.AchievementLVReward
	tEntityPlayer.CollectList = c.CollectList
	tEntityPlayer.CueHandBook = c.CueHandBook

	c.InitBackpack(tEntityPlayer)
	c.PlayerTaskInit(EntityID)
	tEntityPlayer.AddBoxInit(conf.MaxBoxNum)
	tEntityPlayer.SaveEntity(DBConnect)

	log.Info("-->InitBackPack-->Complete,EntityID:", EntityID)
}

// 初始化角色背包
func (c *_InitPlayer) InitBackpack(tEntityPlayer *entity.EntityPlayer) {
	val := reflect.ValueOf(tEntityPlayer)
	for key, vl := range conf.BaseTableInfo {
		itemKey := strings.Split(key, "_")
		itemType, subType := 0, 0
		itemType, _ = strconv.Atoi(itemKey[0])
		if len(itemKey) > 1 {
			subType, _ = strconv.Atoi(itemKey[1])
		}
		fieldVal := val.Elem().FieldByName(vl)
		if !fieldVal.IsValid() || fieldVal.Uint() == 0 {
			continue
		}

		newItemInfo := new(entity.Item)
		newItemInfo.TableID = uint32(fieldVal.Uint())
		newItemInfo.ItemID = tEntityPlayer.GetMaxUuid()
		newItemInfo.ItemNum = uint32(1)
		newItemInfo.ItemType = uint32(itemType)
		newItemInfo.SubType = uint32(subType)
		newItemInfo.ItemStatus = conf.ItemUse
		if itemType == conf.Cue {
			q, a, _ := c.getKeyByTableId(uint32(fieldVal.Uint()))
			newItemInfo.CueInfo.Quality = uint32(q)
			newItemInfo.CueInfo.Star = uint32(a)
		}
		tEntityPlayer.BagList = append(tEntityPlayer.BagList, *newItemInfo)
	}

	tEntityPlayer.BagList = append(tEntityPlayer.BagList, *c.addDress(tEntityPlayer))

	for _, item := range Table.ClothingCfg {
		if item.TypeN < conf.Clothing_5 {
			continue
		}
		newItemInfo := new(entity.Item)
		newItemInfo.TableID = item.TableID
		newItemInfo.ItemID = tEntityPlayer.GetMaxUuid()
		newItemInfo.ItemNum = uint32(1)
		newItemInfo.ItemType = conf.Clothing
		newItemInfo.SubType = item.TypeN
		tEntityPlayer.BagList = append(tEntityPlayer.BagList, *newItemInfo)
	}

	num, charm := c.getCueNum(tEntityPlayer)
	tEntityPlayer.CharmNum = charm
	cond := make([]conf.ConditionData, 0)
	cond = append(cond, conf.ConditionData{conf.CueNum, num, true}, conf.ConditionData{conf.CharmRating, charm, true}, conf.ConditionData{conf.CharmNum, charm, true})
	TaskDBManger.updateConditional(tEntityPlayer.EntityID, cond)
}

func (c *_InitPlayer) addDress(tEntityPlayer *entity.EntityPlayer) *entity.Item {
	newItemInfo := new(entity.Item)
	newItemInfo.TableID = Table.DefaultDress[conf.USER_WOMEN]
	newItemInfo.ItemID = tEntityPlayer.GetMaxUuid()
	newItemInfo.ItemNum = uint32(1)
	newItemInfo.ItemType = uint32(conf.Dress)
	return newItemInfo
}

// / 获取n秒后的时间戳
func (c *_InitPlayer) getEndTimeStamp(endTime uint32) uint32 {
	if endTime > 0 {
		return endTime + uint32(carbon.Now().Timestamp())
	}
	return 0
}

// 获取阶数和星数以及个位数
func (c *_InitPlayer) getKeyByTableId(tableId uint32) (int, int, int) {
	return int(tableId) / conf.CueQualityDigit % 10, int(tableId) / conf.CueStarDigit % 10, int(tableId) / conf.CueKeyDigit % 100
}

func (c *_InitPlayer) getCueNum(tEntityPlayer *entity.EntityPlayer) (num, charmNum uint32) {
	for _, value := range tEntityPlayer.BagList {
		if value.ItemType == conf.Cue {
			charmNum += Table.GetCueCharmScore(value.TableID)
			num++
		}
	}
	return
}

// 任务初始化
func (c *_InitPlayer) PlayerTaskInit(EntityID uint32) {
	tEntity := Entity.EmEntityPlayer.GetEntityByID(EntityID)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	taskNum := len(tEntityPlayer.TaskList)
	if taskNum > 0 {
		return
	}

	TaskDBManger.setPlayerDailyTaskList(tEntityPlayer)
	TaskDBManger.setDayProgressRewardList(tEntityPlayer)
	TaskDBManger.setWeekProgressRewardList(tEntityPlayer)

	log.Info("-->PlayerTaskInit-->Complete:", EntityID)
}

func (c *_InitPlayer) getTodayBeginTime() int64 {
	return tools.GetTodayBeginTime()
}

// 判断是否周1
func (c *_InitPlayer) isMonDay() bool {
	return tools.GetWeekDay() == 1
}

// 获取当前时间的日期格式
func (c *_InitPlayer) getNowDateString() string {
	return tools.GetNowDateString()
}

// 获取本周周一的日期格式
func (c *_InitPlayer) getThisWeekFirstDateString() string {
	return tools.GetThisWeekFirstDateString()
}

// 获取本周周1的时间戳
func (c *_InitPlayer) getThisWeekFirstTime() int64 {
	return tools.GetThisWeekFirstDate()
}

// 初始化成就等级表
func (c *_InitPlayer) PlayerAchievementLVInit() {
	c.AchievementLVReward = make([]entity.AchievementLVReward, 0)
	for _, vl := range Table.GetAchievementLVCfg() {
		if vl.TableID == 0 {
			continue
		}
		reward := new(entity.AchievementLVReward)
		reward.AchievementLvID = vl.TableID
		reward.Score = uint32(vl.Score)
		reward.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		reward.StateReward = 0
		c.AchievementLVReward = append(c.AchievementLVReward, *reward)
	}

	sort.Slice(c.AchievementLVReward, func(i, j int) bool {
		return c.AchievementLVReward[i].AchievementLvID < c.AchievementLVReward[j].AchievementLvID
	})
}

// 初始化成就表
func (c *_InitPlayer) PlayerAchievementListInit() {
	c.AchievementList = make([]entity.Achievement, 0)
	for _, vl := range Table.GetAchievementCfg() {
		achievement := new(entity.Achievement)
		achievement.AchievementID = vl.TableID
		achievement.TypeN = vl.TypeN
		achievement.ChildList = make([]entity.ChildAchievement, 0)
		for _, v := range vl.Child {
			ele := Table.GetAchievementElementCfg(v)
			if ele == nil {
				continue
			}
			childAchievement := new(entity.ChildAchievement)
			childAchievement.ChildID = v

			elsCondition, _ := strconv.Atoi(ele.Condition[0])
			elsTaskProgress, _ := strconv.Atoi(ele.Condition[1])
			childAchievement.ConditionID = uint32(elsCondition)
			childAchievement.TaskProgress = uint32(elsTaskProgress)
			childAchievement.State = 0
			childAchievement.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			childAchievement.CompleteProgress = 0
			achievement.ChildList = append(achievement.ChildList, *childAchievement)
		}

		c.AchievementList = append(c.AchievementList, *achievement)
	}

	sort.Slice(c.AchievementList, func(i, j int) bool {
		return c.AchievementList[i].AchievementID < c.AchievementList[j].AchievementID
	})
}

// 初始化称号表
func (c *_InitPlayer) PlayerCollectListInit() {
	c.CollectList = make([]entity.Collect, 0)
	for _, vl := range Table.GetCollectCfg() {
		ct := new(entity.Collect)
		ct.CollectID = vl.TableID
		ct.State = 0
		ct.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		ct.Apply = 0
		ct.CompleteProgress = 0

		if len(vl.Conditions) > 1 {
			ct.ConditionID = vl.Conditions[0]
			ct.TaskProgress = vl.Conditions[1]
		}
		c.CollectList = append(c.CollectList, *ct)
	}

	sort.Slice(c.CollectList, func(i, j int) bool {
		return c.CollectList[i].CollectID < c.CollectList[j].CollectID
	})
}

// 初始化球杆图鉴
func (c *_InitPlayer) PlayerCueHandBookListInit() {
	c.CueHandBook = make([]entity.ElemBook, 0)
	for _, vl := range Table.GetCueHandbookCfg() {
		elem := new(entity.ElemBook)
		elem.Key = vl.TableID
		elem.State = 0
		elem.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
		elem.CueID = vl.CueID
		elem.CueQuality = vl.CueQuality
		c.CueHandBook = append(c.CueHandBook, *elem)
	}

	sort.Slice(c.CueHandBook, func(i, j int) bool {
		return c.CueHandBook[i].Key < c.CueHandBook[j].Key
	})
}

// 更新角色默认的道具
func (c *_InitPlayer) checkDefaultItem(entityId uint32) *entity.EntityPlayer {
	tEntity := Entity.EmEntityPlayer.GetEntityByID(entityId)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	tEntityPlayer.SetDBConnect(entity.UnitPlayer)
	resExpireList := c.getExpireItems(tEntityPlayer)
	if len(resExpireList) == 0 {
		log.Info("没有过期道具,直接返回,", entityId)
		return tEntityPlayer
	}

	if len(c.DefaultItem) == 0 {
		log.Error("DefaultItem is nil")
		return tEntityPlayer
	}

	for _, item := range resExpireList {
		newTableId := uint32(0)
		if v, ok := c.DefaultItem[fmt.Sprintf("%d_%d", item.ItemType, tEntityPlayer.Sex)]; ok && conf.Dress == item.ItemType && item.ItemStatus == conf.ItemUse {
			newTableId = v
		}
		if v, ok := c.DefaultItem[fmt.Sprintf("%d_%d", item.ItemType, item.SubType)]; ok && conf.Dress != item.ItemType && item.ItemStatus == conf.ItemUse {
			if conf.Cue == item.ItemType {
				newTableId = c.getPlayerDefaultCue(tEntityPlayer, v)
			} else {
				newTableId = v
			}
		}
		for key, itemData := range tEntityPlayer.BagList {
			if itemData.TableID != item.TableID && itemData.TableID != newTableId {
				continue
			}
			log.Info("item,", item, ",newTableId,", newTableId)
			if itemData.TableID == newTableId {
				structName, ok := conf.BaseTableInfo[fmt.Sprintf("%d_%d", item.ItemType, item.SubType)]
				if !ok {
					log.Error("-->BaseTableInfo is nil;itemID,", newTableId)
					continue
				}
				value := reflect.ValueOf(tEntityPlayer)
				fieldVal := value.Elem().FieldByName(structName)
				if !fieldVal.IsValid() {
					log.Error("-->checkDefaultItem-->reflect-->无效的,", newTableId)
					continue
				}
				fieldVal.SetUint(uint64(newTableId))
				items := tEntityPlayer.BagList[key]
				items.ItemStatus = conf.ItemUse
				tEntityPlayer.BagList[key] = items
				log.Info("-->checkDefaultItem->entityId:", tEntityPlayer.EntityID, "-->newTableId->", items)
			} else if itemData.TableID == item.TableID {
				tEntityPlayer.BagList = append(tEntityPlayer.BagList[:key], tEntityPlayer.BagList[(key+1):]...)
				log.Info("-->checkDefaultItem->entityId:", tEntityPlayer.EntityID, "-->oldTableId->", item)
			}
		}
	}

	tEntityPlayer.SaveEntity(DBConnect)
	return tEntityPlayer
}

// 获取使用中并且过期的道具
func (c *_InitPlayer) getExpireItems(tEntityPlayer *entity.EntityPlayer) (res []entity.Item) {
	nowTime := uint32(time.Now().Unix())
	for _, item := range tEntityPlayer.BagList {
		if item.EndTime > uint32(0) && item.EndTime < nowTime {
			res = append(res, item)
		}
	}
	return
}

// 获取角色默认的球杆
func (c *_InitPlayer) getPlayerDefaultCue(tEntityPlayer *entity.EntityPlayer, defaultKey uint32) uint32 {
	for _, val := range tEntityPlayer.BagList {
		if val.ItemType != conf.Cue {
			continue
		}
		cueId, errs := tools.GetCueIDByTableID(val.TableID)
		if errs != nil {
			continue
		}
		if cueId == defaultKey {
			return val.TableID
		}
	}
	return 0
}
