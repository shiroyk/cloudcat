// Package http the http JS implementation
package http

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"text/template"

	"github.com/dop251/goja"
	"github.com/shiroyk/cloudcat/di"
	"github.com/shiroyk/cloudcat/fetch"
	"github.com/shiroyk/cloudcat/js/modules"
)

// Module js module
type Module struct{}

// Exports returns module instance
func (*Module) Exports() any {
	return &Http{di.MustResolve[fetch.Fetch]()}
}

func init() {
	modules.Register("http", &Module{})
	modules.Register("FormData", &NativeFormData{})
	modules.Register("URLSearchParams", &NativeURLSearchParams{})
}

// Http module for fetching resources (including across the network).
type Http struct { //nolint
	fetch fetch.Fetch
}

// handleBody process the send request body and set the content-type
func handleBody(body any, header map[string]string) (any, error) {
	switch data := body.(type) {
	case FormData:
		buf := &bytes.Buffer{}
		mpw := multipart.NewWriter(buf)
		for k, v := range data.data {
			for _, ve := range v {
				if f, ok := ve.(FileData); ok {
					// Creates a new form-data header with the provided field name and file name.
					fw, err := mpw.CreateFormFile(k, f.Filename)
					if err != nil {
						return nil, err
					}
					// Write bytes to the part
					if _, err := fw.Write(f.Data); err != nil {
						return nil, err
					}
				} else {
					// Write string value
					if err := mpw.WriteField(k, fmt.Sprintf("%v", v)); err != nil {
						return nil, err
					}
				}
			}
		}
		header["Content-Type"] = mpw.FormDataContentType()
		if err := mpw.Close(); err != nil {
			return nil, err
		}
		return buf, nil
	case URLSearchParams:
		header["Content-Type"] = "application/x-www-form-url"
		return data.encode(), nil
	case goja.ArrayBuffer:
		return data.Bytes(), nil
	case []byte, map[string]any, string, nil:
		return body, nil
	default:
		return nil, fmt.Errorf("unsupported request body type %v", body)
	}
}

// Get Make a GET request with URL and optional headers.
func (h *Http) Get(u string, header map[string]string) (*Response, error) {
	res, err := h.fetch.Get(u, header)
	if err != nil {
		return nil, err
	}

	return NewResponse(res), nil
}

// Post Make a POST request with URL, optional body, optional headers.
// Send POST with multipart:
// http.post(url, new FormData({'bytes': new Uint8Array([0]).buffer}))
// Send POST with x-www-form-urlencoded:
// http.post(url, new URLSearchParams({'key': 'foo', 'value': 'bar'}))
// Send POST with json:
// http.post(url, {'key': 'foo'})
func (h *Http) Post(u string, body any, header map[string]string) (*Response, error) {
	if header == nil {
		header = make(map[string]string)
	}

	var err error
	body, err = handleBody(body, header)
	if err != nil {
		return nil, err
	}

	res, err := h.fetch.Post(u, body, header)
	if err != nil {
		return nil, err
	}

	return NewResponse(res), nil
}

// Head Make a HEAD request with URL and optional headers.
func (h *Http) Head(u string, header map[string]string) (*Response, error) {
	res, err := h.fetch.Head(u, header)
	if err != nil {
		return nil, err
	}

	return NewResponse(res), nil
}

// Request Make a request with method and URL, optional body, optional headers.
func (h *Http) Request(method, u string, body any, header map[string]string) (*Response, error) {
	if header == nil {
		header = make(map[string]string)
	}

	var err error
	body, err = handleBody(body, header)
	if err != nil {
		return nil, err
	}

	res, err := h.fetch.Request(method, u, body, header)
	if err != nil {
		return nil, err
	}

	return NewResponse(res), nil
}

// Template Make a request with an HTTP template, template argument.
func (h *Http) Template(tpl string, arg map[string]any) (*Response, error) {
	funcs, _ := di.Resolve[template.FuncMap]()

	req, err := fetch.NewTemplateRequest(funcs, tpl, arg)
	if err != nil {
		return nil, err
	}

	res, err := h.fetch.DoRequest(req)
	if err != nil {
		return nil, err
	}

	return NewResponse(res), nil
}

// SetProxy set the proxy URLs for the specified URL.
func (h *Http) SetProxy(u string, proxyURL ...string) {
	fetch.AddRoundRobinProxy(u, proxyURL...)
}
