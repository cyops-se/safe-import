package routes

import (
	"net/http"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(auth *gin.RouterGroup) {
	auth.GET("/user", GetAllUsers)
	auth.GET("/user/id/:id", GetUserByID)
	auth.GET("/user/field/:name/:value", GetUserByField)

	auth.POST("/user", NewUser)
	auth.PUT("/user/:id", UpdateUser)
	auth.PATCH("/user/:id", UpdateUser)
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
	ID       string `form:"id" json:"id" binding:"required"`
	Fullname string `form:"fullname" json:"fullname" binding:"required"`
	Username string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
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
	var data userdata
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusNotModified, gin.H{"error": err})
		return
	}

	var user db.User
	if err := db.DB.First(&user, "ID = ?", data.ID); err != nil {
		c.JSON(http.StatusNotModified, gin.H{"error": err})
		return
	}

	db.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}
