package helpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func MakeHttpRequest(params data.RequestParams) (*data.ResponseParams, error) {
	req, err := http.NewRequest(params.Method, params.Link, bytes.NewReader(params.Body))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create request")
	}

	ctx, cancel := context.WithTimeout(context.Background(), params.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	if params.Header != nil {
		for key, value := range params.Header {
			req.Header.Set(key, value)
		}
	}

	if params.Query != nil {
		q := req.URL.Query()
		for key, value := range params.Query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making http request")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	return &data.ResponseParams{
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     response.Header,
		StatusCode: response.StatusCode,
	}, nil
}

func HandleHttpResponseStatusCode(response *data.ResponseParams, params data.RequestParams) (*data.ResponseParams, error) {
	switch status := response.StatusCode; {
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		return response, nil
	case status == http.StatusNotFound:
		return nil, nil
	case status == http.StatusForbidden:
		return HandleTooManyRequests(response.Header, params)
	case status < http.StatusOK || status >= http.StatusMultipleChoices:
		return nil, errors.New(fmt.Sprintf("error in response `%s`", http.StatusText(response.StatusCode)))
	default:
		return nil, errors.New(fmt.Sprintf("unexpected response status `%s`", http.StatusText(response.StatusCode)))
	}
}

func HandleTooManyRequests(header http.Header, params data.RequestParams) (*data.ResponseParams, error) {
	timeoutDuration, err := GetDuration(header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get time duration from response")
	}

	log.Printf("we need to wait `%s`", timeoutDuration.String())
	time.Sleep(timeoutDuration)

	return MakeHttpRequest(params)
}

func GetDuration(header http.Header) (time.Duration, error) {
	if header.Get("Retry-After") != "" {
		durationInSeconds, err := strconv.ParseInt(header.Get("Retry-After"), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse `retry-after `header")
		}
		return time.Duration(durationInSeconds), nil
	}

	if header.Get("x-ratelimit-reset") != "" {
		timeInUTCSeconds, err := strconv.ParseInt(header.Get("x-ratelimit-reset"), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse `retry-after `header")
		}

		parsedTime := time.Unix(timeInUTCSeconds, 0)
		currentTime := time.Now()

		duration := parsedTime.Sub(currentTime)

		return duration, nil
	}

	return 0, errors.New("failed to retrieve time from headers")
}
