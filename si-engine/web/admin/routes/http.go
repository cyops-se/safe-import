package routes

import (
	"net/http"
	"strconv"

	"github.com/cyops-se/safe-import/si-inner/types"
	innertypes "github.com/cyops-se/safe-import/si-inner/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
)

var httpSvc *usvc.UsvcStub

func RegisterHttpRoutes(auth *gin.RouterGroup, broker *usvc.UsvcBroker) {
	auth.GET("/http", GetAllHttp)
	auth.GET("/http/id/:id", GetHttpByID)
	auth.GET("/http/field/:name/:value", GetHttpByField)
	auth.GET("/http/prune", PruneHttp)

	auth.PUT("/http/:id", UpdateHttp)
	auth.PATCH("/http/:id", UpdateHttp)

	auth.DELETE("/http", DeleteAllHttp)
	httpSvc = usvc.CreateStub(broker, "http", "si-inner", 1)
}

func GetAllHttp(c *gin.Context) {
	r, _ := httpSvc.Request("allitems")
	c.JSON(http.StatusOK, gin.H{"items": r})
}

func GetHttpByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	args := &innertypes.ByIdRequest{ID: uint(id)}
	r, _ := httpSvc.RequestMessage("byid", &args)
	c.JSON(http.StatusOK, gin.H{"item": r})
}

func GetHttpByField(c *gin.Context) {
	f := c.Params.ByName("name")
	v := c.Params.ByName("value")
	args := &innertypes.ByNameRequest{Name: f, Value: v}
	r, _ := httpSvc.RequestMessage("byfieldname", args)
	c.JSON(http.StatusOK, gin.H{"items": r})
}

func UpdateHttp(c *gin.Context) {
	var item types.HttpRequest
	if err := c.ShouldBind(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to unmarshal request", "error": err})
		return
	}

	r, err := httpSvc.RequestMessage("update", item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "'update' request to si-inner failed", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": r})
}

func PruneHttp(c *gin.Context) {
	r, err := httpSvc.Request("prune")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "'prune' request to si-inner failed", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": r})
}

func DeleteAllHttp(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"data": "not yet implemented"})
}
