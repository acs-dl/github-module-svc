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

	response := struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Invitee struct {
			Login     string `json:"login"`
			Id        int64  `json:"id"`
			AvatarUrl string `json:"avatar_url"`
		} `json:"invitee"`
		Permissions string `json:"permissions"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Permission{
		Link:        response.Repository.FullName,
		Username:    response.Invitee.Login,
		GithubId:    response.Invitee.Id,
		AccessLevel: response.Permissions,
		Type:        data.Repository,
		AvatarUrl:   response.Invitee.AvatarUrl,
	}, nil

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
		return nil, nil
	}

	response := struct {
		Repository struct {
			FullName string `json:"login"`
		} `json:"organization"`
		Invitee struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
		Role string `json:"role"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Permission{
		Link:        response.Repository.FullName,
		Username:    response.Invitee.Login,
		GithubId:    response.Invitee.Id,
		AccessLevel: response.Role,
		Type:        data.Organization,
	}, nil

}
