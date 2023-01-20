package github

import (
	"fmt"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

func (g *github) FindType(link string) (string, error) {
	repo, err := g.CheckRepoFromApi(link)
	if err != nil {
		return "", err
	}
	if repo {
		return data.Repository, err
	}

	org, err := g.CheckOrgFromApi(link)
	if err != nil {
		return "", err
	}
	if org {
		return data.Organization, nil
	}

	return "", nil
}

func (g *github) CheckRepoFromApi(link string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s", link), nil)
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

	return false, nil
}

func (g *github) CheckOrgFromApi(link string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/orgs/%s", link), nil)
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

	return false, nil
}
