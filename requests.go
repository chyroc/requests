package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	cookiejar "github.com/chyroc/persistent-cookiejar"
)

var httpClient = &http.Client{Timeout: time.Second * 3}

var httpClientNoTimeout = &http.Client{}

type Request struct {
	// internal
	cachedURL     atomic.Pointer[string] // 请求 URL
	isRequest     atomic.Bool            // 是否已经发起请求
	isRead        atomic.Bool            // 是否已经 read 了 response.Body
	persistentJar *cookiejar.Jar         // cookie
	err           error                  // err
	logger        Logger                 // 日志

	// request
	reqLock    sync.RWMutex        // reqLock, 尝试修改为无锁
	context    context.Context     // request context
	ignoreSSL  bool                // request ignore ssl verify
	header     http.Header         // request header
	query      map[string][]string // request query
	noRedirect bool                // request ignore redirect
	timeout    time.Duration       // request timeout
	url        string              // request url
	method     string              // request method
	rawBody    []byte              // []byte of body
	body       io.Reader           // request body

	// resp
	wrapResponse func(resp *http.Response) (*http.Response, error) // wrap response
	resp         *http.Response
	bytes        []byte
	cancel       context.CancelFunc
}

func New(method, url string, options ...Option) *Request {
	r := &Request{
		url:     url,
		method:  method,
		header:  map[string][]string{},
		query:   make(map[string][]string),
		context: context.Background(),
		logger:  StdoutLogger(),
		timeout: time.Second * 3,
	}
	r.header.Set("user-agent", fmt.Sprintf("chyroc-requests/%s (https://github.com/chyroc/requests)", version))
	applyOpt(r, options...)
	return r
}

func Get(url string, options ...Option) *Request {
	return New(http.MethodGet, url, options...)
}

func Post(url string, options ...Option) *Request {
	return New(http.MethodPost, url, options...)
}

func Put(url string, options ...Option) *Request {
	return New(http.MethodPut, url, options...)
}

func Patch(url string, options ...Option) *Request {
	return New(http.MethodPatch, url, options...)
}

func Delete(url string, options ...Option) *Request {
	return New(http.MethodDelete, url, options...)
}

func (r *Request) SetError(err error) *Request {
	r.err = err
	return r
}
