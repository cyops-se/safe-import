package routes

import (
	"net/http"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

func RegisterPKIRoutes(auth *gin.RouterGroup) {
	auth.GET("/pki/cert", GetAllCertificates)
	auth.GET("/pki/cert/id/:id", GetCertByID)
	auth.GET("/pki/cert/serial/:sno", GetCertBySerial)

	auth.POST("/pki/csr/:caid", NewCSR)
}

func GetAllCertificates(c *gin.Context) {
	var certs []db.Certificate
	result := db.DB.Find(&certs)
	if result.Error != nil {
		// fmt.Println("ERROR GetAllCertificates:", result.Error)
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"certificates": certs})
}

func GetCertByID(c *gin.Context) {
	var user db.UserData
	id := c.Params.ByName("id")
	result := db.DB.Model(&db.User{}).First(&user, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetCertBySerial(c *gin.Context) {
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

func NewCSR(c *gin.Context) {
	var data userdata
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusOK, "Not OK")
		return
	}

	user := &db.User{UserName: data.Username, Password: data.Password, FullName: data.Fullname}
	db.DB.Create(&user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}
