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

func (g *github) GetProjectsFromApi(link string) ([]data.Sub, error) {
	response, err := helpers.MakeRequestWithPagination(data.RequestParams{
		Method: http.MethodGet,
		Link:   fmt.Sprintf("https://api.github.com/orgs/%s/repos", link),
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

	var result []data.Sub
	if err = json.Unmarshal(response, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return result, nil
}
