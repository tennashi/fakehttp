package fakehttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestJSONHandler_checkPath_unmatchedPathes(t *testing.T) {
	type input struct {
		pathFmt   string
		reqPathes []string
	}

	cases := []struct {
		input input
	}{
		{
			input: input{
				pathFmt: "/users/1",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/1/",
					"/users/2",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/*",
				reqPathes: []string{
					"/users",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/?",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/11",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/[1]",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/2",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/\\1",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/2",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/[0-9]",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/11",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/[^a-z]",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/a",
					"/users/11",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/[123]",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/a",
					"/users/11",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/[\\123]",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/a",
					"/users/11",
					"/users/hoge",
					"/users/1/",
					"hoge",
				},
			},
		},
		{
			input: input{
				pathFmt: "/users/[hjk]oge",
				reqPathes: []string{
					"/users",
					"/users/",
					"/users/h",
					"/users/hogehoge",
					"/users/11",
					"/users/1/",
					"hoge",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.input.pathFmt, func(t *testing.T) {
			h := JSONHandler{PathFmt: tt.input.pathFmt}
			for _, reqPath := range tt.input.reqPathes {
				t.Run(reqPath, func(t *testing.T) {
					got, err := h.checkPath(reqPath)
					if err == nil {
						t.Fatalf("should be error, but not")
					}
					if len(got) != 0 {
						t.Fatalf("want length 0, but got: %v", got)
					}
				})
			}
		})
	}
}

func TestJSONHandler_checkPath_parsedParams(t *testing.T) {
	type input struct {
		pathFmt string
		reqPath string
	}

	cases := []struct {
		input input
		want  []string
	}{
		{
			input: input{pathFmt: "", reqPath: "/users/1"},
			want:  []string{"", "users", "1"},
		},
		{
			input: input{pathFmt: "/users/1", reqPath: "/users/1"},
			want:  []string{},
		},
		{
			input: input{pathFmt: "/users/*", reqPath: "/users/"},
			want:  []string{""},
		},
		{
			input: input{pathFmt: "/users/*", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/?", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/[1]", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/\\1", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/[0-9]", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/[^a-z]", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/[123]", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/[\\123]", reqPath: "/users/1"},
			want:  []string{"1"},
		},
		{
			input: input{pathFmt: "/users/[hjk]oge", reqPath: "/users/hoge"},
			want:  []string{"hoge"},
		},
		{
			input: input{pathFmt: "/groups/*/users/*", reqPath: "/groups/testgroup/users/1"},
			want:  []string{"testgroup", "1"},
		},
		{
			input: input{pathFmt: "/hoge/*/*", reqPath: "/hoge/1/abc"},
			want:  []string{"1", "abc"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.input.pathFmt, func(t *testing.T) {
			h := JSONHandler{PathFmt: tt.input.pathFmt}
			got, err := h.checkPath(tt.input.reqPath)
			if err != nil {
				t.Fatalf("should not be error, but: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("want %v, but got: %v", tt.want, got)
			}
		})
	}
}

func TestJSONHandler_checkMethod(t *testing.T) {
	type input struct {
		wantMethod string
		reqMethod  string
	}

	cases := []struct {
		input input
		err   bool
	}{
		{input: input{wantMethod: "GET", reqMethod: "GET"}, err: false},
		{input: input{wantMethod: "GET", reqMethod: "POST"}, err: true},
		{input: input{wantMethod: "HOGE", reqMethod: "HOGE"}, err: false},
		{input: input{wantMethod: "HOGE", reqMethod: "FUGA"}, err: true},
	}

	for _, tt := range cases {
		t.Run(tt.input.wantMethod+"/"+tt.input.reqMethod, func(t *testing.T) {
			h := JSONHandler{Method: tt.input.wantMethod}
			err := h.checkMethod(tt.input.reqMethod)
			if !tt.err && err != nil {
				t.Fatalf("should not be error, but: %v", err)
			}
			if tt.err && err == nil {
				t.Fatalf("should be error, but not")
			}
		})
	}
}

func TestJSONHandler_checkContentType(t *testing.T) {
	cases := []struct {
		input string
		err   bool
	}{
		{input: "application/json", err: false},
		{input: "application/xml", err: true},
		{input: "hoge", err: true},
	}
	for _, tt := range cases {
		t.Run("nil_request_body/"+tt.input, func(t *testing.T) {
			h := JSONHandler{RequestBody: "not-nil"}
			err := h.checkContentType(tt.input)
			if !tt.err && err != nil {
				t.Fatalf("should not be error, but: %v", err)
			}
			if tt.err && err == nil {
				t.Fatalf("should be error, but not")
			}
		})
	}

	cases = []struct {
		input string
		err   bool
	}{
		{input: "application/json", err: false},
		{input: "application/xml", err: false},
		{input: "hoge", err: false},
	}
	for _, tt := range cases {
		t.Run("not_nil_request_body/"+tt.input, func(t *testing.T) {
			h := JSONHandler{RequestBody: nil}
			err := h.checkContentType(tt.input)
			if !tt.err && err != nil {
				t.Fatalf("should not be error, but: %v", err)
			}
			if tt.err && err == nil {
				t.Fatalf("should be error, but not")
			}
		})
	}
}

func TestJSONHandler_errorResponse(t *testing.T) {
	type input struct {
		err        error
		statusCode int
	}
	type want struct {
		err        string
		statusCode int
	}

	cases := []struct {
		input input
		want  want
	}{
		{
			input: input{err: errors.New("error occurred"), statusCode: 404},
			want:  want{err: "error occurred", statusCode: 404},
		},
		{
			input: input{err: nil, statusCode: 404},
			want:  want{err: "nil", statusCode: 404},
		},
		{
			input: input{err: errors.New("error occurred"), statusCode: 10},
			want:  want{err: "error occurred", statusCode: 10},
		},
		{
			input: input{err: errors.New("error occurred"), statusCode: 0},
			want:  want{err: "error occurred", statusCode: 200},
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			h := JSONHandler{ErrResponseFn: nil}
			w := httptest.NewRecorder()
			h.errorResponse(w, tt.input.err, tt.input.statusCode)
			got := w.Result()
			if got.StatusCode != tt.want.statusCode {
				t.Fatalf("want %v, but got %v", tt.want.statusCode, got.StatusCode)
			}
			r := errorResponse{}
			json.NewDecoder(got.Body).Decode(&r)
			if r.Message != tt.want.err {
				t.Fatalf("want %v, but got %v", tt.want.err, r.Message)
			}
		})
	}
}

func TestJSONHandler_errorResponse_specifyErrResponseFn(t *testing.T) {
	h := JSONHandler{
		ErrResponseFn: func(w http.ResponseWriter, err error, statusCode int) {
			w.WriteHeader(statusCode)
			w.Write([]byte(err.Error()))
		},
	}
	w := httptest.NewRecorder()
	h.errorResponse(w, errors.New("error occurred"), 400)
	got := w.Result()
	if got.StatusCode != 400 {
		t.Fatalf("want 404, but got %v", got.StatusCode)
	}
	b, _ := ioutil.ReadAll(got.Body)
	if string(b) != "error occurred" {
		t.Fatalf("want 'error occurred', but got %v", string(b))
	}
}

func TestJSONHandler_ServeHTTP_queryParams(t *testing.T) {
	cases := []struct {
		input string
		want  map[string][]string
	}{
		{
			input: "a=hoge",
			want: map[string][]string{
				"a": []string{"hoge"},
			},
		},
		{
			input: "a=hoge&b=fuga",
			want: map[string][]string{
				"a": []string{"hoge"},
				"b": []string{"fuga"},
			},
		},
		{
			input: "a=hoge&b=fuga&a=hogehoge",
			want: map[string][]string{
				"a": []string{"hoge", "hogehoge"},
				"b": []string{"fuga"},
			},
		},
		{
			input: "",
			want:  map[string][]string{},
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			h := JSONHandler{
				Method:        "GET",
				PathFmt:       "/users",
				ErrResponseFn: nil,
				ResponseCode:  200,
				RequestBody:   nil,
				ResponseFn: func(_ interface{}, _ []string, params url.Values) (interface{}, error) {
					return params, nil
				},
			}

			req := httptest.NewRequest("GET", "http://localhost/users?"+tt.input, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != 200 {
				t.Fatalf("want 200, but got %v", res.StatusCode)
			}

			var got map[string][]string
			json.NewDecoder(res.Body).Decode(&got)

			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("want %v, but got %v", tt.want, got)
			}
		})
	}
}

func TestJSONHandler_ServeHTTP_requestBody(t *testing.T) {
	type user struct {
		Name string
	}
	testUser := user{
		Name: "test-user",
	}

	h := JSONHandler{
		Method:        "POST",
		PathFmt:       "/users",
		ErrResponseFn: nil,
		ResponseCode:  200,
		RequestBody:   &user{},
		ResponseFn:    nil,
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(&testUser)

	req := httptest.NewRequest("POST", "http://localhost/users", &b)
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Fatalf("want 200, but got %v", res.StatusCode)
	}

	var got user
	json.NewDecoder(res.Body).Decode(&got)

	if !reflect.DeepEqual(testUser, got) {
		t.Fatalf("want %v, but got %v", testUser, got)
	}
}

func TestNewMultipleHandler(t *testing.T) {
	cases := []struct {
		input []JSONHandler
		want  *MultipleHandler
	}{
		{
			input: []JSONHandler{
				{Method: "GET", PathFmt: "/users/*"},
				{Method: "PUT", PathFmt: "/users/*"},
				{Method: "POST", PathFmt: "/users"},
			},
			want: &MultipleHandler{
				handlers: []JSONHandler{
					{Method: "GET", PathFmt: "/users/*"},
					{Method: "PUT", PathFmt: "/users/*"},
					{Method: "POST", PathFmt: "/users"},
				},
			},
		},
		{
			input: []JSONHandler{
				{Method: "GET", PathFmt: "/users/*"},
				{Method: "GET", PathFmt: "/users/*"},
				{Method: "POST", PathFmt: "/users"},
			},
			want: &MultipleHandler{
				handlers: []JSONHandler{
					{Method: "GET", PathFmt: "/users/*"},
					{Method: "GET", PathFmt: "/users/*"},
					{Method: "POST", PathFmt: "/users"},
				},
			},
		},
		{
			input: []JSONHandler{
				{Method: "", PathFmt: "/users/*"},
			},
			want: &MultipleHandler{
				handlers: []JSONHandler{},
			},
		},
		{
			input: []JSONHandler{
				{Method: "GET", PathFmt: ""},
			},
			want: &MultipleHandler{
				handlers: []JSONHandler{},
			},
		},
		{
			input: []JSONHandler{
				{Method: "", PathFmt: ""},
			},
			want: &MultipleHandler{
				handlers: []JSONHandler{},
			},
		},
		{
			input: []JSONHandler{},
			want: &MultipleHandler{
				handlers: []JSONHandler{},
			},
		},
		{
			input: nil,
			want: &MultipleHandler{
				handlers: []JSONHandler{},
			},
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			got := NewMultipleHandler(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("want %#v, but got: %#v", tt.want, got)
			}
		})
	}
}

func TestMultipleHandler_AddHandler_nilHandlers(t *testing.T) {
	cases := []struct {
		input JSONHandler
		want  MultipleHandler
	}{
		{
			input: JSONHandler{Method: "GET", PathFmt: "/users/*"},
			want: MultipleHandler{
				handlers: []JSONHandler{
					{Method: "GET", PathFmt: "/users/*"},
				},
			},
		},
		{
			input: JSONHandler{Method: "", PathFmt: "/users/*"},
			want:  MultipleHandler{},
		},
		{
			input: JSONHandler{Method: "GET", PathFmt: ""},
			want:  MultipleHandler{},
		},
		{
			input: JSONHandler{Method: "", PathFmt: ""},
			want:  MultipleHandler{},
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			h := MultipleHandler{}
			h.AddHandler(tt.input)
			if !reflect.DeepEqual(h, tt.want) {
				t.Fatalf("want %#v, but got: %#v", tt.want, h)
			}
		})
	}
}

func TestMultipleHandler_ServeHTTP_matchHandler(t *testing.T) {
	h := NewMultipleHandler([]JSONHandler{
		{
			Method:       "GET",
			PathFmt:      "/users/*",
			ResponseCode: 200,
			ResponseFn: func(_ interface{}, pParams []string, _ url.Values) (interface{}, error) {
				return map[string]interface{}{"called": pParams[0]}, nil
			},
		},
		{
			Method:       "GET",
			PathFmt:      "/users/*",
			ResponseCode: 200,
			ResponseFn: func(_ interface{}, pParams []string, _ url.Values) (interface{}, error) {
				return map[string]interface{}{"never_called": pParams[0]}, nil
			},
		},
	})

	req := httptest.NewRequest("GET", "http://localhost/users/1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Fatalf("want 200, but got %v", res.StatusCode)
	}

	var got map[string]interface{}
	err := json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}

	if got["called"] != "1" {
		t.Fatalf("want '1', but got %v", got["called"])
	}
}

func TestMultipleHandler_ServeHTTP_unmatched(t *testing.T) {
	cases := []struct {
		method       string
		path         string
		responseCode int
	}{
		{
			method:       "GET",
			path:         "/hoge",
			responseCode: 404,
		},
		{
			method:       "POST",
			path:         "/users/1",
			responseCode: 404,
		},
	}

	h := NewMultipleHandler([]JSONHandler{
		{
			Method:       "GET",
			PathFmt:      "/users/*",
			ResponseCode: 200,
		},
	})

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "http://localhost"+tt.path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != tt.responseCode {
				t.Fatalf("want %v, but got %v", tt.responseCode, res.StatusCode)
			}
		})
	}
}
