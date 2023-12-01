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
func (r *Request[T]) Context() context.Context {
	if r.context != nil {
		return r.context
	}
	return context.Background()
}

// Timeout request timeout
func (r *Request[T]) Timeout() time.Duration {
	return r.timeout
}

// URL request url
func (r *Request[T]) URL() string {
	return r.url // todo: url prased
}

// RequestFullURL request full url, contain query param
func (r *Request[T]) RequestFullURL() string {
	r.lock.RLock() // todo: remove lock
	defer r.lock.RUnlock()

	return r.parseRequestURL()
}

// Method request method
func (r *Request[T]) Method() string {
	return r.method
}

// Header request header
func (r *Request[T]) Header() http.Header {
	return r.header
}

// --- 设置参数 ---

// WithContext setup request context.Context
func (r *Request[T]) WithContext(ctx context.Context) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.context = ctx
	})
}

// WithTimeout setup request timeout
func (r *Request[T]) WithTimeout(timeout time.Duration) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.timeout = timeout
	})
}

// WithIgnoreSSL ignore ssl verify
func (r *Request[T]) WithIgnoreSSL(ignore bool) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.isIgnoreSSL = ignore
	})
}

// WithWrapResponse set round tripper response wrap
func (r *Request[T]) WithWrapResponse(f func(resp *http.Response) (*http.Response, error)) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.wrapResponse = f
	})
}

// WithHeader set one header k-v map
func (r *Request[T]) WithHeader(k, v string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.configHeader(k, v)
	})
}

// WithHeaders set multi header k-v map
func (r *Request[T]) WithHeaders(kv map[string]string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		for k, v := range kv {
			r.configHeader(k, v)
		}
	})
}

// WithHeaders set multi header k-v map
func (r *Request[T]) configHeader(k, v string) {
	if strings.ToLower(k) == "user-agent" {
		r.header.Set(k, v)
	} else {
		r.header.Add(k, v)
	}
}

// WithRedirect set allow or not-allow redirect with Location header
func (r *Request[T]) WithRedirect(redirect bool) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.isNoRedirect = !redirect
	})
}

// WithQuery set one query k-v map
func (r *Request[T]) WithQuery(k, v string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.querys[k] = append(r.querys[k], v)
	})
}

// WithQuerys set multi query k-v map
func (r *Request[T]) WithQuerys(kv map[string]string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		for k, v := range kv {
			r.querys[k] = append(r.querys[k], v)
		}
	})
}

// WithQueryStruct set multi query k-v map
func (r *Request[T]) WithQueryStruct(v any) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		kv, err := queryToMap(v)
		if err != nil {
			r.err = err
			return
		}
		for k, v := range kv {
			r.querys[k] = append(r.querys[k], v...)
		}
	})
}

// WithBody set request body, support: io.Reader, []byte, string, any(as json format)
func (r *Request[T]) WithBody(body any) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.rawBody, r.body, r.err = toBody(body)
	})
}

// WithJSON set body same as WithBody, and set Content-Type to application/json
func (r *Request[T]) WithJSON(body any) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.rawBody, r.body, r.err = toBody(body)
		if r.err != nil {
			return
		}
		r.header.Set("Content-Type", "application/json")
	})
}

// WithForm set body and set Content-Type to multiform
func (r *Request[T]) WithForm(body map[string]string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		buf := bytes.Buffer{}
		f := multipart.NewWriter(&buf)
		for k, v := range body {
			if err := f.WriteField(k, v); err != nil {
				r.err = err
				return
			}
		}

		r.rawBody, r.body = buf.Bytes(), strings.NewReader(buf.String())
		r.header.Set("Content-Type", f.FormDataContentType())
	})
}

// WithFormURLEncoded set body and set Content-Type to application/x-www-form-urlencoded
func (r *Request[T]) WithFormURLEncoded(body map[string]string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		u := url.Values{}
		for k, v := range body {
			u.Add(k, v)
		}

		r.rawBody, r.body = []byte(u.Encode()), strings.NewReader(u.Encode())
		r.header.Set("Content-Type", "application/x-www-form-urlencoded")
	})
}

// WithFile set file to body and set some multi-form k-v map
func (r *Request[T]) WithFile(filename string, file io.Reader, fileKey string, params map[string]string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
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
func (r *Request[T]) WithURLCookie(uri string) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
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
func (r *Request[T]) WithLogger(logger Logger) *Request[T] {
	return r.configParamFactor(func(r *Request[T]) {
		r.logger = logger
	})
}

// WithHeader set one header k-v map
func (r *Request[T]) configParamFactor(f func(*Request[T])) *Request[T] {
	r.lock.Lock() // todo: cas
	defer r.lock.Unlock()

	if r.isRequest {
		r.SetError(fmt.Errorf("request %s %s alreday sended, cannot set request params", r.method, r.cachedurl))
		return r
	}

	if r.err != nil {
		return r
	}

	f(r)

	return r
}
