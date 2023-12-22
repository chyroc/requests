package requests

import (
	"time"
)

type Option func(req *Request)

func Options(options ...Option) []Option {
	return options
}

func WithLogger(logger Logger) Option {
	return func(req *Request) {
		req.WithLogger(logger)
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(req *Request) {
		req.WithTimeout(timeout)
	}
}

func WithHeader(key, val string) Option {
	return func(req *Request) {
		req.WithHeader(key, val)
	}
}

func WithHeaders(kv map[string]string) Option {
	return func(req *Request) {
		req.WithHeaders(kv)
	}
}

func WithQuery(key, val string) Option {
	return func(req *Request) {
		req.WithQuery(key, val)
	}
}

func WithQueries(queries any) Option {
	return func(req *Request) {
		req.WithQueries(queries)
	}
}

func applyOpt(r *Request, options ...Option) {
	for _, opt := range options {
		opt(r)
	}
}
