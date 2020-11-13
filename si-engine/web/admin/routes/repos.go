package routes

import (
	"net/http"
	"strconv"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	innertypes "github.com/cyops-se/safe-import/si-inner/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
)

var repoSvc *usvc.UsvcStub

func RegisterReposRoutes(auth *gin.RouterGroup, broker *usvc.UsvcBroker) {
	auth.GET("/repo", GetAllRepo)
	auth.GET("/repo/id/:id", GetRepoByID)
	auth.GET("/repo/field/:name/:value", GetRepoByField)

	auth.POST("/repo", NewRepo)
	auth.POST("/repo/approve", ApproveRepo)

	auth.PUT("/repo/:id", UpdateRepo)
	auth.PATCH("/repo/:id", UpdateRepo)

	auth.DELETE("/repo", DeleteAllRepos)
	auth.DELETE("/repo/:id", DeleteRepo)
	repoSvc = usvc.CreateStub(broker, "repos", "si-outer", 1)
}

func GetAllRepo(c *gin.Context) {
	r, err := repoSvc.Request("allitems")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": r, "error": err})
}

func GetRepoByID(c *gin.Context) {
	var item db.NetRepos
	id := c.Params.ByName("id")
	result := db.DB.First(&item, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func GetRepoByField(c *gin.Context) {
	var items []db.NetRepos
	f := c.Params.ByName("name")
	v := c.Params.ByName("value")
	result := db.DB.Where(map[string]interface{}{f: v}).Find(&items)
	if result.Error != nil {
		c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func NewRepo(c *gin.Context) {
	var item innertypes.Repository
	if err := c.ShouldBind(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	r, err := repoSvc.RequestMessage("create", item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r, "error": err})
}

func ApproveRepo(c *gin.Context) {
	var args innertypes.ApproveRequest
	if err := c.ShouldBind(&args); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	r, err := repoSvc.RequestMessage("approve", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r, "error": err})
}

func UpdateRepo(c *gin.Context) {
	var item innertypes.Repository
	if err := c.ShouldBind(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	r, err := repoSvc.RequestMessage("update", item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r, "error": err})
}

func DeleteAllRepos(c *gin.Context) {
	r, err := repoSvc.Request("deleteall")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "result": r, "error": err})
}

func DeleteRepo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	msg := &innertypes.ByIdRequest{ID: uint(id)}
	r, err := repoSvc.RequestMessage("deletebyid", msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "result": r, "error": err})
}
