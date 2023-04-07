package github

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

func (g *github) getDuration(response *http.Response) (time.Duration, error) {
	if response.Header.Get("Retry-After") != "" {
		durationInSeconds, err := strconv.ParseInt(response.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse `retry-after `header")
		}
		return time.Duration(durationInSeconds), nil
	}

	if response.Header.Get("x-ratelimit-reset") != "" {
		timeInUTCSeconds, err := strconv.ParseInt(response.Header.Get("x-ratelimit-reset"), 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse `retry-after `header")
		}

		parsedTime := time.Unix(timeInUTCSeconds, 0)
		currentTime := time.Now()

		durationInSeconds := parsedTime.Sub(currentTime)

		return durationInSeconds, nil
	}

	return 0, errors.Errorf("failed to retrive time from headers")
}
