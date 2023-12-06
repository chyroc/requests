package requests

import (
	"net/http"
	"sync"

	cookiejar "github.com/chyroc/persistent-cookiejar"
)

type Session struct {
	jar        *cookiejar.Jar
	err        error
	cookieFile string
	options    []Option
}

func (r *Session) New(method, url string, options ...Option) *Request {
	req := New(method, url, options...)
	req.persistentJar = r.jar
	req.SetError(r.err)

	applyOpt(req)

	return req
}

func (r *Session) Get(url string, options ...Option) *Request {
	return r.New(http.MethodGet, url, r.allOptions(options)...)
}

func (r *Session) Post(url string, options ...Option) *Request {
	return r.New(http.MethodPost, url, r.allOptions(options)...)
}

func (r *Session) Put(url string, options ...Option) *Request {
	return r.New(http.MethodPut, url, r.allOptions(options)...)
}

func (r *Session) Patch(url string, options ...Option) *Request {
	return r.New(http.MethodPatch, url, r.allOptions(options)...)
}

func (r *Session) Delete(url string, options ...Option) *Request {
	return r.New(http.MethodDelete, url, r.allOptions(options)...)
}

func (r *Session) allOptions(options []Option) []Option {
	if len(r.options) == 0 {
		return options
	}
	if len(options) == 0 {
		return r.options
	}
	res := make([]Option, 0, len(r.options)+len(options))
	for _, v := range r.options {
		res = append(res, v)
	}
	for _, v := range options {
		res = append(res, v)
	}
	return res
}

func (r *Session) Jar() http.CookieJar {
	return r.jar
}

func (r *Session) CookieFile() string {
	return r.cookieFile
}

var (
	sessionLock sync.Mutex
	sessionMap  = map[string]*Session{}
)

// same cookie-file has same session instance
func NewSession(cookieFile string, options ...Option) *Session {
	sessionLock.Lock()
	defer sessionLock.Unlock()

	v := sessionMap[cookieFile]
	if v != nil {
		return v
	}

	v = newSession(cookieFile, options)
	sessionMap[cookieFile] = v
	return v
}

func newSession(cookieFile string, options []Option) *Session {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename:   cookieFile,
		Persistent: true,
	})
	if err != nil {
		return &Session{err: err, cookieFile: cookieFile, options: options}
	} else {
		return &Session{jar: jar, cookieFile: cookieFile, options: options}
	}
}
