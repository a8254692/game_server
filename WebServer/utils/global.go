package utils

import (
	"BilliardServer/Util/db/mongodb"
	"github.com/beego/beego/v2/server/web"
	"github.com/pkg/errors"
)

var GameDB *mongodb.DBConnect //GameDB库
var LogDB *mongodb.DBConnect  //LogDB库

// 初始化数据库
func InitMongoDB(dBPath string) error {
	if dBPath == "" {
		return errors.Errorf("-->InitMongoDB-----------dBPath is empty")
	}

	mongoConn, err := mongodb.CreateDBConnect(dBPath, 10)
	if err != nil {
		return err
	}

	GameDB = new(mongodb.DBConnect)
	LogDB = new(mongodb.DBConnect)

	GameDB.Context = mongoConn
	GameDB.DBName, _ = web.AppConfig.String("mongodb::dbName")

	LogDB.Context = mongoConn
	LogDB.DBName, _ = web.AppConfig.String("mongodb::dbLogName")
	return nil
}
