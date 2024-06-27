package mongodb

import (
	"BilliardServer/Util/event"
	"container/heap"
	"sync"
	"time"

	"BilliardServer/Util/log"
	"BilliardServer/Util/stack"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// session
type Session struct {
	*mgo.Session
	ref   int
	index int
}

// session heap
type SessionHeap []*Session

func (h SessionHeap) Len() int {
	return len(h)
}

func (h SessionHeap) Less(i, j int) bool {
	return h[i].ref < h[j].ref
}

func (h SessionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *SessionHeap) Push(s interface{}) {
	s.(*Session).index = len(*h)
	*h = append(*h, s.(*Session))
}

func (h *SessionHeap) Pop() interface{} {
	l := len(*h)
	s := (*h)[l-1]
	s.index = -1
	*h = (*h)[:l-1]
	return s
}

type DialContext struct {
	sync.Mutex
	sessions SessionHeap
}

// 数据库事件
type DBEvent struct {
	CallBack func(interface{}) //回调函数
	Param    interface{}       //参数
	Data     interface{}       //数据
}

// goroutine safe
func Dial(url string, sessionNum int) (*DialContext, error) {
	c, err := DialWithTimeout(url, sessionNum, 10*time.Second, 5*time.Minute, mgo.Primary)
	return c, err
}

// goroutine safe
func DialWithMode(url string, sessionNum int, mode mgo.Mode) (*DialContext, error) {
	c, err := DialWithTimeout(url, sessionNum, 10*time.Second, 5*time.Minute, mode)
	return c, err
}

// goroutine safe
func DialWithTimeout(url string, sessionNum int, dialTimeout time.Duration, timeout time.Duration, mode mgo.Mode) (*DialContext, error) {
	if sessionNum <= 0 {
		sessionNum = 100
		log.Waring("invalid sessionNum, reset to %v", sessionNum)
	}

	s, err := mgo.DialWithTimeout(url, dialTimeout)
	if err != nil {
		return nil, err
	}
	s.SetSyncTimeout(timeout)
	s.SetSocketTimeout(timeout)
	s.SetPoolLimit(sessionNum)
	s.SetMode(mode, true)

	c := new(DialContext)

	// sessions
	c.sessions = make(SessionHeap, sessionNum)
	c.sessions[0] = &Session{s, 0, 0}
	for i := 1; i < sessionNum; i++ {
		c.sessions[i] = &Session{s.New(), 0, i}
	}
	heap.Init(&c.sessions)

	return c, nil
}

// goroutine safe
func (c *DialContext) Close() {
	c.Lock()
	for _, s := range c.sessions {
		s.Close()
		if s.ref != 0 {
			log.Error("session ref = %v", s.ref)
		}
	}
	c.Unlock()
}

// goroutine safe
func (c *DialContext) Ref() *Session {
	c.Lock()
	s := c.sessions[0]
	if s.ref == 0 {
		s.Refresh()
	}
	s.ref++
	heap.Fix(&c.sessions, 0)
	c.Unlock()

	return s
}

// goroutine safe
func (c *DialContext) UnRef(s *Session) {
	c.Lock()
	s.ref--
	heap.Fix(&c.sessions, s.index)
	c.Unlock()
}

// //////////////////////////////////////////////////////////////////////////////
func (c *DialContext) GetData(db string, collection string, key string,
	val interface{}, i interface{}) error {

	s := c.Ref()
	defer c.UnRef(s)
	err := s.DB(db).C(collection).Find(bson.M{key: val}).One(i)
	if err != nil {
		return err
	}

	return nil
}

// goroutine safe
// 获取多个数据
func (c *DialContext) GetDataAll(db string, collection string, key string, val interface{}, i interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Find(bson.M{key: val}).All(i)
}

func (c *DialContext) GetDataByKeyAndTime(db string, collection string, key string, val interface{},
	startTimeKey string, endTimeKey string, targetTime time.Time, i interface{}) error {

	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Find(bson.M{key: val, startTimeKey: bson.M{"$lt": targetTime},
		endTimeKey: bson.M{"$gt": targetTime}}).All(i)
}

// 获取整个表的数据
func (c *DialContext) GetTableDataAll(db string, collection string, i interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Find(nil).All(i)
}

// 获取整个表的数据
func (c *DialContext) GetTableCount(db string, collection string) int {
	s := c.Ref()
	defer c.UnRef(s)
	count, err := s.DB(db).C(collection).Count()
	if err != nil {
		return 0
	}
	return count
}

// 删除数据
func (c *DialContext) RemoveData(db string, collection string, key string, val interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	err := s.DB(db).C(collection).Remove(bson.M{key: val})

	return err
}

// 删除数据
func (c *DialContext) RemoveAllByQuery(db string, collection string, query bson.M) error {
	s := c.Ref()
	defer c.UnRef(s)
	_, err := s.DB(db).C(collection).RemoveAll(query)

	return err
}

// 异步删除数据
func (c *DialContext) RemoveDataAsync(db string, collection string, key string, val interface{},
	fun func(param interface{}), param interface{}) {

	//启动协程
	go func() {
		s := c.Ref()
		defer c.UnRef(s)
		s.DB(db).C(collection).Remove(bson.M{key: val})

		//函数回调
		if fun != nil {
			fun(param)
		}
	}()
}

// 异步删除数据
func (c *DialContext) RemoveDataByTimeAsync(db string, collection string, key string, timeCondition string, val interface{},
	fun func(param interface{}), param interface{}) {

	//启动协程
	go func() {
		s := c.Ref()
		defer c.UnRef(s)

		s.DB(db).C(collection).Remove(bson.M{key: bson.M{timeCondition: val}})

		//函数回调
		if fun != nil {
			fun(param)
		}
	}()

}

// 异步删除数据
func (c *DialContext) RemoveDataByTimeAndConditionAsync(db string, collection string, key string, timeCondition string, val interface{}, key2 string, val2 interface{},
	fun func(param interface{}), param interface{}) {

	//启动协程
	go func() {
		s := c.Ref()
		defer c.UnRef(s)

		s.DB(db).C(collection).Remove(bson.M{key: bson.M{timeCondition: val}, key2: val2})

		//函数回调
		if fun != nil {
			fun(param)
		}
	}()

}

// 删除所有数据
func (c *DialContext) RemoveAllData(db string, collection string) {
	s := c.Ref()
	defer c.UnRef(s)
	s.DB(db).C(collection).RemoveAll(bson.M{})
}

func (c *DialContext) RemoveAllDataAsync(db string, collection string, fun func(param interface{}), param interface{}) {
	go func() {
		s := c.Ref()
		defer c.UnRef(s)
		s.DB(db).C(collection).RemoveAll(bson.M{})
		if fun != nil {
			fun(param)
		}
	}()
}

// goroutine safe
func (c *DialContext) SaveData(db string, collection string, key string, val interface{}, i interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	_, err := s.DB(db).C(collection).Upsert(bson.M{key: val}, i)
	if err != nil {
		DB_Error(err)
	}
	return err
}

func (c *DialContext) SaveData2(db string, collection string, key string, val interface{}, i interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	_, err := s.DB(db).C(collection).Upsert(bson.M{key: val}, i)
	if err != nil {
		DB_Error(err)
	}
	return err
}

// 异步保存数据
func (c *DialContext) SaveDataAsync(db string, collection string, key string,
	val interface{}, i interface{},
	fun func(param interface{}), param interface{}) {

	saveData := stack.DeepCloneForDB(i)
	//启动协程
	go func() {
		c.SaveData(db, collection, key, val, saveData)

		//函数回调
		if fun != nil {
			event.Fire("OnCallBackFun", fun)
		}
	}()
}

// goroutine safe
func (c *DialContext) InsertData(db string, collection string, i ...interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Insert(i...)
}

// 异步保存数据
func (c *DialContext) InsertDataAsync(db string, collection string, fun func(param interface{}), param interface{}, i ...interface{}) {

	//启动协程
	go func() {
		s := c.Ref()
		defer c.UnRef(s)
		err := s.DB(db).C(collection).Insert(i...)
		if err != nil {
			log.Error("异步插入数据错误:", err)
		}

		//函数回调
		if fun != nil {
			fun(param)
		}
	}()
}

// 夺宝获取机器人
func (c *DialContext) SnatchPart_GetPlayerFromDb(db string, collection string,
	key string, val interface{}, i interface{}, level int, account string) error {

	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Find(bson.M{key: val, "Account": bson.M{"$ne": account},
		"PlayerBase.intattr.0": bson.M{"$gte": level - 10, "$lte": level + 10}}).All(i)
}

func (c *DialContext) FriendPart_GetPlayerFromDb(db string, collection string,
	i interface{}, level int, account string) error {

	s := c.Ref()
	defer c.UnRef(s)
	minLevel := 0
	maxLevel := 0
	if level%10 == 0 {
		minLevel = level - 10 + 1
		maxLevel = level
	} else {
		minLevel = level - level%10 + 1
		maxLevel = minLevel + 10 - 1
	}

	return s.DB(db).C(collection).Find(bson.M{"Account": bson.M{"$ne": account},
		"PlayerBase.intattr.0": bson.M{"$gte": minLevel, "$lte": maxLevel}}).All(i)
}

// goroutine safe
func (c *DialContext) GetCount(db string, collection string) (int, error) {
	s := c.Ref()
	defer c.UnRef(s)

	n, err := s.DB(db).C(collection).Count()
	if err != nil {
		DB_Error(err)
	}

	return n, err
}

// 根据条件查找
func (c *DialContext) SearchData(db string, collection string, key string, val interface{},
	i interface{}, limitNum int, searchKey string, searchValue bson.M) {

	s := c.Ref()
	defer c.UnRef(s)

	err := s.DB(db).C(collection).Find(bson.M{key: val, searchKey: searchValue}).Limit(limitNum).All(i)
	if err != nil {
		DB_Error(err)
	}
}

// 获取指定符合条件的指定条数的数据
func (c *DialContext) GetLimitDataAndSort(db string, collection string, limit int, searchValue interface{}, i interface{}, fields ...string) error {
	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Find(searchValue).Sort(fields...).Limit(limit).All(i)
}

// 根据条件查找 如查玩家vip大于0 则searchValue传bson.M{PlayerBase.intattr.17:bson.M{"$gt": 0}}
func (c *DialContext) Search(db string, collection string, i interface{}, searchValue bson.M) error {
	s := c.Ref()
	defer c.UnRef(s)

	err := s.DB(db).C(collection).Find(searchValue).All(i)
	if err != nil {
		DB_Error(err)
	}
	return err
}

func (c *DialContext) MapReduce(db string, collection string, bm bson.M, ret interface{}, mp *mgo.MapReduce) error {
	s := c.Ref()
	defer c.UnRef(s)
	_, err := s.DB(db).C(collection).Find(bm).MapReduce(mp, ret)
	return err
}
func DB_Error(err interface{}) {
	log.Error("DB Error:", err)
	stack.PrintCallStack()
}

// 通用获取函数接口
func (c *DialContext) GetOne(db string, collection string, query bson.M, selector bson.M, i interface{}) error {

	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).Find(query).Select(selector).One(i)
}

// 通用获取函数接口
func (c *DialContext) GetOneAsync(db string, collection string, query bson.M, selector bson.M, i interface{}, callback func(i interface{}, err error)) {

	//启动协程
	go func() {
		s := c.Ref()
		defer c.UnRef(s)

		err := s.DB(db).C(collection).Find(query).Select(selector).One(i)

		if callback != nil {
			callback(i, err)
		}
	}()

}

// goroutine safe
func (c *DialContext) SaveOne(db string, collection string, query bson.M, i interface{}) {
	s := c.Ref()
	defer c.UnRef(s)
	_, err := s.DB(db).C(collection).Upsert(query, i)
	if err != nil {
		DB_Error(err)
	}
	return
}

// goroutine safe
func (c *DialContext) SaveOneAsync(db string, collection string, query bson.M, i interface{}) {

	//启动协程
	go func() {
		s := c.Ref()
		defer c.UnRef(s)
		_, err := s.DB(db).C(collection).Upsert(query, i)
		if err != nil {
			DB_Error(err)
		}
	}()

}
func (c *DialContext) GetAll(db string, collection string, query bson.M, selector bson.M, i interface{}) error {

	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).Find(query).Select(selector).All(i)
}

// 删除一个
func (c *DialContext) RemoveOne(db string, collection string, query bson.M) error {
	s := c.Ref()
	defer c.UnRef(s)
	err := s.DB(db).C(collection).Remove(query)

	return err
}

func (c *DialContext) GetAllCount(db string, collection string, query bson.M) (int, error) {

	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).Find(query).Count()
}

func (c *DialContext) GetDataAllByQuery(db string, collection string, query bson.M, field bson.M, i interface{}) error {
	s := c.Ref()
	defer c.UnRef(s)
	return s.DB(db).C(collection).Find(query).Select(field).All(i)
}

func (c *DialContext) UpdateDataById(db string, collection string, ObjId bson.ObjectId, update bson.M) (*mgo.ChangeInfo, error) {
	s := c.Ref()
	defer c.UnRef(s)
	ret, err := s.DB(db).C(collection).UpsertId(ObjId, update)
	return ret, err
}

// goroutine safe
// 更新所有纪录
func (c *DialContext) UpdateAllData(db string, collection string, update bson.M) (*mgo.ChangeInfo, error) {
	s := c.Ref()
	defer c.UnRef(s)
	ret, err := s.DB(db).C(collection).UpdateAll(nil, update)
	return ret, err
}

// goroutine safe
func (c *DialContext) EnsureIndex(db string, collection string, key []string) error {
	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).EnsureIndex(mgo.Index{
		Key:    key,
		Unique: false,
		Sparse: true,
	})
}

// goroutine safe
func (c *DialContext) EnsureUniqueIndex(db string, collection string, key []string) error {
	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).EnsureIndex(mgo.Index{
		Key:    key,
		Unique: true,
		Sparse: true,
	})
}
