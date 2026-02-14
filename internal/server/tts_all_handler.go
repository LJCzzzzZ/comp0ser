package server

import (
	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	TTSAllChain = []gin.HandlerFunc{
		BindJSON[TTSGenAllReq](),
		preTTSAll(),
		Submit(),
		Convert(),
	}

	preTTSAll = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[TTSGenAllReq](c)

			s := MustScope(c)
			s.Type = worker.GenTTSAll
			s.Payload = &worker.GenTTSPayLoad{
				Folder: req.Folder,
			}
			c.Next()
		}
	}
)
