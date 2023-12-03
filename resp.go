package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"unsafe"

	. "github.com/chyroc/anyhow"
)

// Bytes get request body response as []byte
func (r *Request) Bytes() Result1[[]byte] {
	if err := r.doRead(); err != nil {
		return Err1[[]byte](err)
	}

	return Ok1(r.bytes)
}

// JSON convert request body to T as json type
func JSON[T any](r *Request) Result1[*T] {
	return Then1(r.Bytes(), func(data []byte) Result1[*T] {
		var resp T
		if err := json.Unmarshal(data, &resp); err != nil {
			// todo: 统一错误格式
			return Err1[*T](fmt.Errorf("[requests] %s %s unmarshal %s to %s failed: %w",
				r.method, r.cachedRequestURL(), data, reflect.TypeOf(resp).Name(), err))
		}
		return Ok1(&resp)
	})
}

// Map convert request body to map
func (r *Request) Map() Result1[map[string]any] {
	return Then1(r.Bytes(), func(data []byte) Result1[map[string]any] {
		resp := make(map[string]any)
		if err := json.Unmarshal(data, &resp); err != nil {
			return Err1[map[string]any](
				fmt.Errorf("[requests] %s %s unmarshal %s to map failed: %w",
					r.method, r.cachedRequestURL(), data, err))
		}
		return Ok1(resp)
	})
}

// Map convert request body to str
func (r *Request) Text() Result1[string] {
	return Then1(r.Bytes(), func(data []byte) Result1[string] {
		return Ok1(*(*string)(unsafe.Pointer(&data)))
	})
}

// Response get http response
func (r *Request) Response() Result1[*http.Response] {
	if err := r.doRequest(); err != nil {
		return Err1[*http.Response](err)
	}
	return Ok1(r.resp)
}

// Response get http response status
func (r *Request) Status() Result1[int] {
	return Then1(r.Response(), func(data *http.Response) Result1[int] {
		return Ok1(data.StatusCode)
	})
}

// Header get http response header
func (r *Request) Header() Result1[http.Header] {
	return Then1(r.Response(), func(data *http.Response) Result1[http.Header] {
		return Ok1(data.Header)
	})
}

// HeadersByKey get specific http header response with key
func (r *Request) HeadersByKey(key string) Result1[[]string] {
	return Then1(r.Response(), func(data *http.Response) Result1[[]string] {
		return Ok1(data.Header.Values(key))
	})
}

// CookiesByKey get specific http cookie response with key
func (r *Request) CookiesByKey(key string) Result1[[]string] {
	return Then1(r.Response(), func(data *http.Response) Result1[[]string] {
		var resp []string
		for _, v := range r.resp.Cookies() {
			if v.Name == key {
				resp = append(resp, v.Value)
			}
		}
		return Ok1(resp)
	})
}

// HeaderByKey get specific http header response with key
func (r *Request) HeaderByKey(key string) Result1[string] {
	return Then1(r.Response(), func(data *http.Response) Result1[string] {
		return Ok1(data.Header.Get(key))
	})
}
