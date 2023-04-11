package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) FindType(link string) (*TypeSub, error) {
	repo, err := g.GetRepoFromApi(link)
	if err != nil {
		return nil, err
	}
	if repo != nil {
		return &TypeSub{data.Repository, *repo}, err
	}

	org, err := g.GetOrgFromApi(link)
	if err != nil {
		return nil, err
	}
	if org != nil {
		return &TypeSub{data.Organization, *org}, nil
	}

	return nil, errors.Errorf("failed to check type for `%s`", link)
}

func (g *github) GetRepoFromApi(link string) (*data.Sub, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s", link), nil)
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.superToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.GetRepoFromApi(link)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get group status `%s`", res.Status)
	}

	var returned data.Sub
	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}
	returned.Type = data.Repository

	return &returned, nil
}

func (g *github) GetOrgFromApi(link string) (*data.Sub, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/orgs/%s", link), nil)
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.superToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.GetOrgFromApi(link)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get group status `%s`", res.Status)
	}

	returned := struct {
		Id    int64  `json:"id"`
		Login string `json:"login"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return &data.Sub{
		Id:   returned.Id,
		Link: returned.Login,
		Path: returned.Login,
		Type: data.Organization,
	}, nil
}
