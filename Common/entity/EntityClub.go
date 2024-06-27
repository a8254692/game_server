package entity

import (
	conf "BilliardServer/GameServer/initialize/consts"
	"BilliardServer/Util/db/mongodb"
	"BilliardServer/Util/event"
	"BilliardServer/Util/stack"
	"BilliardServer/Util/tools"
	"errors"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Club struct {
	CollectionName  string         `bson:"-"`               //数据集名称
	FlagChange      bool           `bson:"-"`               //是否被修改
	ObjID           bson.ObjectId  `bson:"_id,omitempty"`   //唯一ID
	ClubID          uint32         `bson:"ClubID"`          //俱乐部id
	ClubName        string         `bson:"ClubName"`        //俱乐部名称
	ClubNotice      string         `bson:"ClubNotice"`      //俱乐部公告
	ClubLV          uint32         `bson:"ClubLV"`          //俱乐部等级
	ClubBadge       uint32         `bson:"ClubBadge"`       //俱乐部徽章
	ClubRate        uint32         `bson:"ClubRate"`        //俱乐部评级
	ProfitGold      uint64         `bson:"ProfitGold"`      //俱乐部周盈利(每周6清0)
	JoinLevel       uint32         `bson:"JoinLevel"`       //加入要求等级
	NumExp          uint32         `bson:"NumExp"`          //经验数量
	ClubScore       uint32         `bson:"ClubScore"`       //俱乐部本周评分
	TotalScore      uint32         `bson:"TotalScore"`      //俱乐部累计评分
	TimeCreate      string         `bson:"TimeCreate"`      //创建时间
	TimeUpdate      string         `bson:"TimeUpdate"`      //更新时间
	MaxNum          uint32         `bson:"MaxNum"`          //俱乐部最大人数
	Num             uint32         `bson:"Num"`             //俱乐部人数
	MasterEntityID  uint32         `bson:"MasterEntityID"`  //俱乐部部长
	IsOpen          bool           `bson:"IsOpen"`          //新人加入自动审批
	Members         []Member       `bson:"Members"`         //成员
	ReqList         []uint32       `bson:"ReqList"`         //申请加入列表
	RedEnvelopeList []RedEnvelope  `bson:"RedEnvelopeList"` //红包列表
	ShopList        []ClubShopItem `bson:"ShopList"`        //俱乐部商店列表
	ClubActiveValue uint32         `bson:"ClubActiveValue"` //俱乐部活跃值(每周6清0)
	LastWeekRank    uint32         `bson:"LastWeekRank"`    //上周同级别排名
}

type Member struct {
	ObjID       bson.ObjectId `bson:"ID"` //唯一ID
	EntityID    uint32        `bson:"EntityID"`
	AddTime     string        `bson:"AddTime"`  //加入时间
	Position    uint32        `bson:"Position"` //职别
	ActiveValue uint32        `bson:"ActiveValue"`
}

type RedEnvelope struct {
	RedEnvelopeID                 bson.ObjectId       `bson:"RedEnvelopeID"`
	ClubRedEnvelopeRecordList     []RedEnvelopeRecord `bson:"ClubRedEnvelopeRecordList"`     //领取记录
	SendCoinNum                   uint32              `bson:"SendCoinNum"`                   //发送金额
	SendTime                      int64               `bson:"SendTime"`                      //发送时间
	TotalSendNum                  uint32              `bson:"TotalSendNum"`                  //总发送个数
	NumDelivered                  uint32              `bson:"NumDelivered"`                  //已拆包数量
	AmountDelivered               uint32              `bson:"AmountDelivered"`               //已发放金额
	SendEnvelopeEntityID          uint32              `bson:"SendEnvelopeEntityID"`          //发送人ID
	SendEnvelopeEntityName        string              `bson:"SendEnvelopeEntityName"`        //发送红包人名字
	SendEnvelopeEntityAvatarID    uint32              `bson:"SendEnvelopeEntityAvatarID"`    //发送红包人头像ID
	SendEnvelopeEntityIconFrameID uint32              `bson:"SendEnvelopeEntityIconFrameID"` //发送红包人头像框ID
	BlessWorld                    string              `bson:"BlessWorld"`                    //祝福语
}

type RedEnvelopeRecord struct {
	EntityID   uint32 `bson:"EntityID"`
	EntityName string `bson:"EntityName"` //领取人名字
	GetCoinNum uint32 `bson:"GetCoinNum"` //领取金额
	GetTime    int64  `bson:"GetTime"`    //领取时间
}

type ClubFundItem struct {
	FundID                uint32 `bson:"FundID"`
	FundEntityID          uint32 `bson:"FundEntityID"`          //发送人ID
	FundEntityName        string `bson:"FundEntityName"`        //发送红包人名字
	FundEntitySex         uint32 `bson:"FundEntitySex"`         //发送红包人性别
	FundEntityAvatarID    uint32 `bson:"FundEntityAvatarID"`    //发送红包人等级ID
	FundEntityIconFrameID uint32 `bson:"FundEntityIconFrameID"` //发送红包人头像框ID
	Contribution          uint32 `bson:"Contribution"`          //贡献
}

// 数据实体 物品
type ClubShopItem struct {
	ItemID    uint32 `bson:"ItemID"`    //商店ItemID
	TableID   uint32 `bson:"TableID"`   //商店配置表ID
	MaxBuyNum uint32 `bson:"MaxBuyNum"` //购买次数
	AddTime   string `bson:"AddTime"`   //添加时间
	Sort      int    `bson:"Sort"`      //排序
	Unlock    uint32 `bson:"Unlock"`    //解锁条件
}

func (this *Club) InitByFirst(collectionName string, tEntityID uint32) {
	this.CollectionName = collectionName
	this.FlagChange = false
	this.ObjID = bson.NewObjectId()
	this.ClubID = 0
	this.ClubName = ""
	this.ClubBadge = 0
	this.ClubNotice = ""
	this.ClubLV = 1
	this.ClubRate = 1
	this.JoinLevel = 0
	this.NumExp = 0
	this.TimeCreate = tools.GetTimeByTimeStamp(time.Now().Unix())
	this.TimeUpdate = this.TimeCreate
	this.MaxNum = 30
	this.Num = 0
	this.MasterEntityID = tEntityID
	this.IsOpen = false
	this.Members = make([]Member, 0)
	this.ClubActiveValue = 0
	this.LastWeekRank = 0
	this.ShopList = make([]ClubShopItem, 0)
}

// 获取ObjID
func (this *Club) GetObjID() string {
	return this.ObjID.String()
}

// 获取id
func (this *Club) GetEntityID() uint32 {
	return this.ClubID
}

// 设置DBConnect
func (this *Club) SetDBConnect(collectionName string) {
	this.CollectionName = collectionName
}

// 初始化 by数据结构
func (this *Club) InitByData(playerData interface{}) {
	stack.SimpleCopyProperties(this, playerData)
}

// 初始化 by数据库
func (this *Club) InitFormDB(clubID uint32, tDBConnect *mongodb.DBConnect) (bool, error) {
	if tDBConnect == nil {
		return false, errors.New("tDBConnect == nil")
	}
	err := tDBConnect.GetData(this.CollectionName, "ClubID", clubID, this)
	if err != nil {
		return false, err
	}

	return true, err
}

// 插入数据库
func (this *Club) InsertEntity(tDBConnect *mongodb.DBConnect) error {
	if tDBConnect == nil {
		return nil
	}
	return tDBConnect.InsertData(this.CollectionName, this)
}

// 保存致数据库
func (this *Club) SaveEntity(tDBConnect *mongodb.DBConnect) {
	if tDBConnect == nil {
		return
	}
	tDBConnect.SaveData(this.CollectionName, "_id", this.ObjID, this)
}

// 清理实体
func (this *Club) ClearEntity() {
	this.CollectionName = ""
}

// 同步实体
// typeSave: 0定时同步 1根据环境默认 2立即同步
func (this *Club) SyncEntity(typeSave uint32) {
	evEntity := new(EntityEvent)
	evEntity.TypeSave = typeSave
	evEntity.TypeEntity = EntityTypeClub
	evEntity.Entity = this
	event.Emit(UnitSyncentity, evEntity)
}

func (this *Club) IsMasterEntityID(entityID uint32) bool {
	return this.MasterEntityID == entityID
}

func (this *Club) SetNewMaster(EntityID uint32) {
	this.MasterEntityID = EntityID
}

func (this *Club) IsMember(EntityID uint32) *Member {
	for _, vl := range this.Members {
		if vl.EntityID == EntityID {
			return &vl
		}
	}
	return nil
}

func (this *Club) MemberPosition(EntityID uint32) uint32 {
	member := this.IsMember(EntityID)
	if member == nil {
		return 0
	}
	return member.Position
}

func (this *Club) AddMembers(tEntityID uint32, position uint32) {
	member := new(Member)
	member.ObjID = bson.NewObjectId()
	member.EntityID = tEntityID
	member.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
	member.Position = position
	this.Members = append(this.Members, *member)

	this.Num = uint32(len(this.Members))
}

func (this *Club) SetMemberPosition(EntityMap map[uint32]uint32) {
	for _, vl := range this.Members {
		if p, ok := EntityMap[vl.EntityID]; ok {
			vl.Position = p
		}
	}
}

func (this *Club) GetMembers() []Member {
	return this.Members
}

func (this *Club) ClearReqList() {
	this.ReqList = nil
}

func (this *Club) ReMoveMember(EntityID uint32) {
	for key, vl := range this.Members {
		if vl.EntityID == EntityID {
			this.Members = append(this.Members[:key], this.Members[(key+1):]...)
			break
		}
	}
	this.Num = uint32(len(this.Members))
}

func (this *Club) RemoveOneReqList(EntityID uint32) {
	for key, vl := range this.ReqList {
		if EntityID == vl {
			this.ReqList = append(this.ReqList[:key], this.ReqList[(key+1):]...)
			break
		}
	}
}

func (this *Club) RemoveReqList(EntityID []uint32) {
	list := make([]uint32, 0)
	for _, id := range EntityID {
		for _, vl := range this.ReqList {
			if id != vl {
				list = append(list, vl)
			}
		}
	}
	this.ReqList = list
}

func (this *Club) AddReqList(EntityID uint32) {
	this.ReqList = append(this.ReqList, EntityID)
}

func (this *Club) LessClubMaxNum() bool {
	return this.MaxNum > this.Num
}

func (this *Club) TotalPosition(position uint32) uint32 {
	total := uint32(0)
	for _, vl := range this.Members {
		if vl.Position == position {
			total++
		}
	}
	return total
}

func (this *Club) IsInReqList(EntityID uint32) bool {
	for _, vl := range this.ReqList {
		if vl == EntityID {
			return true
		}
	}
	return false
}

// 通过红包ID获取红包数据
func (this *Club) GetRedEnvelopeFromRedEnvelopeID(redEnvelopeID string) *RedEnvelope {
	for _, vl := range this.RedEnvelopeList {
		if vl.RedEnvelopeID.Hex() == redEnvelopeID {
			return &vl
		}
	}
	return nil
}

// 通过红包ID获取红包记录
func (this *Club) GetRedEnvelopeRecordFromRedEnvelopeID(redEnvelopeID string) ([]RedEnvelopeRecord, error) {
	redEnvelope := this.GetRedEnvelopeFromRedEnvelopeID(redEnvelopeID)
	if redEnvelope != nil {
		return redEnvelope.ClubRedEnvelopeRecordList, nil
	}
	return nil, errors.New("获取红包记录失败")
}

// 判断是否领取过红包
func (this *Club) CheckHadOpenRedEnvelopeByRedEnvelopeID(redEnvelopeID string, EntityID uint32) (bool, error) {
	redEnvelope := this.GetRedEnvelopeFromRedEnvelopeID(redEnvelopeID)
	if redEnvelope == nil {
		return false, errors.New("获取红包数据失败")
	}
	return this.CheckHadOpenRedEnvelopeByRedEnvelope(redEnvelope, EntityID), nil
}

// 判断是否领取过红包
func (this *Club) CheckHadOpenRedEnvelopeByRedEnvelope(redEnvelope *RedEnvelope, EntityID uint32) bool {
	for _, redEnvelopeRecord := range redEnvelope.ClubRedEnvelopeRecordList {
		if redEnvelopeRecord.EntityID == EntityID {
			return true
		}
	}
	return false
}

// 保存红包数据
func (this *Club) SaveClubRedEnvelope(redEnvelope *RedEnvelope) {
	for k, vl := range this.RedEnvelopeList {
		if vl.RedEnvelopeID == redEnvelope.RedEnvelopeID {
			this.RedEnvelopeList[k] = *redEnvelope
		}
	}
}

func (this *Club) GetClubActiveValue() uint32 {
	return this.ClubActiveValue
}

func (this *Club) AddClubActiveValueNumExp(value uint32) {
	this.NumExp += value
	this.ClubActiveValue += value
}

func (this *Club) GetClubNumExp() uint32 {
	return this.NumExp
}

func (this *Club) UpgradeClub(maxNum uint32) {
	this.ClubLV += 1
	this.MaxNum = maxNum
}

func (this *Club) AddClubScore(score uint32) {
	this.ClubScore += score
	this.TotalScore += score
}

func (this *Club) AddProfitGold(gold uint32) {
	this.ProfitGold += uint64(gold)
}

func (this *Club) UpdateMemberActive(EntityID, value uint32) {
	for k, vl := range this.Members {
		if vl.EntityID == EntityID {
			v := vl
			v.ActiveValue += value
			v.AddTime = tools.GetTimeByTimeStamp(time.Now().Unix())
			this.Members[k] = v
		}
	}
}

func (this *Club) UpgradeClubFromScore(rate, rank uint32) {
	this.ClubRate = rate
	this.LastWeekRank = rank + 1
}

func (this *Club) ReSetClubScore() {
	this.ClubScore = 0
}

func (this *Club) ReSetClubActiveValue() {
	this.ClubActiveValue = 0
}

func (this *Club) ReSeProfitGold() {
	this.ProfitGold = 0
}

func (this *Club) ResetClubMember(promoteEliteActive uint32) {
	for k, vl := range this.Members {
		m := vl
		if m.Position == conf.General && m.ActiveValue >= promoteEliteActive {
			m.Position = conf.Elite
		} else if m.Position == conf.Elite && m.ActiveValue < promoteEliteActive {
			m.Position = conf.General
		}
		m.ActiveValue = 0
		this.Members[k] = m
	}
}

func (this *Club) GetClubShopItem(itemID uint32) *ClubShopItem {
	for _, val := range this.ShopList {
		if val.ItemID == itemID {
			return &val
		}
	}
	return nil
}
