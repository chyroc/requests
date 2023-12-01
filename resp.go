package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

func (r *Request[T]) Bytes() Result[[]byte] {
	if err := r.doRequest(); err != nil {
		return Err[[]byte](err)
	}
	if err := r.doRead(); err != nil {
		return Err[[]byte](err)
	}

	return Ok(r.bytes)
}

func (r *Request[T]) JSON() Result[T] {
	bs := r.Bytes()
	return andThen(bs, func(bsData []byte) Result[T] {
		var data T
		if err := json.Unmarshal(bsData, &data); err != nil {
			// todo: 统一错误格式
			return Err[T](fmt.Errorf("[requests] %s %s unmarshal %s to %s failed: %w",
				r.method, r.cachedurl, bsData, reflect.TypeOf(data).Name(), err))
		}
		return Ok(data)
	})
}

func (r *Request[T]) Map() Result[map[string]any] {
	bs := r.Bytes()
	return andThen(bs, func(bsData []byte) Result[map[string]any] {
		m := make(map[string]any)
		if err := json.Unmarshal(bsData, &m); err != nil {
			return Err[map[string]any](fmt.Errorf("[requests] %s %s unmarshal %s to map failed: %w",
				r.method, r.cachedurl, bsData, err))
		}
		return Ok(m)
	})
}

func (r *Request[T]) Text() Result[string] {
	bs := r.Bytes()
	return andThen(bs, func(t []byte) Result[string] {
		return Ok(string(t)) // todo: 高性能写法
	})
}

func (r *Request[T]) Response() Result[*http.Response] {
	if err := r.doRequest(); err != nil {
		return Err[*http.Response](err)
	}

	return Ok(r.resp)
}

func (r *Request[T]) Status() Result[int] {
	if err := r.doRequest(); err != nil {
		return Err[int](err)
	}

	return Ok(r.resp.StatusCode)
}

func (r *Request[T]) RespHeaders() Result[http.Header] {
	if err := r.doRequest(); err != nil {
		return Err[http.Header](err)
	}
	return Ok(r.resp.Header)
}

func (r *Request[T]) RespHeadersByKey(key string) Result[[]string] {
	if err := r.doRequest(); err != nil {
		return Err[[]string](err)
	}
	return Ok(r.resp.Header.Values(key))
}

func (r *Request[T]) RespCookiesByKey(key string) Result[[]string] {
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

func (r *Request[T]) RespHeaderByKey(key string) Result[string] {
	if err := r.doRequest(); err != nil {
		return Err[string](err)
	}
	return Ok(r.resp.Header.Get(key))
}
