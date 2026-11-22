package admin

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/cyops-se/safe-import/si-engine/web/admin/routes"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/nats-io/nats.go"

	"github.com/gin-gonic/gin"
)

// The UI is built into the static folder by build.sh (or the Dockerfile). The folder
// only contains a placeholder in version control.
//
//go:embed all:static
var static embed.FS

// Run starts the admin web server and blocks until it fails
func Run(broker *usvc.UsvcBroker, cfg *Config) error {
	if err := db.ConnectDatabase(cfg.DBPath); err != nil {
		return err
	}

	r, err := NewRouter(cfg, broker)
	if err != nil {
		return err
	}

	Log("info", "STARTING ADMIN WEB", cfg.ListenAddr)
	return r.Run(cfg.ListenAddr)
}

// NewRouter sets up all HTTP routes. The database must be connected. If broker is nil
// (unit tests), the routes and subscriptions that need the message bus are left out.
func NewRouter(cfg *Config, broker *usvc.UsvcBroker) (*gin.Engine, error) {
	subFS, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}

	if _, err := fs.Stat(subFS, "index.html"); err != nil {
		log.Println("WARNING: the web UI is not included in this build, run si-engine/build.sh")
	}

	hub := NewHub()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), securityHeaders)
	r.StaticFS("/ui", http.FS(subFS))

	authMiddleware, err := NewJWTMiddleware(cfg)
	if err != nil {
		return nil, err
	}

	r.GET("/ws", func(c *gin.Context) {
		WShandler(c.Writer, c.Request, hub, authMiddleware, cfg)
	})

	routes.RegisterAuthRoutes(r, authMiddleware)

	api := r.Group("/api")
	api.Use(authMiddleware.MiddlewareFunc())
	{
		routes.RegisterUserRoutes(api)
		routes.RegisterLogRoutes(api)
		routes.RegisterSystemRoutes(api)
		if broker != nil {
			routes.RegisterDnsRoutes(api, broker)
			routes.RegisterHttpRoutes(api, broker)
			routes.RegisterReposRoutes(api, broker)
			routes.RegisterJobsRoutes(api, broker)
		}
	}

	if broker != nil {
		// Forward bus messages to the (authenticated) websocket clients
		broker.Subscribe(">", func(m *nats.Msg) {
			if hub.Count() == 0 {
				return
			}
			hub.Broadcast(gin.H{"topic": m.Subject, "data": gin.H{"message": string(m.Data)}})
		})

		broker.Subscribe("log.>", func(m *nats.Msg) {
			var entry db.Log
			if err := json.Unmarshal(m.Data, &entry); err != nil {
				log.Println("ERROR: Failed to unmarshal log entry:", string(m.Data), err)
				return
			}

			db.DB.Create(&entry)
		})
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	return r, nil
}

// securityHeaders adds basic security related HTTP headers to all responses
func securityHeaders(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "no-referrer")
	c.Next()
}
