package mongodb

import (
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DBConnect struct {
	Context *DialContext
	DBName  string
}

// 缓存链接
type DBContextWithIp struct {
	Context *DialContext
	Ip      string
}

var dbContextList []*DBContextWithIp

// TODO : gopkg.in/mgo.v2不支持mongo5.0+
// 通过mongo地址创建数据库连接
// 参数1，mongoIP地址，如"mongodb://192.168.1.1:27017"
// 参数2，数据库名
// 参数3，事件数量
func CreateDBConnect(ip string, sessionNum int) (*DialContext, error) {
	if dbContextList == nil {
		dbContextList = make([]*DBContextWithIp, 0)
	}

	//判断是否已有链接，如果有直接返还
	for i := 0; i < len(dbContextList); i++ {
		dbConnect := dbContextList[i]
		if dbConnect.Ip == ip && nil != dbConnect.Context {
			return dbConnect.Context, nil
		}
	}

	//新建一个链接
	context, err := DialWithMode(ip, sessionNum, 5)
	if err != nil {
		return nil, err
	}

	//添加链接缓存
	dbContextWithIp := new(DBContextWithIp)
	dbContextWithIp.Ip = ip
	dbContextWithIp.Context = context
	dbContextList = append(dbContextList, dbContextWithIp)
	return dbContextWithIp.Context, nil
}

// 获取一个数据
// 参数1，数据集名
// 参数2，Key名
// 参数3，Value值
// 参数4，获取到的值
func (c *DBConnect) GetData(collection string, key string,
	val interface{}, i interface{}) error {
	err := c.Context.GetData(c.DBName, collection, key, val, i)
	if err != nil {
		return err
	}

	return nil
}

// goroutine safe
// 获取多个数据
// 参数1，数据集名
// 参数2，Key名
// 参数3，Value值
// 参数4，获取到的值
func (c *DBConnect) GetDataAll(collection string, key string, val interface{}, i interface{}) error {
	return c.Context.GetDataAll(c.DBName, collection, key, val, i)
}

// 根据条件查找 如查玩家vip大于0 则searchKey传："PlayerBase.intattr.17": bson.M{"$gt": 0}
func (c *DBConnect) SearchData(collection string,
	i interface{}, searchKey string, searchValue bson.M) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	err := s.DB(c.DBName).C(collection).Find(bson.M{searchKey: searchValue}).All(i)
	if err != nil {
		return err
	}
	return nil
}

// 根据条件查找 如查玩家vip大于0 则searchValue传bson.M{PlayerBase.intattr.17:bson.M{"$gt": 0}}
func (c *DBConnect) Search(collection string, i interface{}, searchValue bson.M) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	err := s.DB(c.DBName).C(collection).Find(searchValue).All(i)
	if err != nil {
		return err
	}
	return nil
}

// 根据条件查找 如查玩家vip大于0 则searchValue传bson.M{PlayerBase.intattr.17:bson.M{"$gt": 0}}
func (c *DBConnect) SearchByProjection(collection string, i interface{}, searchValue bson.M, projection bson.M) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	err := s.DB(c.DBName).C(collection).Find(searchValue).All(i)
	if err != nil {
		return err
	}
	return nil
}

// 根据条件查找 如查玩家vip大于0 则searchValue传bson.M{PlayerBase.intattr.17:bson.M{"$gt": 0}}
func (c *DBConnect) SearchCount(collection string, searchValue bson.M) int {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	count, err := s.DB(c.DBName).C(collection).Find(searchValue).Count()
	if err != nil {
		return 0
	}
	return count
}

// 根据条件查找 如查玩家vip大于0 则searchValue传bson.M{PlayerBase.intattr.17:bson.M{"$gt": 0}}
func (c *DBConnect) SearchCountByProjection(collection string, searchValue bson.M, projection bson.M) int {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	count, err := s.DB(c.DBName).C(collection).Find(searchValue).Select(projection).Count()
	if err != nil {
		return 0
	}
	return count
}

// 获取某个时间间隔内的数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间
// 参数4，结束时间
// 参数5，获取到的值
func (c *DBConnect) GetDataByTime(collection string, timeKey string, startTime time.Time, endTime time.Time, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": startTime, "$lt": endTime}}).All(i)

}

func (c *DBConnect) GetAcctSystemGameLogAll(collection string, system int, action int, acct string, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{"System": system, "Action": action, "Account": acct}).All(i)
}

// 获取某个时间间隔(时间戳)内的数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间戳
// 参数4，结束时间戳
// 参数5，获取到的值
func (c *DBConnect) GetDataByTimeStamp(collection string, timeKey string,
	startTime int64, endTime int64, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": strconv.FormatInt(startTime, 10),
		"$lt": strconv.FormatInt(endTime, 10)}}).All(i)
}

// 获取某个时间间隔内的数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间
// 参数4，结束时间
// 参数5，获取到的值
func (c *DBConnect) GetDataAfterTime(collection string, timeKey string, startTime time.Time, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": startTime}}).All(i)
}

// 获取某个时间间隔内的数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间
// 参数4，结束时间
// 参数5，获取到的值
func (c *DBConnect) GetDataAfterTimeInt64(collection string, timeKey string, time int64, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": time}}).All(i)
}

// 获取某个时间间隔内的某个数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间
// 参数4，结束时间
// 参数5，同时还需要满足的Key
// 参数6，同时还需要满足的Value
// 参数7，获取到的值
func (c *DBConnect) GetDataByTimeAndValue(collection string, timeKey string,
	startTime time.Time, endTime time.Time, key string, val interface{}, i interface{}) error {

	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": startTime, "$lt": endTime}, key: val}).All(i)
}

// 获取某个时间间隔内的某个数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间
// 参数4，结束时间
// 参数5，同时还需要满足的Key
// 参数6，同时还需要满足的Value
// 参数7，获取到的值
func (c *DBConnect) GetDataByTimeAndValues(collection string, timeKey string,
	startTime time.Time, endTime time.Time, key1 string, val1 interface{}, key2 string, val2 interface{}, key3 string, val3 interface{}, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": startTime, "$lt": endTime}, key1: val1, key2: val2, key3: val3}).All(i)
}

// 获取某个时间间隔内(时间戳)的数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间戳
// 参数4，结束时间戳
// 参数5，同时还需要满足的Key
// 参数6，同时还需要满足的Value
// 参数7，获取到的值
func (c *DBConnect) GetDataByTimeStampAndValue(collection string, timeKey string, startTime int64, endTime int64, key string, val interface{}, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": strconv.FormatInt(startTime, 10), "$lt": strconv.FormatInt(endTime, 10)}, key: val}).All(i)
}

// 获取某个时间间隔内(时间戳)的数据
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间戳
// 参数4，结束时间戳
// 参数5，获取到的值
func (c *DBConnect) GetDataByTimeStampAndValueWithoutKey(collection string, timeKey string, startTime int64, endTime int64, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": strconv.FormatInt(startTime, 10), "$lt": strconv.FormatInt(endTime, 10)}}).All(i)
}

// 检查某个时间间隔内的某个数据有多少条
// 参数1，数据集名
// 参数2，时间字段名
// 参数3，开始时间
// 参数4，结束时间
// 参数5，同时还需要满足的Key
// 参数6，同时还需要满足的Value
func (c *DBConnect) GetDataCountByTimeAndValue(collection string, timeKey string,
	startTime time.Time, endTime time.Time, key string, val interface{}) int {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	count, err := s.DB(c.DBName).C(collection).Find(bson.M{timeKey: bson.M{"$gte": startTime, "$lt": endTime}, key: val}).Count()
	if err != nil {
		return 0
	}
	return count
}

// 根据条件查找并排序 如查玩家vip大于0 则searchKey传："PlayerBase.intattr.17": bson.M{"$gt": 0}
func (c *DBConnect) SearchAndSortData(collection string,
	i interface{}, searchKey string, searchValue bson.M, sort string, limitCount int) error {

	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	err := s.DB(c.DBName).C(collection).Find(bson.M{searchKey: searchValue}).Sort(sort).Limit(limitCount).All(i)
	if err != nil {
		return err
	}
	return nil
}

// 获取指定符合条件的指定条数的数据
func (c *DBConnect) GetLimitDataAndSort(collection string, limit int, searchValue interface{}, i interface{}, fields ...string) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(c.DBName).C(collection).Find(searchValue).Sort(fields...).Limit(limit).All(i)
}

func (c *DBConnect) Query(collection string, searchValue, selector interface{}, limit int, i interface{}, sort ...string) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	return s.DB(c.DBName).C(collection).Find(searchValue).Select(selector).Sort(sort...).Limit(limit).All(i)
}

// 分页查询
// collection:数据集名称
// i:数据结果
// sort:排序条件 如根据玩家等级排序 正序：PlayerBase.intattr.0 降序：-PlayerBase.intattr.0
// startIndex:起始索引
// dataCount:需要获取的数据数量
func (c *DBConnect) GetDatByPage(collection string, i interface{}, sort string, startIndex int, dataCount int) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	return s.DB(c.DBName).C(collection).Find(nil).Sort(sort).Skip(startIndex).Limit(dataCount).All(i)
}

// 获取某个数据集总共有多少条
// 参数1，数据集名
func (c *DBConnect) GetDataCountTotal(collection string) (int, error) {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	count, err := s.DB(c.DBName).C(collection).Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 获取某个数据集有多少条
// 参数1，数据集名
// 参数2，满足的Key
// 参数3，满足的Value
func (c *DBConnect) GetDataCount(collection string, key string, searchValue bson.M) int {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	count, err := s.DB(c.DBName).C(collection).Find(bson.M{key: searchValue}).Count()
	if err != nil {
		return 0
	}
	return count
}

// 获取自定义查询数据的数量 则searchValue传bson.M{PlayerBase.intattr.17:bson.M{"$gt": 0}}
func (c *DBConnect) GetCustomDataCount(collection string, searchValue bson.M) int {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	count, err := s.DB(c.DBName).C(collection).Find(searchValue).Count()
	if err != nil {
		return 0
	}
	return count
}

// 获取整个表的数据
// 参数1，数据集名
// 参数2，获取到的值
func (c *DBConnect) GetTableDataAll(collection string, i interface{}) error {
	return c.Context.GetTableDataAll(c.DBName, collection, i)
}

// 获取整个表的数据 异步
func (c *DBConnect) GetTableDataAllAsync(collection string, i interface{}, callback func(i interface{}, err error)) {

	go func() {
		s := c.Context.Ref()
		defer c.Context.UnRef(s)
		err := s.DB(c.DBName).C(collection).Find(nil).All(i)

		if callback != nil {
			callback(i, err)
		}
	}()
}

// 获得整个表的项目统计数
func (c *DBConnect) GetTableCount(collection string) (int, error) {
	return c.Context.GetCount(c.DBName, collection)
}

// goroutine safe
// 保存一个数据
// 参数1，数据集名
// 参数2，Key名
// 参数3，Value值
// 参数4，需要保存的数据
func (c *DBConnect) SaveData(collection string, key string, val interface{}, i interface{}) error {
	return c.Context.SaveData(c.DBName, collection, key, val, i)
}

// goroutine safe
// 插入一个数据
// 参数1，数据集名
// 参数2，插入的数据
func (c *DBConnect) InsertData(collection string, i interface{}) error {
	return c.Context.InsertData(c.DBName, collection, i)
}

// 删除数据
func (c *DBConnect) RemoveData(collection string, key string, i interface{}) error {
	return c.Context.RemoveData(c.DBName, collection, key, i)
}

func (c *DBConnect) RemoveAllData(collection string) {
	c.Context.RemoveAllData(c.DBName, collection)
}

func (c *DBConnect) MapReduce(collection string, s bson.M, ret interface{}, mp *mgo.MapReduce) error {
	return c.Context.MapReduce(c.DBName, collection, s, ret, mp)
}

// goroutine safe
func (c *DBConnect) Close() {
	c.Context.Close()
}

// 获取指定数据集一条数据
func (c *DBConnect) GetOne(collection string, query bson.M, selector bson.M, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	return s.DB(c.DBName).C(collection).Find(query).Select(selector).One(i)
}

// 获取指定数据集一条数据 异步
func (c *DBConnect) GetOneAsync(collection string, query bson.M, selector bson.M, i interface{}, callback func(i interface{}, err error)) {

	//启动协程
	go func() {
		s := c.Context.Ref()
		defer c.Context.UnRef(s)

		err := s.DB(c.DBName).C(collection).Find(query).Select(selector).One(i)

		if callback != nil {
			callback(i, err)
		}
	}()

}
func (c *DBConnect) GetAll(collection string, query bson.M, selector bson.M, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	return s.DB(c.DBName).C(collection).Find(query).Select(selector).All(i)
}

func (c *DBConnect) GetPipe(db string, collection string, i interface{}) *mgo.Pipe {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)
	return s.DB(db).C(collection).Pipe(i)
}

func (c *DBConnect) GetFindIn(collection string, query string, slice []uint32, i interface{}) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	return s.DB(c.DBName).C(collection).Find(bson.M{query: bson.M{"$in": slice}}).All(i)
}

func (c *DBConnect) GetDataLimitAndPage(collection string, query bson.M, limit, skip int, i interface{}, fields ...string) error {
	s := c.Context.Ref()
	defer c.Context.UnRef(s)

	return s.DB(c.DBName).C(collection).Find(query).Sort(fields...).Limit(limit).Skip(skip).All(i)
}
