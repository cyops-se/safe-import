package admin

import (
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/cyops-se/safe-import/si-engine/web/admin/routes"
	"github.com/gin-gonic/gin"
)

type login struct {
	Username string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

const identityKey = "email"
const roleKey = "role"

// dummyHash is compared against when the user does not exist, so that the
// response time does not reveal which user names exist
var dummyHash, _ = db.HashPassword("dummy-password-for-timing")

// readOnlyMethods are the HTTP methods a viewer is allowed to use
var readOnlyMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// IsAuthorized decides if a user with the given role may perform a request with the given method
func IsAuthorized(role string, method string) bool {
	switch role {
	case db.RoleAdmin:
		return true
	case db.RoleViewer:
		return readOnlyMethods[method]
	default:
		return false
	}
}

// NewJWTMiddleware creates the JWT middleware from the configuration
func NewJWTMiddleware(cfg *Config) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "si-engine",
		Key:         cfg.JWTKey,
		Timeout:     cfg.TokenTimeout,
		MaxRefresh:  cfg.TokenTimeout,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*db.User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
					roleKey:     v.Role,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			name, _ := claims[identityKey].(string)
			role, _ := claims[roleKey].(string)
			return &db.User{UserName: name, Role: role}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			var user db.User
			if result := db.DB.Where("user_name = ?", loginVals.Username).First(&user); result.Error != nil {
				db.CheckPassword(dummyHash, loginVals.Password)
				Log("warning", "Failed login attempt", loginVals.Username)
				return nil, jwt.ErrFailedAuthentication
			}

			if !db.CheckPassword(user.Password, loginVals.Password) {
				Log("warning", "Failed login attempt", loginVals.Username)
				return nil, jwt.ErrFailedAuthentication
			}

			return &user, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			v, ok := data.(*db.User)
			if !ok {
				return false
			}

			// The role in the token may be outdated, always use the current one
			var user db.User
			if err := db.DB.Where("user_name = ?", v.UserName).First(&user).Error; err != nil {
				return false
			}

			if routes.SelfServicePaths[c.Request.Method+" "+c.FullPath()] {
				return true
			}

			return IsAuthorized(user.Role, c.Request.Method)
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
}
