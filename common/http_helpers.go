package common

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func JSONResponse(w http.ResponseWriter, respObj interface{}) {
	if err := JSONResponseC(w, respObj); err != nil {
		ErrorResponse(w, err)
	}
}

func JSONResponseC(w http.ResponseWriter, respObj interface{}) error {
	respJson, err := json.Marshal(respObj)
	if err != nil {
		return errors.Wrap(err, "failed to marshal JSON")
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respJson); err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func ErrorResponse(w http.ResponseWriter, err error) {
	// TODO: Except in debug mode, don't serve errors to clients.
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// --------------
// XXX: Code below this line was moved here from `cmd/publisherd/util.go` and probably needs some refactoring.  In
// particular, many things don't seem to be exported.
// --------------

type WebError interface {
	error

	StatusCode() int
}

type webError struct {
	status int
	msg    string
}

func (err *webError) StatusCode() int {
	return err.status
}

func (err *webError) Error() string {
	msg := err.msg
	if msg == "" {
		msg = http.StatusText(err.status)
	}
	return msg
}

var _ WebError = (*webError)(nil)

var (
	ErrMethodNotAllowed = &webError{status: http.StatusMethodNotAllowed}
)

type MyHandlerFunc func(w http.ResponseWriter, req *http.Request) (respObj interface{}, err WebError)
