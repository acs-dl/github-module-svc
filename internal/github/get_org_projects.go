package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) GetProjectsFromApi(link string) ([]data.Sub, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/orgs/%s/repos", link), nil)
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.superToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	var result []data.Sub

	for page := 1; ; page++ {
		q := req.URL.Query()
		q.Set("per_page", "100")
		q.Set("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, " error making http request")
		}

		if res.StatusCode == http.StatusForbidden {
			timeoutDuration, err := g.getDuration(res)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get time duration from response")
			}
			g.log.Warnf("we need to wait `%d`", timeoutDuration)
			time.Sleep(timeoutDuration)
			page--
			continue
		}

		if res.StatusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("error in response, status %s", res.Status))
		}

		var returned []data.Sub
		if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
			return nil, errors.Wrap(err, " failed to unmarshal body")
		}

		if len(returned) == 0 {
			break
		}

		result = append(result, returned...)
	}

	return result, nil
}
