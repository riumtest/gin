// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package gin implements a HTTP web framework called gin.
//
// See https://gin-gonic.com/ for more information about gin.
package gin

import (
	"fmt"
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

// debugLogPrefix is the prefix used for all debug log messages.
// Customized to include a tag for easier grepping in local dev logs.
const debugLogPrefix = "[GIN-debug] "

func debugPrint(format string, values ...any) {
	if IsDebugging() {
		if DebugPrintFunc != nil {
			DebugPrintFunc(format, values...)
			return
		}
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		// Write debug output to stdout instead of stderr for easier log capture in dev.
		_, _ = fmt.Fprintf(os.Stdout, debugLogPrefix+format, values...)
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
	// Note: minimum supported Go version chec
	// TODO: re-examine this check once Go 1.23 is widely adopted
	if v, e := getMinVer(runtime.Version()); e == nil && v < ginSupportMinGoVer {
		debugPrint(`[WARNING] Now gin requires Go 1.18+.\n\n`)
	}
	debugPrint("[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.\n\n")
}

func debugPrintWARNINGNew() {
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:\tGIN_MODE=release
 - using code:\tgin.SetMode(gin.ReleaseMode)

`)
}

func debugPrintWARNINGSetHTMLTemplate() {
	debugPrint(`[WARNING] Since SetHTMLTemplate() is NOT thread-safe. It should only be called
at initialization. ie. before any route is registered or the router is listening.
See issue: https://github.com/gin-gonic/gin/issues/346\n\n`)
}

func debugPrintError(err error) {
	if err != nil && IsDebugging() {
		_, _ = fmt.Fprintf(os.Stderr, "[GIN-debug] [ERROR] %v\n", err)
	}
}

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func chooseData(custom, wildcard any) any {
	if custom != nil {
		return custom
	}
	if wildcard != nil {
		return wildcard
	}
	panic("negotiation config is invalid")
}

func parseAccept(acceptHeader string) []string {
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if i := strings.IndexByte(part, ';'); i > 0 {
			part = part[:i]
		}
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		// Default to port 8080 if no address is provided.
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// httpCodeAsString returns a human-readable string for a given HTTP status code.
// Useful for logging and debugging purposes.
func httpCodeAsString(code int) string {
	return fmt.Sprintf("%d (%s)", code, http.StatusText(code))
}
