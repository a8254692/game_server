package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	gmsg "BilliardServer/Proto/gmsg"
	"BilliardServer/Util/log"
	"math/rand"
	"sort"
	"time"
)

// 随机奖励结构
type RandRewardInfo struct {
	itemTableId int
	weight      uint32
	minNum      uint32 //最小数量
	maxNum      uint32 //最大数量
	expTimeId   uint32 //过期时间id
	num         uint32 //最终发放数量
}

type _RewardManager struct {
}

var RewardManager _RewardManager

// 通用同步奖励列表消息
func (s *_RewardManager) CommonSendRewardAndSendMsg(entityID uint32, rewardList []entity.RewardEntity) {
	if entityID <= 0 || len(rewardList) <= 0 {
		return
	}

	msgList := make([]*gmsg.RewardInfo, 0)
	for _, v := range rewardList {
		info := gmsg.RewardInfo{
			ItemTableId:  v.ItemTableId,
			Num:          v.Num,
			ExpireTimeId: v.ExpireTimeId,
		}
		msgList = append(msgList, &info)
	}

	resp := &gmsg.CommonSendRewardSync{
		RewardList: msgList,
	}
	//广播同步消息
	ConnectManager.SendMsgPbToGateBroadCast(gmsg.MsgTile_Reward_CommonSendRewardSync, resp, []uint32{entityID})
	return
}

// 将配置中二维数组转换为列表并直接发奖
func (s *_RewardManager) AddRewardByRegularList(entityID uint32, itemReward [][]uint32, resParam entity.ResParam) []entity.RewardEntity {
	resp := make([]entity.RewardEntity, 0)
	if len(itemReward) <= 0 || entityID <= 0 {
		return resp
	}

	rewardList := s.GetRewardByRegularList(itemReward)
	if len(rewardList) > 0 {
		resp = s.AddReward(entityID, rewardList, resParam)
	}

	return resp
}

// 将配置中二维数值转换随机一个并发奖
func (s *_RewardManager) AddRewardByRandomOne(entityID uint32, itemReward [][]uint32, resParam entity.ResParam) []entity.RewardEntity {
	resp := make([]entity.RewardEntity, 0)
	if len(itemReward) <= 0 {
		return resp
	}
	rewardList := s.GetRewardByRandomOne(itemReward)
	if len(rewardList) > 0 {
		resp = s.AddReward(entityID, rewardList, resParam)
	}
	return resp
}

// 将配置中二维数组转换为列表并直接发奖
func (s *_RewardManager) AddRewardByRegularListForItem(entityID uint32, itemReward [][]uint32, resParam entity.ResParam) ([]entity.RewardEntity, []*entity.Item) {
	resp := make([]*entity.Item, 0)
	if len(itemReward) <= 0 || entityID <= 0 {
		return nil, resp
	}

	rewardList := s.GetRewardByRegularList(itemReward)
	if len(rewardList) > 0 {
		resp = s.AddRewardForItem(entityID, rewardList, resParam)
	}

	return rewardList, resp
}

// 将配置中二维数值转换随机一个并发奖
func (s *_RewardManager) AddRewardByRandomOneForItem(entityID uint32, itemReward [][]uint32, resParam entity.ResParam) ([]entity.RewardEntity, []*entity.Item) {
	resp := make([]*entity.Item, 0)
	if len(itemReward) <= 0 {
		return nil, resp
	}
	rewardList := s.GetRewardByRandomOne(itemReward)
	if len(rewardList) > 0 {
		resp = s.AddRewardForItem(entityID, rewardList, resParam)
	}
	return rewardList, resp
}

// 将配置中二维数组转换为列表
func (s *_RewardManager) GetRewardByRegularList(itemReward [][]uint32) []entity.RewardEntity {
	resp := make([]entity.RewardEntity, 0)
	if len(itemReward) <= 0 {
		return resp
	}

	for _, v := range itemReward {
		arrLen := len(v)
		var itemTableId, minNum, expTimeId uint32
		if arrLen > 0 {
			itemTableId = v[0]
		}
		if arrLen > 1 {
			minNum = v[1]
		}
		if arrLen > 2 {
			expTimeId = v[2]
		}

		if itemTableId <= 0 || minNum <= 0 {
			continue
		}

		info := entity.RewardEntity{
			ItemTableId:  itemTableId,
			Num:          minNum,
			ExpireTimeId: expTimeId,
		}
		resp = append(resp, info)
	}

	return resp
}

// 获取配置中二维数值转换随机一个
func (s *_RewardManager) GetRewardByRandomOne(itemReward [][]uint32) []entity.RewardEntity {
	resp := make([]entity.RewardEntity, 0)
	if len(itemReward) <= 0 {
		return resp
	}

	itemRandRewardList := make([]RandRewardInfo, 0)
	for _, v := range itemReward {
		arrLen := len(v)
		var itemTableId, weight, minNum, maxNum, expTimeId uint32
		if arrLen > 0 {
			itemTableId = v[0]
		}
		if arrLen > 1 {
			weight = v[1]
		}
		if arrLen > 2 {
			minNum = v[2]
		}
		if arrLen > 3 {
			maxNum = v[3]
		}
		if arrLen > 4 {
			expTimeId = v[4]
		}

		if itemTableId <= 0 || minNum <= 0 {
			continue
		}

		info := RandRewardInfo{
			itemTableId: int(itemTableId),
			weight:      weight,
			minNum:      minNum,
			maxNum:      maxNum,
			expTimeId:   expTimeId,
			num:         0,
		}
		itemRandRewardList = append(itemRandRewardList, info)
	}

	if len(itemRandRewardList) <= 0 {
		return resp
	}

	rs := randomDraw(itemRandRewardList)

	rewardList := make([]entity.RewardEntity, 0)

	if rs.itemTableId > 0 {
		rewardInfo := entity.RewardEntity{
			ItemTableId:  uint32(rs.itemTableId),
			Num:          rs.minNum,
			ExpireTimeId: rs.expTimeId,
		}

		if rs.maxNum > 0 && rs.maxNum > rs.minNum {
			//TODO:需要随机第二次
			rand := rand.New(rand.NewSource(time.Now().UnixNano()))
			randomNum := uint32(rand.Intn(int(rs.maxNum-rs.minNum)) + int(rs.minNum))
			rewardInfo.Num = uint32(randomNum)
		}

		rewardList = append(rewardList, rewardInfo)
	}

	return rewardList
}

func (s *_RewardManager) AddReward(entityID uint32, rewardList []entity.RewardEntity, resParam entity.ResParam) []entity.RewardEntity {
	if len(rewardList) <= 0 || entityID <= 0 {
		return nil
	}

	itemList := make([]entity.RewardEntity, 0)
	for _, reward := range rewardList {
		if reward.ExpireTimeId == 0 {
			reward.ExpireTimeId = conf.TABLE_ITEM_TIME_PERMANENTLY
		}

		err, _, code := Backpack.BackpackAddOneItemAndSave(entityID, reward, resParam)
		if err != nil {
			log.Waring("-->logic--_RewardManager--AddReward--err:", err.Error(), "--", entityID, "--", reward.ItemTableId)
			continue
		}

		if code == 0 {
			rewardInfo := entity.RewardEntity{
				ItemTableId:  reward.ItemTableId,
				Num:          reward.Num,
				ExpireTimeId: reward.ExpireTimeId,
			}
			itemList = append(itemList, rewardInfo)
		}
	}

	return itemList
}

func (s *_RewardManager) AddRewardForItem(entityID uint32, rewardList []entity.RewardEntity, resParam entity.ResParam) []*entity.Item {
	if len(rewardList) <= 0 || entityID <= 0 {
		return nil
	}

	itemList := make([]*entity.Item, 0)
	for _, reward := range rewardList {
		if reward.ExpireTimeId == 0 {
			reward.ExpireTimeId = conf.TABLE_ITEM_TIME_PERMANENTLY
		}

		err, newItems, _ := Backpack.BackpackAddOneItemAndSave(entityID, reward, resParam)
		if err != nil {
			log.Waring("-->logic--_RewardManager--AddReward--err:", err.Error(), "--", entityID, "--", reward.ItemTableId)
			continue
		}

		itemList = append(itemList, newItems)
	}

	return itemList
}

// 奖励数组去重
func (s *_RewardManager) RewardListDelDupl(rewardList []entity.RewardEntity) []entity.RewardEntity {
	list := make([]entity.RewardEntity, 0, len(rewardList))
	for i := 0; i < len(rewardList); i++ {
		sameIndex := -1
		for j := 0; j < len(list); j++ {
			if list[j].ItemTableId == rewardList[i].ItemTableId {
				sameIndex = j
				break
			}
		}
		if sameIndex >= 0 {
			list[sameIndex].Num += rewardList[i].Num
		} else {
			list = append(list, rewardList[i])
		}
	}

	return list
}

// 权重随机抽奖
func randomDraw(prizes []RandRewardInfo) RandRewardInfo {
	//权重累加求和
	var weightSum uint32
	concatWeightArr := make([]RandRewardInfo, 0)
	for _, v := range prizes {
		weightSum += v.weight

		concatWeightArr = append(concatWeightArr, RandRewardInfo{
			itemTableId: v.itemTableId,
			weight:      v.weight,
		})
	}

	//生成一个权重随机数，介于0-weightSum之间
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(int(weightSum))
	randomNumUint := uint32(randomNum)

	////权重数组重组并排序
	//randomNumTmp := RandRewardInfo{itemTableId: -1, weight: randomNumUint}
	//concatWeightArr = append(concatWeightArr, randomNumTmp) //将随机数加入权重数组
	//
	////将包含随机数的新权重数组按从小到大（升序）排序
	//sort.Slice(concatWeightArr, func(i, j int) bool {
	//	return concatWeightArr[i].weight < concatWeightArr[j].weight
	//})
	sort.Slice(prizes, func(i, j int) bool {
		return prizes[i].weight < prizes[j].weight
	})

	//索引权重随机数的数组下标
	var randomNumIndex = -1 //索引随机数在新权重数组中的位置
	//for p, v := range concatWeightArr {
	//	if v.weight == randomNumUint {
	//		randomNumIndex = p
	//	}
	//}
	//
	//randomNumIndexTmp := math.Min(float64(randomNumIndex), float64(len(prizes)-1)) //权重随机数的下标不得超过奖项数组的长度-1，重新计算随机数在奖项数组中的索引位置
	//randomNumIndex = int(randomNumIndexTmp)
	weight := uint32(0)
	for index, vl := range prizes {
		weight += vl.weight
		if weight >= randomNumUint {
			randomNumIndex = index
			break
		}
	}
	//取出对应奖项
	resp := prizes[randomNumIndex] //从奖项数组中取出本次抽奖结果

	return resp
}
