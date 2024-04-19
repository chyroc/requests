package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// doRequest 发送请求
func (r *Request) doRequest() error {
	return r.doRequestFactor(r.doInternalRequest)
}

// doRead send request and read response
func (r *Request) doRead() error {
	return r.doRequestFactor(func() error {
		if err := r.doInternalRequest(); err != nil {
			return err
		}

		return r.doInternalRead()
	})
}

func (r *Request) doInternalRequest() error {
	if !r.isRequest.CompareAndSwap(false, true) {
		// 已经请求过, 返回
		return nil
	}

	requestURL := r.requestURL() // 这里不可能有值
	r.cachedURL.Store(&requestURL)

	r.logger.Info(r.Context(), "[requests] %s: %s, body=%s, header=%+v",
		r.method, requestURL, r.rawBody, r.header)

	if r.persistentJar != nil {
		defer func() {
			if err := r.persistentJar.Save(); err != nil {
				r.logger.Error(r.Context(), "save cookie failed: %s", err)
			}
		}()
	}

	r.context, r.cancel = context.WithTimeout(r.Context(), r.timeout)

	req, err := http.NewRequestWithContext(r.context, r.method, requestURL, r.body)
	if err != nil {
		return fmt.Errorf("[requests] %s %s new request failed: %w",
			r.method, requestURL, err)
	}

	req.Header = r.header

	resp, err := r.httpCli().Do(req)
	r.resp = resp
	if err != nil {
		return fmt.Errorf("[requests] %s %s send request failed: %w",
			r.method, requestURL, err)
	}
	return nil
}

func (r *Request) httpCli() *http.Client {
	if !r.ignoreSSL && r.wrapResponse == nil && r.persistentJar == nil && !r.noRedirect {
		if r.timeout > 0 {
			return httpClientNoTimeout // 用 req 的 ctx 来控制
		}
		return httpClient
	}

	c := &http.Client{}
	if r.ignoreSSL || r.wrapResponse != nil {
		c.Transport = &wrapRoundTripper{
			isIgnoreSSL: r.ignoreSSL,
			wrapResp:    r.wrapResponse,
		}
	}
	if r.persistentJar != nil {
		c.Jar = r.persistentJar
	}
	if r.noRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return c
}

// doRead send request and read response
func (r *Request) doInternalRead() error {
	if !r.isRead.CompareAndSwap(false, true) {
		// 已经 read 过
		return nil
	}
	if r.cancel != nil {
		defer r.cancel()
	}

	requestURL := *r.cachedURL.Load()

	// todo: write file
	var err error
	r.bytes, err = io.ReadAll(r.resp.Body)
	if err != nil {
		return fmt.Errorf("[requests] %s %s read response failed: %w",
			r.method, requestURL, err)
	}

	r.logger.Info(r.Context(), "[requests] %s: %s, doRead: %s",
		r.method, requestURL, r.bytes)
	return nil
}

func (r *Request) requestURL() string {
	reqURL, err := url.Parse(r.url)
	if err != nil {
		return r.url
	}
	q := reqURL.Query()
	for k, v := range r.query {
		q[k] = append(q[k], v...)
	}
	reqURL.RawQuery = q.Encode()
	return reqURL.String()
}

func (r *Request) cachedRequestURL() string {
	return *r.cachedURL.Load()
}

func (r *Request) doRequestFactor(f func() error) error {
	if r.err != nil {
		return r.err
	}

	r.err = f()
	return r.err
}
