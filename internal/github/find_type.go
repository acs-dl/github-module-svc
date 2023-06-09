package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) FindType(link string) (*TypeSub, error) {
	repo, err := g.GetRepositoryFromApi(link)
	if err != nil {
		return nil, err
	}
	if repo != nil {
		return &TypeSub{data.Repository, *repo}, err
	}

	org, err := g.GetOrganizationFromApi(link)
	if err != nil {
		return nil, err
	}
	if org != nil {
		return &TypeSub{data.Organization, *org}, nil
	}

	return nil, nil
}

func (g *github) GetRepositoryFromApi(link string) (*data.Sub, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/repos/%s", link),
		Body:   nil,
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

	var response data.Sub
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}
	response.Type = data.Repository

	return &response, nil
}

func (g *github) GetOrganizationFromApi(link string) (*data.Sub, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/orgs/%s", link),
		Body:   nil,
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

	return populateGetOrganizationResponse(res)
}
