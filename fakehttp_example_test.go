package fakehttp_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"

	"github.com/tennashi/fakehttp"
)

func ExampleJSONHandler_get() {
	type user struct {
		ID    int
		Name  string
		Email string
	}
	type errRes struct {
		Msg string
	}

	// Setup fake handler
	h := fakehttp.JSONHandler{
		Method:       http.MethodGet,
		PathFmt:      "/users/*",
		ResponseCode: http.StatusOK,
		ResponseFn: func(_ interface{}, pParams []string, qParams url.Values) (interface{}, error) {
			userID, _ := strconv.Atoi(pParams[0])
			return &user{
				ID:    userID,
				Name:  "test-user-" + pParams[0],
				Email: "test-user-" + pParams[0] + "@example.com",
			}, nil
		},
		ErrResponseFn: func(w http.ResponseWriter, err error, statusCode int) {
			errRes := errRes{Msg: err.Error()}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(errRes)
		},
	}

	// Setup HTTP test server
	ts := httptest.NewServer(h)
	defer ts.Close()

	// Test target function
	sut := func(userID int) (*user, error) {
		res, err := http.Get(ts.URL + "/users/" + strconv.Itoa(userID))
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			eRes := errRes{}
			json.NewDecoder(res.Body).Decode(&eRes)
			return nil, errors.New(eRes.Msg)
		}
		u := user{}
		if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
			return nil, err
		}
		return &u, nil
	}

	// Test
	cases := []struct {
		input int
		want  *user
	}{
		{
			input: 1,
			want:  &user{ID: 1, Name: "test-user-1", Email: "test-user-1@example.com"},
		},
		{
			input: 2,
			want:  &user{ID: 2, Name: "test-user-2", Email: "test-user-2@example.com"},
		},
		{
			input: 3,
			want:  &user{ID: 3, Name: "test-user-3", Email: "test-user-3@example.com"},
		},
	}

	for _, tt := range cases {
		got, err := sut(tt.input)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			fmt.Printf("want %v, but got: %v\n", tt.want, got)
			return
		}
		fmt.Println("test success")
	}

	// Output:
	// test success
	// test success
	// test success
}

func ExampleJSONHandler_post() {
	type userReq struct {
		Name  string
		Email string
	}
	type user struct {
		ID    int
		Name  string
		Email string
	}
	type errRes struct {
		Msg string
	}

	idSrc := 0
	// Setup fake handler
	h := fakehttp.JSONHandler{
		Method:       http.MethodPost,
		PathFmt:      "/users",
		RequestBody:  &userReq{},
		ResponseCode: http.StatusOK,
		ResponseFn: func(req interface{}, _ []string, _ url.Values) (interface{}, error) {
			r, ok := req.(*userReq)
			if !ok {
				return nil, errors.New("invalid request")
			}
			idSrc += 1
			return &user{
				ID:    idSrc,
				Name:  r.Name,
				Email: r.Email,
			}, nil
		},
		ErrResponseFn: func(w http.ResponseWriter, err error, statusCode int) {
			errRes := errRes{Msg: err.Error()}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(errRes)
		},
	}

	// Setup HTTP test server
	ts := httptest.NewServer(h)
	defer ts.Close()

	// Test target function
	sut := func(req userReq) (*user, error) {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(req); err != nil {
			return nil, err
		}
		res, err := http.Post(ts.URL+"/users", "application/json", &b)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			eRes := errRes{}
			json.NewDecoder(res.Body).Decode(&eRes)
			return nil, errors.New(eRes.Msg)
		}

		u := user{}
		if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
			return nil, err
		}
		return &u, nil
	}

	// Test
	cases := []struct {
		input userReq
		want  *user
	}{
		{
			input: userReq{Name: "test-user-1", Email: "test-user-1@example.com"},
			want:  &user{ID: 1, Name: "test-user-1", Email: "test-user-1@example.com"},
		},
		{
			input: userReq{Name: "test-user-2", Email: "test-user-2@example.com"},
			want:  &user{ID: 2, Name: "test-user-2", Email: "test-user-2@example.com"},
		},
		{
			input: userReq{Name: "test-user-3", Email: "test-user-3@example.com"},
			want:  &user{ID: 3, Name: "test-user-3", Email: "test-user-3@example.com"},
		},
	}

	for _, tt := range cases {
		got, err := sut(tt.input)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			fmt.Printf("want %v, but got: %v\n", tt.want, got)
			return
		}
		fmt.Println("test success")
	}

	// Output:
	// test success
	// test success
	// test success
}
