package github

import (
	"fmt"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

func (g *github) CheckRepoCollaborator(link, username string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", link, username), nil)
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

	if res.StatusCode == 204 {
		return true, nil
	}

	if res.StatusCode == 404 {
		return false, nil
	}

	return false, errors.Errorf("unexpected status %s", res.Status)
}

func (g *github) CheckOrgCollaborator(link, username string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username), nil)
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

	if res.StatusCode == 404 {
		return false, nil
	}

	return false, errors.Errorf("unexpected status %s", res.Status)
}
