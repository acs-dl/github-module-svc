package github

import (
	"fmt"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

// GetRolesFromApi it doesn't implemented yet, cause custom roles are for premium users
func (g *github) GetRolesFromApi(id string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/organizations/%s/custom_roles", id), nil)
	if err != nil {
		return false, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, " error making http request")
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("status: %s", res.Status))
}
