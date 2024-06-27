package tools

import (
	"errors"
)

func GetNumLenForUint32(n uint32) uint8 {
	count := 0
	for n > 0 {
		n = n / 10
		count++
	}
	return uint8(count)
}

// 校验tableId是否合法
func CheckItemTableIdLegal(tableId uint32) bool {
	var isLegal bool
	if GetNumLenForUint32(tableId) == 8 {
		isLegal = true
	}
	return isLegal
}

// 通过物品tableId获取物品类型
func GetItemTypeByTableId(tableId uint32) (uint32, error) {
	if !CheckItemTableIdLegal(tableId) {
		return 0, errors.New("tableId len is not 8")
	}

	return tableId / 10000000 % 10, nil
}

// 获取球杆后三位
func GetCueIDByTableID(tableId uint32) (uint32, error) {
	if !CheckItemTableIdLegal(tableId) {
		return 0, errors.New("tableId len is not 8")
	}
	return tableId % 1000, nil
}
