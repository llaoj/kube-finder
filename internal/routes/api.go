package routes

import (
	"github.com/gin-contrib/cors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llaoj/kube-finder/internal/config"
	"github.com/llaoj/kube-finder/internal/finder"
	"github.com/llaoj/kube-finder/internal/middleware"
)

func Api(router *gin.Engine, finderController *finder.Controller) {
	router.Use(cors.Default())
	router.Group("/apis/v1/auth", gin.BasicAuth(gin.Accounts(config.Get().Clients))).GET("/token", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": middleware.NewJWT(c.GetString(gin.AuthUserKey))})
		return
	})
	v1 := router.Group("/apis/v1").Use(middleware.JWT())
	{
		v1.GET("/namespaces/:namespace/pods/:pod/containers/:container/files", finderController.ProxyHandler)
		v1.POST("/namespaces/:namespace/pods/:pod/containers/:container/files", finderController.ProxyHandler)
		v1.GET("/containers/:containerid/files", finderController.ListHandler)
		v1.POST("/containers/:containerid/files", finderController.CreateHandler)
	}

	return
}
