package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jekyulll/url_shortener/internal/middleware"
)

func (a *Application) initRouter() {

	a.r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 用户认证路由组
	u := a.r.Group("/api/auth")

	// 用户认证相关端点
	u.POST("/login", a.userHandler.Login)                  // 用户登录
	u.POST("/register", a.userHandler.Register)            // 用户注册
	u.POST("/forget", a.userHandler.ForgetPassword)        // 忘记密码
	u.GET("/register/:email", a.userHandler.SendEmailCode) // 发送注册验证码

	// URL缩短服务相关路由
	a.r.GET("/:code", a.urlHandler.RedirectURL) // 短链接重定向

	// 需要JWT认证的URL管理API
	url := a.r.Group("/api", middleware.JWTAuther(a.jwt))
	url.POST("/url", a.urlHandler.CreateURL)                // 创建短链接
	url.GET("/urls", a.urlHandler.GetURLs)                  // 获取用户的所有短链接
	url.DELETE("/url/:code", a.urlHandler.DeleteURL)        // 删除短链接
	url.PATCH("/url/:code", a.urlHandler.UpdateURLDuration) // 更新短链接的有效期
}
