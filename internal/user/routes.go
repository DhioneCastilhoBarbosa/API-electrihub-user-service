package user

import (
	"user-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	group := r.Group("/user")
	{
		group.POST("/register", RegisterUser)
		group.POST("/login", LoginUser)
		group.GET("/public/installers", ListPublicInstallers)
		group.GET("/list", middlewares.AuthMiddleware(), ListUsers)
		group.GET("/installers/pending", middlewares.AuthMiddleware(), ListPendingInstallers)
		group.PATCH("/:id/authorize", middlewares.AuthMiddleware(), AuthorizeUser)
		group.PUT("/:id/password", middlewares.AuthMiddleware(), UpdatePassword)
		group.PUT("/:id", middlewares.AuthMiddleware(), UpdateUser)
		group.PUT("/:id/photo", middlewares.AuthMiddleware(), UpdateUserPhoto)
		group.DELETE("/:id", middlewares.AuthMiddleware(), DeleteUser)
		group.GET("/public/installers/nearby", ListNearbyInstallers)

	}
}
