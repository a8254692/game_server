
//-----------------------------------------------
//              生成代码不要修改
//-----------------------------------------------

package table

type AchievementElementCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //成就名字
    Icon string `json:"Icon"`    //成就图标
    Condition []string `json:"Condition"`    //成就条件
    Score uint32 `json:"Score"`    //成就积分
}
type AchievementLvCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    Score uint32 `json:"Score"`    //积分
    Lv string `json:"Lv"`    //奖励等级
    Reward [][]uint32 `json:"Reward"`    //奖励
}
type AchievementCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Child []uint32 `json:"Child"`    //成就字元素
    Jump []uint32 `json:"Jump"`    //跳转界面
    TypeN uint32 `json:"TypeN"`    //类型
}
type AudioCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Pass uint32 `json:"Pass"`    //音效通道
    Loop uint32 `json:"Loop"`    //是否循环
    Voice uint32 `json:"Voice"`    //音量
    Path string `json:"Path"`    //路径
}
type BattleClubProgressRewardCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    TypeN uint32 `json:"TypeN"`    //类型
    RoomType uint32 `json:"RoomType"`    //房间类型
    WinReward uint32 `json:"WinReward"`    //赢的奖励
    LoseReward uint32 `json:"LoseReward"`    //输的奖励
}
type BattleClubScoreRewardCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    TypeN uint32 `json:"TypeN"`    //类型
    RoomType uint32 `json:"RoomType"`    //房间类型
    WinScore uint32 `json:"WinScore"`    //赢的奖励
    LoseScore uint32 `json:"LoseScore"`    //输的奖励
}
type BattleEffectCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    TypeN uint32 `json:"TypeN"`    //类型
    ConditionID uint32 `json:"ConditionID"`    //条件ID
    Name string `json:"Name"`    //效果名称
    Icon string `json:"Icon"`    //效果图标
}
type BoxCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    TypeN uint32 `json:"TypeN"`    //类型
    Weight uint32 `json:"Weight"`    //获得权重
    Icon string `json:"Icon"`    //图标
    Interval uint32 `json:"Interval"`    //宝箱加速间隔
    LifeTime uint32 `json:"LifeTime"`    //开箱需要时间
    Cost []uint32 `json:"Cost"`    //宝箱开启消耗
    Desc string `json:"Desc"`    //宝箱描述
    FixedReward [][]uint32 `json:"FixedReward"`    //固定奖励
    RandomReward []uint32 `json:"RandomReward"`    //随机奖励
    SecretReward []uint32 `json:"SecretReward"`    //神秘奖励
}
type ChatDescCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Content string `json:"Content"`    //表情
    AudioID uint32 `json:"AudioID"`    //音效ID
}
type ClothingCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    TypeN uint32 `json:"TypeN"`    //类型
    Icon string `json:"Icon"`    //图标
    QuaIcon string `json:"QuaIcon"`    //品质
    LifeTime uint32 `json:"LifeTime"`    //可使用时间
    Desc string `json:"Desc"`    //描述
    GetWay uint32 `json:"GetWay"`    //获取途径
}
type ClubDailySignCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    TypeN uint32 `json:"TypeN"`    //类型
    Reward uint32 `json:"Reward"`    //活跃值
}
type ClubProgressRewardCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Progress uint32 `json:"Progress"`    //进度值
    Rewards []uint32 `json:"Rewards"`    //奖励
}
type ClubRateRewardCfg struct {
    Level uint32 `json:"Level"`    //评级
    LevelSymbol string `json:"LevelSymbol"`    //评级符号
    Reward []uint32 `json:"Reward"`    //奖励俱乐部币
    Upgrade uint32 `json:"Upgrade"`    //晋升评级分数
    KeepGrade uint32 `json:"KeepGrade"`    //保持评级分数
    ItemReward [][]uint32 `json:"ItemReward"`    //道具奖励
}
type ClubShopCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    ItemTableID uint32 `json:"ItemTableID"`    //道具ID
    MaxBuyNum uint32 `json:"MaxBuyNum"`    //最多购买数量
    Unlock uint32 `json:"Unlock"`    //解锁条件
    PriceType uint32 `json:"PriceType"`    //货币类型
    Price uint32 `json:"Price"`    //售价
    Sort uint32 `json:"Sort"`    //排序
}
type ClubTaskProgressCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Progress uint32 `json:"Progress"`    //进度值
    Rewards []uint32 `json:"Rewards"`    //奖励
}
type ClubTaskCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    WeekCondition []uint32 `json:"WeekCondition"`    //周任务条件
    DayCondition []uint32 `json:"DayCondition"`    //天任务条件
    Score uint32 `json:"Score"`    //积分
    Jump string `json:"Jump"`    //跳转界面
}
type ClubCfg struct {
    Level uint32 `json:"Level"`    //等级
    Exp uint32 `json:"Exp"`    //升到下一等级需要经验
    Num uint32 `json:"Num"`    //人数
    MasterNum uint32 `json:"MasterNum"`    //俱乐部副部长人数
}
type CollectCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    Conditions []uint32 `json:"Conditions"`    //达成条件
    Icon string `json:"Icon"`    //图标
}
type ConditionalCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //条件名
    TypeN uint32 `json:"TypeN"`    //类型
}
type ConstImageCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    ImageCN string `json:"ImageCN"`    //中文
    ImageEN string `json:"ImageEN"`    //English
}
type ConstTextCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    TextCN string `json:"TextCN"`    //中文
    TextEN string `json:"TextEN"`    //English
}
type ConstCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Paramater1 uint32 `json:"Paramater1"`    //参数1
    Paramater10 []uint32 `json:"Paramater10"`    //参数10
    Paramater20 [][]uint32 `json:"Paramater20"`    //参数20
}
type CueHandbookCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    CueQuality uint32 `json:"CueQuality"`    //球杆阶级
    CueID uint32 `json:"CueID"`    //球杆ID
    FixedReward [][]uint32 `json:"FixedReward"`    //解锁固定奖励
}
type CueCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    ShredNum uint32 `json:"ShredNum"`    //合成所需碎片数量
    NextID uint32 `json:"NextID"`    //下一级ID
    FullID uint32 `json:"FullID"`    //满级ID
    Height float32 `json:"Height"`    //高度
    TypeN uint32 `json:"TypeN"`    //类型
    BallArm string `json:"BallArm"`    //球杆模型
    InitQuality uint32 `json:"InitQuality"`    //初始品质
    MaxQuality uint32 `json:"MaxQuality"`    //最高品质
    UpgradeConst map[int]int `json:"UpgradeConst"`    //升阶（品质）升星需要物品
    Icon string `json:"Icon"`    //球杆图标
    Model string `json:"Model"`    //球杆预支
    QuaIcon string `json:"QuaIcon"`    //品质图标
    BgIcon string `json:"BgIcon"`    //背景品质
    Side uint32 `json:"Side"`    //加塞
    Force uint32 `json:"Force"`    //力度
    AimingLine uint32 `json:"AimingLine"`    //瞄准线
    CharmScore uint32 `json:"CharmScore"`    //魅力评分
    Price uint32 `json:"Price"`    //价格
    LifeTime uint32 `json:"LifeTime"`    //可使用时间
    Desc string `json:"Desc"`    //描述
}
type DailySignCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Day uint32 `json:"Day"`    //天数
    Rewards [][]uint32 `json:"Rewards"`    //奖励列表
    TypeN uint32 `json:"TypeN"`    //类型
}
type DanCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    UpgradeStar uint32 `json:"UpgradeStar"`    //升至下一级的累计星星
    Icon string `json:"Icon"`    //图标
    Lv uint32 `json:"Lv"`    //等级
}
type DressCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    Color string `json:"Color"`    //
    TypeN uint32 `json:"TypeN"`    //类型
    ModelPath string `json:"ModelPath"`    //模型
    Icon string `json:"Icon"`    //图标
    LifeTime uint32 `json:"LifeTime"`    //可使用时间
    Sex uint32 `json:"Sex"`    //性别
    QuailyIcon string `json:"QuailyIcon"`    //服装品质图标
    Desc string `json:"Desc"`    //描述
}
type EffectCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    ShopID uint32 `json:"ShopID"`    //对应商品表ID
    TypeN uint32 `json:"TypeN"`    //类型
    Model string `json:"Model"`    //模型
    Icon string `json:"Icon"`    //图标
    QuaIcon string `json:"QuaIcon"`    //品质图标
    BGQuaIcon string `json:"BGQuaIcon"`    //背景图标
    LifeTime uint32 `json:"LifeTime"`    //可使用时间
    Desc string `json:"Desc"`    //道具描述
}
type EightBallRoomCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    Level uint32 `json:"Level"`    //房间类型
    TableFee uint32 `json:"TableFee"`    //台费
    WinCoin uint32 `json:"WinCoin"`    //赢得金币数量
    MinCoin uint32 `json:"MinCoin"`    //进入最低金币
    MaxCoin uint32 `json:"MaxCoin"`    //进入最高金币
    WinExp uint32 `json:"WinExp"`    //胜利获得经验
    TransporterExp uint32 `json:"TransporterExp"`    //失败获得经验
}
type EmojiCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Emoji string `json:"Emoji"`    //表情
    Price string `json:"Price"`    //价格
    TypeN uint32 `json:"TypeN"`    //花费类型
    AudioID uint32 `json:"AudioID"`    //音效ID
}
type FirstRechargeCfg struct {
    TableID uint32 `json:"TableID"`    //名称
    RewardList [][]uint32 `json:"RewardList"`    //物品
    Dan uint32 `json:"Dan"`    //档位
    Price uint32 `json:"Price"`    //金额
}
type FreeStoreCfg struct {
    TableID uint32 `json:"TableID"`    //名称
    Product []uint32 `json:"Product"`    //商品
    BuyType uint32 `json:"BuyType"`    //货币类型
    RandomType uint32 `json:"RandomType"`    //随机标识
    Discount uint32 `json:"Discount"`    //折扣
    Price uint32 `json:"Price"`    //价格
}
type GuideCfg struct {
    TableID uint32 `json:"TableID"`    //等级
    TypeN uint32 `json:"TypeN"`    //引导组
    GuideType uint32 `json:"GuideType"`    //引导步骤
    Conditions []uint32 `json:"Conditions"`    //触发条件
    Panel string `json:"Panel"`    //引导所在窗口
}
type ItemTimeCfg struct {
    TableID uint32 `json:"TableID"`    //等级
    TypeN uint32 `json:"TypeN"`    //时限类型
    Time uint32 `json:"Time"`    //时限时间(秒)
}
type ItemCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    TypeN uint32 `json:"TypeN"`    //类型
    Icon string `json:"Icon"`    //图标
    LifeTime uint32 `json:"LifeTime"`    //可使用时间
    CueID uint32 `json:"CueID"`    //合成球杆ID
    Desc string `json:"Desc"`    //道具描述
    QuaIcon string `json:"QuaIcon"`    //道具品质图标
    JumpID uint32 `json:"JumpID"`    //使用跳转
    FixedReward [][]uint32 `json:"FixedReward"`    //固定奖励
    RandomReward [][]uint32 `json:"RandomReward"`    //随机奖励
}
type KingCfg struct {
    TableID uint32 `json:"TableID"`    //名称
    RewardID1 []uint32 `json:"RewardID1"`    //精英版奖励
    RewardID2 []uint32 `json:"RewardID2"`    //进阶版奖励
    Count []uint32 `json:"Count"`    //s
    TxtID uint32 `json:"TxtID"`    //匹配文本ID
}
type PlayerLevelCfg struct {
    Level uint32 `json:"Level"`    //等级
    Exp uint32 `json:"Exp"`    //升至下一级的累计经验
}
type PropertyItemCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    TypeN uint32 `json:"TypeN"`    //类型
    Icon string `json:"Icon"`    //图标
    Desc string `json:"Desc"`    //道具描述
    QuaIcon string `json:"QuaIcon"`    //道具品质图标
    JumpID uint32 `json:"JumpID"`    //使用跳转
}
type RandNameCfg struct {
    TableID string `json:"TableID"`    //主key
    SC string `json:"SC"`    //简体中文
    EN string `json:"EN"`    //英文
}
type RandPreNameCfg struct {
    TableID string `json:"TableID"`    //主key
    SC string `json:"SC"`    //简体中文
    EN string `json:"EN"`    //英文
}
type RoomCfg struct {
    TableID uint32 `json:"TableID"`    //房间ID
    Name string `json:"Name"`    //房间名
    WinExp uint32 `json:"WinExp"`    //赢一局获得的经验
    LoseExp uint32 `json:"LoseExp"`    //输一局获得的经验
}
type SeasonCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    NextSeasonID uint32 `json:"NextSeasonID"`    //下一个赛季ID
    StartTime uint32 `json:"StartTime"`    //开始时间
    EndTime uint32 `json:"EndTime"`    //结束时间
    AwardTime uint32 `json:"AwardTime"`    //可领奖时间
}
type ShopCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    ItemID uint32 `json:"ItemID"`    //道具ID
    TypeN uint32 `json:"TypeN"`    //类型
    SubType uint32 `json:"SubType"`    //子类型
    ShowStartTime int64 `json:"ShowStartTime"`    //展示开始时间
    ShowEndTime int64 `json:"ShowEndTime"`    //展示结束时间
    TokenType uint32 `json:"TokenType"`    //代币类型
    Price uint32 `json:"Price"`    //售价
    Discount uint32 `json:"Discount"`    //折扣
    Sex uint32 `json:"Sex"`    //性别
    Sort uint32 `json:"Sort"`    //排序
}
type SpecialShopCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Item []uint32 `json:"Item"`    //道具数量
    GiftNum []uint32 `json:"GiftNum"`    //赠送数量
    PayType uint32 `json:"PayType"`    //支付类型
    Price uint32 `json:"Price"`    //售价
    Icon string `json:"Icon"`    //图标
}
type TaskProgressCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Progress uint32 `json:"Progress"`    //进度值
    Rewards []uint32 `json:"Rewards"`    //奖励
}
type TaskCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Condition []uint32 `json:"Condition"`    //条件
    Rewards [][]uint32 `json:"Rewards"`    //奖励
    Jump uint32 `json:"Jump"`    //跳转界面
}
type TollgateExpertAdvancedCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    FinishCondition string `json:"FinishCondition"`    //技巧要点
    Rewards [][]uint32 `json:"Rewards"`    //奖励
}
type TollgateCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    Name string `json:"Name"`    //名称
    FinishCondition map[int]int `json:"FinishCondition"`    //完成条件
    Rewards [][]uint32 `json:"Rewards"`    //奖励
}
type ViewCfg struct {
    TableID uint32 `json:"TableID"`    //ID
    TypeN uint32 `json:"TypeN"`    //类型
    UIName string `json:"UIName"`    //UI名称
    Condition []uint32 `json:"Condition"`    //类型
    Text string `json:"Text"`    //提示文本
    Sound uint32 `json:"Sound"`    //背景音乐
    Display uint32 `json:"Display"`    //未开启时是否显示
    NeedShow bool `json:"NeedShow"`    //是否弹新功能提示框
}
type VipCfg struct {
    Level uint32 `json:"Level"`    //等级
    Name string `json:"Name"`    //名称
    Exp uint32 `json:"Exp"`    //升至下一级的累计经验
    Reward [][]uint32 `json:"Reward"`    //限购礼包
    Price []uint32 `json:"Price"`    //购买价格
    Box [][]uint32 `json:"Box"`    //每日礼包
    Desc string `json:"Desc"`    //描述
}
