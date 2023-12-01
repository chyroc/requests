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
