package routers

import (
	"BilliardServer/WebServer/controllers"
	"github.com/beego/beego/v2/server/web"
)

func GetNsRouter() *web.Namespace {
	ns := web.NewNamespace("/v1",
		web.NSGet("/login", controllers.Login),
		web.NSGet("/login/google", controllers.GoogleLogin),
		web.NSGet("/login/apple", controllers.AppleLogin),
		web.NSGet("/order/create", controllers.CreateOrder),
		web.NSGet("/google/callback", controllers.GooglePayCallback),

		//web.NSPost("/register", controllers.Register),
	)

	return ns
}
