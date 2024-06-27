package logic

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Common/resp_code"
	"BilliardServer/Common/table"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/event"
	"BilliardServer/Util/log"
	"BilliardServer/Util/network"
	"BilliardServer/Util/tools"
	"errors"
	"fmt"
	"gitee.com/go-package/carbon/v2"
	"google.golang.org/protobuf/proto"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type _Backpack struct {
	lock         sync.Mutex
	ChangeSexRaw map[int]int
}

var Backpack _Backpack

type CueList struct {
	CueTableId []uint32 //球杆iD
	CharmNum   uint32   //球杆魅力评分
}

func (s *_Backpack) Init() {
	s.ChangeSexRaw = make(map[int]int, 0)
	s.initChangeSexRaw()

	event.OnNet(gmsg.MsgTile_Hall_CueUpgradeQualityStarRequest, reflect.ValueOf(s.OnCueUpgradeQualityStarRequest))
	event.OnNet(gmsg.MsgTile_Hall_UseItemRequest, reflect.ValueOf(s.OnUseItemRequest))
}

// 切换性别消耗配置
func (s *_Backpack) initChangeSexRaw() {
	config, ok := Table.GetConstMap()["5"]
	if !ok || ok && (len(config.Paramater10) <= 1 || config.Paramater10[1] == 0) {
		log.Error("initChangeSexRaw is err!")
		return
	}
	s.ChangeSexRaw[int(config.Paramater10[0])] = int(config.Paramater10[1])
}

// 查询性别切换消耗数量
func (s *_Backpack) GetPlayerChangeSexRaw(tEntityPlayer *entity.EntityPlayer) (uint32, int) {
	if len(s.ChangeSexRaw) == 0 {
		log.Error("ChangeSexRaw is nil!")
		return uint32(0), 0
	}
	if len(Table.DefaultDress) == 0 {
		log.Error("DefaultDress is nil!")
		return uint32(0), 0
	}
	for id, num := range s.ChangeSexRaw {
		res, index := s.GetItemByTableID(tEntityPlayer, uint32(id))
		if res != nil && res.ItemNum >= uint32(num) {
			return res.ItemNum, index
		}
	}

	return uint32(0), 0
}

// 更新扣减材料，更新道具
func (s *_Backpack) DeductPlayerChangeSexRes(tEntityPlayer *entity.EntityPlayer, index int, sex, delNum uint32) error {
	var res []*entity.Item
	resParam := GetResParam(conf.SYSTEM_ID_BAG, conf.ChangeSex)
	delitem := s.deductByItemIndex(tEntityPlayer, index, delNum, *resParam)
	log.Info("delitem", delitem)
	replaceItem, err := s.replaceDressBySex(tEntityPlayer, sex)
	if err != nil {
		return err
	}
	log.Info("replaceItem", replaceItem)
	res = append(res, delitem, replaceItem)
	s.BackpackUpdateItemSync(res, tEntityPlayer.EntityID)
	return nil
}

// 修改默认时装 并绑定使用
func (s *_Backpack) replaceDressBySex(tEntityPlayer *entity.EntityPlayer, sex uint32) (item *entity.Item, err error) {
	dressId, ok := Table.DefaultDress[sex]
	if !ok {
		return nil, errors.New("DefaultDress is err")
	}
	if dressId == tEntityPlayer.PlayerDress {
		return nil, errors.New("PlayerDress and dressId is the same")
	}
	item = s.updateBagListStatusByOld(tEntityPlayer, dressId, tEntityPlayer.PlayerDress)
	return item, nil
}

// 获取全部背包
func (s *_Backpack) GetAllItem(tEntityPlayer *entity.EntityPlayer) (item []*gmsg.ItemInfo) {
	for _, itemData := range tEntityPlayer.BagList {
		a := new(gmsg.ItemInfo)
		a.ItemID = proto.Uint32(itemData.ItemID)
		a.TableID = itemData.TableID
		a.ItemType = itemData.ItemType
		a.SubType = proto.Uint32(itemData.SubType)
		a.ItemNum = proto.Uint32(itemData.ItemNum)
		a.EndTime = proto.Uint32(itemData.EndTime)
		a.ItemStatus = proto.Uint32(itemData.ItemStatus)
		a.CueInfo = new(gmsg.CueInfo)
		a.CueInfo.Quality = *proto.Uint32(itemData.CueInfo.Quality)
		a.CueInfo.Star = *proto.Uint32(itemData.CueInfo.Star)
		item = append(item, a)
	}
	return
}

// 球杆升阶升星，前端->游戏服->DB服
func (s *_Backpack) OnCueUpgradeQualityStarRequest(msgEV *network.MsgBodyEvent) {
	s.lock.Lock()
	defer s.lock.Unlock()
	msgBody := &gmsg.CueUpgradeQualityStarRequest{}
	err := msgEV.Unmarshal(msgBody)
	if err != nil {
		return
	}

	log.Info("-->OnCueUpgradeQualityStarRequest-->msgBody,", msgBody)

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	response := &gmsg.CueUpgradeQualityStarResponse{}
	response.Code = *proto.Uint32(0)

	code, newTableID, itemList := s.checkCueUpgradeBagList(tEntityPlayer, msgBody.ItemID)
	if newTableID == 0 {
		response.Code = *proto.Uint32(code)
	} else {
		response.NextTableID = newTableID
		response.Code = *proto.Uint32(code)
		response.ItemID = msgBody.ItemID
	}

	tEntityPlayer.SyncEntity(1)
	log.Info("-->OnCueUpgradeQualityStarRequest-->response:", response)
	if len(itemList) > 0 {
		s.BackpackUpdateItemSync(itemList, msgBody.EntityID)
	}
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_CueUpgradeQualityStarResponse, response, []uint32{msgBody.EntityID})
}

// 获取球杆材料(必须所有材料都足够才能满足)，满足就+1 resUpgrade map[index]num
func (s *_Backpack) getCueUpgradeItem(tEntityPlayer *entity.EntityPlayer, upgradeMap map[int]int) (resUpgrade map[int]int) {
	resUpgrade = make(map[int]int, 0)
	for index, itemData := range tEntityPlayer.BagList {
		upgradeNum, ok := upgradeMap[int(itemData.TableID)]
		if ok && itemData.ItemNum >= uint32(upgradeNum) {
			resUpgrade[index] = upgradeNum
		}
	}
	return
}

// 扣减材料多个材料
func (s *_Backpack) deductItemMore(tEntityPlayer *entity.EntityPlayer, upgradeMap map[int]int, resParam entity.ResParam) (err error, itemList []*entity.Item) {
	for index, upgradeNum := range upgradeMap {
		itemData := tEntityPlayer.BagList[index]
		if itemData.ItemNum >= uint32(upgradeNum) {
			itemData.ItemNum -= uint32(upgradeNum)
			itemList = append(itemList, &itemData)
			if itemData.ItemNum == 0 {
				tEntityPlayer.BagList = append(tEntityPlayer.BagList[:index], tEntityPlayer.BagList[(index+1):]...)
			} else {
				tEntityPlayer.BagList[index] = itemData
			}
			//更新消耗
			SendConsumeResourceLogToDb(resParam.Uuid, tEntityPlayer.EntityID, itemData.ItemType, itemData.SubType, itemData.TableID, conf.RES_TYPE_DECR, uint64(upgradeNum), itemData.ItemNum, resParam.SysID, resParam.ActionID)
			log.Info("-->deductItem-->", tEntityPlayer.EntityID, "-->item-->", itemData, "-->update-->", itemData.TableID)
		} else {
			log.Error("-->deductItem--err-->道具数量不足。", tEntityPlayer.EntityID, "-->item-->", itemData.ItemID, "-->update-->", itemData.TableID)
			err = errors.New("-->deductItem--err-->道具数量不足。")
			break
		}
	}

	return err, itemList
}

// 扣减单个道具数量
func (s *_Backpack) deductByItemIndex(tEntityPlayer *entity.EntityPlayer, index int, num uint32, resParam entity.ResParam) *entity.Item {
	item := tEntityPlayer.BagList[index]
	item.ItemNum = item.ItemNum - num
	if item.ItemNum == 0 {
		tEntityPlayer.BagList = append(tEntityPlayer.BagList[:index], tEntityPlayer.BagList[(index+1):]...)
	} else {
		tEntityPlayer.BagList[index] = item
	}
	//更新消耗
	SendConsumeResourceLogToDb(resParam.Uuid, tEntityPlayer.EntityID, item.ItemType, item.SubType, item.TableID, conf.RES_TYPE_DECR, uint64(num), item.ItemNum, resParam.SysID, resParam.ActionID)
	log.Info("-->deductItem-->", tEntityPlayer.EntityID, "-->item-->", item, "-->update-->", item.TableID)
	tEntityPlayer.SyncEntity(0)
	return &item
}

// 检查背包并扣减消耗品
func (s *_Backpack) checkCueUpgradeBagList(tEntityPlayer *entity.EntityPlayer, itemID uint32) (errCode, nextCueID uint32, itemList []*entity.Item) {
	errCode = uint32(1)
	item, index := s.GetItemByItemID(tEntityPlayer, itemID)
	if item == nil {
		log.Error("GeCueCfg is nil,tableID:", itemID)
		return errCode, 0, nil
	}

	cueCfg := Table.GetCueCfg(fmt.Sprintf("%d", item.TableID))
	if cueCfg == nil || (cueCfg.NextID > uint32(0) && len(cueCfg.UpgradeConst) == 0) {
		log.Error("cfg is err,不可强化,", item.TableID)
		return errCode, 0, nil
	}

	if cueCfg.NextID == 0 {
		log.Info("满阶不可强化:", item.TableID)
		errCode = uint32(3)
		return errCode, 0, nil
	}

	//查询消耗材料数量
	resUpgrade := s.getCueUpgradeItem(tEntityPlayer, cueCfg.UpgradeConst)
	if len(cueCfg.UpgradeConst) > len(resUpgrade) {
		return uint32(2), 0, nil
	}
	resParam := GetResParam(conf.SYSTEM_ID_BAG, conf.CueUpgrade)
	// 更新球杆数据
	newTableID := s.updateQualityStarByIndex(index, item.TableID, tEntityPlayer, *resParam)
	if newTableID == 0 {
		errCode = uint32(3)
		return errCode, 0, nil
	}

	// 扣减材料,并返回材料信息
	deductErr, itemList := s.deductItemMore(tEntityPlayer, resUpgrade, *resParam)
	if deductErr != nil {
		return uint32(2), 0, nil
	}

	//同步图鉴
	s.UpdateConditionalAndCueHandBook(tEntityPlayer, newTableID)
	return uint32(0), newTableID, itemList
}

func (s *_Backpack) updateQualityStarByIndex(index int, tableID uint32, tEntityPlayer *entity.EntityPlayer, resParam entity.ResParam) uint32 {
	cueItem := tEntityPlayer.BagList[index]

	maxQuality := Table.GetCueCfg(fmt.Sprintf("%d", tableID)).MaxQuality

	switch {
	case maxQuality > cueItem.CueInfo.Quality && cueItem.CueInfo.Star+1 == maxQuality:
		cueItem.CueInfo.Quality = cueItem.CueInfo.Quality + 1
		cueItem.CueInfo.Star = 0
	case cueItem.CueInfo.Star+1 < maxQuality:
		cueItem.CueInfo.Star = cueItem.CueInfo.Star + 1
	default:
		return 0
	}

	// 查找下一阶id
	newCue := Table.GetCueCfg(fmt.Sprintf("%d", Table.GetCueCfg(fmt.Sprintf("%d", cueItem.TableID)).NextID))
	if newCue == nil || !s.compareQualityAndStar(int(cueItem.CueInfo.Quality), int(cueItem.CueInfo.Star), newCue.TableID) {
		log.Error(fmt.Sprintf("tabelid:%d,下一阶,%d;星,%d为空", cueItem.TableID, cueItem.CueInfo.Quality, cueItem.CueInfo.Star))
		return 0
	}
	cueItem.TableID = newCue.TableID
	cueItem.EndTime = s.getEndTimeStamp(newCue.LifeTime)
	cueItem.SubType = newCue.TypeN
	tEntityPlayer.BagList[index] = cueItem
	//本球杆升级升星更新tEntityPlayer
	if CheckSameHundredId(tEntityPlayer.CueTableId, newCue.TableID) {
		Player.ChangeCueTableID(tEntityPlayer.EntityID, newCue.TableID)
	}
	//更新消耗
	SendConsumeResourceLogToDb(resParam.Uuid, tEntityPlayer.EntityID, cueItem.ItemType, cueItem.SubType, tableID, conf.RES_TYPE_DECR, uint64(1), 0, resParam.SysID, resParam.ActionID)
	//更新产出
	SendProductionResourceLogToDb(resParam.Uuid, tEntityPlayer.EntityID, cueItem.ItemType, cueItem.SubType, newCue.TableID, conf.RES_TYPE_INCR, uint64(1), 1, resParam.SysID, resParam.ActionID)
	log.Info("-->updateQualityStarByIndex,update index:", index, "newTableID:", newCue.TableID)
	return newCue.TableID
}

// / 获取n秒后的时间戳
func (s *_Backpack) getEndTimeStamp(endTime uint32) uint32 {
	if endTime > 0 {
		return endTime + uint32(carbon.Now().Timestamp())
	}
	return 0
}

// 获取阶数和星数以及个位数
func (s *_Backpack) getCueQualityAndStarByTableId(tableId uint32) (int, int, int) {
	return int(tableId) / conf.CueQualityDigit % 10, int(tableId) / conf.CueStarDigit % 10, int(tableId) / conf.CueKeyDigit % 100
}

// 根据阶数和星数获取tableId
func (s *_Backpack) getCueUpByTableId(cueQualityDigit, cueStarDigit, tableIDKey int) *table.CueCfg {
	for _, cueCfg := range Table.GetCueCfgMap() {
		q, a, g := s.getCueQualityAndStarByTableId(cueCfg.TableID)
		if q == cueQualityDigit && a == cueStarDigit && g == tableIDKey {
			log.Info("-->getCueUpByTableId:", cueCfg.TableID)
			return cueCfg
		}
	}
	return nil
}

// 根据阶数和星数，与tableid校验
func (s *_Backpack) compareQualityAndStar(q, a int, tableID uint32) bool {
	n_q, n_a, _ := s.getCueQualityAndStarByTableId(tableID)
	return n_q == q && n_a == a
}

func (s *_Backpack) useItemByItemIdRequest(tEntityPlayer *entity.EntityPlayer, msgBody *gmsg.UseItemRequest) {
	msgUseItemResponse := &gmsg.UseItemResponse{}
	item, code, resItems := s.UpdatePlayInfoByItemType(tEntityPlayer, msgBody)
	if code > uint32(0) {
		msgUseItemResponse.Code = code
		msgUseItemResponse.ItemType = item.ItemType
		log.Info("-->OnUseItemRequest-->msgUseItemResponse->", msgUseItemResponse)
		ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UseItemResponse, msgUseItemResponse, []uint32{msgBody.EntityID})
		return
	}

	//角色数据同步
	if item.ItemType == conf.Cue {
		Player.ChangeCueTableID(msgBody.EntityID, item.TableID)
	} else if item.ItemType == conf.Dress {
		Player.ChangePlayerDress(msgBody.EntityID, item.TableID)
	} else if item.ItemType == conf.Clothing {
		Player.ChangeClothing(msgBody.EntityID, item.SubType, item.TableID)
	} else if item.ItemType == conf.Effect {
		Player.ChangeEffect(msgBody.EntityID, item.SubType, item.TableID)
	}

	tEntityPlayer.SyncEntity(1)
	msgUseItemResponse.Code = *proto.Uint32(0)
	msgUseItemResponse.ItemType = item.ItemType

	resItems = append(resItems, item)
	s.BackpackUpdateItemSync(resItems, msgBody.EntityID)

	log.Info("-->OnUseItemRequest-->msgUseItemResponse->", msgUseItemResponse, "-->resItems-->", resItems)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_UseItemResponse, msgUseItemResponse, []uint32{msgBody.EntityID})
}

// 使用道具
func (s *_Backpack) OnUseItemRequest(msgEV *network.MsgBodyEvent) {
	s.lock.Lock()
	defer s.lock.Unlock()
	msgBody := &gmsg.UseItemRequest{}
	if err := msgEV.Unmarshal(msgBody); err != nil {
		return
	}

	log.Info("-->OnUseItemRequest--------------begin-------", msgBody)

	tEntity := Entity.EmPlayer.GetEntityByID(msgBody.EntityID)
	if tEntity == nil {
		return
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	s.useItemByItemIdRequest(tEntityPlayer, msgBody)
}

func (s *_Backpack) UpdatePlayInfoByItemType(tEntityPlayer *entity.EntityPlayer, useItemReq *gmsg.UseItemRequest) (*entity.Item, uint32, []*entity.Item) {
	code := uint32(1)
	// 道具超过有效期直接返回
	item, index := s.GetItemByItemID(tEntityPlayer, useItemReq.ItemID)
	if item == nil || (uint32(carbon.Now().Timestamp()) > item.EndTime && item.EndTime > 0) {
		log.Info("-->道具为空或者已过期,item:", item)
		return nil, code, nil
	}

	// 道具相关
	if item.ItemType == conf.Item {
		var (
			resItems []*entity.Item
			delNum   uint32
			resParam *entity.ResParam
		)
		if item.SubType == conf.Item_5 {
			log.Info("----->OnUseItemRequest-->礼包-->", item.TableID)
			resParam = GetResParam(conf.SYSTEM_ID_BAG, conf.BagGift)
			err, items := s.BackpackItemReward(tEntityPlayer.EntityID, item.TableID, *resParam)
			if err != nil {
				log.Error("BackpackItemReward err", err)
				return item, code, nil
			}
			item.ItemNum -= uint32(1)
			resItems, delNum = items, uint32(1)
		} else if item.SubType == conf.Item_2 {
			log.Info("----->OnUseItemRequest-->碎片-->", item.TableID)
			resParam = GetResParam(conf.SYSTEM_ID_BAG, conf.ComposeItem)
			err, dushredNum, itemInfo := s.BackpackItemShred(tEntityPlayer.EntityID, item.TableID, item.ItemNum, *resParam)
			if err != nil {
				log.Error("BackpackItemDebris err", err, item.TableID)
				return item, uint32(2), nil
			}
			item.ItemNum -= dushredNum
			resItems, delNum = []*entity.Item{itemInfo}, dushredNum
		} else if item.SubType == conf.Item_7 {
			log.Info("----->OnUseItemRequest-->赠送礼物-->", useItemReq)
			if useItemReq.ToEntityID == 0 || useItemReq.Number <= 0 || useItemReq.EntityID == useItemReq.ToEntityID {
				return item, code, nil
			}
			if useItemReq.Number > item.ItemNum {
				return item, uint32(2), nil
			}
			//todo 暂时扣减材料，晚点同步到db
			delNum = useItemReq.Number
			resParam = GetResParam(conf.SYSTEM_ID_BAG, conf.GiveGifts)
			s.deductByItemIndex(tEntityPlayer, index, delNum, *resParam)
			err := GiftsMr.GiveGiftRequest(useItemReq, item.TableID, *resParam)
			if err != nil {
				log.Error("GiveGiftRequest err", err, useItemReq)
				return item, code, nil
			}
			item.ItemNum -= useItemReq.Number
			return item, 0, resItems
		}

		//扣减材料，同步到db
		s.deductByItemIndex(tEntityPlayer, index, delNum, *resParam)
		tEntityPlayer.SyncEntity(1)
		return item, 0, resItems
	} else if item.ItemType == conf.Dress {
		if Table.GetDressCfg(tools.FormatUint32(item.TableID)).Sex != tEntityPlayer.Sex {
			return item, code, nil
		}
	}

	structName, ok := conf.BaseTableInfo[fmt.Sprintf("%d_%d", item.ItemType, item.SubType)]
	if !ok {
		log.Error("-->BaseTableInfo is nil;itemID,", useItemReq.ItemID)
		return item, 0, nil
	}

	val := reflect.ValueOf(tEntityPlayer)
	fieldVal := val.Elem().FieldByName(structName)
	if !fieldVal.IsValid() {
		log.Error("-->updatePlayInfoByItemType-->reflect-->无效的,", tEntityPlayer.EntityID)
		return item, code, nil
	}

	oldTableId := uint32(fieldVal.Uint())

	s.updateBagListStatusByOld(tEntityPlayer, item.TableID, oldTableId)
	item.ItemStatus = conf.ItemUse

	return item, 0, nil
}

// 背包数据同步
func (s *_Backpack) BackpackUpdateItemSync(resItems []*entity.Item, entityID uint32) {
	if len(resItems) == 0 {
		return
	}
	msgResponse := &gmsg.BackpackUpdateItemSync{}
	msgResponse.EntityID = entityID
	msgResponse.Items = make([]*gmsg.ItemInfo, 0)
	for _, val := range resItems {
		newItem := new(gmsg.ItemInfo)
		newItem.ItemID = proto.Uint32(val.ItemID)
		newItem.TableID = val.TableID
		newItem.ItemType = val.ItemType
		newItem.SubType = proto.Uint32(val.SubType)
		newItem.ItemNum = proto.Uint32(val.ItemNum)
		newItem.EndTime = proto.Uint32(val.EndTime)
		newItem.ItemStatus = proto.Uint32(val.ItemStatus)
		newItem.CueInfo = new(gmsg.CueInfo)
		newItem.CueInfo.Quality = *proto.Uint32(val.CueInfo.Quality)
		newItem.CueInfo.Star = *proto.Uint32(val.CueInfo.Star)
		msgResponse.Items = append(msgResponse.Items, newItem)
	}

	log.Info("--->BackpackUpdateItemSync-->data:", msgResponse)
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Hall_BackpackUpdateItemSync, msgResponse, []uint32{msgResponse.EntityID})
}

// 查询道具
func (s *_Backpack) GetItemByItemID(tEntityPlayer *entity.EntityPlayer, itemID uint32) (item *entity.Item, index int) {
	for key, itemData := range tEntityPlayer.BagList {
		if itemData.ItemID == itemID {
			items := tEntityPlayer.BagList[key]
			item = &items
			index = key
			break
		}
	}

	return item, index
}

// 查询道具
func (s *_Backpack) GetItemByTableID(tEntityPlayer *entity.EntityPlayer, tableID uint32) (item *entity.Item, index int) {
	for key, itemData := range tEntityPlayer.BagList {
		if itemData.TableID == tableID {
			items := tEntityPlayer.BagList[key]
			item = &items
			index = key
			break
		}
	}

	return item, index
}

// 重置道具使用
func (s *_Backpack) updateBagListStatusByOld(tEntityPlayer *entity.EntityPlayer, nID, oID uint32) (resNewItem *entity.Item) {
	for index, itemData := range tEntityPlayer.BagList {
		if itemData.TableID == nID {
			items := tEntityPlayer.BagList[index]
			items.ItemStatus = conf.ItemUse
			tEntityPlayer.BagList[index] = items
			resNewItem = &items
		} else if itemData.TableID == oID {
			items := tEntityPlayer.BagList[index]
			items.ItemStatus = conf.ItemNoUse
			tEntityPlayer.BagList[index] = items
		}
	}

	tEntityPlayer.SyncEntity(0)
	return
}

// 增加code返回；todo 0添加物品成功，1添加失败，2已有物品不能添加
func (s *_Backpack) BackpackAddOneItemAndSave(entityId uint32, rewardEntity entity.RewardEntity, resParam entity.ResParam) (error, *entity.Item, uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityId)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	err, newItemInfo, code := s.BackpackAddOneItem(entityId, rewardEntity, resParam)
	if err != nil {
		log.Error("-->entityId--->", entityId, rewardEntity, "--err ", err.Error())
		return err, nil, code
	}
	tEntityPlayer.SyncEntity(1)
	if newItemInfo != nil {
		s.BackpackUpdateItemSync([]*entity.Item{newItemInfo}, entityId)
		//更新产出
		SendProductionResourceLogToDb(resParam.Uuid, entityId, newItemInfo.ItemType, newItemInfo.SubType, newItemInfo.TableID, conf.RES_TYPE_INCR, uint64(rewardEntity.Num), newItemInfo.ItemNum, resParam.SysID, resParam.ActionID)
	}

	return nil, newItemInfo, code
}

func (s *_Backpack) BackpackAddOneItem(entityId uint32, rewardEntity entity.RewardEntity, resParam entity.ResParam) (error, *entity.Item, uint32) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityId)
	if tEntity == nil {
		return errors.New("角色为空。"), nil, resp_code.CODE_ERR
	}
	tEntityPlayer := tEntity.(*entity.EntityPlayer)

	//属性道具
	if s.isPropertyItem(rewardEntity.ItemTableId) {
		log.Info("--->BackpackAddOneItem-->", "-->添加属性道具-->", "-->entityId-->", entityId, "-->", rewardEntity)
		Player.UpdatePlayerPropertyItem(entityId, rewardEntity.ItemTableId, int32(rewardEntity.Num), resParam)
		return nil, nil, resp_code.CODE_SUCCESS
	}
	//判断tableid类型
	itemType, err := tools.GetItemTypeByTableId(rewardEntity.ItemTableId)
	if err != nil {
		return err, nil, resp_code.CODE_ERR
	}
	//球杆替换或者新增判断
	if itemType == conf.Cue {
		checkCode, oldId, resErr := s.checkCueList(tEntityPlayer, rewardEntity.ItemTableId)
		if checkCode > uint32(0) {
			if checkCode == uint32(1) {
				log.Error("--->entityId", entityId, "--oldId->", oldId, "->newItemTableId->", rewardEntity.ItemTableId, "购买同一根球杆，替换处理；")
				newitem := s.ReplaceCueIDByTableID(tEntityPlayer, oldId, rewardEntity.ItemTableId)
				return nil, newitem, resp_code.CODE_SUCCESS
			} else {
				log.Error("--->entityId-->", entityId, "--oldId->", oldId, "->newItemTableId->", rewardEntity.ItemTableId, "--err：", resErr.Error())
				return errors.New(fmt.Sprintf("%s", resErr.Error())), nil, uint32(2)
			}
		}
	}

	if tEntityPlayer.BagMax <= uint32(len(tEntityPlayer.BagList)) {
		return errors.New("背包到达最大值。"), nil, resp_code.CODE_ERR
	}

	log.Info("-->BackpackAddOneItem-->begin-->", entityId, "-->items-->", rewardEntity)

	item, index := tEntityPlayer.GetItemFromTableID(rewardEntity.ItemTableId)
	if item != nil && (item.ItemType == conf.Dress || item.ItemType == conf.Clothing || item.ItemType == conf.Effect) {
		if item.EndTime == 0 {
			return errors.New("服装和装扮不能重复购买。"), nil, resp_code.CODE_ERR
		} else {
			if item.EndTime > uint32(time.Now().Unix()) {
				item.EndTime += Table.GetItemTimeCfgById(rewardEntity.ExpireTimeId).Time
			} else {
				item.EndTime = s.getEndTimeStamp(Table.GetItemTimeCfgById(rewardEntity.ExpireTimeId).Time)
			}
			tEntityPlayer.BagList[index] = *item
			return nil, item, resp_code.CODE_SUCCESS
		}
	}

	// 道具叠加
	if item != nil && item.ItemType == conf.Item {
		newItem := tEntityPlayer.BagList[index]
		newItem.ItemNum += rewardEntity.Num
		tEntityPlayer.BagList[index] = newItem
		log.Info("-->BackpackAddOneItem-->道具叠加->item:", tEntityPlayer.BagList[index])
		return nil, &newItem, resp_code.CODE_SUCCESS
	}

	err, newItemInfo := s.addItemFromRewardEntity(entityId, rewardEntity)
	if err != nil {
		return err, nil, resp_code.CODE_ERR
	}

	log.Info("-->BackpackAddOneItem-->newItemInfo:", newItemInfo)
	tEntityPlayer.BagList = append(tEntityPlayer.BagList, *newItemInfo)
	if newItemInfo.ItemType == conf.Cue {
		s.UpdateConditionalAndCueHandBook(tEntityPlayer, newItemInfo.TableID)
	}
	return nil, newItemInfo, resp_code.CODE_SUCCESS
}

// 添加多个道具，固定奖励可以直接调这个
func (s *_Backpack) BackpackAddItemListAndSave(entityId uint32, rewardEntityList []entity.RewardEntity, resParam entity.ResParam) (error, []*entity.Item) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityId)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	itemList := make([]*entity.Item, 0)
	for _, rewardEntity := range rewardEntityList {
		err, item, _ := s.BackpackAddOneItem(entityId, rewardEntity, resParam)
		if err != nil {
			return err, nil
		}
		if item != nil {
			itemList = append(itemList, item)
			SendProductionResourceLogToDb(resParam.Uuid, entityId, item.ItemType, item.SubType, item.TableID, conf.RES_TYPE_INCR, uint64(rewardEntity.Num), item.ItemNum, resParam.SysID, resParam.ActionID)
		}
	}
	if len(itemList) > 0 {
		s.BackpackUpdateItemSync(itemList, entityId)
	}

	tEntityPlayer.SyncEntity(1)
	if len(rewardEntityList) > 0 {
		RewardManager.CommonSendRewardAndSendMsg(entityId, rewardEntityList)
	}

	return nil, itemList
}

// 发放随机奖励
func (s *_Backpack) BackpackAddOneItemSaveFromRandReward(entityID uint32, randomReward [][]uint32, resParam entity.ResParam) (RandomRewardNum uint32) {
	list := RewardManager.AddRewardByRandomOne(entityID, randomReward, resParam)
	if len(list) > 0 {
		for _, v := range list {
			RandomRewardNum += v.Num
		}
	}

	log.Info("-->BackpackAddOneItemSaveFromRandReward-->fromName-->rewardEntity-->", list, "-->entityID-->", entityID)
	return RandomRewardNum
}

// 发送奖励，不调用通用奖励
func (s *_Backpack) BackpackAddItemListAndUpdateItemSync(entityId uint32, rewardEntityList []entity.RewardEntity, resParam entity.ResParam) (error, []*entity.Item) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityId)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	itemList := make([]*entity.Item, 0)
	for _, rewardEntity := range rewardEntityList {
		err, item, _ := s.BackpackAddOneItem(entityId, rewardEntity, resParam)
		if err != nil {
			return err, nil
		}
		if item != nil {
			itemList = append(itemList, item)
			SendProductionResourceLogToDb(resParam.Uuid, entityId, item.ItemType, item.SubType, item.TableID, conf.RES_TYPE_INCR, uint64(rewardEntity.Num), item.ItemNum, resParam.SysID, resParam.ActionID)
		}
	}
	if len(itemList) > 0 {
		s.BackpackUpdateItemSync(itemList, entityId)
	}

	tEntityPlayer.SyncEntity(1)

	return nil, itemList
}

func (s *_Backpack) isPropertyItem(tableID uint32) bool {
	return 60000000 < int(tableID) && int(tableID) < 70000000
}

// 礼包
func (s *_Backpack) BackpackItemReward(entityID, itemTableID uint32, resParam entity.ResParam) (error, []*entity.Item) {
	itemCfg := Table.GetItemCfgById(strconv.Itoa(int(itemTableID)))
	if itemCfg == nil {
		return errors.New("item tableid is err"), nil
	}
	rewardItem, rewardEntityList := make([]*entity.Item, 0), make([]entity.RewardEntity, 0)
	if len(itemCfg.FixedReward) > 0 {
		//获取固定奖励列表
		fixeReward, fixeItems := RewardManager.AddRewardByRegularListForItem(entityID, itemCfg.FixedReward, resParam)
		rewardItem = fixeItems
		rewardEntityList = fixeReward
	}
	if len(itemCfg.RandomReward) > 0 {
		//获取随机奖励列表
		randReward, randItem := RewardManager.AddRewardByRandomOneForItem(entityID, itemCfg.RandomReward, resParam)
		rewardItem = s.removeRepeatedRewardItem(rewardItem, randItem)
		rewardEntityList = s.removeRepeatedReward(rewardEntityList, randReward)
	}

	if len(rewardEntityList) > 0 {
		RewardManager.CommonSendRewardAndSendMsg(entityID, rewardEntityList)
	}
	log.Info("-->BackpackItemReward-->end-->", rewardEntityList)
	return nil, rewardItem
}

// 合成碎片
func (s *_Backpack) BackpackItemShred(entityID, itemTableID, itemNum uint32, resParam entity.ResParam) (error, uint32, *entity.Item) {
	itemCfg := Table.GetItemCfgById(strconv.Itoa(int(itemTableID)))
	if itemCfg == nil || itemCfg.CueID == 0 {
		return errors.New("item tableid ShredNum is nil"), 0, nil
	}
	cueCfg := Table.GetCueCfg(fmt.Sprintf("%d", itemCfg.CueID))
	if cueCfg == nil || cueCfg.ShredNum > itemNum {
		log.Error("材料不足，不能合成", itemTableID)
		return errors.New("材料不足"), 0, nil
	}
	rewardEntity := new(entity.RewardEntity)
	rewardEntity.ItemTableId = itemCfg.CueID
	rewardEntity.Num = 1
	rewardEntity.ExpireTimeId = 0
	err, newItemInfo, _ := s.BackpackAddOneItem(entityID, *rewardEntity, resParam)
	if err != nil {
		log.Error("-->entityId--->", entityID, rewardEntity, "err ", err.Error())
		return err, 0, nil
	}
	rewardEntityList := make([]entity.RewardEntity, 0)
	rewardEntityList = append(rewardEntityList, *rewardEntity)
	if len(rewardEntityList) > 0 {
		RewardManager.CommonSendRewardAndSendMsg(entityID, rewardEntityList)
	}
	//更新产出
	SendProductionResourceLogToDb(resParam.Uuid, entityID, conf.Cue, 0, newItemInfo.TableID, conf.RES_TYPE_INCR, 1, newItemInfo.ItemNum, resParam.SysID, resParam.ActionID)
	return nil, cueCfg.ShredNum, newItemInfo
}

func (s *_Backpack) addItemFromRewardEntity(entityId uint32, rewardEntity entity.RewardEntity) (error, *entity.Item) {
	tEntity := Entity.EmPlayer.GetEntityByID(entityId)
	tEntityPlayer := tEntity.(*entity.EntityPlayer)
	cfg := Table.IsExistTable(strconv.Itoa(int(rewardEntity.ItemTableId)))
	if cfg.ItemType == 0 {
		return errors.New(fmt.Sprintf("配置异常>>>>>>>>非法的TableID:%d", rewardEntity.ItemTableId)), nil
	}

	//将ExpireTimeId转为对应时间
	if rewardEntity.ExpireTimeId == 0 {
		rewardEntity.ExpireTimeId = conf.TABLE_ITEM_TIME_PERMANENTLY
	}
	expTime := Table.GetItemTimeCfgById(rewardEntity.ExpireTimeId).Time

	newItemInfo := new(entity.Item)
	newItemInfo.ItemID = tEntityPlayer.GetMaxUuid()
	newItemInfo.TableID = rewardEntity.ItemTableId
	newItemInfo.ItemNum = rewardEntity.Num
	newItemInfo.ItemType = cfg.ItemType
	newItemInfo.EndTime = s.getEndTimeStamp(expTime)
	newItemInfo.SubType = cfg.SubType
	newItemInfo.ItemStatus = 0
	if cfg.ItemType == conf.Cue {
		q, a, _ := s.getCueQualityAndStarByTableId(rewardEntity.ItemTableId)
		newItemInfo.CueInfo.Quality = uint32(q)
		newItemInfo.CueInfo.Star = uint32(a)
	}

	return nil, newItemInfo

}

func (s *_Backpack) getCueNum(tEntityPlayer *entity.EntityPlayer) (res *CueList) {
	cueList := new(CueList)
	for _, value := range tEntityPlayer.BagList {
		if value.ItemType == conf.Cue {
			cueList.CharmNum += Table.GetCueCharmScore(value.TableID)
			cueList.CueTableId = append(cueList.CueTableId, value.TableID)
		}
	}
	return cueList
}

func (s *_Backpack) checkCueList(tEntityPlayer *entity.EntityPlayer, tableID uint32) (code, oldId uint32, resErr error) {
	key, err := tools.GetCueIDByTableID(tableID)
	if err != nil {
		code = uint32(3)
		resErr = errors.New("球杆数据异常。")
		return
	}
	cueList := s.getCueNum(tEntityPlayer)
	for _, val := range cueList.CueTableId {
		cueId, errs := tools.GetCueIDByTableID(val)
		if errs != nil {
			continue
		}
		if key > 0 && key == cueId {
			if tableID > val {
				code = uint32(1)
				oldId = val
			} else {
				code = uint32(2)
				resErr = errors.New("不能买低阶或同阶球杆")
			}
		}
	}
	return
}

// 去重返回的道具奖励
func (s *_Backpack) removeRepeatedRewardItem(fixeRewardItem, randRewardItem []*entity.Item) []*entity.Item {
	randMap := make(map[uint32]*entity.Item, 0)
	for _, v := range randRewardItem {
		randMap[v.ItemID] = v
	}

	reward := make([]*entity.Item, 0)
	for _, fixval := range fixeRewardItem {
		value, ok := randMap[fixval.ItemID]
		if ok {
			val := fixval
			val.ItemNum = value.ItemNum
			reward = append(reward, val)
		} else {
			reward = append(reward, fixval)
		}
	}
	if len(fixeRewardItem) == 0 {
		for _, v := range randRewardItem {
			reward = append(reward, v)
		}
	}

	return reward
}

// 去重返回的奖励
func (s *_Backpack) removeRepeatedReward(fixeReward, randReward []entity.RewardEntity) []entity.RewardEntity {
	randMap := make(map[uint32]entity.RewardEntity, 0)
	for _, v := range randReward {
		randMap[v.ItemTableId] = v
	}

	reward := make([]entity.RewardEntity, 0)
	for _, fixval := range fixeReward {
		value, ok := randMap[fixval.ItemTableId]
		if ok {
			val := fixval
			val.Num = val.Num + value.Num
			reward = append(reward, val)
		} else {
			reward = append(reward, fixval)
		}
	}

	if len(fixeReward) == 0 {
		for _, v := range randReward {
			reward = append(reward, v)
		}
	}

	return reward
}

// 检查背包是否存在当前永久道具
func (s *_Backpack) CheckIsHaveItemByTableID(tEntityPlayer *entity.EntityPlayer, tableID uint32) bool {
	var isHave bool
	for _, v := range tEntityPlayer.BagList {
		if v.TableID == tableID && v.EndTime == 0 {
			isHave = true
			break
		}
	}

	return isHave
}

// 检查背包是否存在相同球杆
func (s *_Backpack) CheckIsHaveCueByTableID(tEntityPlayer *entity.EntityPlayer, tableID uint32) bool {
	var isHave bool
	for _, v := range tEntityPlayer.BagList {
		if CheckSameHundredId(v.TableID, tableID) && v.EndTime == 0 {
			isHave = true
			break
		}
	}

	return isHave
}

// 球杆ID替换，todo 记得同步客户端
func (s *_Backpack) ReplaceCueIDByTableID(tEntityPlayer *entity.EntityPlayer, oldId, tableID uint32) (item *entity.Item) {
	q, a, _ := s.getCueQualityAndStarByTableId(tableID)
	itemData, index := s.GetItemByTableID(tEntityPlayer, oldId)
	itemData.TableID = tableID
	itemData.CueInfo.Quality = uint32(q)
	itemData.CueInfo.Star = uint32(a)
	tEntityPlayer.BagList[index] = *itemData
	item = itemData

	//同步球杆
	if CheckSameHundredId(tEntityPlayer.CueTableId, tableID) {
		Player.ChangeCueTableID(tEntityPlayer.EntityID, tableID)
	}
	return
}

// 更新成就和图鉴
func (s *_Backpack) UpdateConditionalAndCueHandBook(tEntityPlayer *entity.EntityPlayer, tableID uint32) {
	cueList := s.getCueNum(tEntityPlayer)
	tEntityPlayer.CharmNum = cueList.CharmNum
	cond := make([]conf.ConditionData, 0)
	cond = append(cond,
		conf.ConditionData{conf.CueNum, uint32(len(cueList.CueTableId)), true},
		conf.ConditionData{conf.CharmRating, cueList.CharmNum, true},
		conf.ConditionData{conf.CharmNum, cueList.CharmNum, true},
		conf.ConditionData{conf.XYCue, 1, false})
	CueHandBookMr.UpdateCueHandBook(tEntityPlayer.EntityID, tableID)
	ConditionalMr.SyncConditional(tEntityPlayer.EntityID, cond)
}
