package admin

import (
	"fmt"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

type login struct {
	Username string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "email"

var JWT = &jwt.GinJWTMiddleware{
	Realm:       "si-engine",
	Key:         []byte("RGV05HJoZWx0b3Ryb2xpZ3R2aWxrZXRzduVydGz2c2Vub3JkZGV0dGHkcg=="),
	Timeout:     time.Hour * 24 * 365 * 30,
	MaxRefresh:  time.Hour,
	IdentityKey: identityKey,
	PayloadFunc: func(data interface{}) jwt.MapClaims {
		fmt.Println("Payload:", data)
		if v, ok := data.(*db.User); ok {
			return jwt.MapClaims{
				identityKey: v.UserName,
			}
		}
		return jwt.MapClaims{}
	},
	IdentityHandler: func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return &db.User{
			UserName: claims[identityKey].(string),
		}
	},
	Authenticator: func(c *gin.Context) (interface{}, error) {
		var loginVals login
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}
		userID := loginVals.Username
		password := loginVals.Password

		// TODO: save salted and hashed passwords rather than plain text
		var user db.User
		result := db.DB.Where("user_name = ? and password = ?", userID, password).First(&user)
		if result.Error != nil {
			return nil, jwt.ErrFailedAuthentication
		}
		return &user, nil
	},
	Authorizator: func(data interface{}, c *gin.Context) bool {
		if v, ok := data.(*db.User); ok && v.UserName == "admin@acme.com" {
			return true
		}

		return true
	},
	Unauthorized: func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	},
	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	// - "param:<name>"
	TokenLookup: "header: Authorization",
	// TokenLookup: "query:token",

	// TokenHeadName is a string in the header. Default value is "Bearer"
	TokenHeadName: "Bearer",

	// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
	TimeFunc: time.Now,
}
