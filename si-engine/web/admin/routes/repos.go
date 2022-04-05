package routes

import (
	"log"
	"net/http"
	"strconv"

	innertypes "github.com/cyops-se/safe-import/si-inner/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
)

var repoSvc *usvc.UsvcStub
var jobsSvc *usvc.UsvcStub

func RegisterReposRoutes(auth *gin.RouterGroup, broker *usvc.UsvcBroker) {
	auth.GET("/repo", GetAllRepo)
	auth.GET("/repo/download/:id", GetDownloadRepo)
	auth.GET("/repo/id/:id", GetRepoByID)
	auth.GET("/repo/field/:name/:value", GetRepoByField)

	auth.POST("/repo", NewRepo)
	auth.POST("/repo/approve", ApproveRepo)

	auth.PUT("/repo/:id", UpdateRepo)
	auth.PATCH("/repo/:id", UpdateRepo)

	auth.DELETE("/repo", DeleteAllRepos)
	auth.DELETE("/repo/:id", DeleteRepo)

	repoSvc = usvc.CreateStub(broker, "repos", "si-inner", 1)
	jobsSvc = usvc.CreateStub(broker, "jobs", "si-outer", 1)
}

func GetAllRepo(c *gin.Context) {
	r, err := repoSvc.Request("allitems")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": r})
}

func GetDownloadRepo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	msg := &innertypes.ByIdRequest{ID: uint(id)}
	r, err := jobsSvc.RequestMessage("requestrepodownload", msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r})
}

func GetRepoByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	msg := &innertypes.ByIdRequest{ID: uint(id)}
	r, err := repoSvc.RequestMessage("byid", msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r})
}

func GetRepoByField(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Method not yet implemented"})
}

func NewRepo(c *gin.Context) {
	var item innertypes.Repository
	if err := c.ShouldBind(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r, err := repoSvc.RequestMessage("create", item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r})
}

func ApproveRepo(c *gin.Context) {
	var args innertypes.ApproveRequest
	if err := c.ShouldBind(&args); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r, err := repoSvc.RequestMessage("approve", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r})
}

func UpdateRepo(c *gin.Context) {
	var item innertypes.Repository
	if err := c.ShouldBind(&item); err != nil {
		log.Printf("UpdateRepo: Failed to bind PUT body: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r, err := repoSvc.RequestMessage("update", item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": r})
}

func DeleteAllRepos(c *gin.Context) {
	r, err := repoSvc.Request("deleteall")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "result": r})
}

func DeleteRepo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	msg := &innertypes.ByIdRequest{ID: uint(id)}
	r, err := repoSvc.RequestMessage("deletebyid", msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "result": r})
}
