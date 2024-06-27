package tools

import (
	"BilliardServer/Common"
	"container/list"
	"crypto/md5"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

func GetArgsMode(args []string) string {
	var mode string
	for _, v := range args {
		if v == Common.ModeLocal || v == Common.ModeDev || v == Common.ModeProd {
			mode = v
		}
	}

	return mode
}

func IsModeProd(mode string) bool {
	var isProd bool

	if mode == Common.ModeProd {
		isProd = true
	}

	return isProd
}

func GetModeConfPath(mode string) string {
	var cfgPath string

	if mode == Common.ModeLocal {
		cfgPath = Common.LocalConfPath
	} else if mode == Common.ModeDev {
		cfgPath = Common.DevConfPath
	} else if mode == Common.ModeProd {
		cfgPath = Common.ProdConfPath
	}

	return cfgPath
}

func GetModeTablePath(mode string) string {
	var cfgPath string

	if mode == Common.ModeLocal {
		cfgPath = Common.LocalTablePath
	} else if mode == Common.ModeDev {
		cfgPath = Common.DevConfPath
	} else if mode == Common.ModeProd {
		cfgPath = Common.ProdConfPath
	}

	return cfgPath
}

func ArrShuffle(arr []int) []int32 {
	var rs []int32

	if len(arr) <= 0 {
		return rs
	}

	rand.NewSource(time.Now().UnixNano())
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})

	if len(arr) <= 0 {
		return rs
	}

	for _, v := range arr {
		rs = append(rs, int32(v))
	}

	return rs
}

func LimitChinese(str string) bool {
	res, err := regexp.Compile("^[\u4e00-\u9fa5]{3,8}$")
	if err != nil {
		return false
	}
	return res.MatchString(str)
}

func FormatTimeStr(str, s string) string {
	return strings.Replace(str, s, ".", -1)
}

func StringReplace(str, s, news string) string {
	strs := strings.Replace(str, "\\n", "\n", 1)
	return strings.Replace(strs, s, news, 1)
}

// slice或map里是否包含某元素
func Contains(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}
	return false, errors.New("not in")
}

// 获取某个路径下的所有文件名
func GetAllFileNamesByPath(path string) []string {
	//获取绝对路径
	absolutePath, _ := filepath.Abs(path)
	listStr := list.New()

	//遍历获取该路径下的所有文件
	filepath.Walk(absolutePath, func(path string, fi os.FileInfo, err error) error {
		if nil == fi {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		name := fi.Name()
		listStr.PushBack(name)

		return nil
	})

	//转换成切片数据
	fileNamelist := make([]string, 0)
	for el := listStr.Front(); nil != el; el = el.Next() {
		fileNamelist = append(fileNamelist, el.Value.(string))
	}
	return fileNamelist
}

// 转换一个Bson
func BsonObjectID(s string) bson.ObjectId {
	if s == "" {
		return bson.NewObjectId()
	}

	if bson.IsObjectIdHex(s) {
		return bson.ObjectIdHex(s)
	}

	return bson.ObjectId(s)
}

// 获取结构体中字段的名称
func GetStructFieldName(structName interface{}) ([]string, error) {
	t := reflect.TypeOf(structName)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("Check type error not Struct")
	}
	fieldNum := t.NumField()
	result := make([]string, 0, fieldNum)
	for i := 0; i < fieldNum; i++ {
		result = append(result, t.Field(i).Name)
	}
	return result, nil
}

// 获取结构体中Tag的值，如果没有tag则返回字段值
func GetStructTagName(structName interface{}) ([]string, error) {
	t := reflect.TypeOf(structName)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("Check type error not Struct")
	}
	fieldNum := t.NumField()
	result := make([]string, 0, fieldNum)
	for i := 0; i < fieldNum; i++ {
		tagName := t.Field(i).Name
		tags := strings.Split(string(t.Field(i).Tag), "\"")
		if len(tags) > 1 {
			tagName = tags[1]
		}
		result = append(result, tagName)
	}
	return result, nil
}

// GetEntityID 并发环境下生成一个增长的id,按需设置局部变量或者全局变量
func GetEntityID(ID *int32) int32 {
	var n, v int32
	for {
		v = atomic.LoadInt32(ID)
		n = v + 1
		if atomic.CompareAndSwapInt32(ID, v, n) {
			break
		}
	}
	return n
}

func FormatUint32(sNum uint32) string {
	return strconv.FormatUint(uint64(sNum), 10)
}
func FormatUint64(sNum uint64) string {
	return strconv.FormatUint(sNum, 10)
}
func FormatInt32(sNum int32) string {
	return strconv.FormatInt(int64(sNum), 10)
}
func FormatInt64(sNum int64) string {
	return strconv.FormatInt(sNum, 10)
}

func StringToInt(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

// 截取字符串 start 起点下标 end 终点下标(不包括)
func GetSubString(str string, start int, end int) string {
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

// 截取小数位数
func FloatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}

// 生成随机字符串
func GetRandomString(l int) string {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func RemoveElement(nums []uint32, val uint32) []uint32 {
	var newNums []uint32
	for _, num := range nums {
		if num != val {
			newNums = append(newNums, num)
		}
	}
	return newNums
}

func GetFirstDateOfWeek(t time.Time) time.Time {

	offset := int(time.Monday - t.Weekday())
	if offset > 0 {
		offset = -6
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).
		AddDate(0, 0, offset)

}

func GetUint(i int32) uint32 {
	if i <= 0 {
		return uint32(0)
	}
	return uint32(i)
}

func MD5(str string) string {
	data := []byte(str) //切片
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}
