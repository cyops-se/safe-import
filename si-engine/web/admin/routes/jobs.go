package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	innertypes "github.com/cyops-se/safe-import/si-inner/types"
	outertypes "github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
)

var jobSvc *usvc.UsvcStub

func RegisterJobsRoutes(auth *gin.RouterGroup, broker *usvc.UsvcBroker) {
	auth.GET("/job", GetAllJobs)
	auth.GET("/job/id/:id", GetJobById)
	auth.DELETE("/job/:id", DeleteJob)

	jobSvc = usvc.CreateStub(broker, "jobs", "si-outer", 1)
}

func GetAllJobs(c *gin.Context) {
	var items []*outertypes.Job
	r, err := jobSvc.Request("allitems")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if err = json.Unmarshal([]byte(r.Payload), &items); err != nil {
		// fmt.Println("Unable to unmarshal payload:", r.Payload, ", error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "error": err})
}

func GetJobById(c *gin.Context) {

	// r, err := jobSvc.Request("allitems")
	// c.JSON(http.StatusOK, gin.H{"items": r, "error": err})

	// id, _ := strconv.Atoi(c.Params.ByName("id"))
	// msg := &innertypes.ByIdRequest{ID: id}
	// r, err := jobSvc.RequestMessage("byid", msg)

	// var job innertypes.JobInfo{}

	// result := db.DB.First(&log, "id = ?", id)
	// if result.Error != nil {
	// 	c.JSON(http.StatusNoContent, gin.H{"error": result.Error})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"job": "nodata"})
}

func DeleteJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("id"))
	msg := &innertypes.ByIdRequest{ID: uint(id)}
	r, err := jobSvc.RequestMessage("deletebyid", msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "result": r, "error": err})
}

func NewJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("repoid"))
	msg := &outertypes.ByIdRequest{ID: id}
	r, err := jobSvc.RequestMessage("runjob", msg)
	c.JSON(http.StatusOK, gin.H{"result": r, "error": err})
}
