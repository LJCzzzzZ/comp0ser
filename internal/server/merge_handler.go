package server

import (
	"fmt"

	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	MergeChain = []gin.HandlerFunc{
		BindJSON[MergeReq](),
		preMerge(),
		Submit(),
		Convert(),
	}

	preMerge = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[MergeReq](c)
			fmt.Println(req)
			s := MustScope(c)

			s.Type = worker.Merge
			s.Payload = &worker.MergePayLoad{
				AudioPath: req.AudioPath,
				VideoPath: req.VideoPath,
				OutPath:   req.OutPath,
			}

			c.Next()
		}
	}
)

