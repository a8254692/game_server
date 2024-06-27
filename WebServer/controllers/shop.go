package controllers

import (
	"BilliardServer/WebServer/controllers/response"
	"BilliardServer/WebServer/models"
	"BilliardServer/WebServer/utils"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	bectx "github.com/beego/beego/v2/server/web/context"
	"math/rand"
	"strconv"
	"time"
)

// 生成订单号
func getOrderID(entityID uint32) string {
	pre := "CM"
	now := time.Now().Unix()
	randNum := rand.Int31n(9999) + 1000
	return fmt.Sprintf("%s-%d-%d%d", pre, entityID, now, randNum)
}

func CreateOrder(ctx *bectx.Context) {
	entityID := ctx.Request.FormValue("entity_id")
	itemId := ctx.Request.FormValue("item_id")
	//price := ctx.Request.FormValue("price")
	//payChannel := ctx.Request.FormValue("pay_channel")
	//payType := ctx.Request.FormValue("pay_type")

	entityIDInt, _ := strconv.Atoi(entityID)
	itemIdInt, _ := strconv.Atoi(itemId)
	//priceInt, _ := strconv.Atoi(price)

	if entityIDInt <= 0 {
		_ = ctx.JSONResp(utils.Result(1, nil, "userId is empty"))
		return
	}
	if itemIdInt <= 0 {
		_ = ctx.JSONResp(utils.Result(2, nil, "itemId is empty"))
		return
	}

	orderId := getOrderID(uint32(entityIDInt))
	//增加登录日志
	info := models.ShopOrder{
		OrderId:  orderId,
		EntityID: uint32(entityIDInt),
		ItemId:   uint32(itemIdInt),
		Price:    0,
	}
	models.CreateShopOrder(&info)

	resp := response.Order{
		EntityID: uint32(entityIDInt),
		OrderId:  orderId,
	}

	_ = ctx.JSONResp(utils.Result(0, resp, "OK"))
	return
}

// 假设这是你从谷歌支付控制台获取到的商户密钥
var merchantKey = "YOUR_MERCHANT_KEY"

// GoogleWalletNotification 是回调数据的结构
type GoogleWalletNotification struct {
	MessageType   string         `json:"messageType"`
	MerchantID    string         `json:"merchantId"`
	WalletObjects []WalletObject `json:"walletObjects"`
	// 其他字段...
}

// WalletObject 是回调数据中钱包对象的结构
type WalletObject struct {
	ID                  string `json:"id"`
	GoogleTransactionID string `json:"googleTransactionId"`
	State               string `json:"state"`
	// 其他字段...
}

// VerifySignature 验证回调的签名
func VerifySignature(payload, sign, merchantKey string) bool {
	decodePublic, err := base64.StdEncoding.DecodeString(merchantKey)
	if err != nil {
		return false
	}
	pubInterface, err := x509.ParsePKIXPublicKey(decodePublic)
	if err != nil {
		return false
	}

	pub := pubInterface.(*rsa.PublicKey)

	decodeSign, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	sh1 := sha1.New()
	sh1.Write([]byte(payload))
	hashData := sh1.Sum(nil)

	result := rsa.VerifyPKCS1v15(pub, crypto.SHA1, hashData, decodeSign)
	if result != nil {
		return false
	}

	return true
}

func GooglePayCallback(ctx *bectx.Context) {
	logs.Warning("-->controllers--GooglePayCallback--RequestBody--", string(ctx.Input.RequestBody))
	logs.Warning("-->controllers--GooglePayCallback--Header--", ctx.Request.Header)

	// 解析回调数据
	var notification GoogleWalletNotification
	if err := ctx.BindJSON(&notification); err != nil {
		_ = ctx.JSONResp(utils.Result(1, nil, "userId is empty"))
		return
	}

	// 验证签名
	if !VerifySignature(string(ctx.Input.RequestBody), ctx.Request.Header.Get("X-Goog-Signature"), merchantKey) {
		_ = ctx.JSONResp(utils.Result(2, nil, "userId is empty"))
		return
	}

	// 处理业务逻辑
	for _, wo := range notification.WalletObjects {
		if wo.State == "SUCCESS" {
			fmt.Printf("Payment successful for wallet object ID: %s\n", wo.ID)
			// 在这里更新订单状态、处理付款等
		} else {
			fmt.Printf("Payment failed for wallet object ID: %s\n", wo.ID)
			// 在这里处理支付失败的情况
		}
	}

	_ = ctx.JSONResp(utils.Result(0, nil, "OK"))
	return
}
