//go:build go1.22

// Package main provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/runtime"
)

// Error defines model for Error.
type Error struct {
	// Code Error code static for the error type
	Code string `json:"code"`

	// Message Error message with further details
	Message string `json:"message"`
}

// HealthCheckResult defines model for HealthCheckResult.
type HealthCheckResult struct {
	// DurationSeconds Duration of the health check in seconds
	DurationSeconds float32 `json:"duration_seconds"`

	// Error Error message if the health check failed
	Error *string `json:"error,omitempty"`

	// Healthy Whether the target is considered healthy
	Healthy bool `json:"healthy"`

	// Id Target identifier
	Id string `json:"id"`

	// Status HTTP status code from the health check
	Status int `json:"status"`

	// Timestamp When the health check was performed
	Timestamp time.Time `json:"timestamp"`

	// Url The monitored URL
	Url string `json:"url"`
}

// Target defines model for Target.
type Target struct {
	// Id Unique identifier for the target
	Id string `json:"id"`

	// Url The URL to be monitored
	Url string `json:"url"`
}

// BadRequest defines model for BadRequest.
type BadRequest = Error

// InternalServerError defines model for InternalServerError.
type InternalServerError = Error

// NotAllowed defines model for NotAllowed.
type NotAllowed = Error

// NotFound defines model for NotFound.
type NotFound = Error

// RegisterTargetJSONRequestBody defines body for RegisterTarget for application/json ContentType.
type RegisterTargetJSONRequestBody = Target

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get the health status of the API
	// (GET /health)
	GetHealth(w http.ResponseWriter, r *http.Request)
	// Register a new URL for health checking
	// (POST /register)
	RegisterTarget(w http.ResponseWriter, r *http.Request)
	// Get health check status for all registered targets
	// (GET /status)
	GetStatus(w http.ResponseWriter, r *http.Request)
	// Unregister a URL from health checking
	// (DELETE /unregister/{id})
	UnregisterTarget(w http.ResponseWriter, r *http.Request, id string)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetHealth operation middleware
func (siw *ServerInterfaceWrapper) GetHealth(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetHealth(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// RegisterTarget operation middleware
func (siw *ServerInterfaceWrapper) RegisterTarget(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.RegisterTarget(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetStatus operation middleware
func (siw *ServerInterfaceWrapper) GetStatus(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetStatus(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// UnregisterTarget operation middleware
func (siw *ServerInterfaceWrapper) UnregisterTarget(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameterWithOptions("simple", "id", r.PathValue("id"), &id, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UnregisterTarget(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{})
}

// ServeMux is an abstraction of http.ServeMux.
type ServeMux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type StdHTTPServerOptions struct {
	BaseURL          string
	BaseRouter       ServeMux
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, m ServeMux) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseRouter: m,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, m ServeMux, baseURL string) http.Handler {
	return HandlerWithOptions(si, StdHTTPServerOptions{
		BaseURL:    baseURL,
		BaseRouter: m,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options StdHTTPServerOptions) http.Handler {
	m := options.BaseRouter

	if m == nil {
		m = http.NewServeMux()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	m.HandleFunc("GET "+options.BaseURL+"/health", wrapper.GetHealth)
	m.HandleFunc("POST "+options.BaseURL+"/register", wrapper.RegisterTarget)
	m.HandleFunc("GET "+options.BaseURL+"/status", wrapper.GetStatus)
	m.HandleFunc("DELETE "+options.BaseURL+"/unregister/{id}", wrapper.UnregisterTarget)

	return m
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8RX32/cNgz+VwRtj1583VKg8Fu6Xw2QDcE1wR6KYFAs+qxOllyKTnAI/L8PlGyfr3aW",
	"tGu3p7Nj6iP58SOpPMjSN6134CjI4kEihNa7APHltdJb+NBBIH4rvSNw8VG1rTWlIuNd/j54x38LZQ2N",
	"4qdvESpZyG/yA3Sevob8Z0SPsu/7TGoIJZqWQWTBvgQOzvpMnjsCdMq+BbwDTKe+egyjUxGiVwHJMJO/",
	"ezqz1t+D/vpB/AZUey2cJ6EGnymCX3zn/gP/Wwi+wxJiBFX0yUbDOYadqtGibwHJJLWUXgP/HsNFY8Hf",
	"RCBFphSVR0E1JHYF7VuQmYw/hQyExu044QZCULtHAYfP4t5QLaoOqQYUGkgZG5ZwfSZZWwa5gO9SpAcX",
	"N5O9v30PZdTfG1CW6h9rKP/aQugsLfPVHUbi/wxQeqfDMtSfBgvhq5hxHUFFyajCODEezGTlsVEkC1lZ",
	"r+iQgOuaW4gShJH0f2LDrPiplLGg1yhOZvsl6B81RD4ZixTugIQJovQuGA0IWownJ9Bb7y0ox6hGLwGv",
	"BhANjkxlANfCYXl0Kyy+ubq6FOlj0lGFvlnkeUA0jmCXSCPTQCDVtKs5uiVZ9yqIFpDLETmb6qIVwXcM",
	"txZ5h3Yl5xpE450hz4xdby/mcB2aJ1VqOACGnqg5lGyeWrZU4pqiUw2WMl6r17UzHzqY1Wtq2iSHTyLh",
	"enshyIvbGR2fSkViweiVxNjUuMovfZ9dnsewEXYmEDCwUE6PUfArhzavfxwdhiw7SBNAxBEAKM4uz2Um",
	"7wBDQn9xsjl5wXn7FpxqjSzkDyebk43MZKuojtTmCZsfB+qZ+Fiqcy0L+StQ8iKz48X7/eZ0PR8Tpt7r",
	"M/lys3lszE9w+domjQO9axqF+xTGvBOGVhtmFufN5vlIY1SQDyvpbAeLq1Ejwz5/7fX+i+2tAbw/lghh",
	"B/2CxM2jkyh0ZQkhVJ21+0kgadOePofU2c0oHnn59JHZHeLLlW6kXCjh4D7qmSU/1/TYT/lhvj4mxrfj",
	"mFnj8dnVMwRNeKqMy/3aT62tENV+7WpyYQKxLo9GNsbj4bPqsGiDI+ShEZhQZe1MJsMUDInXzo1f8gej",
	"+6Q5CwRLjq8n06lHWoWqAQIMsni3Nj67xSweOnPYzOTFIQKeknyQR5DMpFNN3Ihaftws2axgH8/em89t",
	"pEMc/6KVTp9VwnQZ/l9771BMoVLn8b1k2Xp8KIKsVfjCl8oKDXdgfduAo+F/j2H1F7Imaos8t2xX+0DF",
	"q82rjexv+r8DAAD//7AE+ne6DQAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
