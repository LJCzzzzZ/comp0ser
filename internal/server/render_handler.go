package server

import (
	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	RenderChain = []gin.HandlerFunc{
		BindJSON[RenderReq](),
		preRender(),
		Submit(),
		Convert(),
	}

	preRender = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[RenderReq](c)
			s := MustScope(c)

			s.Type = worker.Render
			s.Payload = &worker.RenderPayLoad{
				Folder:  req.Folder,
				Dur:     req.Dur,
				TailCut: req.TailCut,
				Loop:    req.Loop,
				Out:     req.Out,
			}

			c.Next()
		}
	}
)
