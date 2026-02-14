package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

var (
	MixdownChain = []gin.HandlerFunc{
		BindForm[MixdownReq](),
		saveMixdownUploads(),
		Submit(),
		Convert(),
	}

	saveMixdownUploads = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			s := MustScope(c)
			req := MustReq[MixdownReq](c)

			dir, err := os.MkdirTemp(s.Deps.TmpRoot, "mixdown-*")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "mkdir temp failed", "detail": err.Error()})
				return
			}

			s.FailCleanup = func() { _ = os.RemoveAll(dir) }

			audioName := filepath.Base(req.Audio.Filename)
			bgmName := filepath.Base(req.BGM.Filename)

			audioPath := filepath.Join(dir, audioName)
			bgmPath := filepath.Join(dir, bgmName)

			if err := c.SaveUploadedFile(req.Audio, audioPath); err != nil {
				s.FailCleanup()
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "save audio failed", "detail": err.Error()})
				return
			}

			if err := c.SaveUploadedFile(req.BGM, bgmPath); err != nil {
				s.FailCleanup()
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "save bgm failed", "detail": err.Error()})
				return
			}

			s.Type = worker.Mixdown
			s.Payload = &worker.MixdownPayLoad{
				AudioPath: audioPath,
				BGMPath:   bgmPath,
				Filename:  strings.TrimSpace(req.Filename),
				Loop:      req.Loop,
			}

			c.Next()
		}
	}
)
