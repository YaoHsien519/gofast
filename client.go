package gofast

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

type Header map[string]string

type Client struct {
	fastClient      *fasthttp.Client
	errorHandler    ErrorHandler
	requestEncoder  RequestEncoder
	responseDecoder ResponseDecoder
}

func New(config ...Config) *Client {
	var cfg Config

	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Name == "" {
		cfg.Name = defaultUserAgent
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaultTimeout
	}
	if cfg.RequestEncoder == nil {
		cfg.RequestEncoder = jsonRequestEncoder
	}
	if cfg.ResponseDecoder == nil {
		cfg.ResponseDecoder = jsonResponseDecoder
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = defaultErrorHandler
	}

	return &Client{
		fastClient: &fasthttp.Client{
			Name:                     cfg.Name,
			NoDefaultUserAgentHeader: cfg.NoDefaultUserAgentHeader,
			ReadTimeout:              cfg.ReadTimeout,
			WriteTimeout:             cfg.WriteTimeout,
		},
		errorHandler:    cfg.ErrorHandler,
		requestEncoder:  cfg.RequestEncoder,
		responseDecoder: cfg.ResponseDecoder,
	}
}

func (c *Client) Get(uri string, out interface{}, header Header) error {
	return c.do(uri, fasthttp.MethodGet, nil, out, header)
}

func (c *Client) Post(uri string, in, out interface{}, header Header) error {
	return c.do(uri, fasthttp.MethodPost, in, out, header)
}

func (c *Client) Put(uri string, in, out interface{}, header Header) error {
	return c.do(uri, fasthttp.MethodPut, in, out, header)
}

func (c *Client) Patch(uri string, in, out interface{}, header Header) error {
	return c.do(uri, fasthttp.MethodPatch, in, out, header)
}

func (c *Client) Delete(uri string, in, out interface{}, header Header) error {
	return c.do(uri, fasthttp.MethodDelete, in, out, header)
}

func (c *Client) do(uri string, method string, in, out interface{}, header Header) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.SetRequestURI(uri)
	req.Header.SetMethod(method)
	for k, v := range header {
		req.Header.Set(k, v)
	}

	if in != nil {
		if err := c.requestEncoder(req, in); err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
	}

	if err := c.fastClient.Do(req, resp); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return c.errorHandler(resp)
	}

	if out != nil {
		if err := c.responseDecoder(resp, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
