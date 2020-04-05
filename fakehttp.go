package fakehttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type JSONHandler struct {
	PathFmt       string
	Method        string
	RequestBody   interface{}
	ResponseCode  int
	ResponseFn    func(interface{}, []string, url.Values) (interface{}, error) `json:"-"`
	ErrResponseFn func(http.ResponseWriter, error, int)                        `json:"-"`
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
