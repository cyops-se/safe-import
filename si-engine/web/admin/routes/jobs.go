package routes

import (
	"net/http"
	"strconv"

	innertypes "github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gin-gonic/gin"
)

var jobSvc *usvc.UsvcStub

func RegisterJobsRoutes(auth *gin.RouterGroup, broker *usvc.UsvcBroker) {
	auth.GET("/job", GetAllJobs)
	auth.GET("/job/id/:id", GetJobByID)
	// auth.GET("/job/:id/start", StartJob)
	// auth.GET("/job/:id/stop", StopJob)
	auth.GET("/job/new/:repoid", NewJob)

	// auth.DELETE("/log/:id", DeleteJobById)
	jobSvc = usvc.CreateStub(broker, "jobs", "si-outer", 1)
}

func GetAllJobs(c *gin.Context) {
	r, err := jobSvc.Request("allitems")
	c.JSON(http.StatusOK, gin.H{"items": r, "error": err})
}

func GetJobByID(c *gin.Context) {

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

func NewJob(c *gin.Context) {
	id, _ := strconv.Atoi(c.Params.ByName("repoid"))
	msg := &innertypes.ByIdRequest{ID: id}
	r, err := jobSvc.RequestMessage("runjob", msg)
	c.JSON(http.StatusOK, gin.H{"result": r, "error": err})
}
