package routes

import (
	"net/http"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/cyops-se/safe-import/si-inner/types"
	innertypes "github.com/cyops-se/safe-import/si-inner/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
)

var dnsSvc *usvc.UsvcStub

func RegisterDnsRoutes(auth *gin.RouterGroup, broker *usvc.UsvcBroker) {
	auth.GET("/dns", GetAllDns)
	auth.GET("/dns/id/:id", GetDnsByID)
	auth.GET("/dns/field/:name/:value", GetDnsByField)
	auth.GET("/dns/prune", PruneDns)

	auth.PUT("/dns/:id", UpdateDns)
	auth.PATCH("/dns/:id", UpdateDns)

	auth.DELETE("/dns", DeleteAllDns)
	dnsSvc = usvc.CreateStub(broker, "dns", "si-inner", 1)
}

func GetAllDns(c *gin.Context) {
	r, err := dnsSvc.Request("allitems")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "'allitems' request to si-inner failed", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": r})
}

func GetDnsByID(c *gin.Context) {
	var item db.NetCapture
	id := c.Params.ByName("id")
	result := db.DB.First(&item, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func GetDnsByField(c *gin.Context) {
	f := c.Params.ByName("name")
	v := c.Params.ByName("value")
	args := &innertypes.ByNameRequest{Name: f, Value: v}
	r, err := dnsSvc.RequestMessage("byfieldname", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "'byfieldname' request to si-inner failed", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": r})
}

func UpdateDns(c *gin.Context) {
	var item types.DnsRequest
	if err := c.ShouldBind(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to unmarshal request", "error": err})
		return
	}

	r, err := dnsSvc.RequestMessage("update", item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "'update' request to si-inner failed", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": r})
}

func PruneDns(c *gin.Context) {
	r, err := dnsSvc.Request("prune")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "'prune' request to si-inner failed", "error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": r})
}

func DeleteAllDns(c *gin.Context) {
	db.DB.Where("1 = 1").Delete(&db.NetCapture{})
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
