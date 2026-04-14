// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package binding provides request binding utilities for decoding
// HTTP request data (JSON, XML, form, query, etc.) into Go structs.
package binding

import (
	"net/http"
)

// Content-Type MIME types for various data formats.
const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEProtobuf          = "application/x-protobuf"
	MIMEMsgpack           = "application/x-msgpack"
	MIMEMsgpack2          = "application/msgpack"
	MIMEYAML              = "application/x-yaml"
	MIMEYAMLText          = "text/yaml"
	MIMETOML              = "application/toml"
)

// Binding describes the interface for binding request data into a struct.
// Implementations should be stateless and safe for concurrent use.
type Binding interface {
	// Name returns the name of the binding.
	Name() string
	// Bind reads data from the request and populates the given struct.
	Bind(req *http.Request, obj any) error
}

// BindingBody extends Binding with the ability to bind from raw bytes.
type BindingBody interface {
	Binding
	BindBody(body []byte, obj any) error
}

// BindingUri describes the interface for binding URI parameters.
type BindingUri interface {
	Name() string
	BindUri(values map[string][]string, obj any) error
}

// StructValidator is the minimal interface for struct validation.
// A custom validator can be set via binding.Validator.
type StructValidator interface {
	// ValidateStruct validates the given struct and returns an error
	// if any field fails its validation constraint.
	ValidateStruct(obj any) error
	// Engine returns the underlying validation engine.
	Engine() any
}

// Validator is the default validator used by gin bindings.
// It can be replaced with a custom implementation.
var Validator StructValidator = &defaultValidator{}

// Predefined binding instances for common content types.
var (
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	Form          = formBinding{}
	Query         = queryBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = formMultipartBinding{}
	Header        = headerBinding{}
)

// Default returns the appropriate Binding implementation based on
// the HTTP method and the request's Content-Type header.
// Note: HEAD requests are treated the same as GET (query/form binding).
func Default(method, contentType string) Binding {
	if method == http.MethodGet || method == http.MethodDelete || method == http.MethodHead {
		return Form
	}

	switch contentType {
	case MIMEHTML:
		return Form
	case MIMEXML, "text/xml":
		return XML
	case MIMEPOSTForm:
		return FormPost
	case MIMEMultipartPOSTForm:
		return FormMultipart
	default: // case MIMEJSON:
		return JSON
	}
}

// validate runs the struct validator against obj if obj is a struct or
// pointer to a struct. It is a no-op for ot
