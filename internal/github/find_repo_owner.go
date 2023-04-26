package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) FindRepositoryOwner(link string) (string, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/repos/%s", link),
		Body:   nil,
		Query:  nil,
		Header: map[string]string{
			"Accept":               "application/vnd.Github+json",
			"Authorization":        "Bearer " + g.superToken,
			"X-GitHub-Api-Version": "2022-11-28",
		},
		Timeout: time.Second * 30,
	}

	res, err := helpers.MakeHttpRequest(params)
	if err != nil {
		return "", errors.Wrap(err, "failed to make http request")
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return "", errors.Wrap(err, "failed to check response status code")
	}
	if res == nil {
		return "", errors.Errorf("error in response, status %v", res.StatusCode)
	}

	response := struct {
		Owner struct {
			Type string `json:"type"`
		} `json:"owner"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal body")
	}

	return response.Owner.Type, nil
}
