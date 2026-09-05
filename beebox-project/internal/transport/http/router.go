package httpapi

import "github.com/gin-gonic/gin"

func New() *gin.Engine {
	engine := gin.Default()
	engine.Use(ErrorHandler())

	engine.GET("/healthz", HealthHandler)

	return engine
}