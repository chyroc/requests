package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	cookiejar "github.com/chyroc/persistent-cookiejar"
)

type Request[T any] struct {
	// internal
	cachedurl     string
	persistentJar *cookiejar.Jar
	lock          sync.RWMutex
	err           error
	logger        Logger

	// request
	context      context.Context     // request context
	isIgnoreSSL  bool                // request  ignore ssl verify
	header       http.Header         // request header
	querys       map[string][]string // request query
	isNoRedirect bool                // request ignore redirect
	timeout      time.Duration       // request timeout
	url          string              // request url
	method       string              // request method
	rawBody      []byte              // []byte of body
	body         io.Reader           // request body

	// resp
	wrapResponse func(resp *http.Response) (*http.Response, error) // wrap response
	resp         *http.Response
	bytes        []byte
	isRead       bool
	isRequest    bool
}

func New[T any](method, url string) *Request[T] {
	r := &Request[T]{
		url:     url,
		method:  method,
		header:  map[string][]string{},
		querys:  make(map[string][]string),
		context: context.Background(),
		logger:  StdoutLogger(),
	}
	r.header.Set("user-agent", fmt.Sprintf("chyroc-requests/%s (https://github.com/chyroc/requests)", version))
	return r
}

func (r *Request[T]) SetError(err error) *Request[T] {
	r.err = err
	return r
}
