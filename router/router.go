package router

import (
	"github.com/ProgramZheng/base/middleware"

	"github.com/ProgramZheng/base/controller/admin"
	"github.com/ProgramZheng/base/controller/auth"
	"github.com/ProgramZheng/base/controller/post"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func init() {
	Router = gin.Default()
	setMiddleware()
}

func setMiddleware() {
	Router.Use(middleware.CORSMiddleware())
}

func SetRouter() {
	apiGroup := Router.Group("/api")
	adminGroup := apiGroup.Group("/admin")
	{
		adminGroup.POST("register", admin.Register)
		adminGroup.POST("login", admin.Login)
		adminGroup.POST("auth", auth.Vaild)
	}
	apiGroup.Use(middleware.ValidJSONWebToken())
	{
		postGroup := apiGroup.Group("/post")
		{
			postGroup.POST("", post.Add)
			postGroup.GET("/:id", post.GetForID)
			postGroup.GET("", post.Get)
			postGroup.PATCH("/:id", post.SaveForID)
			postGroup.PATCH("", post.Save)
			postGroup.DELETE("/:id", post.DelForID)
			postGroup.DELETE("", post.Del)
		}
	}

}
