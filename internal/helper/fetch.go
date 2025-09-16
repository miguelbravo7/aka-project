package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type APIResponse struct {
	Info struct {
		Next  string `json:"next"`
		Prev  string `json:"prev"`
		Count int    `json:"count"`
		Pages int    `json:"pages"`
	} `json:"info"`
	Results []json.RawMessage `json:"results"`
}

func FetchPage(ctx context.Context, url string) (*APIResponse, error) {
	tracer := otel.Tracer("aka-project/internal/helper")
	ctx, span := tracer.Start(ctx, "FetchPage",
		trace.WithAttributes(attribute.String("http.url", url)))
	defer span.End()

	client := otelhttp.DefaultClient

	var resp *http.Response
	var err error
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err = client.Do(req)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("fetch.error", err.Error()))
			fmt.Println("Transient error, retrying:", err)
			time.Sleep(time.Duration(attempt+1) * time.Second) // simple backoff
			continue
		}
		if resp.StatusCode == 429 { // rate limit
			span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
			retryAfter := resp.Header.Get("Retry-After")
			wait := 5 * time.Second
			if dur, err := time.ParseDuration(retryAfter + "s"); err == nil {
				wait = dur
			}
			fmt.Println("Rate limited, waiting:", wait)
			time.Sleep(wait)
			continue
		}
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			// server error, retry
			span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
			fmt.Println("Server error, retrying...")
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		break
	}
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("fetch.final_error", err.Error()))
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close response body:", err)
		}
	}()
	body, _ := io.ReadAll(resp.Body)

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("json.unmarshal_error", err.Error()))
		return nil, err
	}

	return &apiResp, nil
}
