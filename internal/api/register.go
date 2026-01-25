package api

import "github.com/gin-gonic/gin"

type Register interface {
	Register(*gin.Engine)
}
