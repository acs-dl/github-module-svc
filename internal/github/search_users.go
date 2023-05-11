package github

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) SearchByFromApi(username string) ([]data.User, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   "https://api.github.com/search/users",
		Body:   nil,
		Query: map[string]string{
			"q": username + " in:login",
		},
		Header: map[string]string{
			"Accept":               data.AcceptHeader,
			"Authorization":        "Bearer " + g.superUserToken,
			"X-GitHub-Api-Version": data.GithubApiVersionHeader,
		},
		Timeout: time.Second * 30,
	}

	res, err := helpers.MakeHttpRequest(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make http request")
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check response status code")
	}
	if res == nil {
		return nil, errors.Errorf("error in response, status %v", res.StatusCode)
	}

	var response struct {
		Items []data.User `json:"items"`
	}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	if len(response.Items) == 0 {
		return nil, nil
	}

	return response.Items, nil
}
