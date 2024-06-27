package controllers

import (
	"BilliardServer/Common/entity"
	"BilliardServer/Util/jwt"
	"BilliardServer/Util/tools"
	"BilliardServer/WebServer/controllers/response"
	"BilliardServer/WebServer/initialize/consts"
	"BilliardServer/WebServer/models"
	"BilliardServer/WebServer/utils"
	"BilliardServer/WebServer/utils/apple_login"
	"context"
	"errors"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	bectx "github.com/beego/beego/v2/server/web/context"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

func OnMongoDBInItComplete() int {
	serverZone, _ := web.AppConfig.Int("mongodb::serverzone")
	serverId, _ := web.AppConfig.Int("mongodb::serverid")
	entityId, _ := web.AppConfig.Int("mongodb::entityid")
	AccEntityIdNow := serverZone*entityId + serverId*(entityId/1000)

	//获得当前数据加中的最新EntityId
	tEntityMainType := entity.UnitAcc
	entityCount, err := utils.GameDB.GetDataCountTotal(tEntityMainType)
	if err != nil {
		logs.Warning("-->WebServer----OnMongoDBInItComplete----Err:", err.Error())
		return AccEntityIdNow
	}

	return AccEntityIdNow + entityCount
}

func registerAcc(accUnique string, password string, isIPhone bool, platform uint32, loginPlatform uint32, channel uint32, deviceId string, machine string, remoteAddr string, packageName string, language uint32) (*entity.EntityAcc, error) {
	if accUnique == "" {
		return nil, errors.New("acc is empty")
	}

	accIdNow := OnMongoDBInItComplete()
	if accIdNow <= 0 {
		return nil, errors.New("acc Id now is empty")
	}

	tEntityAcc := new(entity.EntityAcc)

	query := bson.M{"AccUnique": accUnique, "LoginPlatform": loginPlatform, "Channel": channel}
	err := utils.GameDB.GetOne(entity.UnitAcc, query, nil, tEntityAcc)
	if err != nil && !errors.Is(err, mgo.ErrNotFound) {
		return nil, err
	}

	if tEntityAcc.EntityID > 0 {
		return nil, errors.New("acc is register")
	}

	tEntityID32 := int32(accIdNow)
	tEntityID := tools.GetEntityID(&tEntityID32)
	tEntityAcc.InitByFirst(entity.UnitAcc, uint32(tEntityID))
	tEntityAcc.AccUnique = accUnique
	tEntityAcc.PassWord = tools.MD5(password)
	tEntityAcc.IsIPhone = isIPhone
	tEntityAcc.Platform = platform
	tEntityAcc.LoginPlatform = loginPlatform
	tEntityAcc.Channel = channel
	tEntityAcc.DeviceId = deviceId
	tEntityAcc.Machine = machine
	tEntityAcc.RemoteAddr = remoteAddr
	tEntityAcc.PackageName = packageName
	tEntityAcc.Language = language

	err = tEntityAcc.InsertEntity(utils.GameDB)
	if err != nil {
		return nil, err
	}

	//增加注册日志
	log := models.UserCreateLog{
		Time:          time.Now().Unix(),
		Account:       accUnique,
		EntityID:      uint32(tEntityID),
		IsIPhone:      isIPhone,
		Platform:      platform,
		LoginPlatform: loginPlatform,
		Channel:       channel,
		DeviceId:      deviceId,
		Machine:       machine,
		RemoteAddr:    remoteAddr,
		PackageName:   packageName,
		Language:      language,
	}
	models.AddCreateAccountLog(&log)

	return tEntityAcc, nil
}

//func Register(ctx *bectx.Context) {
//	req := &request.Register{}
//	err := json.Unmarshal(ctx.Input.RequestBody, req)
//	if err != nil {
//		_ = ctx.JSONResp(utils.Result(1, nil, "json unmarshal err"))
//		return
//	}
//
//	if req.UserName == "" {
//		_ = ctx.JSONResp(utils.Result(2, nil, "username is nil"))
//		return
//	}
//
//	if req.LoginPlatform <= 0 {
//		_ = ctx.JSONResp(utils.Result(3, nil, "login platform is nil"))
//		return
//	}
//
//	if req.Channel <= 0 {
//		_ = ctx.JSONResp(utils.Result(4, nil, "username is nil"))
//		return
//	}
//
//	entityInfo, err := registerAcc(req.UserName, req.IsIPhone, req.Platform, req.LoginPlatform, req.Channel, req.DeviceId, req.Machine, ctx.Request.RemoteAddr, req.PackageName, req.Language)
//	if err != nil {
//		logs.Warning("-->controllers--Register--registerAcc--err--", err.Error())
//		_ = ctx.JSONResp(utils.Result(5, nil, "insert err"))
//		return
//	}
//
//	if entityInfo == nil || entityInfo.EntityID <= 0 {
//		logs.Warning("-->controllers--Register--registerAcc--entityID is nil")
//		_ = ctx.JSONResp(utils.Result(6, nil, "insert entityID is nil"))
//		return
//	}
//
//	_ = ctx.JSONResp(utils.Result(0, nil, "OK"))
//	return
//}

func Login(ctx *bectx.Context) {
	userName := ctx.Request.FormValue("user_name")
	password := ctx.Request.FormValue("password")
	isIPhone := ctx.Request.FormValue("is_iphone")
	platform := ctx.Request.FormValue("platform")
	loginPlatform := ctx.Request.FormValue("login_platform")
	channel := ctx.Request.FormValue("channel")
	deviceId := ctx.Request.FormValue("device_id")
	machine := ctx.Request.FormValue("machine")
	packageName := ctx.Request.FormValue("package_name")
	language := ctx.Request.FormValue("language")

	if userName == "" {
		_ = ctx.JSONResp(utils.Result(1, nil, "username is empty"))
		return
	}

	if password == "" {
		_ = ctx.JSONResp(utils.Result(1, nil, "password is empty"))
		return
	}

	isIPhoneBool, _ := strconv.ParseBool(isIPhone)
	platformInt, _ := strconv.Atoi(platform)
	loginPlatformInt, _ := strconv.Atoi(loginPlatform)
	channelInt, _ := strconv.Atoi(channel)
	languageInt, _ := strconv.Atoi(language)

	if loginPlatformInt <= 0 {
		_ = ctx.JSONResp(utils.Result(2, nil, "login platform is empty"))
		return
	}

	if channelInt <= 0 {
		_ = ctx.JSONResp(utils.Result(3, nil, "channel is empty"))
		return
	}

	tEntityAcc := new(entity.EntityAcc)
	query := bson.M{"AccUnique": userName, "LoginPlatform": loginPlatformInt, "Channel": channelInt}
	err := utils.GameDB.GetOne(entity.UnitAcc, query, nil, tEntityAcc)
	if err != nil && !errors.Is(err, mgo.ErrNotFound) {
		_ = ctx.JSONResp(utils.Result(4, nil, "Err"))
		return
	}

	//查看是否不存在数据
	if tEntityAcc.EntityID <= 0 || errors.Is(err, mgo.ErrNotFound) {
		tEntityAcc, err = registerAcc(userName, password, isIPhoneBool, uint32(platformInt), uint32(loginPlatformInt), uint32(channelInt), deviceId, machine, ctx.Request.RemoteAddr, packageName, uint32(languageInt))
		if err != nil {
			logs.Warning("-->controllers--Login--registerAcc--err--", err.Error())
			_ = ctx.JSONResp(utils.Result(5, nil, "insert err"))
			return
		}
	} else {
		//判断密码
		if tEntityAcc.PassWord == "" {
			logs.Warning("-->controllers--Login--tEntityAcc.PassWord is nil")
			_ = ctx.JSONResp(utils.Result(6, nil, "password err"))
			return
		}

		inPwd := tools.MD5(password)
		if tEntityAcc.PassWord != inPwd {
			logs.Warning("-->controllers--Login--tEntityAcc.PassWord != inPwd")
			_ = ctx.JSONResp(utils.Result(7, nil, "password err1"))
			return
		}
	}

	//需要再次校验账号数据
	if tEntityAcc == nil || tEntityAcc.EntityID <= 0 {
		logs.Warning("-->controllers--Login--tEntityAcc is nil")
		_ = ctx.JSONResp(utils.Result(8, nil, "tEntityAcc is nil"))
		return
	}

	if tEntityAcc.State == consts.USER_STATUS_BAN_ACC || tEntityAcc.State == consts.USER_STATUS_BAN_IP {
		_ = ctx.JSONResp(utils.Result(9, nil, "Err"))
		return
	}

	token, err := jwt.GenToken(tEntityAcc.EntityID)
	if err != nil {
		_ = ctx.JSONResp(utils.Result(10, nil, "Err"))
		return
	}

	portGateSocket, _ := web.AppConfig.String("gatesocket::port")
	portGateServer, _ := web.AppConfig.String("gatewsserver::port")
	ipGateServer, _ := web.AppConfig.String("gatewsserver::ip")

	//增加登录日志
	log := models.UserLoginLog{
		Time:          time.Now().Unix(),
		Account:       userName,
		EntityID:      tEntityAcc.EntityID,
		IsIPhone:      isIPhoneBool,
		Platform:      uint32(platformInt),
		LoginPlatform: uint32(loginPlatformInt),
		Channel:       uint32(channelInt),
		DeviceId:      deviceId,
		Machine:       machine,
		RemoteAddr:    ctx.Request.RemoteAddr,
		PackageName:   packageName,
		Language:      uint32(languageInt),
	}
	models.AddLoginLog(&log)

	resp := response.Login{
		Token:         token,
		EntityID:      tEntityAcc.EntityID,
		GateAdr:       ipGateServer + ":" + portGateServer,
		GateSocketAdr: ipGateServer + ":" + portGateSocket,
	}

	_ = ctx.JSONResp(utils.Result(0, resp, "OK"))
	return
}

func GoogleLogin(ctx *bectx.Context) {
	idToken := ctx.Request.FormValue("id_token")
	if idToken == "" {
		_ = ctx.JSONResp(utils.Result(1, nil, "id_token is nil"))
		return
	}

	isIPhone := ctx.Request.FormValue("is_iphone")
	platform := ctx.Request.FormValue("platform")
	loginPlatform := ctx.Request.FormValue("login_platform")
	channel := ctx.Request.FormValue("channel")
	deviceId := ctx.Request.FormValue("device_id")
	machine := ctx.Request.FormValue("machine")
	packageName := ctx.Request.FormValue("package_name")
	language := ctx.Request.FormValue("language")

	isIPhoneBool, _ := strconv.ParseBool(isIPhone)
	platformInt, _ := strconv.Atoi(platform)
	loginPlatformInt, _ := strconv.Atoi(loginPlatform)
	channelInt, _ := strconv.Atoi(channel)
	languageInt, _ := strconv.Atoi(language)

	if loginPlatformInt <= 0 {
		_ = ctx.JSONResp(utils.Result(2, nil, "login platform is empty"))
		return
	}

	if channelInt <= 0 {
		_ = ctx.JSONResp(utils.Result(3, nil, "channel is empty"))
		return
	}

	oauthService, err := oauth2.NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		logs.Warning("-->login--GoogleLogin--NewService--err--", err.Error())
		_ = ctx.JSONResp(utils.Result(4, nil, "new oauth service err"))
		return
	}
	tokenInfoCall := oauthService.Tokeninfo()
	tokenInfoCall.IdToken(idToken)
	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		logs.Warning("-->login--GoogleLogin--tokenInfoCall--err--", err.Error())
		_ = ctx.JSONResp(utils.Result(5, nil, "token info call err"))
		return
	}
	if tokenInfo.UserId == "" {
		_ = ctx.JSONResp(utils.Result(6, nil, "UserId is nil"))
		return
	}

	tEntityAcc := new(entity.EntityAcc)
	query := bson.M{"AccUnique": tokenInfo.UserId, "LoginPlatform": loginPlatformInt, "Channel": channelInt}
	err = utils.GameDB.GetOne(entity.UnitAcc, query, nil, tEntityAcc)
	if err != nil && !errors.Is(err, mgo.ErrNotFound) {
		_ = ctx.JSONResp(utils.Result(7, nil, "Err"))
		return
	}

	//查看是否不存在数据
	if tEntityAcc.EntityID <= 0 || errors.Is(err, mgo.ErrNotFound) {
		tEntityAcc, err = registerAcc(tokenInfo.UserId, "", isIPhoneBool, uint32(platformInt), uint32(loginPlatformInt), uint32(channelInt), deviceId, machine, ctx.Request.RemoteAddr, packageName, uint32(languageInt))
		if err != nil {
			logs.Warning("-->controllers--Login--registerAcc--err--", err.Error())
			_ = ctx.JSONResp(utils.Result(8, nil, "insert err"))
			return
		}
	}

	//需要再次校验账号数据
	if tEntityAcc == nil || tEntityAcc.EntityID <= 0 {
		logs.Warning("-->controllers--Login--tEntityAcc is nil")
		_ = ctx.JSONResp(utils.Result(9, nil, "tEntityAcc is nil"))
		return
	}

	if tEntityAcc.State == consts.USER_STATUS_BAN_ACC || tEntityAcc.State == consts.USER_STATUS_BAN_IP {
		_ = ctx.JSONResp(utils.Result(10, nil, "Err"))
		return
	}

	token, err := jwt.GenToken(tEntityAcc.EntityID)
	if err != nil {
		_ = ctx.JSONResp(utils.Result(11, nil, "Err"))
		return
	}

	portGateServer, _ := web.AppConfig.String("gatewsserver::port")
	ipGateServer, _ := web.AppConfig.String("gatewsserver::ip")

	//增加登录日志
	log := models.UserLoginLog{
		Time:          time.Now().Unix(),
		Account:       tokenInfo.UserId,
		EntityID:      tEntityAcc.EntityID,
		IsIPhone:      isIPhoneBool,
		Platform:      uint32(platformInt),
		LoginPlatform: uint32(loginPlatformInt),
		Channel:       uint32(channelInt),
		DeviceId:      deviceId,
		Machine:       machine,
		RemoteAddr:    ctx.Request.RemoteAddr,
		PackageName:   packageName,
		Language:      uint32(languageInt),
	}
	models.AddLoginLog(&log)

	resp := response.Login{
		Token:    token,
		EntityID: tEntityAcc.EntityID,
		GateAdr:  ipGateServer + ":" + portGateServer,
	}

	_ = ctx.JSONResp(utils.Result(0, resp, "OK"))
	return
}

func AppleLogin(ctx *bectx.Context) {
	idToken := ctx.Request.FormValue("id_token")
	if idToken == "" {
		_ = ctx.JSONResp(utils.Result(1, nil, "id_token is nil"))
		return
	}

	isIPhone := ctx.Request.FormValue("is_iphone")
	platform := ctx.Request.FormValue("platform")
	loginPlatform := ctx.Request.FormValue("login_platform")
	channel := ctx.Request.FormValue("channel")
	deviceId := ctx.Request.FormValue("device_id")
	machine := ctx.Request.FormValue("machine")
	packageName := ctx.Request.FormValue("package_name")
	language := ctx.Request.FormValue("language")

	isIPhoneBool, _ := strconv.ParseBool(isIPhone)
	platformInt, _ := strconv.Atoi(platform)
	loginPlatformInt, _ := strconv.Atoi(loginPlatform)
	channelInt, _ := strconv.Atoi(channel)
	languageInt, _ := strconv.Atoi(language)

	if loginPlatformInt <= 0 {
		_ = ctx.JSONResp(utils.Result(2, nil, "login platform is empty"))
		return
	}

	if channelInt <= 0 {
		_ = ctx.JSONResp(utils.Result(3, nil, "channel is empty"))
		return
	}

	oauthService := apple_login.NewAppleLogin("")
	claims, err := oauthService.VerifyIdToken(idToken)
	if err != nil {
		logs.Warning("-->login--AppleLogin--VerifyIdToken--err--", err.Error())
		_ = ctx.JSONResp(utils.Result(4, nil, "get claims err"))
		return
	}

	sub, err := claims.GetSubject()
	if err != nil {
		logs.Warning("-->login--AppleLogin--GetSubject--err--", err.Error())
		_ = ctx.JSONResp(utils.Result(5, nil, "get subject err"))
		return
	}

	if sub == "" {
		_ = ctx.JSONResp(utils.Result(6, nil, "sub is nil"))
		return
	}

	tEntityAcc := new(entity.EntityAcc)
	query := bson.M{"AccUnique": sub, "LoginPlatform": loginPlatformInt, "Channel": channelInt}
	err = utils.GameDB.GetOne(entity.UnitAcc, query, nil, tEntityAcc)
	if err != nil && !errors.Is(err, mgo.ErrNotFound) {
		_ = ctx.JSONResp(utils.Result(7, nil, "Err"))
		return
	}

	//查看是否不存在数据
	if tEntityAcc.EntityID <= 0 || errors.Is(err, mgo.ErrNotFound) {
		tEntityAcc, err = registerAcc(sub, "", isIPhoneBool, uint32(platformInt), uint32(loginPlatformInt), uint32(channelInt), deviceId, machine, ctx.Request.RemoteAddr, packageName, uint32(languageInt))
		if err != nil {
			logs.Warning("-->controllers--Login--registerAcc--err--", err.Error())
			_ = ctx.JSONResp(utils.Result(8, nil, "insert err"))
			return
		}
	}

	//需要再次校验账号数据
	if tEntityAcc == nil || tEntityAcc.EntityID <= 0 {
		logs.Warning("-->controllers--Login--tEntityAcc is nil")
		_ = ctx.JSONResp(utils.Result(9, nil, "tEntityAcc is nil"))
		return
	}

	if tEntityAcc.State == consts.USER_STATUS_BAN_ACC || tEntityAcc.State == consts.USER_STATUS_BAN_IP {
		_ = ctx.JSONResp(utils.Result(10, nil, "Err"))
		return
	}

	token, err := jwt.GenToken(tEntityAcc.EntityID)
	if err != nil {
		_ = ctx.JSONResp(utils.Result(11, nil, "Err"))
		return
	}

	portGateServer, _ := web.AppConfig.String("gatewsserver::port")
	ipGateServer, _ := web.AppConfig.String("gatewsserver::ip")

	//增加登录日志
	log := models.UserLoginLog{
		Time:          time.Now().Unix(),
		Account:       sub,
		EntityID:      tEntityAcc.EntityID,
		IsIPhone:      isIPhoneBool,
		Platform:      uint32(platformInt),
		LoginPlatform: uint32(loginPlatformInt),
		Channel:       uint32(channelInt),
		DeviceId:      deviceId,
		Machine:       machine,
		RemoteAddr:    ctx.Request.RemoteAddr,
		PackageName:   packageName,
		Language:      uint32(languageInt),
	}
	models.AddLoginLog(&log)

	resp := response.Login{
		Token:    token,
		EntityID: tEntityAcc.EntityID,
		GateAdr:  ipGateServer + ":" + portGateServer,
	}

	_ = ctx.JSONResp(utils.Result(0, resp, "OK"))
	return
}
