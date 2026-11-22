package routes

import (
	"log"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

// userFieldColumns maps the field names accepted by GetUserByField to database columns
var userFieldColumns = map[string]string{
	"id":       "id",
	"email":    "user_name",
	"username": "user_name",
	"fullname": "full_name",
	"role":     "role",
}

func RegisterUserRoutes(auth *gin.RouterGroup) {
	auth.GET("/user", GetAllUsers)
	auth.GET("/user/current", GetCurrentUser)
	auth.GET("/user/id/:id", GetUserByID)
	auth.GET("/user/field/:name/:value", GetUserByField)

	auth.POST("/user", NewUser)
	auth.PUT("/user/:id", UpdateUser)
	auth.PATCH("/user/:id", UpdateUser)
	auth.DELETE("/user/:id", DeleteUser)

	// Any authenticated user may change his/her own password
	auth.PUT("/password", ChangeOwnPassword)
}

// SelfServicePaths are API paths (method + path) that every authenticated user may call,
// regardless of role
var SelfServicePaths = map[string]bool{
	"PUT /api/password": true,
}

type passwordChange struct {
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"newpassword" binding:"required"`
}

func ChangeOwnPassword(c *gin.Context) {
	var data passwordChange
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := jwt.ExtractClaims(c)
	name, _ := claims["email"].(string)

	var user db.User
	if err := db.DB.Where("user_name = ?", name).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !db.CheckPassword(user.Password, data.Password) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Current password is wrong"})
		return
	}

	hash, err := db.HashPassword(data.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Password = hash
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func toUserData(user *db.User) db.UserData {
	return db.UserData{ID: user.ID, UserName: user.UserName, FullName: user.FullName, Role: user.Role}
}

func validRole(role string) bool {
	return role == db.RoleAdmin || role == db.RoleViewer
}

func GetAllUsers(c *gin.Context) {
	var users []db.UserData
	result := db.DB.Model(&db.User{}).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func GetUserByID(c *gin.Context) {
	var user db.UserData
	id := c.Params.ByName("id")
	result := db.DB.Model(&db.User{}).First(&user, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetCurrentUser(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	c.JSON(http.StatusOK, gin.H{"claims": claims})
}

func GetUserByField(c *gin.Context) {
	column, ok := userFieldColumns[c.Params.ByName("name")]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown field"})
		return
	}

	var user db.UserData
	v := c.Params.ByName("value")
	result := db.DB.Model(&db.User{}).Where(column+" = ?", v).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

type userdata struct {
	ID       uint   `form:"id" json:"ID"`
	Fullname string `form:"fullname" json:"fullname" binding:"required"`
	Username string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password"`
	Role     string `form:"role" json:"role"`
}

func NewUser(c *gin.Context) {
	var data userdata
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if data.Role == "" {
		data.Role = db.RoleViewer
	}
	if !validRole(data.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	hash, err := db.HashPassword(data.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &db.User{UserName: data.Username, Password: hash, FullName: data.Fullname, Role: data.Role}
	if err := db.DB.Create(user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toUserData(user)})
}

func UpdateUser(c *gin.Context) {
	var user db.User
	var data userdata

	if err := c.ShouldBind(&data); err != nil {
		log.Printf("UpdateUser: bind() failed %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if err := db.DB.First(&user, c.Params.ByName("id")).Error; err != nil {
		log.Printf("UpdateUser: first() failed %s", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": err.Error()})
		return
	}

	user.FullName = data.Fullname
	user.UserName = data.Username

	if data.Role != "" {
		if !validRole(data.Role) {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid role"})
			return
		}
		if user.Role == db.RoleAdmin && data.Role != db.RoleAdmin && isLastAdmin(&user) {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Cannot remove the last administrator"})
			return
		}
		user.Role = data.Role
	}

	if data.Password != "" {
		hash, err := db.HashPassword(data.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
			return
		}
		user.Password = hash
	}

	result := db.DB.Save(&user)
	if result.Error != nil {
		log.Printf("UpdateUser: save() failed %s", result.Error.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toUserData(&user), "rows": result.RowsAffected})
}

func DeleteUser(c *gin.Context) {
	var user db.User
	if err := db.DB.First(&user, c.Params.ByName("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if user.Role == db.RoleAdmin && isLastAdmin(&user) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Cannot delete the last administrator"})
		return
	}

	result := db.DB.Unscoped().Delete(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "rows": result.RowsAffected})
}

func isLastAdmin(user *db.User) bool {
	var count int64
	db.DB.Model(&db.User{}).Where("role = ? AND id <> ?", db.RoleAdmin, user.ID).Count(&count)
	return count == 0
}
