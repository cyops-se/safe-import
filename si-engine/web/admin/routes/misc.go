package routes

import (
	"container/list"
	"net/http"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func RegisterMiscRoutes(r *gin.Engine, broker *usvc.UsvcBroker, connections *list.List) {

	r.POST("/test", func(c *gin.Context) {
		results := &db.KeyValuePair{Key: "KEY", Value: "VALUE"}
		db.DB.Create(&results)
		c.JSON(http.StatusOK, gin.H{"data": results})
	})

	r.GET("/test", func(c *gin.Context) {
		var results []db.KeyValuePair
		db.DB.Find(&results)
		c.JSON(http.StatusOK, gin.H{"data": results})
	})

	r.GET("/uix", func(c *gin.Context) {
		var results []db.KeyValuePair
		db.DB.Find(&results)
		c.JSON(http.StatusOK, gin.H{"data": results})
	})

	r.GET("/msg/:text", func(c *gin.Context) {
		text := c.Params.ByName("text")
		msg := gin.H{"topic": "chat", "data": gin.H{"message": text}}
		c.JSON(http.StatusOK, msg)

		for e := connections.Front(); e != nil; e = e.Next() {
			conn := e.Value.(*websocket.Conn)
			conn.WriteJSON(msg)
		}
	})

	r.GET("/chat/:text", func(c *gin.Context) {
		// text := fmt.Sprintf(`{"topic": "chat", "data": {"message": "%s"}}`, c.Params.ByName("text"))
		text := c.Params.ByName("text")
		c.JSON(http.StatusOK, []byte(text))
		broker.PublishString("chat", text)
	})
}
