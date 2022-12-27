package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/venus-tool/service"
	"github.com/ipfs-force-community/venus-tool/version"
)

func registerRoute(s *service.Service) http.Handler {
	router := gin.Default()
	router.Use(CorsMiddleWare())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "dashboard is developing...")
	})

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Version": version.Version})
	})

	apiV0Group := router.Group("/api/v0")

	chainGroup := apiV0Group.Group("/chain")
	chainGroup.GET("/head/", Wrap(s.ChainHead))

	msgGroup := apiV0Group.Group("/msg")
	msgGroup.GET("query", Wrap(s.MsgQuery))
	msgGroup.GET(":ID", Wrap(s.MsgQuery))
	msgGroup.POST("send", Wrap(s.MsgSend))
	msgGroup.POST("replace", Wrap(s.MsgReplace))

	addrGroup := apiV0Group.Group("/addr")
	addrGroup.GET("list", Wrap(s.AddrList))
	addrGroup.POST("operate", Wrap(s.AddrOperate))

	minerGroup := apiV0Group.Group("/miner")
	minerGroup.POST("create", Wrap(s.MinerCreate))
	storageAskGroup := minerGroup.Group("/ask/storage")
	storageAskGroup.GET("", Wrap(s.MinerGetStorageAsk))
	storageAskGroup.POST("", Wrap(s.MinerSetStorageAsk))
	retrievalAskGroup := minerGroup.Group("/ask/retrieval")
	retrievalAskGroup.GET("", Wrap(s.MinerGetRetrievalAsk))
	retrievalAskGroup.POST("", Wrap(s.MinerSetRetrievalAsk))

	dealGroup := apiV0Group.Group("/deal")
	dealGroup.GET("storage", Wrap(s.DealStorageList))
	dealGroup.GET("retrieval", Wrap(s.DealRetrievalList))
	dealGroup.POST("storage/state", Wrap(s.DealStorageUpdateState))

	sectorGroup := apiV0Group.Group("/sector")
	sectorGroup.POST("extend", Wrap(s.SectorExtend))
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
