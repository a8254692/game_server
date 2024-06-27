package tools

import (
	"BilliardServer/Util/log"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

// 获取时间之间的天数
func GetTimesDistanceDays(startTime int64, endTime int64) int64 {
	//最小时间
	lastTime := time.Now().AddDate(-10, 0, 0)
	if startTime < lastTime.Unix() || endTime < lastTime.Unix() {
		return 0
	}
	days := (GetBeginTime(endTime) - GetBeginTime(startTime)) / 86400
	return days
}

// 通过时间戳获取时间字符串
func GetTimeByTimeStamp(timeStamp int64) string {
	return time.Unix(timeStamp, 0).Format("2006-01-02 15:04:05")
}

// 获取0点时间戳
func GetBeginTime(t64 int64) int64 {
	t := time.Unix(t64, 0)
	return t.Unix() - int64(t.Hour()*3600+t.Minute()*60+t.Second())
}

// 获取今天的零点时间戳
func GetTodayBeginTime() int64 {
	return GetBeginTime(time.Now().Unix())
}

// 获取明天的零点时间戳
func GetTomorrowBeginTime() int64 {
	return GetBeginTime(time.Now().Unix()) + 86400
}

func GetTimeByUnix(utc int64) time.Time {
	return GetTimeByString(GetTimeStringByUTC(utc))
}

func GetTimeByString(tstr string) time.Time {
	if tstr <= "1970-01-01 00:00:00" {
		return time.Unix(0, 0)
	}
	//string转化为时间，layout必须为 "2006-01-02 15:04:05"
	t, err := time.Parse("2006-01-02 15:04:05", tstr)
	if err != nil {
		fmt.Printf("GetTimeByString err %v", err)
	}
	return time.Unix(t.Unix()-int64(GetLocalDiffTime()), 0)
}

func GetLocalDiffTime() int {
	t := time.Unix(0, 0)
	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}

// 时间戳toString
func GetTimeStringByUTC(utc int64) string {
	t := time.Unix(utc, 0)
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

// 时间戳toString
func GetTimeDayStringByUTC(utc int64) string {
	t := time.Unix(utc, 0)
	return fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

// 获取两天的时间间隔（天数）
func Tool_GetTimeSubDays(now, subTime time.Time) int {
	hours := subTime.Sub(now).Hours()
	if hours <= 0 {
		return -1
	}
	// sub hours less than 24
	if hours < 24 {
		// may same day
		t1y, t1m, t1d := now.Date()
		t2y, t2m, t2d := subTime.Date()
		isSameDay := (t1y == t2y && t1m == t2m && t1d == t2d)
		if isSameDay {
			return 0
		} else {
			return 1
		}
	} else { // equal or more than 24
		if (hours/24)-float64(int(hours/24)) == 0 { // just 24's times
			return int(hours / 24)
		} else { // more than 24 hours
			return int(hours/24) + 1
		}
	}
}

// 获取0点时间
func Tool_GetZeroTime(_time time.Time) time.Time {
	return time.Date(_time.Year(), _time.Month(), _time.Day(), 0, 0, 0, 0, time.Local)
}

// 获取时间间隔，如果已经过了，就拿下一天的
// 参数1，小时
// 参数2，分钟
// 返回值，以秒计算的时间间隔
func Tool_GetTimeGap(hour, minute int) int64 {

	timeNow := time.Now()
	targetTime := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), hour, minute, 0, 0, time.Local)
	//当前的小时数大于预定的小时数
	//当前的小时数等于预定的小时数，并且分钟大于预定的分钟数

	if timeNow.Hour() > hour || (timeNow.Hour() == hour && timeNow.Minute() > minute) {
		//明天
		targetTime = targetTime.Add(24 * time.Hour)

	}
	return targetTime.Unix() - timeNow.Unix()
}

// 返回当天时间戳
func Tool_ReturnTime() int64 {
	year := strconv.Itoa(time.Now().Year())
	mongth := int(time.Now().Month())
	day := time.Now().Day()
	loc, _ := time.LoadLocation("Local")
	zeroTime, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf(`%s-%02d-%02d 00:00:00`, year, mongth, day), loc)
	if err != nil {
		log.Error(err)
		return 0
	}

	return zeroTime.Local().Unix()
}

func Tool_2StartAndEndTime(startTimeStr, endTimeStr string) (startTime, endTime time.Time, err error) {
	if startTimeStr == "" || endTimeStr == "" {
		err = errors.New("startTimeStr & endTimeStr is nil")
		return
	}

	//处理时间
	startTimeArr := strings.Split(startTimeStr, " ")
	endTimeArr := strings.Split(endTimeStr, " ")
	if len(startTimeArr) < 2 || len(endTimeArr) < 2 {
		err = errors.New("开始或结束时间异常：" + startTimeStr + "," + endTimeStr)
		return
	}

	//获取两个时间点
	startTime = ConvertTime(startTimeArr[0])
	endTime = ConvertTime(endTimeArr[0])

	//如果开始时间大于结束时间则时间对换
	if startTime.Unix() > endTime.Unix() {
		temp := startTime
		startTime = endTime
		endTime = temp
	}

	//开始时间都是从当天0点开始，结束时间都为当天23:59:59，
	startTime = Tool_GetZeroTime(startTime)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
	return
}

// 将HTML5时间设置成下格式
// xxxx-xx-xx 00:00:00
func ConvertTime(timeStr string) time.Time {
	//分割出日期和时间
	strArray := strings.Split(timeStr, " ")

	//分割出年月日字符
	dateStrArray := strings.Split(strArray[0], "-")
	if len(dateStrArray) < 3 {
		return time.Time{}
	}

	//转化为整数
	date := make([]int, 3)
	for i := 0; i < 3; i++ {
		date[i], _ = strconv.Atoi(dateStrArray[i])
	}

	location, _ := time.LoadLocation("Local")
	//判断是否有时分秒
	if len(strArray) >= 2 {
		//分割出时分秒字符
		dateStrArray2 := strings.Split(strArray[1], ":")
		//转化为整数
		date2 := make([]int, 3)
		for i := 0; i < 3; i++ {
			date2[i], _ = strconv.Atoi(dateStrArray2[i])
		}

		result := time.Date(date[0], time.Month(date[1]), date[2],
			date2[0], date2[1], date2[2], 0, location)

		return result
	} else {
		result := time.Date(date[0], time.Month(date[1]), date[2],
			0, 0, 0, 0, location)
		return result
	}
}

// 秒数转换为时间（天时分秒）
func ConvertSecond2Time(seconds int64) string {
	//获得秒
	second := seconds % 60

	//获得分
	mins := seconds / 60
	min := mins % 60

	//获得时
	hours := mins / 60
	hour := hours % 24

	//获得天数
	days := hour / 24

	if days > 0 {
		return fmt.Sprintf("%d天%d时%d分%d秒", days, hour, min, second)
	} else {
		return fmt.Sprintf("%d时%d分%d秒", hour, min, second)
	}

}

// 获取某个time的日期字符串
func GetDate(date time.Time) string {
	return strconv.Itoa(date.Year()) + "年" + strconv.Itoa(int(date.Month())) + "月" + strconv.Itoa(date.Day()) + "日"
}

// 获取某个time的日期字符串
func GetTime(date time.Time) string {
	return strconv.Itoa(date.Year()) + "年" + strconv.Itoa(int(date.Month())) + "月" + strconv.Itoa(date.Day()) + "日" + strconv.Itoa(date.Hour()) + "时" + strconv.Itoa(date.Minute()) + "分"
}

func GetUnixFromStr(str string) int64 {
	timeLayout := "2006-01-02 15:04:05"                      //转化所需模板
	loc, _ := time.LoadLocation("Local")                     //重要：获取时区
	theTime, _ := time.ParseInLocation(timeLayout, str, loc) //使用模板在对应时区转化为time.time类型
	return theTime.Unix()
}

func GetDateOfWeekUnix() (weekMonday int64, endDay int64) {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	endN := offset + 7
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekEndDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, endN)
	weekMonday = weekStartDate.Unix()
	endDay = weekEndDate.Unix()
	return
}

func GetDateOfWeek() (weekMonday string, endDay string) {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	endN := offset + 7
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekEndDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, endN)
	weekMonday = weekStartDate.Format("2006-01-02")
	endDay = weekEndDate.Format("2006-01-02")
	return
}

func GetDayDateFormat(date string) string {
	dateParse, err := time.ParseInLocation("2006-01-02 15:04:05", date, time.Local)
	if err != nil {
		return ""
	}
	dateKey := dateParse.Format("2006-01-02")
	return dateKey
}

func GetNowDateString() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02")
}

func GetLeftSecondByTomorrow() int {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return int(t2.AddDate(0, 0, 1).Unix() - time.Now().Unix())
}

func GetWeekDay() int {
	now := time.Now()
	return int(now.Weekday()) // 转换为数字，从1开始
}

func GetThisWeekFirstDate() (weekMonday int64) {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)

	return weekStartDate.Unix()
}

func GetThisWeekFirstDateString() string {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)

	return weekStartDate.Format("2006-01-02")
}

func GetNowTimeMonthAndUnix() (m, n int, t int64) {
	timeNow := time.Now()
	month := timeNow.Month()
	day := timeNow.Day()
	return int(month), day, timeNow.Unix()
}

func FormatToNMonthInt(m int) int {
	timeStr := time.Now().Format("2006-01")
	t2, _ := time.ParseInLocation("2006-01", timeStr, time.Local)
	return int(t2.AddDate(0, m, 0).Month())
}

func FormatToNdayString(d int) string {
	timeStr := time.Now().Format("2006-01-02")
	t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)

	return t2.AddDate(0, 0, d).Format("2006-01-02")
}

func GetBeforeNDayString(n int) []string {
	var beforeDays []string
	for i := 0; i < n; i++ {
		beforeDays = append(beforeDays, FormatToNdayString(-i))
	}

	return beforeDays
}

func GetThisWeekSaturday() (weekSaturday int64) {
	now := time.Now()

	offset := int(time.Saturday - now.Weekday())
	if offset == 6 {
		offset = -1
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)

	return weekStartDate.Unix()
}

func GetTimeMinFormat() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02  15:04:00")
}

func GetHourMinuteInt(hour, minu int) (t int) {
	return hour*100 + minu
}
