package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) GetUsersFromApi(link, typeTo string) ([]data.Permission, error) {
	resultLink := fmt.Sprintf("https://api.github.com/repos/%s/collaborators", link)
	if typeTo == data.Organization {
		resultLink = fmt.Sprintf("https://api.github.com/orgs/%s/members", link)
	}
	req, err := http.NewRequest(http.MethodGet, resultLink, nil)
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	var result []data.Permission

	for page := 1; ; page++ {
		q := req.URL.Query()
		q.Add("per_page", "100")
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, " error making http request")
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, errors.New(fmt.Sprintf("failed to process request: bad status (%s)", res.Status))
		}

		if res == nil {
			return nil, errors.New("failed to process request: response is nil")
		}

		var returned []data.Permission
		if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
			return nil, errors.Wrap(err, " failed to unmarshal body")
		}

		if len(returned) == 0 {
			break
		}

		result = append(result, returned...)
	}

	return result, nil
}
