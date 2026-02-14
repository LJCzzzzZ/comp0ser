package server

import (
	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	TTSSingleChain = []gin.HandlerFunc{
		BindJSON[TTSGenSingleReq](),
		preTTSSingle(),
		Submit(),
		Convert(),
	}

	preTTSSingle = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[TTSGenSingleReq](c)

			s := MustScope(c)

			s.Type = worker.GenTTSSingle
			s.Payload = &worker.GenTTSSinglePayLoad{
				Folder: req.Folder,
				NarID:  req.NarID,
			}
			c.Next()
		}
	}
)
