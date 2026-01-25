package api

import (
	"net/http"
	"strings"

	"comp0ser/worker"

	"github.com/gin-gonic/gin"
)

type cmdHandler struct {
	worker worker.Worker
}

func NewCmdHandler(worker worker.Worker) Register {
	return &cmdHandler{
		worker: worker,
	}
}

func (h *cmdHandler) Register(r *gin.Engine) {
	r.POST("/gen", h.genScript)
	r.POST("/tts/genall", h.ttsGenAll)
}

func (h *cmdHandler) ttsGenAll(ctx *gin.Context) {
	var req TTSGenAllReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	taskID, err := h.worker.Submit(ctx, worker.GenTTSAll, &worker.GenTTSPlayLoad{
		FileID: req.FileID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"taskId": taskID,
	})
}

func (h *cmdHandler) genScript(ctx *gin.Context) {
	var req GenScriptReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	req.RawText = strings.TrimSpace(req.RawText)
	if req.RawText == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "rawText is empty"})
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

	taskID, err := h.worker.Submit(ctx, worker.GenScript, &worker.GenScriptPlayLoad{
		RawText:  req.RawText,
		Subject:  req.Subject,
		Segments: req.Segments,
		MinChars: req.MinChars,
		MaxChars: req.MaxChars,
		Focus:    req.Focus,
		Hook:     req.Hook,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"taskId": taskID,
	})
}
