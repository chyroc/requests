package requests_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chyroc/requests"
)

func Example_new() {
	text := requests.New(http.MethodGet, "https://httpbin.org/get").WithTimeout(time.Second * 10).Text()
	fmt.Println("text", text.Unwrap())
}

func Example_factory() {
	// I hope to set fixed parameters every time I initiate a request

	// Then, every request created by this factory will not log
	opt := requests.Options(
		requests.WithLogger(requests.DiscardLogger()),
		requests.WithTimeout(time.Second*10),
	)

	// Send sample request
	text := requests.New(
		http.MethodGet, "https://httpbin.org/get", opt...,
	).Text()
	fmt.Println("text", text.Unwrap())
}

func Example_newSession() {
	session := requests.NewSession("/tmp/requests-session.txt")
	text := session.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").WithTimeout(time.Second * 10).Text()
	fmt.Println("text", text.Unwrap())
}
