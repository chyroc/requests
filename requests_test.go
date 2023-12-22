package requests_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chyroc/requests"
	"github.com/stretchr/testify/assert"
)

func joinHttpBinURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "https://httpbin.org" + path
}

func Test_Real(t *testing.T) {
	as := assert.New(t)

	t.Run("/ip", func(t *testing.T) {
		type IP struct {
			Origin string `json:"origin"`
		}
		res := requests.JSON[IP](
			requests.New(http.MethodGet, joinHttpBinURL("/ip")),
		).Unwrap()
		as.NotEmpty(res.Origin)
	})

	t.Run("/user-agent", func(t *testing.T) {
		type Resp struct {
			UserAgent string `json:"user-agent"`
		}
		resp := requests.JSON[Resp](
			requests.New(http.MethodGet, joinHttpBinURL("/user-agent")),
		).Unwrap()
		as.True(regexp.MustCompile(`chyroc-requests/v\d+.\d+.\d+ \(https://github.com/chyroc/requests\)`).MatchString(resp.UserAgent),
			fmt.Sprintf("%s not match user-agent", resp.UserAgent))
	})

	t.Run("/headers", func(t *testing.T) {
		type Resp struct {
			Headers struct {
				A string `json:"A"`
				B string `json:"B"`
			} `json:"headers"`
		}
		resp := requests.JSON[Resp](
			requests.New(http.MethodGet, joinHttpBinURL("/headers")).WithHeader(
				"a", "1",
			).WithHeaders(map[string]string{
				"a": "2",
				"b": "3",
			}),
		).Unwrap()
		as.Equal("1,2", resp.Headers.A)
		as.Equal("3", resp.Headers.B)
	})

	t.Run("/get", func(t *testing.T) {
		type Resp struct {
			Args struct {
				A []string `json:"a"`
				B string   `json:"b"`
			} `json:"args"`
		}
		resp := requests.JSON[Resp](
			requests.New(http.MethodGet, joinHttpBinURL("/get")).
				WithQuery("a", "1").
				WithQueries(map[string]string{
					"a": "2",
					"b": "3",
				}),
		).Unwrap()
		as.Equal([]string{"1", "2"}, resp.Args.A)
		as.Equal("3", resp.Args.B)
	})

	t.Run("/status", func(t *testing.T) {
		status := requests.New(http.MethodGet, joinHttpBinURL("/status/403")).Status().Unwrap()
		as.Equal(403, status)
	})

	t.Run("/delay/3", func(t *testing.T) {
		err := requests.New(
			http.MethodGet, joinHttpBinURL("/delay/4"),
		).WithTimeout(time.Second).Text().Err()
		as.NotNil(err)
		as.Contains(err.Error(), "context deadline exceeded")
	})

	t.Run("/image", func(t *testing.T) {
		t.Skip()

		requests.New(http.MethodGet, joinHttpBinURL("/image")).Text()
	})

	t.Run("/post file", func(t *testing.T) {
		type Resp struct {
			Files struct {
				File string `json:"file"`
			} `json:"files"`
			Form map[string]string `json:"form"`
		}
		resp := requests.JSON[Resp](
			requests.New(
				http.MethodPost, joinHttpBinURL("/post"),
			).
				WithFile("1.txt", strings.NewReader("hi"), "file", map[string]string{"field1": "val1", "field2": "val2"}).
				WithTimeout(time.Second * 3),
		).Unwrap()
		as.Equal("hi", resp.Files.File)
		as.Equal("val1", resp.Form["field1"])
	})

	t.Run("fail", func(t *testing.T) {
		err := requests.New(
			http.MethodGet, "",
		).WithTimeout(time.Second).Text().Err()
		as.NotNil(err)
	})

	// https://github.com/postmanlabs/httpbin/issues/653
	t.Run("session", func(t *testing.T) {
		t.Skip()

		go newTestHttpServer()
		time.Sleep(time.Second * 2)

		file := ""
		{
			sessionFile, err := ioutil.TempFile(os.TempDir(), "session-*")
			as.Nil(err)
			t.Logf("session file: %s", sessionFile.Name())
			as.Nil(ioutil.WriteFile(sessionFile.Name(), []byte("[]"), 0o666))
			file = sessionFile.Name()
			t.Logf("file: %s", file)

			s := requests.NewSession(sessionFile.Name())

			fmt.Println(s.New(http.MethodGet, "http://127.0.0.1:5100/set-cookies?a=b&c=d").Header())

			resp := s.New(http.MethodGet, "http://127.0.0.1:5100/get-cookies").Map().Unwrap()
			as.Equal("b", resp["a"])
		}

		{
			as.Nil(os.Rename(file, file+".bak"))
			s := requests.NewSession(file + ".bak")
			resp := s.New(http.MethodGet, "http://127.0.0.1:5100/get-cookies").Map().Unwrap()
			as.Equal("b", resp["a"])
		}
	})
}

func Test_Factory(t *testing.T) {
	as := assert.New(t)

	t.Run("", func(t *testing.T) {
		opt := requests.Options(
			requests.WithTimeout(time.Second * 10),
		)
		type Resp struct {
			Origin string `json:"origin"`
		}
		resp := requests.JSON[Resp](
			requests.New(http.MethodGet, joinHttpBinURL("/ip"), opt...),
		).Unwrap()
		as.NotEmpty(resp.Origin)
	})

	t.Run("", func(t *testing.T) {
		opt := requests.Options(
			requests.WithTimeout(time.Second*10),
			func(req *requests.Request) {
				req.SetError(fmt.Errorf("must fail"))
			},
		)
		err := requests.New(http.MethodGet, joinHttpBinURL("/ip"), opt...).Text().Err()
		as.NotNil(err)
		as.Equal("must fail", err.Error())
	})
}

func newTestHttpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/get-cookies", func(writer http.ResponseWriter, request *http.Request) {
		m := map[string][]string{}
		for _, v := range request.Cookies() {
			m[v.Name] = append(m[v.Name], v.Value)
		}
		bs, _ := json.Marshal(m)
		if _, err := writer.Write(bs); err != nil {
			panic(err)
		}
		writer.WriteHeader(200)
	})
	mux.HandleFunc("/set-cookies", func(writer http.ResponseWriter, request *http.Request) {
		for k, v := range request.URL.Query() {
			for _, vv := range v {
				writer.Header().Add("cookie", fmt.Sprintf("%s=%s; Path=/; Host=127.0.0.1:5100; Max-Age=99999", k, vv))
			}
		}

		writer.WriteHeader(200)
	})
	err := http.ListenAndServe("127.0.0.1:5100", mux)
	if err != nil {
		panic(err)
	}
}
