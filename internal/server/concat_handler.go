package server

import (
	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	ConcatChain = []gin.HandlerFunc{
		BindJSON[ConcatReq](),
		preConcat(),
		Submit(),
		Convert(),
	}

	preConcat = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[ConcatReq](c)
			s := MustScope(c)

			s.Type = worker.Concat
			s.Payload = &worker.ConcatPayLoad{
				Folder: req.Folder,
			}

			c.Next()
		}
	}
)
