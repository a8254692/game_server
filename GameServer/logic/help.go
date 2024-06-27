package logic

import (
	"BilliardServer/Common/entity"
	conf "BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/uuid"
	"errors"
	"strconv"
)

// 获取EntityPlayer
func GetEntityPlayerById(entityId uint32) (*entity.EntityPlayer, error) {
	if entityId <= 0 {
		return nil, errors.New("entityId is nil")
	}

	tEntity := Entity.EmPlayer.GetEntityByID(entityId)
	if tEntity == nil {
		return nil, errors.New("tEntity is nil")
	}

	tEntityPlayer, ok := tEntity.(*entity.EntityPlayer)
	if !ok {
		return nil, errors.New("tEntity change err")
	}

	return tEntityPlayer, nil
}

// 校验两个ID的个十百位及最高位是否都相同
func CheckSameHundredId(id uint32, otherId uint32) bool {
	var resp bool

	if id < 100 && otherId < 100 {
		return resp
	}

	//先校验最高位
	highIdStr := strconv.Itoa(int(id))
	highId := CheckAndSubStr(highIdStr, 0, 1)
	highOtherIdStr := strconv.Itoa(int(otherId))
	highOtherId := CheckAndSubStr(highOtherIdStr, 0, 1)
	if highId != highOtherId {
		return resp
	}

	//校验个十百位
	idB := id / 100 % 10
	idS := id / 10 % 10
	idG := id / 1 % 10

	otherIdB := otherId / 100 % 10
	otherIdS := otherId / 10 % 10
	otherIdG := otherId / 1 % 10

	if idG == otherIdG && idS == otherIdS && idB == otherIdB {
		resp = true
	}

	return resp
}

// 截取字符串 start 起点下标 end 终点下标(不包括)
func CheckAndSubStr(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return ""
	}

	if end < 0 || end > length {
		return ""
	}
	return string(rs[start:end])
}

func GetResParam(sysID, actionID uint32) *entity.ResParam {
	return &entity.ResParam{Uuid: uuid.Next(), SysID: sysID, ActionID: actionID}
}

// 获取Table后两位数
func GetKeyByTableId(tableId uint32) int {
	return int(tableId) / conf.CueKeyDigit % 100
}
