package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"unsafe"
)

// Bytes get request body response as []byte
func (r *Request) Bytes() Result[[]byte] {
	if err := r.doRead(); err != nil {
		return Err[[]byte](err)
	}

	return Ok(r.bytes)
}

// JSON convert request body to T as json type
func JSON[T any](r *Request) Result[*T] {
	return Then(r.Bytes(), func(data []byte) Result[*T] {
		var resp T
		if err := json.Unmarshal(data, &data); err != nil {
			// todo: 统一错误格式
			return Err[*T](fmt.Errorf("[requests] %s %s unmarshal %s to %s failed: %w",
				r.method, r.cachedRequestURL(), data, reflect.TypeOf(resp).Name(), err))
		}
		return Ok(&resp)
	})
}

// Map convert request body to map
func (r *Request) Map() Result[map[string]any] {
	return Then(r.Bytes(), func(data []byte) Result[map[string]any] {
		resp := make(map[string]any)
		if err := json.Unmarshal(data, &resp); err != nil {
			return Err[map[string]any](
				fmt.Errorf("[requests] %s %s unmarshal %s to map failed: %w",
					r.method, r.cachedRequestURL(), data, err))
		}
		return Ok(resp)
	})
}

// Map convert request body to str
func (r *Request) Text() Result[string] {
	return Then(r.Bytes(), func(data []byte) Result[string] {
		return Ok(*(*string)(unsafe.Pointer(&data)))
	})
}

// Response get http response
func (r *Request) Response() Result[*http.Response] {
	if err := r.doRequest(); err != nil {
		return Err[*http.Response](err)
	}
	return Ok(r.resp)
}

// Response get http response status
func (r *Request) Status() Result[int] {
	return Then(r.Response(), func(data *http.Response) Result[int] {
		return Ok(data.StatusCode)
	})
}

// Header get http response header
func (r *Request) Header() Result[http.Header] {
	return Then(r.Response(), func(data *http.Response) Result[http.Header] {
		return Ok(data.Header)
	})
}

// HeadersByKey get specific http header response with key
func (r *Request) HeadersByKey(key string) Result[[]string] {
	return Then(r.Response(), func(data *http.Response) Result[[]string] {
		return Ok(data.Header.Values(key))
	})
}

// CookiesByKey get specific http cookie response with key
func (r *Request) CookiesByKey(key string) Result[[]string] {
	return Then(r.Response(), func(data *http.Response) Result[[]string] {
		var resp []string
		for _, v := range r.resp.Cookies() {
			if v.Name == key {
				resp = append(resp, v.Value)
			}
		}
		return Ok(resp)
	})
}

// HeaderByKey get specific http header response with key
func (r *Request) HeaderByKey(key string) Result[string] {
	return Then(r.Response(), func(data *http.Response) Result[string] {
		return Ok(data.Header.Get(key))
	})
}
