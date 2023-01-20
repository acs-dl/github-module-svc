package github

import (
	"encoding/json"
	"fmt"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

func (g *github) GetUserIdFromApi(username string) (*int64, error) {
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

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("error in response from API, status %s", res.Status))
	}

	var returned data.Permission
	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return &returned.GithubId, nil
}
