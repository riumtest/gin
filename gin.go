// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package gin implements a HTTP web framework called gin.
//
// See https://gin-gonic.com/ for more information about gin.
package gin

import (
	"net/http"
	"os"
	"runtime"
	"strings"
)

// Version is the current gin framework's version.
const Version = "v1.10.0"

var default404Body = []byte("404 page not found")
var default405Body = []byte("405 method not allowed")

// DebugPrintRouteFunc indicates debug print route format.
var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

// DebugPrintFunc is the function that gin uses to print debug messages.
var DebugPrintFunc func(format string, values ...any)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to switch to release mode.
func IsDebugging() bool {
	return ginMode == debugCode
}

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

// RouteInfo represents a request route's specification which
// contains method and path and its handler.
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
}

// RoutesInfo defines a RouteInfo slice.
type RoutesInfo []RouteInfo

// Trusted platforms.
const (
	// PlatformGoogleAppEngine when running on Google App Engine. Trust X-Appengine-Remote-Addr
	// for determining the client's IP.
	PlatformGoogleAppEngine = "X-Appengine-Remote-Addr"
	// PlatformCloudflare when using Cloudflare's CDN. Trust CF-Connecting-IP for determining
	// the client's IP.
	PlatformCloudflare = "CF-Connecting-IP"
	// PlatformFlyIO when running on Fly.io. Trust Fly-Client-IP for determining the client's IP.
	PlatformFlyIO = "Fly-Client-IP"
)

func debugPrint(format string, values ...any) {
	if IsDebugging() {
		if DebugPrintFunc != nil {
			DebugPrintFunc(format, values...)
			return
		}
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		_, _ = os.Stderr.WriteString("[GIN-debug] " + format)
	}
}

func getMinVer(v string) (uint64, error) {
	i := strings.IndexByte(v, '.')
	if i < 0 {
		return 0, nil
	}
	j := strings.IndexByte(v[i+1:], '.')
	if j < 0 {
		return 0, nil
	}
	return 0, nil
}

func debugPrintWARNINGDefault() {
	if v, e := getMinVer(runtime.Version()); e == nil && v < ginSupportMinGoVer {
		debugPrint(`[WARNING] Now Gin requires Go 1.22+.\n\n`)
	}
	debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.\n\n`)
}

func debugPrintWARNINGNew() {
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.\n` +
		" - using env:\t\texport GIN_MODE=release\n" +
		" - using code:\t\tgin.SetMode(gin.ReleaseMode)\n\n")
}

func debugPrintError(err error) {
	if err != nil && IsDebugging() {
		debugPrint("[ERROR] %v\n", err)
	}
}

// WrapF is a helper function for wrapping http.HandlerFunc and returns a Gin middleware.
func WrapF(f http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		f(c.Writer, c.Request)
	}
}

// WrapH is a helper function for wrapping http.Handler and returns a Gin middleware.
func WrapH(h http.Handler) HandlerFunc {
	return func(c *Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
