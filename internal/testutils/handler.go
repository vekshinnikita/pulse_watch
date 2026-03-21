package testutils

import (
	"bytes"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

type RequestHeader struct {
	Key   string
	Value string
}

type MakeRequestOptions struct {
	Headers     []RequestHeader
	Middlewares []gin.HandlerFunc
	Handlers    []gin.HandlerFunc
	Body        string
}

func MakeTestRequest(options *MakeRequestOptions) *httptest.ResponseRecorder {
	// Test Server
	r := gin.New()
	r.POST("/test", options.Handlers...)
	r.Use(options.Middlewares...)

	// Test Request
	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/test",
		bytes.NewBufferString(options.Body),
	)

	// Set headers
	for _, header := range options.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	// Perform request
	r.ServeHTTP(w, req)

	return w
}
