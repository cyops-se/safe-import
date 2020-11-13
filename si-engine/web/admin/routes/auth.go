package routes

import (
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	r.POST("/api/auth/login", authMiddleware.LoginHandler)

	r.POST("/api/auth/logout", authMiddleware.LogoutHandler)

	r.POST("/api/auth/register", func(c *gin.Context) {
		var user db.User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusOK, "Not OK")
			return
		}

		// user := &db.User{UserName: data.Username, Password: data.Password, FullName: data.Fullname}
		db.DB.Create(&user)
		c.JSON(http.StatusOK, gin.H{"data": user})
	})
}
