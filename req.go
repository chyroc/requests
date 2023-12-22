package requests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// --- 获取参数 ---

// Context request context.Context
func (r *Request) Context() context.Context {
	if r.context != nil {
		return r.context
	}
	return context.Background()
}

// Timeout request timeout
func (r *Request) Timeout() time.Duration {
	return r.timeout
}

// URL request url
func (r *Request) URL() string {
	return r.url
}

// FullURL request full url, contain query param
func (r *Request) FullURL() string {
	if cachedURL := r.cachedURL.Load(); cachedURL != nil {
		return *cachedURL
	}
	return r.requestURL()
}

// Method request method
func (r *Request) Method() string {
	return r.method
}

// ReqHeader request header
func (r *Request) ReqHeader() http.Header {
	return r.header
}

// --- 设置参数 ---

// WithContext setup request context.Context
func (r *Request) WithContext(ctx context.Context) *Request {
	return r.setRequestParam(func(r *Request) {
		r.context = ctx
	})
}

// WithTimeout setup request timeout
func (r *Request) WithTimeout(timeout time.Duration) *Request {
	return r.setRequestParam(func(r *Request) {
		r.timeout = timeout
	})
}

// WithIgnoreSSL ignore ssl verify
func (r *Request) WithIgnoreSSL(ignore bool) *Request {
	return r.setRequestParam(func(r *Request) {
		r.ignoreSSL = ignore
	})
}

// WithWrapResponse set round tripper response wrap
func (r *Request) WithWrapResponse(f func(resp *http.Response) (*http.Response, error)) *Request {
	return r.setRequestParam(func(r *Request) {
		r.wrapResponse = f
	})
}

// WithHeader set one header k-v map
func (r *Request) WithHeader(k, v string) *Request {
	return r.setRequestParam(func(r *Request) {
		r.configHeader(k, v)
	})
}

// WithHeaders set multi header k-v map
func (r *Request) WithHeaders(kv map[string]string) *Request {
	return r.setRequestParam(func(r *Request) {
		for k, v := range kv {
			r.configHeader(k, v)
		}
	})
}

func (r *Request) configHeader(k, v string) {
	if strings.ToLower(k) == "user-agent" {
		r.header.Set(k, v)
	} else {
		r.header.Add(k, v)
	}
}

// WithRedirect set allow or not-allow redirect with Location header
func (r *Request) WithRedirect(redirect bool) *Request {
	return r.setRequestParam(func(r *Request) {
		r.noRedirect = !redirect
	})
}

// WithQuery set one query k-v map
func (r *Request) WithQuery(k, v string) *Request {
	return r.setRequestParam(func(r *Request) {
		r.query[k] = append(r.query[k], v)
	})
}

// WithQueries set multi query k-v
func (r *Request) WithQueries(queries any) *Request {
	return r.setRequestParam(func(r *Request) {
		kvs, err := toQueryMapSlice(queries)
		if err != nil {
			r.SetError(err)
			return
		}
		for k, vv := range kvs {
			r.query[k] = append(r.query[k], vv...)
		}
	})
}

// WithBody set request body, support: io.Reader, []byte, string, any(as json format)
func (r *Request) WithBody(body any) *Request {
	return r.setRequestParam(func(r *Request) {
		r.rawBody, r.body, r.err = toBody(body)
	})
}

// WithJSON set body same as WithBody, and set Content-Type to application/json
func (r *Request) WithJSON(body any) *Request {
	return r.setRequestParam(func(r *Request) {
		r.rawBody, r.body, r.err = toBody(body)
		if r.err != nil {
			return
		}
		r.header.Set("Content-Type", "application/json")
	})
}

// WithForm set body and set Content-Type to multiform
func (r *Request) WithForm(body map[string]string) *Request {
	return r.setRequestParam(func(r *Request) {
		buf := bytes.Buffer{}
		f := multipart.NewWriter(&buf)
		for k, v := range body {
			if err := f.WriteField(k, v); err != nil {
				r.err = err
				return
			}
		}

		r.rawBody, r.body = buf.Bytes(), bytes.NewReader(buf.Bytes())
		r.header.Set("Content-Type", f.FormDataContentType())
	})
}

// WithFormURLEncoded set body and set Content-Type to application/x-www-form-urlencoded
func (r *Request) WithFormURLEncoded(body map[string]string) *Request {
	return r.setRequestParam(func(r *Request) {
		u := url.Values{}
		for k, v := range body {
			u.Add(k, v)
		}

		r.rawBody, r.body = []byte(u.Encode()), strings.NewReader(u.Encode())
		r.header.Set("Content-Type", "application/x-www-form-urlencoded")
	})
}

// WithFile set file to body and set some multi-form k-v map
func (r *Request) WithFile(filename string, file io.Reader, fileKey string, params map[string]string) *Request {
	return r.setRequestParam(func(r *Request) {
		contentType, bod, err := newFileUploadRequest(params, fileKey, filename, file)
		if err != nil {
			r.err = err
			return
		}
		r.rawBody, r.body = nil, bod
		r.header.Set("Content-Type", contentType)
	})
}

// WithURLCookie set cookie of uri
func (r *Request) WithURLCookie(uri string) *Request {
	return r.setRequestParam(func(r *Request) {
		if r.persistentJar == nil {
			return
		}

		uriParse, err := url.Parse(uri)
		if err != nil {
			r.err = err
			return
		}

		cookies := []string{}
		for _, v := range r.persistentJar.Cookies(uriParse) {
			cookies = append(cookies, v.Name+"="+v.Value)
		}
		if len(cookies) > 0 {
			r.header.Add("cookie", strings.Join(cookies, "; ")) // use add not set
		}
	})
}

// WithLogger set logger
func (r *Request) WithLogger(logger Logger) *Request {
	return r.setRequestParam(func(r *Request) {
		r.logger = logger
	})
}

// WithHeader set one header k-v map
func (r *Request) setRequestParam(f func(*Request)) *Request {
	r.reqLock.Lock()
	defer r.reqLock.Unlock()

	if r.isRequest.Load() {
		r.SetError(fmt.Errorf("request %s %s alreday sended, cannot set request params",
			r.method, *r.cachedURL.Load()))
		return r
	}

	if r.err != nil {
		return r
	}

	f(r)

	return r
}
