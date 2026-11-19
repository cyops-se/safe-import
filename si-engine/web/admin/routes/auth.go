package routes

import (
	"net/http"
	"sync"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

type registration struct {
	UserName string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	FullName string `form:"fullname" json:"fullname" binding:"required"`
}

// registerLock serializes bootstrap registrations so that only one first admin can be created
var registerLock sync.Mutex

func RegisterAuthRoutes(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	r.POST("/api/auth/login", authMiddleware.LoginHandler)
	r.POST("/api/auth/logout", authMiddleware.LogoutHandler)
	r.GET("/api/auth/refresh", authMiddleware.RefreshHandler)

	// Self registration is only possible as long as there are no users at all. The first
	// registered user becomes administrator. All other users are created by an administrator.
	r.GET("/api/auth/register", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"open": registrationOpen()})
	})

	r.POST("/api/auth/register", func(c *gin.Context) {
		registerLock.Lock()
		defer registerLock.Unlock()

		if !registrationOpen() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Registration is closed, ask an administrator to create an account"})
			return
		}

		var data registration
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := db.HashPassword(data.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user := &db.User{UserName: data.UserName, Password: hash, FullName: data.FullName, Role: db.RoleAdmin}
		if err := db.DB.Create(user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": toUserData(user)})
	})
}

func registrationOpen() bool {
	var count int64
	if err := db.DB.Model(&db.User{}).Count(&count).Error; err != nil {
		return false
	}
	return count == 0
}
