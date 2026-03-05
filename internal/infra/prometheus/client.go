package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type Client interface {
	Query(ctx context.Context, query string, ts time.Time) (*QueryResult, error)
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error)
	Metadata(ctx context.Context, metric string) ([]MetadataItem, error)
}

type HTTPClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	retryCount int
}

func NewClient(cfg Config) (*HTTPClient, error) {
	normalized := cfg.Normalize()
	if strings.TrimSpace(normalized.Address) == "" {
		return nil, fmt.Errorf("prometheus address is empty")
	}
	u, err := url.Parse(normalized.Address)
	if err != nil {
		return nil, err
	}
	return &HTTPClient{
		baseURL: u,
		httpClient: &http.Client{
			Timeout: normalized.Timeout,
		},
		retryCount: normalized.RetryCount,
	}, nil
}

func (c *HTTPClient) Query(ctx context.Context, query string, ts time.Time) (*QueryResult, error) {
	q := url.Values{}
	q.Set("query", query)
	q.Set("time", strconv.FormatFloat(float64(ts.Unix()), 'f', -1, 64))
	body, err := c.doGet(ctx, "/api/v1/query", q)
	if err != nil {
		return nil, err
	}
	env, err := parseQueryEnvelope(body)
	if err != nil {
		return nil, err
	}
	var data queryData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, err
	}
	return &QueryResult{ResultType: data.ResultType, Vector: data.Result}, nil
}

func (c *HTTPClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResult, error) {
	q := url.Values{}
	q.Set("query", query)
	q.Set("start", strconv.FormatFloat(float64(start.Unix()), 'f', -1, 64))
	q.Set("end", strconv.FormatFloat(float64(end.Unix()), 'f', -1, 64))
	q.Set("step", strconv.Itoa(int(step.Seconds())))
	body, err := c.doGet(ctx, "/api/v1/query_range", q)
	if err != nil {
		return nil, err
	}
	env, err := parseQueryEnvelope(body)
	if err != nil {
		return nil, err
	}
	var data rangeData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, err
	}
	return &QueryResult{ResultType: data.ResultType, Matrix: data.Result}, nil
}

func (c *HTTPClient) Metadata(ctx context.Context, metric string) ([]MetadataItem, error) {
	q := url.Values{}
	if strings.TrimSpace(metric) != "" {
		q.Set("metric", metric)
	}
	body, err := c.doGet(ctx, "/api/v1/metadata", q)
	if err != nil {
		return nil, err
	}
	env, err := parseQueryEnvelope(body)
	if err != nil {
		return nil, err
	}
	var raw map[string][]MetadataItem
	if err := json.Unmarshal(env.Data, &raw); err != nil {
		return nil, err
	}
	items := make([]MetadataItem, 0)
	for metricName, defs := range raw {
		for _, d := range defs {
			d.Metric = metricName
			items = append(items, d)
		}
	}
	return items, nil
}

func (c *HTTPClient) doGet(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	var lastErr error
	for i := 0; i < c.retryCount; i++ {
		u := *c.baseURL
		u.Path = path.Join(c.baseURL.Path, endpoint)
		u.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("prometheus http %d", resp.StatusCode)
			continue
		}
		return body, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("prometheus request failed")
	}
	return nil, lastErr
}
