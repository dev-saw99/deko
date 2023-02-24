package router

import (
	"net/http"

	"github.com/dev-saw99/deko/handler"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		{
			v1.GET("/ping", func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{
					"message": "pong",
				})
			})
		}

		ws := v1.Group("/ws")
		{
			ws.GET("/codecompile/:connid", handler.CodeCompiler)

		}
	}
	return router
}
