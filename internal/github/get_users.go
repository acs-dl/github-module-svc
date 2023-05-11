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

func (g *github) GetUsersFromApi(link, typeTo string) ([]data.Permission, error) {
	resultLink := fmt.Sprintf("https://api.github.com/repos/%s/collaborators", link)
	if typeTo == data.Organization {
		resultLink = fmt.Sprintf("https://api.github.com/orgs/%s/members", link)
	}

	response, err := helpers.MakeRequestWithPagination(data.RequestParams{
		Method: http.MethodGet,
		Link:   resultLink,
		Body:   nil,
		Query: map[string]string{
			"per_page": "100",
		},
		Header: map[string]string{
			"Accept":               data.AcceptHeader,
			"Authorization":        "Bearer " + g.superUserToken,
			"X-GitHub-Api-Version": data.GithubApiVersionHeader,
		},
		Timeout: time.Second * 30,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request with pagination")
	}

	var result []data.Permission
	if err = json.Unmarshal(response, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return result, nil
}
