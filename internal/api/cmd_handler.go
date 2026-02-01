package api

import (
	"context"
	"net/http"
	"path/filepath"
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
	r.POST("/tts/all", h.ttsGenAll)
	r.POST("/tts/single", h.ttsGenSingle)

	r.POST("/mix", h.mixdown)
	r.POST("/concat", h.concat)

	r.POST("/render", h.render)

	r.POST("/merge", h.merge)

	r.POST("/brun", h.brun)
}

func (h *cmdHandler) brun(ctx *gin.Context) {
	var req BrunReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	taskID, err := h.worker.Submit(context.Background(), worker.Brun, &worker.BrunSubtitlePayLoad{
		VideoPath: req.VideoPath,
		SubPath:   req.SubPath,
		OutPath:   req.OutPath,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"taskID": taskID,
	})
}

func (h *cmdHandler) merge(ctx *gin.Context) {
	var req MergeReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	taskID, err := h.worker.Submit(context.Background(), worker.Merge, &worker.MergePayLoad{
		AudioPath: req.AudioPath,
		VideoPath: req.VideoPath,
		OutPath:   req.OutPath,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"taskID": taskID,
	})
}

func (h *cmdHandler) render(ctx *gin.Context) {
	var req RenderReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	taskID, err := h.worker.Submit(context.Background(), worker.Render, &worker.RenderPayLoad{
		Folder:  req.Folder,
		Dur:     req.Dur,
		TailCut: req.TailCut,
		Loop:    req.Loop,
		Out:     req.Out,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"taskID": taskID,
	})
}

func (h *cmdHandler) concat(ctx *gin.Context) {
	var req ConcatReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	taskID, err := h.worker.Submit(context.Background(), worker.Concat, &worker.ConcatPayLoad{
		Folder: req.Folder,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"taskID": taskID,
	})
}

func (h *cmdHandler) mixdown(ctx *gin.Context) {
	filename := strings.TrimSpace(ctx.PostForm("filename"))
	if filename == "" {
		ctx.JSON(400, gin.H{"error": "missing field: filename"})
		return
	}

	audio, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(400, gin.H{"error": "missing form file: audio", "detail": err.Error()})
		return
	}

	bgm, err := ctx.FormFile("bgm")
	if err != nil {
		ctx.JSON(400, gin.H{"error": "missing form file: bgm", "detail": err.Error()})
		return
	}

	audioPath := filepath.Join("/tmp/comp0ser", audio.Filename)
	bgmPath := filepath.Join("/tmp/comp0ser", bgm.Filename)

	if err := ctx.SaveUploadedFile(audio, audioPath); err != nil {
		ctx.JSON(500, gin.H{"error": "save audio failed", "detail": err.Error()})
		return
	}

	if err := ctx.SaveUploadedFile(bgm, bgmPath); err != nil {
		ctx.JSON(500, gin.H{"error": "save bgm failed", "detail": err.Error()})
		return
	}
	taskID, err := h.worker.Submit(ctx, worker.Mixdown, &worker.MixdownPayLoad{
		AudioPath: audioPath,
		BGMPath:   bgmPath,
		Filename:  filename,
		Loop:      true,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to submit to worker",
			"detail": err.Error(),
		})
	}

	ctx.JSON(200, gin.H{
		"taskID": taskID,
	})
}

func (h *cmdHandler) ttsGenSingle(ctx *gin.Context) {
	var req TTSGenSingleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}
	taskID, err := h.worker.Submit(ctx, worker.GenTTSSingle, &worker.GenTTSSinglePayLoad{
		Folder: req.Folder,
		NarID:  req.NarID,
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

func (h *cmdHandler) ttsGenAll(ctx *gin.Context) {
	var req TTSGenAllReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "bad json",
			"detail": err.Error(),
		})
		return
	}

	taskID, err := h.worker.Submit(ctx, worker.GenTTSAll, &worker.GenTTSPayLoad{
		Folder: req.Folder,
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

	taskID, err := h.worker.Submit(ctx, worker.GenScript, &worker.GenScriptPayLoad{
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
