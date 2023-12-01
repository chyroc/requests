package requests

import (
	"crypto/tls"
	"net/http"
)

type wrapRoundTripper struct {
	isIgnoreSSL bool
	wrapResp    func(resp *http.Response) (*http.Response, error)
}

func (lf wrapRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var rt = &http.Transport{}
	if lf.isIgnoreSSL {
		rt.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	resp, err := rt.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if lf.wrapResp != nil {
		resp, err = lf.wrapResp(resp)
		if err != nil {
			return nil, err
		}
	}

	return resp, err
}
