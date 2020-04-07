# fakehttp
[![GoDoc](https://godoc.org/github.com/tennashi/fakehttp?status.svg)](https://godoc.org/github.com/tennashi/fakehttp)
[![CI](https://github.com/tennashi/fakehttp/workflows/CI/badge.svg)](https://github.com/tennashi/fakehttp/actions)

Fake HTTP handler.

## Usage
### HTTP GET
Suppose you want to test a method such as the following:
```go
package app

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type User struct {
	ID   int
	Name string
}

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url}
}

func (c Client) GetUser(userID int) (*User, error) {
	res, err := http.Get(c.url + "/users/" + strconv.Itoa(userID))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	u := User{}
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
```

The test function is as follows:
```go
package app_test

import (
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"


	"github.com/tennashi/fakehttp"

	"example.com/yourapp/app" // import your application
)

// mock handler definition
var getHandler = fakehttp.JSONHandler{
	Method:       "GET",
	PathFmt:      "/users/*",
	ResponseCode: 200,
	ResponseFn: func(
		_ interface{}, // request body
		pathParams []string, // URL path params
		_ url.Values, // URL query params
	) (interface{}, error) {
		userID, _ := strconv.Atoi(pathParams[0])
		return &app.User{
			ID:   userID,
			Name: "test-user-" + pathParams[0],
		}, nil
	},
}

func TestCli_GetUser(t *testing.T) {
	cases := []struct {
		input int
		want  *app.User
		err   bool
	}{
		{
			input: 1,
			want: &app.User{
				ID:   1,
				Name: "test-user-1",
			},
			err: false,
		},
		{
			input: 2,
			want: &app.User{
				ID:   2,
				Name: "test-user-2",
			},
			err: false,
		},
	}

	ts := httptest.NewServer(getHandler)
	defer ts.Close()

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			c := app.NewClient(ts.URL)
			got, err := c.GetUser(tt.input)
			if !tt.err && err != nil {
				t.Fatalf("should not be error, but: %v", err)
			}
			if tt.err && err == nil {
				t.Fatalf("should be error, but not")
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("want %v, but got %v", tt.want, got)
			}
		})
	}
}
```

### HTTP POST
Suppose you want to test a method such as the following:
```go
package app

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type User struct {
	ID   int
	Name string
}

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url}
}

func (c *Client) CreateUser(u User) (*User, error) {
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(u)
	res, err := http.Post(c.url+"/users", "application/json", &b)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ret := User{}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
```

The test function is as follows:
```go
package app_test

import (
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/tennashi/fakehttp"

	"example.com/yourapp/app" // import your application
)

// mock handler definition
var postHandler = fakehttp.JSONHandler{
	Method:       "POST",
	PathFmt:      "/users",
	ResponseCode: 200,
	RequestBody:  &app.User{}, // Specify the type of the request body.
	ResponseFn: func(
		req interface{}, // request body
		_ []string, // URL path params
		_ url.Values, // URL query params
	) (interface{}, error) {
		return req, nil
	},
}

func TestCli_CreateUser(t *testing.T) {
	cases := []struct {
		input app.User
		want  *app.User
		err   bool
	}{
		{
			input: app.User{
				ID:   1,
				Name: "test-user-1",
			},
			want: &app.User{
				ID:   1,
				Name: "test-user-1",
			},
			err: false,
		},
		{
			input: app.User{
				ID:   2,
				Name: "test-user-2",
			},
			want: &app.User{
				ID:   2,
				Name: "test-user-2",
			},
			err: false,
		},
	}

	ts := httptest.NewServer(postHandler)
	defer ts.Close()

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			c := app.NewClient(ts.URL)
			got, err := c.CreateUser(tt.input)
			if !tt.err && err != nil {
				t.Fatalf("should not be error, but: %v", err)
			}
			if tt.err && err == nil {
				t.Fatalf("should be error, but not")
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("want %v, but got %v", tt.want, got)
			}
		})
	}
}
```
