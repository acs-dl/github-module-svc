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

func (g *github) AddOrUpdateUserInRepositoryFromApi(link, username, permission string) (*data.Permission, error) {
	jsonBody, err := json.Marshal(struct {
		Permission string `json:"permission"`
	}{
		Permission: permission,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal body")
	}

	params := data.RequestParams{
		Method: http.MethodPut,
		Link:   fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", link, username),
		Body:   jsonBody,
		Query:  nil,
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
		return nil, nil
	}

	//we updated permission
	if res.StatusCode == http.StatusNoContent {
		return &data.Permission{
			Link:        link,
			Username:    username,
			AccessLevel: permission,
			Type:        data.Repository,
		}, nil
	}

	return populateAddUserInRepositoryResponse(res)
}

func (g *github) AddOrUpdateUserInOrganizationFromApi(link, username, permission string) (*data.Permission, error) {
	jsonBody, err := json.Marshal(struct {
		Permission string `json:"role"`
	}{
		Permission: permission,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal body")
	}

	params := data.RequestParams{
		Method: http.MethodPut,
		Link:   fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username),
		Body:   jsonBody,
		Query:  nil,
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
		return nil, nil
	}

	return populateAddUserInOrganizationResponse(res)
}
