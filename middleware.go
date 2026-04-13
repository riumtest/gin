// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"net/http"
	"time"
)

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. i.e. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// Logger returns a middleware that logs requests.
// It logs the method, path, status code, latency, client IP, and error message (if any).
func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		debugPrint("[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
// The stack trace is also logged.
func Recovery() HandlerFunc {
	return RecoveryWithWriter()
}

// RecoveryWithWriter returns a middleware that recovers from any panics.
// If a panic occurs, it logs the error and returns a 500 Internal Server Error response.
func RecoveryWithWriter() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				debugPrint("[Recovery] panic recovered: %v\n", err)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// BasicAuth returns a middleware that implements HTTP Basic Authentication.
// It takes a map of username/password pairs. If authentication fails,
// it responds with 401 Unauthorized.
func BasicAuth(accounts map[string]string) HandlerFunc {
	return func(c *Context) {
		user, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expectedPassword, ok := accounts[user]
		if !ok || expectedPassword != password {
			c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// CORS returns a middleware that adds Cross-Origin Resource Sharing headers.
// It allows all origins, methods, and headers by default.
func CORS() HandlerFunc {
	return func(c *Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
