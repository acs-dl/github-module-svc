package github

import (
	"fmt"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

func (g *github) RemoveUserFromApi(link, username, typeTo string) error {
	resultLink := fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", link, username)
	if typeTo == data.Organization {
		resultLink = fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username)
	}

	req, err := http.NewRequest(http.MethodDelete, resultLink, nil)
	if err != nil {
		return errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, " error making http request")
	}

	if res.StatusCode != 204 {
		return errors.New(fmt.Sprintf("error in response, status %s", res.Status))
	}

	return nil
}
