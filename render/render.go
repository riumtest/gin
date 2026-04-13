// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package render provides HTTP response rendering utilities for the Gin framework.
// It supports multiple response formats including JSON, XML, HTML, and plain text.
package render

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
)

// Render is the interface that wraps the Render method.
// All response renderers must implement this interface.
type Render interface {
	// Render writes data with custom ContentType to http.ResponseWriter.
	Render(w http.ResponseWriter) error
	// WriteContentType writes custom ContentType to http.ResponseWriter.
	WriteContentType(w http.ResponseWriter)
}

// JSON contains the given interface object and renders it as JSON.
type JSON struct {
	Data any
}

// XML contains the given interface object and renders it as XML.
type XML struct {
	Data any
}

// String contains the given string and its format data.
type String struct {
	Format string
	Data   []any
}

// HTMLTemplate renders an HTML template with the given name, data, and template set.
type HTMLTemplate struct {
	Template *template.Template
	Name     string
	Data     any
}

// Redirect renders an HTTP redirect response.
type Redirect struct {
	Code     int
	Request  *http.Request
	Location string
}

const (
	contentTypeJSON = "application/json; charset=utf-8"
	contentTypeXML  = "application/xml; charset=utf-8"
	contentTypeHTML = "text/html; charset=utf-8"
	contentTypeText = "text/plain; charset=utf-8"
)

func writeContentType(w http.ResponseWriter, value string) {
	header := w.Header()
	if ct := header.Get("Content-Type"); ct == "" {
		header.Set("Content-Type", value)
	}
}

// WriteContentType sets the Content-Type header for JSON responses.
func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, contentTypeJSON)
}

// Render encodes the given data as JSON and writes it to the response.
func (r JSON) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

// WriteContentType sets the Content-Type header for XML responses.
func (r XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, contentTypeXML)
}

// Render encodes the given data as XML and writes it to the response.
func (r XML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	return xml.NewEncoder(w).Encode(r.Data)
}

// WriteContentType sets the Content-Type header for plain text responses.
func (r String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, contentTypeText)
}

// Render formats the string with the given data and writes it to the response.
func (r String) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if len(r.Data) > 0 {
		_, err := fmt.Fprintf(w, r.Format, r.Data...)
		return err
	}
	_, err := fmt.Fprint(w, r.Format)
	return err
}

// WriteContentType sets the Content-Type header for HTML responses.
func (r HTMLTemplate) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, contentTypeHTML)
}

// Render executes the HTML template and writes the result to the response.
func (r HTMLTemplate) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if r.Name == "" {
		return r.Template.Execute(w, r.Data)
	}
	return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

// WriteContentType is a no-op for redirect responses.
func (r Redirect) WriteContentType(_ http.ResponseWriter) {}

// Render performs the HTTP redirect by sending the appropriate status code and Location header.
func (r Redirect) Render(w http.ResponseWriter) error {
	if (r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		return fmt.Errorf("cannot redirect with status code %d", r.Code)
	}
	http.Redirect(w, r.Request, r.Location, r.Code)
	return nil
}
