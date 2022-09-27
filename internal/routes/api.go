package routes

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/llaoj/kube-finder/internal/config"
	"github.com/llaoj/kube-finder/internal/finder"
	"github.com/llaoj/kube-finder/internal/middleware"
)

func Api(router *gin.Engine, finderController *finder.Controller) {

	router.Group("/auth", gin.BasicAuth(gin.Accounts(config.Get().Clients))).GET("/token", func(c *gin.Context) {
		c.String(http.StatusOK, middleware.NewJWT(c.GetString(gin.AuthUserKey)))
		return
	})

	v1 := router.Group("/apis/v1").Use(cors.Default()).Use(middleware.JWT())
	{
		v1.GET("/namespaces/:namespace/pods/:pod/containers/:container/files", finderController.ProxyHandler)
		v1.POST("/namespaces/:namespace/pods/:pod/containers/:container/files", finderController.ProxyHandler)
		v1.GET("/containers/:containerid/files", finderController.ListHandler)
		v1.POST("/containers/:containerid/files", finderController.CreateHandler)
	}

	return
}
