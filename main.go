package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/llaoj/kube-finder/internal/config"
	"github.com/llaoj/kube-finder/internal/finder"
	"github.com/llaoj/kube-finder/internal/routes"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)

	router := gin.Default()
	routes.Api(router, finder.NewController())
	_ = router.Run(":" + config.Get().HttpPort)
}
