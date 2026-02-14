package server

import (
	"net/http"
	"strings"

	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	GenScriptChain = []gin.HandlerFunc{
		BindJSON[GenScriptReq](),
		preGenScript(),
		Submit(),
		Convert(),
	}

	preGenScript = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			req := MustReq[GenScriptReq](c)

			req.RawText = strings.TrimSpace(req.RawText)
			if req.RawText == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad json"})
				return
			}

			if req.Segments <= 0 {
				req.Segments = 30
			}
			if req.MinChars <= 0 {
				req.MinChars = 200
			}
			if req.MaxChars <= 0 {
				req.MaxChars = 300
			}

			s := MustScope(c)
			s.Type = worker.GenScript
			s.Payload = &worker.GenScriptPayLoad{
				RawText:  req.RawText,
				Subject:  req.Subject,
				Segments: req.Segments,
				MinChars: req.MinChars,
				MaxChars: req.MaxChars,
				Focus:    req.Focus,
				Hook:     req.Hook,
				Model:    req.Model,
			}

			c.Next()
		}
	}
)
