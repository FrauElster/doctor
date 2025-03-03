package main

import (
	"fmt"
	"net/http"
	"runtime"
)

type ApiError struct {
	Status       int
	Code         string
	Message      string
	WrappedError error
	Origin       string
}

func (e ApiError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Code, e.Message)
	if e.WrappedError != nil {
		msg += fmt.Sprintf(" (%v)", e.WrappedError)
	}
	if e.Origin != "" {
		msg += " at " + e.Origin
	}

	return msg
}

func apiErrorFactory(status int, code, defaultMessage string) func(message string, err error) *ApiError {
	return func(message string, err error) *ApiError {
		e := &ApiError{Status: status, Code: code, Message: message, WrappedError: err}
		if message == "" {
			e.Message = defaultMessage
		}
		_, file, line, ok := runtime.Caller(1)
		if ok {
			e.Origin = fmt.Sprintf("%s:%d", file, line)
		}

		return e
	}
}

var (
	ErrParseJsonBody  = apiErrorFactory(http.StatusBadRequest, "parse_json_body", "Error parsing JSON body")
	ErrEncodeJsonBody = apiErrorFactory(http.StatusInternalServerError, "encode_json_body", "Error encoding JSON body")
	ErrInvalidUrl     = apiErrorFactory(http.StatusBadRequest, "invalid_url", "Invalid URL")
	ErrAddingTarget   = apiErrorFactory(http.StatusInternalServerError, "adding_target", "Error adding target")
	ErrRemovingTarget = apiErrorFactory(http.StatusInternalServerError, "removing_target", "Error removing target")
)
