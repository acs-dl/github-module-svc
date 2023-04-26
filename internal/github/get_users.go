package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

	var result []data.Permission

	for page := 1; ; page++ {
		params := data.RequestParams{
			Method: http.MethodGet,
			Link:   resultLink,
			Body:   nil,
			Query: map[string]string{
				"per_page": "100",
				"page":     strconv.Itoa(page),
			},
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
			return nil, errors.Errorf("error in response, status %v", res.StatusCode)
		}

		var response []data.Permission
		if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal body")
		}

		if len(response) == 0 {
			break
		}

		result = append(result, response...)
	}

	return result, nil
}
