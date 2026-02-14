package server

import (
	"fmt"

	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	BrunChain = []gin.HandlerFunc{
		BindJSON[BrunReq](),
		preBrun(),
		Submit(),
		Convert(),
	}

	preBrun = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[BrunReq](c)
			s := MustScope(c)

			s.Type = worker.Brun
			fmt.Println(req)
			s.Payload = &worker.BrunSubtitlePayLoad{
				VideoPath:    req.VideoPath,
				SubtitlePath: req.SubtitlePath,
				OutputPath:   req.OutputPath,
			}

			c.Next()
		}
	}
)
