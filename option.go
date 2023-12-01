package requests

import (
	"time"
)

type RequestOption[T any] func(req *Request[T]) error

func WithLogger[T any](logger Logger) RequestOption[T] {
	return func(req *Request[T]) error {
		req.WithLogger(logger)
		return nil
	}
}

func WithTimeout[T any](timeout time.Duration) RequestOption[T] {
	return func(req *Request[T]) error {
		req.WithTimeout(timeout)
		return nil
	}
}

func WithHeader[T any](key, val string) RequestOption[T] {
	return func(req *Request[T]) error {
		req.WithHeader(key, val)
		return nil
	}
}

func WithQuery[T any](key, val string) RequestOption[T] {
	return func(req *Request[T]) error {
		req.WithQuery(key, val)
		return nil
	}
}
