package routes

import (
	"net/http"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

func RegisterLogRoutes(auth *gin.RouterGroup) {
	auth.GET("/log", GetAllLogs)
	auth.GET("/log/id/:id", GetLogByID)
	auth.GET("/log/field/:name/:value", GetLogByField)

	auth.POST("/log", NewLog)
	auth.PUT("/log/:id", UpdateLog)
	auth.PATCH("/log/:id", UpdateLog)

	auth.DELETE("/log", DeleteAllLogs)
}

func GetAllLogs(c *gin.Context) {
	var logs []db.Log
	result := db.DB.Find(&logs)
	if result.Error != nil {
		// fmt.Println("ERROR GetAllLogs:", result.Error)
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func GetLogByID(c *gin.Context) {
	var log db.Log
	id := c.Params.ByName("id")
	result := db.DB.First(&log, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"log": log})
}

func GetLogByField(c *gin.Context) {
	var logs []db.Log
	f := c.Params.ByName("name")
	v := c.Params.ByName("value")
	result := db.DB.Where(map[string]interface{}{f: v}).Find(&logs)
	if result.Error != nil {
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	// fmt.Println("RESULTS:", result)

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func NewLog(c *gin.Context) {
	var data db.Log
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	db.DB.Create(&data)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func UpdateLog(c *gin.Context) {
	var data db.Log
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	db.DB.Save(&data)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func DeleteAllLogs(c *gin.Context) {
	db.DB.Where("1 = 1").Delete(&db.Log{})
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
