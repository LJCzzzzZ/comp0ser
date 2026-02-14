package server

import (
	"context"
	"net/http"

	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

func Routes(ctx context.Context, wk worker.Worker, tmpRoot string) http.Handler {
	mux := gin.Default()

	mux.Use(PrepareScope(Deps{Worker: wk, TmpRoot: tmpRoot}))

	mux.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusAccepted, gin.H{"msg": "pong"})
	})

	// narration
	mux.POST("/gen", GenScriptChain...)

	// tts
	mux.POST("/tts/single", TTSSingleChain...)
	mux.POST("/tts/all", TTSAllChain...)

	// ffmpeg audio
	mux.POST("/mix", MixdownChain...)
	mux.POST("/concat", ConcatChain...)

	// ffmpeg video
	mux.POST("/render", RenderChain...)
	mux.POST("/merge", MergeChain...)

	mux.POST("/subtitle", GenSubtitleChain...)
	mux.POST("/brun", BrunChain...)

	return mux
}
