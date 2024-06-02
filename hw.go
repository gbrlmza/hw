package hw

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type (
	OptionsFn   func(r *RequestOpts) error
	RequestOpts struct {
		method     string
		url        string
		body       io.Reader
		headers    http.Header
		params     url.Values
		httpClient *http.Client
	}
)

// Do sends an HTTP request and returns an HTTP response.
func Do(ctx context.Context, method string, url string, options ...OptionsFn) (*http.Response, error) {
	c := &RequestOpts{
		method:     method,
		url:        url,
		httpClient: http.DefaultClient,
		headers:    make(map[string][]string),
		params:     make(map[string][]string),
	}
	for _, o := range options {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	return do(ctx, c)
}

// WithBody sets the body of the request.
func WithBody(body io.Reader) OptionsFn {
	return func(r *RequestOpts) error {
		r.body = body
		return nil
	}
}

// WithJsonBody sets the body of the request as a JSON object.
func WithJsonBody(body interface{}) OptionsFn {
	return func(r *RequestOpts) error {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return err
		}

		r.body = buf
		r.headers.Set("Content-Type", "application/json")
		return nil
	}
}

// WithParam sets a query parameter.
func WithParam(key, value string) OptionsFn {
	return func(r *RequestOpts) error {
		r.params.Set(key, value)
		return nil
	}
}

// WithHeader sets a header.
func WithHeader(key, value string) OptionsFn {
	return func(r *RequestOpts) error {
		r.headers.Set(key, value)
		return nil
	}
}

// WithHttpClient sets the HTTP client to be used.
func WithHttpClient(httpClient *http.Client) OptionsFn {
	return func(r *RequestOpts) error {
		if httpClient == nil {
			return errors.New("nil httpClient")
		}
		r.httpClient = httpClient
		return nil
	}
}

// do builds an sends an HTTP request and returns an HTTP response.
func do(ctx context.Context, opts *RequestOpts) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, opts.method, opts.url, opts.body)
	if err != nil {
		return nil, err
	}

	req.Header = opts.headers

	if len(opts.params) > 0 {
		if req.URL.RawQuery == "" {
			req.URL.RawQuery = opts.params.Encode()
		} else {
			q := req.URL.Query()
			for k, v := range opts.params {
				q[k] = v
			}
			req.URL.RawQuery = q.Encode()
		}
	}

	return opts.httpClient.Do(req)
}
