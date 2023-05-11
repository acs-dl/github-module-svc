package github

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) RemoveUserFromApi(link, username, typeTo string) error {
	resultLink := fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", link, username)
	if typeTo == data.Organization {
		resultLink = fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username)
	}

	params := data.RequestParams{
		Method: http.MethodDelete,
		Link:   resultLink,
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
		return errors.Wrap(err, "failed to make http request")
	}

	res, err = helpers.HandleHttpResponseStatusCode(res, params)
	if err != nil {
		return errors.Wrap(err, "failed to check response status code")
	}
	if res == nil {
		return errors.Errorf("error in response, status %v", res.StatusCode)
	}

	return nil
}
