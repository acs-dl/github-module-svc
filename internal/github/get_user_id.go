package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) GetUserFromApi(username string) (*data.User, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/users/%s", username), nil)
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

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
		return g.GetUserFromApi(username)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("error in response from API, status %s", res.Status))
	}

	var response data.Permission
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return &data.User{
		Username:  response.Username,
		GithubId:  response.GithubId,
		AvatarUrl: response.AvatarUrl,
	}, nil
}
