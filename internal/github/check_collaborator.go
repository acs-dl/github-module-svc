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

func (g *github) CheckUserFromApi(link, username, typeTo string) (*CheckPermission, error) {
	switch typeTo {
	case data.Repository:
		return g.CheckRepositoryCollaborator(link, username)
	case data.Organization:
		return g.CheckOrganizationCollaborator(link, username)
	default:
		return nil, errors.Errorf("failed to check `%s` with `%s` type", link, typeTo)
	}
}

func (g *github) CheckRepositoryCollaborator(link, username string) (*CheckPermission, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s/permission", link, username),
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
		return nil, errors.Wrap(err, "failed to make http request")
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check response status code")
	}
	if res == nil {
		return &CheckPermission{
			false,
			data.Permission{
				Link: link,
				Type: data.Repository,
			}}, nil
	}

	response := struct {
		RoleName string `json:"role_name"`
		User     struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &CheckPermission{
		true,
		data.Permission{
			Link:        link,
			Type:        data.Repository,
			Username:    username,
			GithubId:    response.User.Id,
			AccessLevel: response.RoleName,
		}}, nil
}

func (g *github) CheckOrganizationCollaborator(link, username string) (*CheckPermission, error) {
	params := data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username),
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
		return nil, errors.Wrap(err, "failed to make http request")
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check response status code")
	}
	if res == nil {
		return &CheckPermission{
			false,
			data.Permission{
				Link: link,
				Type: data.Organization,
			}}, nil
	}

	response := struct {
		Role string `json:"role"`
		User struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &CheckPermission{
		true,
		data.Permission{
			Link:        link,
			Type:        data.Organization,
			Username:    username,
			GithubId:    response.User.Id,
			AccessLevel: response.Role,
		}}, nil
}
