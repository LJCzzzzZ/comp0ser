package server

import (
	"net/http"
	"sync"

	"comp0ser/internal/worker"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	Worker  worker.Worker
	TmpRoot string
}

type Scope struct {
	Req  any
	Deps Deps

	Type    worker.TaskType
	Payload any
	TaskID  any

	FailCleanup func()
}

const scopeKey = "__server.scope__"

var scopePool = sync.Pool{
	New: func() any { return &Scope{} },
}

func PrepareScope(deps Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := scopePool.Get().(*Scope)
		*s = Scope{Deps: deps}
		c.Set(scopeKey, s)
		defer func() {
			scopePool.Put(s)
		}()

		c.Next()
	}
}

func BindJSON[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad json", "detail": err.Error()})
			return
		}
		MustScope(c).Req = &req
		c.Next()
	}
}

func BindForm[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.ShouldBind(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad form", "detail": err.Error()})
		}
		MustScope(c).Req = &req
		c.Next()
	}
}

func Submit() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := MustScope(c)
		taskID, err := s.Deps.Worker.Submit(c.Request.Context(), s.Type, s.Payload)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error", "detail": err.Error()})
			return
		}

		s.TaskID = taskID
		c.Next()
	}
}

func Convert() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"taskID": MustScope(c).TaskID})
	}
}

func MustScope(c *gin.Context) *Scope {
	v := c.MustGet(scopeKey)
	s, ok := v.(*Scope)
	if !ok || s == nil {
		panic("scope missing or type mismatch; did you forget to use ScopeInit?")
	}
	return s
}

func MustReq[T any](c *gin.Context) *T {
	s := MustScope(c)
	req, ok := s.Req.(*T)
	if !ok || req == nil {
		panic("req type mismatch or nil")
	}
	return req
}
