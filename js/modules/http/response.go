package http

import (
	"encoding/json"
	"net/http"

	"github.com/dop251/goja"
	"github.com/shiroyk/cloudcat/fetch"
)

// Response wraps the given response
type Response struct {
	body       []byte      // response body.
	Headers    http.Header // response headers.
	Status     int         // HTTP status code.
	StatusText string      // HTTP status message corresponding to the HTTP status code.
	Ok         bool        // true if response Status in the range 200-299.
}

// NewResponse returns a new Response
func NewResponse(res *fetch.Response) *Response {
	return &Response{
		body:       res.Body,
		Headers:    res.Header,
		Status:     res.StatusCode,
		StatusText: res.Status,
		Ok:         res.StatusCode >= 200 || res.StatusCode < 300,
	}
}

// String body resolves with a string. The response is always decoded using UTF-8.
func (r *Response) String() string {
	return string(r.body)
}

// Json parsing the body text as JSON.
func (r *Response) Json() any { //nolint
	j := make(map[string]any)
	_ = json.Unmarshal(r.body, &j)
	return j
}

// Bytes returns an ArrayBuffer.
func (r *Response) Bytes(_ goja.FunctionCall, vm *goja.Runtime) goja.Value {
	return vm.ToValue(vm.NewArrayBuffer(r.body))
}

// JsonEncodable allows custom JSON encoding by JSON.stringify()
func (r *Response) JsonEncodable() any { //nolint
	j := make(map[string]any)
	_ = json.Unmarshal(r.body, &j)
	return j
}
