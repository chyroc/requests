# requests

## Install

```shell
go get github.com/chyroc/requests
```

## Usage

sample usage

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chyroc/requests"
)

func Example_Method() {
	r := requests.Get("https://httpbin.org/get")

	r = requests.Post("https://httpbin.org/post")

	r = requests.Delete("https://httpbin.org/delete")

	r = requests.New(http.MethodPut, "https://httpbin.org/put")

	fmt.Println(r.Text().Unpack())
}

func Example_factory() {
	// I hope to set fixed parameters every time I initiate a request
	// Then, every request created by this factory will not log
	opt := requests.Options(
		requests.WithLogger(requests.DiscardLogger()),
		requests.WithTimeout(time.Second*10),
		requests.WithQuery("query", "value"),
		requests.WithHeader("Auth", "hey"),
	)

	// Send sample request
	r := requests.Get("https://httpbin.org/get", opt...)

	r = requests.Post("https://httpbin.org/get").
		WithLogger(requests.DiscardLogger()).
		WithTimeout(time.Second * 10)
	fmt.Println(r.Text().Unpack())
}

func Example_newSession() {
	session := requests.NewSession("/tmp/requests-session.txt")
	r := session.Get("https://jsonplaceholder.typicode.com/todos/1").
		WithTimeout(time.Second * 10)
	fmt.Println(r.Text().Unpack())
}

```