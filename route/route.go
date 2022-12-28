package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs-force-community/venus-tool/version"
)

func registerRoute(s *service.Service) http.Handler {
	router := gin.New()
	router.Use(CorsMiddleWare())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "dashboard is developing...")
	})

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Version": version.Version})
	})

	apiV0Group := router.Group("/api/v0")
	apiV0Group.POST("send", Wrap(s.Send))

	return router
}

func CorsMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers",
			"DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,"+
				"If-Modified-Since,Cache-Control,Content-Type,Authorization,X-Forwarded-For,Origin,"+
				"X-Real-Ip,spanId,preHost,svcName")
		c.Header("Content-Type", "application/json")
		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}
		c.Next()
	}
}
