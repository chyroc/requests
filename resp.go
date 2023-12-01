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
	bs := r.Bytes()
	return andThen(bs, func(bsData []byte) Result[*T] {
		var data T
		if err := json.Unmarshal(bsData, &data); err != nil {
			// todo: 统一错误格式
			return Err[*T](fmt.Errorf("[requests] %s %s unmarshal %s to %s failed: %w",
				r.method, r.cachedRequestURL(), bsData, reflect.TypeOf(data).Name(), err))
		}
		return Ok(&data)
	})
}

// Map convert request body to map
func (r *Request) Map() Result[map[string]any] {
	bs := r.Bytes()
	return andThen(bs, func(bsData []byte) Result[map[string]any] {
		m := make(map[string]any)
		if err := json.Unmarshal(bsData, &m); err != nil {
			return Err[map[string]any](fmt.Errorf("[requests] %s %s unmarshal %s to map failed: %w",
				r.method, r.cachedRequestURL(), bsData, err))
		}
		return Ok(m)
	})
}

// Map convert request body to str
func (r *Request) Text() Result[string] {
	bs := r.Bytes()
	return andThen(bs, func(bsData []byte) Result[string] {
		return Ok(*(*string)(unsafe.Pointer(&bsData)))
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
	if err := r.doRequest(); err != nil {
		return Err[int](err)
	}

	return Ok(r.resp.StatusCode)
}

// Header get http response header
func (r *Request) Header() Result[http.Header] {
	r.ReqHeader()
	if err := r.doRequest(); err != nil {
		return Err[http.Header](err)
	}
	return Ok(r.resp.Header)
}

// HeadersByKey get specific http header response with key
func (r *Request) HeadersByKey(key string) Result[[]string] {
	if err := r.doRequest(); err != nil {
		return Err[[]string](err)
	}
	return Ok(r.resp.Header.Values(key))
}

// CookiesByKey get specific http cookie response with key
func (r *Request) CookiesByKey(key string) Result[[]string] {
	if err := r.doRequest(); err != nil {
		return Err[[]string](err)
	}

	var resp []string
	for _, v := range r.resp.Cookies() {
		if v.Name == key {
			resp = append(resp, v.Value)
		}
	}
	return Ok(resp)
}

// HeaderByKey get specific http header response with key
func (r *Request) HeaderByKey(key string) Result[string] {
	if err := r.doRequest(); err != nil {
		return Err[string](err)
	}
	return Ok(r.resp.Header.Get(key))
}
