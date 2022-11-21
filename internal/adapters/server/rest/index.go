package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rendau/dop/adapters/logger"
	dopHttps "github.com/rendau/dop/adapters/server/https"
	"github.com/rendau/fs/internal/domain/core"
	swagFiles "github.com/swaggo/files"
	ginSwag "github.com/swaggo/gin-swagger"
)

type St struct {
	lg   logger.Lite
	core *core.St
}

func GetHandler(lg logger.Lite, core *core.St, withCors bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// middlewares

	r.Use(dopHttps.MwRecovery(lg, nil))
	if withCors {
		r.Use(dopHttps.MwCors())
	}

	// handlers

	// doc
	r.GET("/doc/*any", ginSwag.WrapHandler(swagFiles.Handler, func(c *ginSwag.Config) {
		c.DefaultModelsExpandDepth = 0
		c.DocExpansion = "none"
	}))

	s := &St{lg: lg, core: core}

	// healthcheck
	r.GET("/healthcheck", func(c *gin.Context) { c.Status(http.StatusOK) })

	// main
	r.POST("/static", s.hSave)
	r.GET("/static/*any", s.hGet)

	// clean
	r.GET("/clean", s.hClean)

	return r
}
