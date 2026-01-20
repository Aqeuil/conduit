package curl

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
)

type HttpClientInstance struct {
	Client *resty.Client
	Req    *resty.Request
	logger *log.Helper
	cnt    int
	t      time.Duration
}

type logAttr struct {
	method    string
	url       string
	headers   map[string][]string
	param     interface{}
	resp      *resty.Response
	startTime time.Time

	err    error
	shrink time.Duration
}

func NewHttpClientInstance(t time.Duration, retryCnt int, logger log.Logger) *HttpClientInstance {
	c := resty.New().SetRetryCount(retryCnt)
	if t > time.Second {
		c.SetTimeout(t)
	}

	return &HttpClientInstance{
		Client: c,
		Req:    c.R(),
		logger: log.NewHelper(log.With(logger, "x_module", "pkg/http-client")),
		cnt:    retryCnt,
		t:      t,
	}
}

func (hc *HttpClientInstance) HttpGet(ctx context.Context, url string, query map[string]string, headers map[string][]string) (*resty.Response, error) {
	startTime := time.Now()
	hc.Req = hc.Client.NewRequest()
	if headers != nil {
		hc.Req.SetHeaderMultiValues(headers)
	}

	if query != nil {
		hc.Client.SetQueryParams(query)
	}

	resp, err := hc.Req.Get(url)
	hc.log(ctx, logAttr{
		method:    "http-post",
		startTime: startTime,
		url:       url,
		headers:   headers,
		param:     query,
		resp:      resp,
		err:       err,
	})

	if err != nil || resp == nil {
		return nil, err
	}

	return resp, nil
}

func (hc *HttpClientInstance) HttpPost(ctx context.Context, url string, headers map[string][]string, body interface{}) (*resty.Response, error) {
	startTime := time.Now()
	hc.Req = hc.Client.NewRequest()
	if headers != nil {
		hc.Req.SetHeaderMultiValues(headers)
	}

	if body != nil {
		hc.Req.SetBody(body)
	}

	resp, err := hc.Req.Post(url)
	hc.log(ctx, logAttr{
		method:    "http-post",
		startTime: startTime,
		url:       url,
		headers:   headers,
		param:     body,
		resp:      resp,
		err:       err,
	})
	if err != nil || resp == nil {
		return nil, err
	}

	return resp, nil
}

func (hc *HttpClientInstance) HttpPostForm(ctx context.Context, url string, headers map[string][]string, form map[string]string) (*resty.Response, error) {
	startTime := time.Now()

	c := resty.New().SetRetryCount(hc.cnt)
	if hc.t > time.Second {
		c.SetTimeout(hc.t)
	}

	req := c.R()
	if headers != nil {
		req.SetHeaderMultiValues(headers)
	}

	if len(form) > 0 {
		req.SetFormData(form)
	}

	resp, err := req.Post(url)
	hc.log(ctx, logAttr{
		method:    "http-post-form",
		startTime: startTime,
		url:       url,
		headers:   headers,
		param:     form,
		resp:      resp,
		err:       err,
	})
	if err != nil || resp == nil {
		return nil, err
	}

	return resp, nil
}

func (hc *HttpClientInstance) log(ctx context.Context, attr logAttr) {
	var (
		x_shrink float64
		x_error  string
	)

	x_shrink = attr.shrink.Seconds()
	if attr.err != nil {
		x_error = attr.err.Error()
	}

	hc.logger.WithContext(ctx).Infow(
		"x_action", attr.method,
		"x_url", attr.url,
		"x_headers", attr.headers,
		"x_params", attr.param,
		"x_response", cast.ToString(attr.resp),
		"x_shrink", x_shrink,
		"x_error", x_error,
		"x_duration", time.Since(attr.startTime).Seconds(),
	)
}
