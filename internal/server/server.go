package server

import "github.com/gin-gonic/gin"

func Start() error {
	r := gin.Default()
	return r.Run()
}