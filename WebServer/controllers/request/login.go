package request

type Register struct {
	UserName      string `json:"user_name"`
	IsIPhone      bool   `json:"is_iphone"`
	Platform      uint32 `json:"platform"`       //平台
	LoginPlatform uint32 `json:"login_platform"` //登录平台
	Channel       uint32 `json:"channel"`        //渠道
	DeviceId      string `json:"device_id"`      //设备Id
	Machine       string `json:"machine"`        //机型
	RemoteAddr    string `json:"remote_addr"`    //远端ip
	PackageName   string `json:"package_name"`   //当前客户端包名
	Language      uint32 `json:"language"`       //玩家的语言标记
}
