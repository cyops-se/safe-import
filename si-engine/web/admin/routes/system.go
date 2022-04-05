package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemInformation struct  {
	GitVersion string `json:"gitversion"`
	GitCommit string `json:"gitcommit"`
}

var SysInfo SystemInformation

func RegisterSystemRoutes(auth *gin.RouterGroup) {
	auth.GET("/system/info", GetSysInfo)
}

func GetSysInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"sysinfo": SysInfo})
}