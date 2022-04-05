package routes

import (
	"log"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(auth *gin.RouterGroup) {
	auth.GET("/user", GetAllUsers)
	auth.GET("/user/current", GetCurrentUser)
	auth.GET("/user/id/:id", GetUserByID)
	auth.GET("/user/field/:name/:value", GetUserByField)

	auth.POST("/user", NewUser)
	auth.PUT("/user/:id", UpdateUser)
	auth.PATCH("/user/:id", UpdateUser)
	auth.DELETE("/user/:id", DeleteUser)
}

func GetAllUsers(c *gin.Context) {
	var users []db.UserData
	result := db.DB.Model(&db.User{}).Find(&users)
	if result.Error != nil {
		// fmt.Println("ERROR GetAllUsers:", result.Error)
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func GetUserByID(c *gin.Context) {
	var user db.UserData
	id := c.Params.ByName("id")
	result := db.DB.Model(&db.User{}).First(&user, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetCurrentUser(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	c.JSON(http.StatusOK, gin.H{"claims": claims})
}

func GetUserByField(c *gin.Context) {
	var user db.UserData
	f := c.Params.ByName("name")
	v := c.Params.ByName("value")
	result := db.DB.Model(&db.User{}).First(&user, "? = ?", f, v)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

type userdata struct {
	ID       uint   `form:"id" json:"ID" binding:"required"`
	Fullname string `form:"fullname" json:"fullname" binding:"required"`
	Username string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password"`
}

func NewUser(c *gin.Context) {
	var data userdata
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusOK, "Not OK")
		return
	}

	user := &db.User{UserName: data.Username, Password: data.Password, FullName: data.Fullname}
	db.DB.Create(&user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func UpdateUser(c *gin.Context) {
	var user, data db.User

	if err := c.ShouldBind(&data); err != nil {
		log.Printf("UpdateUser: bind() failed %s", err.Error())
		c.JSON(http.StatusNotModified, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if err := db.DB.First(&user, data.ID).Error; err != nil {
		log.Printf("UpdateUser: first() failed %s", err.Error())
		c.JSON(http.StatusNotModified, gin.H{"status": "error", "error": err.Error()})
		return
	}

	user.FullName = data.FullName
	user.UserName = data.UserName

	result := db.DB.Save(&user)
	if result.Error != nil {
		log.Printf("UpdateUser: save() failed %s", result.Error.Error())
		c.JSON(http.StatusNotModified, gin.H{"status": "error", "error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user, "rows": result.RowsAffected})
}

func DeleteUser(c *gin.Context) {
	id := c.Params.ByName("id")
	result := db.DB.Delete(&db.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusNotModified, gin.H{"status": "error", "error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "rows": result})
}
