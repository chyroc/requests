package requests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest_Bytes(t *testing.T) {
	as := assert.New(t)
	as.Nil(nil)

	res := New[any](http.MethodPost, "https://httpbin.org/status/201")
	as.Equal(int(201), res.Status().Unwrap())
}
