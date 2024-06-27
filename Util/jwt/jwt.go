package jwt

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	EntityId uint32
	jwt.RegisteredClaims
}

// 签名密钥
const SIGN_KEY = "byJ74d3s"

// 随机字符串
var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(strLen int) string {
	randBytes := make([]rune, strLen)
	for i := range randBytes {
		randBytes[i] = letters[rand.Intn(len(letters))]
	}
	return string(randBytes)
}

// GenToken 生成JWT
func GenToken(entityID uint32) (string, error) {
	if entityID <= 0 {
		return "", errors.New("param is empty")
	}

	// 创建一个我们自己声明的数据
	claims := MyCustomClaims{
		entityID, // 自定义字段
		jwt.RegisteredClaims{
			Issuer:    "WebServer",                                       // 签发者
			Subject:   fmt.Sprintf("%d", entityID),                       // 签发对象
			Audience:  jwt.ClaimStrings{"Android_APP", "IOS_APP"},        //签发受众
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)), // 定义过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                    //签发时间
			ID:        randStr(10),                                       // wt ID, 类似于盐值
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 生成签名字符串
	return token.SignedString([]byte(SIGN_KEY))
}

func ParseToken(tokenString string) (*MyCustomClaims, error) {
	if tokenString == "" {
		return nil, errors.New("param is empty")
	}

	// 解析token
	var mc = new(MyCustomClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(SIGN_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	// 对token对象中的Claim进行类型断言
	if token.Valid { // 校验token
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
