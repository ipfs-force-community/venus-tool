package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs-force-community/venus-tool/version"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("route")

func registerRoute(s *service.ServiceImpl) http.Handler {
	router := gin.Default()
	router.Use(corsMiddleWare())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "dashboard is developing...")
	})

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Version": version.Version})
	})

	apiV0Group := router.Group("/api/v0")
	Register(apiV0Group, s, service.IServiceStruct{}.Internal)

	return router
}

func corsMiddleWare() gin.HandlerFunc {
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
