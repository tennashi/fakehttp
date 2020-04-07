package fakehttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// JSONHandler is a mock of an HTTP handler that sends and recieves JSON.
type JSONHandler struct {
	// PathFmt is a pattern of URL paths to bind a handler to.
	// See path.Match() for possible value patterns.  Skip the URL path check if
	// it is an empty string.
	PathFmt string
	// Method is an HTTP request method.  Skip the HTTP method check if it is an
	// empty string.
	Method string
	// RequestBody specifies the type to decode JSON of the HTTP request body.
	RequestBody interface{}
	// ResponseCode is an HTTP response code.
	ResponseCode int
	// ResponseFn is the function to return the response.
	// The first argument is the decoded JSON of the HTTP request body to the
	// value specified in RequestBody.
	// The second argument is an element of the URL path that matches the pattern
	// specified in PathFmt.  For example, If PathFmt is `/groups/*/users/*` and
	// the URL path is `/groups/1/users/2`, then `[]string{"1", "2"}`.
	// The third argument is a URL query parameter.
	// The return value is JSON encoded, so it must be a value that can be
	// specified as an argument to json.Marshal().
	ResponseFn func(interface{}, []string, url.Values) (interface{}, error) `json:"-"`
	// ErrResponseFn specifies how to return an error response.
	// If nil is specified, a JSON response encoded from the following type is
	// returned.
	// ```
	// type errorResponse struct {
	//   Message string
	//   Handler JSONHandler
	// }
	// ```
	ErrResponseFn func(http.ResponseWriter, error, int) `json:"-"`
}

func (h JSONHandler) checkPath(reqPath string) ([]string, error) {
	if h.PathFmt == "" {
		return strings.Split(reqPath, "/"), nil
	}
	ok, err := path.Match(h.PathFmt, reqPath)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unmatch path: want %v, got %v", h.PathFmt, reqPath)
	}

	params := []string{}
	r := strings.Split(reqPath, "/")
	pathFmt := strings.Split(h.PathFmt, "/")
	for i, p := range pathFmt {
		if strings.ContainsAny(p, "*?[]-\\^") {
			params = append(params, r[i])
		}
	}

	return params, nil
}

func (h JSONHandler) checkMethod(reqMethod string) error {
	if h.Method == "" {
		return nil
	}

	if h.Method != reqMethod {
		return fmt.Errorf("unmatch method: want %v, got %v", h.Method, reqMethod)
	}
	return nil
}

func (h JSONHandler) checkContentType(reqContentType string) error {
	if h.RequestBody == nil {
		return nil
	}
	if reqContentType != "application/json" {
		return fmt.Errorf("invalid Content-Type: want application/json, got %v", reqContentType)
	}
	return nil
}

// ServeHTTP is a method to implement http.Handler.
func (h JSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := h.checkPath(r.URL.Path)
	if err != nil {
		h.errorResponse(w, err, http.StatusNotFound)
		return
	}

	if err := h.checkMethod(r.Method); err != nil {
		h.errorResponse(w, err, http.StatusNotFound)
		return
	}

	if err := h.checkContentType(r.Header.Get("Content-Type")); err != nil {
		h.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	if h.RequestBody != nil {
		if err := json.NewDecoder(r.Body).Decode(h.RequestBody); err != nil {
			h.errorResponse(w, err, http.StatusBadRequest)
			return
		}
	}

	if h.ResponseFn == nil {
		h.ResponseFn = defaultResponseFn
	}
	res, err := h.ResponseFn(h.RequestBody, params, r.URL.Query())
	if err != nil {
		h.errorResponse(w, err, http.StatusBadRequest)
		return
	}
	if res != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(h.ResponseCode)
		json.NewEncoder(w).Encode(res)
	}
}

type errorResponse struct {
	Message string
	Handler JSONHandler
}

func (h JSONHandler) errorResponse(w http.ResponseWriter, err error, statusCode int) {
	if h.ErrResponseFn != nil {
		h.ErrResponseFn(w, err, statusCode)
		return
	}
	if err == nil {
		errRes := errorResponse{Message: "nil", Handler: h}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(errRes)
		return
	}

	errRes := errorResponse{Message: err.Error(), Handler: h}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errRes)
}

func defaultResponseFn(res interface{}, _ []string, _ url.Values) (interface{}, error) {
	return res, nil
}

// MultipleHandler is an HTTP handler for handling multiple
// fakehttp.JSONHandler.
type MultipleHandler struct {
	// ErrResponseFn specifies how to return an error response.
	ErrResponseFn func(http.ResponseWriter, error, int)

	handlers []JSONHandler
}

// NewMultipleHandler creates an instance of MultipleHandler.
// The JSONHandler argument specifies that the Method and PathFmt fields must
// not be empty strings.
// Requests are matched in the order of the array.
func NewMultipleHandler(hs []JSONHandler) *MultipleHandler {
	handlers := make([]JSONHandler, 0, len(hs))
	for _, h := range hs {
		if h.Method != "" && h.PathFmt != "" {
			handlers = append(handlers, h)
		}
	}
	return &MultipleHandler{
		handlers: handlers,
	}
}

// AddHandler adds a JSONHandler to mock.
// The JSONHandler argument specifies that the Method and PathFmt fields must
// not be empty strings.
func (h *MultipleHandler) AddHandler(handler JSONHandler) {
	if handler.Method == "" || handler.PathFmt == "" {
		return
	}
	if h.handlers == nil {
		h.handlers = []JSONHandler{handler}
		return
	}
	h.handlers = append(h.handlers, handler)
}

// ServeHTTP is a method to implement http.Handler.
func (h MultipleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, handler := range h.handlers {
		if handler.Method == r.Method {
			ok, err := path.Match(handler.PathFmt, r.URL.Path)
			if err != nil {
				h.errorResponse(w, err, http.StatusInternalServerError)
			}
			if ok {
				handler.ServeHTTP(w, r)
			}
		}
	}

	h.errorResponse(w, errors.New("not found"), http.StatusNotFound)
}

func (h MultipleHandler) errorResponse(w http.ResponseWriter, err error, statusCode int) {
	if h.ErrResponseFn != nil {
		h.ErrResponseFn(w, err, statusCode)
		return
	}

	if len(h.handlers) != 0 {
		h.handlers[0].errorResponse(w, err, statusCode)
	}

	if err == nil {
		w.WriteHeader(statusCode)
		w.Write([]byte{})
		return
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
}
