package requests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// doRequest 发送请求
func (r *Request[T]) doRequest() error {
	return r.doRequestFactor(r.doInternalRequest)
}

func (r *Request[T]) parseRequestURL() string {
	reqURL, err := url.Parse(r.url)
	if err != nil {
		return r.url
	}
	q := reqURL.Query()
	for k, v := range r.querys {
		q[k] = append(q[k], v...)
	}
	reqURL.RawQuery = q.Encode()
	return reqURL.String()
}

// doRequest send request
func (r *Request[T]) doInternalRequest() error {
	if r.isRequest {
		return nil
	}

	r.cachedurl = r.parseRequestURL()

	r.logger.Info(r.Context(), "[gorequests] %s: %s, body=%s, header=%+v", r.method, r.cachedurl, r.rawBody, r.header)

	if r.persistentJar != nil {
		defer func() {
			if err := r.persistentJar.Save(); err != nil {
				r.logger.Error(r.Context(), "save cookie failed: %s", err)
			}
		}()
	}

	req, err := http.NewRequest(r.method, r.cachedurl, r.body)
	if err != nil {
		return fmt.Errorf("[requests] %s %s new request failed: %w", r.method, r.cachedurl, err)
	}

	req.Header = r.header

	// TODO: reuse client
	c := &http.Client{
		Timeout: r.timeout,
	}
	if r.isIgnoreSSL || r.wrapResponse != nil {
		c.Transport = &wrapRoundTripper{
			isIgnoreSSL: r.isIgnoreSSL,
			wrapResp:    r.wrapResponse,
		}
	}
	if r.persistentJar != nil {
		c.Jar = r.persistentJar
	}
	if r.isNoRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := c.Do(req)
	r.resp = resp
	r.isRequest = true
	if err != nil {
		return fmt.Errorf("[requests] %s %s send request failed: %w", r.method, r.cachedurl, err)
	}
	return nil
}

// doRead send request and read response
func (r *Request[T]) doRead() error {
	return r.doRequestFactor(func() error {
		if err := r.doInternalRequest(); err != nil {
			return err
		}

		if r.isRead {
			return nil
		}

		var err error
		r.bytes, err = ioutil.ReadAll(r.resp.Body)
		r.isRead = true
		if err != nil {
			return fmt.Errorf("[requests] %s %s read response failed: %w", r.method, r.cachedurl, err)
		}

		r.logger.Info(r.Context(), "[gorequests] %s: %s, doRead: %s", r.method, r.cachedurl, r.bytes)
		return nil
	})
}

func (r *Request[T]) doRequestFactor(f func() error) error {
	if r.err != nil {
		return r.err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.err = f()
	return r.err
}
