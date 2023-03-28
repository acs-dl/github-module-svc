package github

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) SearchByFromApi(username string) ([]data.User, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/search/users", nil)
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	q := req.URL.Query()
	q.Add("q", fmt.Sprintf("%s in:login", username))
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("error in response from API, status %s", res.Status))
	}

	var returned struct {
		Items []data.User `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	if len(returned.Items) == 0 {
		return nil, nil
	}

	return returned.Items, nil
}
