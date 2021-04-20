package admin

import (
	"container/list"
	"encoding/json"
	"log"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/cyops-se/safe-import/si-engine/web/admin/routes"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/nats-io/nats.go"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Run(broker *usvc.UsvcBroker) {
	db.ConnectDatabase()

	// fmt.Println("STARTING ADMIN WEB")
	Log("info", "STARTING ADMIN WEB", "")
	connections := list.New()

	r := gin.Default()
	r.Static("/ui", "./dist")

	authMiddleware, _ := jwt.New(JWT)
	authMiddleware.MiddlewareInit()

	r.GET("/ws", func(c *gin.Context) {
		WShandler(c.Writer, c.Request, connections)
	})

	routes.RegisterAuthRoutes(r, authMiddleware)
	routes.RegisterMiscRoutes(r, broker, connections)

	api := r.Group("/api")
	api.Use(authMiddleware.MiddlewareFunc())
	{
		routes.RegisterUserRoutes(api)
		routes.RegisterLogRoutes(api)
		routes.RegisterDnsRoutes(api, broker)
		routes.RegisterHttpRoutes(api, broker)
		routes.RegisterReposRoutes(api, broker)
		routes.RegisterJobsRoutes(api, broker)
		routes.RegisterPKIRoutes(api)
	}

	broker.Subscribe(">", func(m *nats.Msg) {
		if m.Subject != "system.heartbeat" {
			// fmt.Println("NATS message: ", m.Subject, string(m.Data))
		}

		msg := gin.H{"topic": m.Subject, "data": gin.H{"message": string(m.Data)}}
		for e := connections.Front(); e != nil; e = e.Next() {
			conn := e.Value.(*websocket.Conn)
			conn.WriteJSON(msg)
		}
	})

	broker.Subscribe("log.>", func(m *nats.Msg) {
		var entry db.Log
		if err := json.Unmarshal(m.Data, &entry); err != nil {
			log.Println("ERROR: Failed to unmarshal log entry:", string(m.Data), err)
		}

		db.DB.Create(&entry)
	})

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	r.Run(":7499")
}
