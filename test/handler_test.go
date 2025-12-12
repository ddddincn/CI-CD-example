package test

import (
	"cicd_example/handler"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHelloWorldEndpoint(t *testing.T) {
	// 设置 gin 为测试模式
	gin.SetMode(gin.TestMode)
	// 创建一个 Handler 实例
	h := handler.NewHandler()
	// 创建一个新的 gin 引擎实例并设置路由
	engine := gin.New()
	engine.GET("/", h.HelloWorld)

	// 创建一个测试请求
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// 执行请求
	engine.ServeHTTP(w, req)

	// 验证响应状态码
	assert.Equal(t, 200, w.Code)

	// 验证响应内容
	expected := `{"message":"Hello World"}`

	assert.JSONEq(t, expected, w.Body.String())
}
