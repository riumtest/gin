// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"math"
	"net/http"
	"sync"
)

// abortIndex represents a typical value over 1<<7 (128), which is the maximum
// number of handlers that can be registered in a single chain.
const abortIndex int8 = math.MaxInt8 >> 1

// Context is the most important part of gin. It allows us to pass variables
// between middleware, manage the flow, validate the JSON of a request and
// render a JSON response for example.
type Context struct {
	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter

	Params   Params
	handlers HandlersChain
	index    int8
	fullPath string

	engine       *Engine
	params       *Params
	skippedNodes *[]skippedNode

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs

	// Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// queryCache caches the query result from c.Request.URL.Query().
	queryCache url.Values

	// formCache caches c.Request.PostForm, which contains the parsed form data
	// from POST, PUT and PATCH body parameters.
	formCache url.Values

	// SameSite allows a server to define a cookie attribute making it impossible
	// for the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite
}

// reset resets the context to its initial state, used by the engine to
// recycle context objects from the pool.
func (c *Context) reset() {
	c.Writer = &c.writermem
	c.Params = c.Params[:0]
	c.handlers = nil
	c.index = -1
	c.fullPath = ""
	c.Keys = nil
	c.Errors = c.Errors[:0]
	c.Accepted = nil
	c.queryCache = nil
	c.formCache = nil
	c.sameSite = 0
	*c.params = (*c.params)[:0]
	*c.skippedNodes = (*c.skippedNodes)[:0]
}

// Copy returns a copy of the current context that can be safely used outside
// the request's scope. This must be used when the context has to be passed to
// a goroutine.
func (c *Context) Copy() *Context {
	cp := Context{
		writermem: c.writermem,
		Request:   c.Request,
		engine:    c.engine,
	}
	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem
	cp.index = abortIndex
	cp.handlers = nil
	cp.fullPath = c.fullPath

	cCopy := make(Params, len(c.Params))
	copy(cCopy, c.Params)
	cp.Params = cCopy

	c.mu.RLock()
	cp.Keys = make(map[string]any, len(c.Keys))
	for k, v := range c.Keys {
		cp.Keys[k] = v
	}
	c.mu.RUnlock()

	return &cp
}

// HandlerName returns the main handler's name. For example if the handler
// is "handleGetUsers()", this function will return "main.handleGetUsers".
func (c *Context) HandlerName() string {
	return nameOfFunction(c.handlers.Last())
}

// HandlerNames returns a list of all registered handlers for this context
// in descending order, following the semantics of HandlerName().
func (c *Context) HandlerNames() []string {
	hn := make([]string, 0, len(c.handlers))
	for _, val := range c.handlers {
		hn = append(hn, nameOfFunction(val))
	}
	return hn
}

// Handler returns the main handler.
func (c *Context) Handler() HandlerFunc {
	return c.handlers.Last()
}

// FullPath returns a matched route's full path. For not found routes
// returns an empty string.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func (c *Context) FullPath() string {
	return c.fullPath
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not
// stop the current handler. Let's say you have an authorization middleware
// that validates that the current request is authorized. If the authorization
// fails (ex: the password does not match), call Abort to ensure the remaining
// handlers for this request are not called.
func (c *Context) Abort() {
	c.index = abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified
// status code. For example, a failed attempt to authenticate a request could
// use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Writer.WriteHeaderNow()
	c.Abort()
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}
