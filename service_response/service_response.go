package service_response

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"
)

type ResponseResult struct {
	Site     string
	Duration time.Duration
	Err      error `json:",omitempty"`
}

func serviceResponseTime(ctx context.Context, url string) (time.Duration, error) {
	var start time.Time
	var firstByte time.Duration

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			firstByte = time.Since(start)
		},
	}
	req, err := http.NewRequest("GET", "https://"+url, nil)
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	client := http.Client{Timeout: 30 * time.Second}

	start = time.Now()
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	}

	return firstByte, err
}

func ServiceResponseTimeWorker(
	ctx context.Context,
	siteChan chan string,
	resultChan chan ResponseResult,
) {
	for site := range siteChan {
		r := ResponseResult{Site: site}
		r.Duration, r.Err = serviceResponseTime(ctx, site)
		resultChan <- r
	}
}
