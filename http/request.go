package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	xhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ability-sh/abi-lib/json"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type HTTPResponse interface {
	Code() int
	Headers() map[string]string
	Body() []byte
	PraseBody() (interface{}, error)
}

type HTTPRequest interface {
	SetURL(baseURL string, query map[string]string) HTTPRequest
	SetHeaders(headers map[string]string) HTTPRequest
	SetBody(body []byte) HTTPRequest
	SetUrlencodeBody(body map[string]string) HTTPRequest
	SetMultipartBody(cb func(w *multipart.Writer)) HTTPRequest
	SetJSONBody(interface{}) HTTPRequest
	SetTimeout(timeout time.Duration) HTTPRequest
	SetOutput(out io.Writer) HTTPRequest
	Send() (HTTPResponse, error)
	SendWithClient(client *xhttp.Client) (HTTPResponse, error)
	SetClient(client *xhttp.Client) HTTPRequest
}

func NewHTTPRequest(method string) HTTPRequest {
	return &httpRequest{method: method}
}

type httpRequest struct {
	url     string
	method  string
	headers map[string]string
	timeout time.Duration
	body    []byte
	output  io.Writer
	client  *xhttp.Client
}

func (r *httpRequest) SetURL(baseURL string, query map[string]string) HTTPRequest {
	if len(query) == 0 {
		r.url = baseURL
	} else {
		b := bytes.NewBuffer(nil)
		b.WriteString(baseURL)
		if strings.HasSuffix(baseURL, "?") || strings.HasSuffix(baseURL, "&") {

		} else if strings.Contains(baseURL, "?") {
			b.WriteString("&")
		} else {
			b.WriteString("?")
		}
		n := 0
		for key, value := range query {
			if n != 0 {
				b.WriteString("&")
			}
			b.WriteString(key)
			b.WriteString("=")
			b.WriteString(url.QueryEscape(value))
			n = n + 1
		}
		r.url = b.String()
	}
	return r
}

func (r *httpRequest) SetUrlencodeBody(body map[string]string) HTTPRequest {

	b := bytes.NewBuffer(nil)
	n := 0

	for key, value := range body {
		if n != 0 {
			b.WriteString("&")
		}
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(url.QueryEscape(value))
		n = n + 1
	}

	if r.headers == nil {
		r.headers = map[string]string{}
	}

	r.headers["Content-Type"] = "application/x-www-form-urlencoded"

	r.body = b.Bytes()

	return r
}

func (r *httpRequest) SetMultipartBody(cb func(w *multipart.Writer)) HTTPRequest {
	b := bytes.NewBuffer(nil)
	w := multipart.NewWriter(b)
	cb(w)
	w.Close()
	if r.headers == nil {
		r.headers = map[string]string{}
	}
	r.headers["Content-Type"] = fmt.Sprintf("multipart/form-data; boundary=%s", w.Boundary())
	r.body = b.Bytes()
	return r
}

func (r *httpRequest) SetJSONBody(body interface{}) HTTPRequest {
	if r.headers == nil {
		r.headers = map[string]string{}
	}
	r.headers["Content-Type"] = "application/json"
	r.body, _ = json.Marshal(body)
	return r
}

func (r *httpRequest) SetHeaders(headers map[string]string) HTTPRequest {
	if r.headers == nil {
		r.headers = headers
	} else {
		for k, v := range headers {
			r.headers[k] = v
		}
	}
	return r
}

func (r *httpRequest) SetBody(body []byte) HTTPRequest {
	r.body = body
	return r
}

func (r *httpRequest) SetTimeout(timeout time.Duration) HTTPRequest {
	r.timeout = timeout
	return r
}

func (r *httpRequest) SetOutput(out io.Writer) HTTPRequest {
	r.output = out
	return r
}

func (r *httpRequest) Send() (HTTPResponse, error) {
	if r.client != nil {
		return r.SendWithClient(r.client)
	}
	return r.SendWithClient(GetClient())
}

func (r *httpRequest) SetClient(client *xhttp.Client) HTTPRequest {
	r.client = client
	return r
}

func (r *httpRequest) SendWithClient(client *xhttp.Client) (HTTPResponse, error) {

	req, err := xhttp.NewRequest(r.method, r.url, bytes.NewReader(r.body))

	if err != nil {
		return nil, err
	}

	if r.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	var body []byte = nil

	if r.output != nil && resp.StatusCode == 200 {

		_, err = io.Copy(r.output, resp.Body)

		resp.Body.Close()

		if err != nil && err != io.EOF {
			return nil, err
		}

	} else {

		body, err = ioutil.ReadAll(resp.Body)

		resp.Body.Close()

		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	headers := map[string]string{}

	for k, vs := range resp.Header {
		headers[k] = vs[len(vs)-1]
	}

	return &httpResponse{code: resp.StatusCode, headers: headers, body: body}, nil
}

type httpResponse struct {
	code    int
	headers map[string]string
	body    []byte
}

func (res *httpResponse) Code() int {
	return res.code
}

func (res *httpResponse) Headers() map[string]string {
	return res.headers
}

func (res *httpResponse) Body() []byte {
	return res.body
}

func (res *httpResponse) PraseBody() (interface{}, error) {

	contentType := res.headers["Content-Type"]

	if contentType == "" {
		contentType = res.headers["content-type"]
	}

	contentType = strings.ToLower(contentType)

	var err error = nil
	var b = res.body

	if strings.Contains(contentType, "charset=gbk") {
		rd := transform.NewReader(bytes.NewReader(res.body), simplifiedchinese.GBK.NewDecoder())
		b, err = ioutil.ReadAll(rd)
		if err != nil {
			return nil, err
		}
	} else if strings.Contains(contentType, "charset=gb2312") {
		rd := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GB18030.NewDecoder())
		b, err = ioutil.ReadAll(rd)
		if err != nil {
			return nil, err
		}
	}

	if strings.Contains(contentType, "json") {
		var rs interface{}
		err = json.Unmarshal(b, &rs)
		if err != nil {
			return nil, err
		}
		return rs, nil
	} else if strings.Contains(contentType, "text") {
		return string(b), nil
	}

	return b, nil
}
