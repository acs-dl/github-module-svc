package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) FindRepoOwner(link string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s", link), nil)
	if err != nil {
		return "", errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.superToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, " error making http request")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return "", errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.FindRepoOwner(link)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", errors.New(fmt.Sprintf("failed to process request: bad status (%s)", res.Status))
	}

	returned := struct {
		Owner struct {
			Type string `json:"type"`
		} `json:"owner"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return "", errors.Wrap(err, " failed to unmarshal body")
	}

	return returned.Owner.Type, nil
}
