package main

import (
	"cicd_example/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	h := handler.NewHandler()
	engine := gin.Default()
	engine.GET("/", h.HelloWorld)
	if err := engine.Run(":8888"); err != nil {
		panic(err)
	}
}
