package requests_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chyroc/anyhow"
	"github.com/chyroc/requests"
)

func Example_Method() {
	r := requests.Get("https://httpbin.org/get")

	r = requests.Post("https://httpbin.org/post")

	r = requests.Delete("https://httpbin.org/delete")

	r = requests.New(http.MethodPut, "https://httpbin.org/put")

	fmt.Println(r.Text().Unpack())
}

func Example_unmarshal() {
	type Data struct{}

	f := func() (Data, error) {
		type Response struct {
			Code int32
			Data Data
		}
		return anyhow.AndThen11(requests.JSON[Response](requests.Post("https://httpbin.org/post")), func(t1 *Response) anyhow.Result1[Data] {
			if t1.Code != 0 {
				return anyhow.Err1[Data](fmt.Errorf("fail: %d", t1.Code))
			}
			return anyhow.Ok1(t1.Data)
		}).Unpack()
	}
	f()
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
