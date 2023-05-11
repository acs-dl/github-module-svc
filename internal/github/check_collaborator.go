package github

import (
	"fmt"
	"net/http"
	"time"

	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) CheckUserFromApi(link, username, typeTo string) (*data.Permission, error) {
	switch typeTo {
	case data.Repository:
		return g.CheckRepositoryCollaborator(link, username)
	case data.Organization:
		return g.CheckOrganizationCollaborator(link, username)
	default:
		return nil, errors.Errorf("failed to check `%s` with `%s` type", link, typeTo)
	}
}

func (g *github) CheckRepositoryCollaborator(link, username string) (*data.Permission, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s/permission", link, username),
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

	return populateCheckRepositoryCollaboratorResponse(res, link, username)
}

func (g *github) CheckOrganizationCollaborator(link, username string) (*data.Permission, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username),
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

	return populateCheckOrganizationCollaboratorResponse(res, link, username)
}
