package server

import (
	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	GenSubtitleChain = []gin.HandlerFunc{
		BindJSON[GenSubtitleReq](),
		preGenSubtitle(),
		Submit(),
		Convert(),
	}

	preGenSubtitle = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[GenSubtitleReq](c)

			s := MustScope(c)
			s.Type = worker.GenSrt
			s.Payload = &worker.GenSubtitlePayload{
				AudioPath:  req.AudioPath,
				OutputPath: req.OutputPath,
				Lang:       req.Lang,
			}
			c.Next()
		}
	}
)
