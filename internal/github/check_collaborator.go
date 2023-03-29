package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) CheckUserFromApi(link, username, typeTo string) (bool, *data.Permission, error) {
	if typeTo == data.Repository {
		return g.CheckRepoCollaborator(link, username)
	}
	if typeTo == data.Organization {
		return g.CheckOrgCollaborator(link, username)
	}

	return false, nil, errors.Errorf("failed to check `%s` with `%s` type", link, typeTo)
}

func (g *github) CheckRepoCollaborator(link, username string) (bool, *data.Permission, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s/permission", link, username), nil)
	if err != nil {
		return false, nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil, errors.Wrap(err, " error making http request")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return false, nil, errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.CheckRepoCollaborator(link, username)
	}

	if res.StatusCode == http.StatusNotFound {
		return false, nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return false, nil, errors.Errorf("unexpected status %s", res.Status)
	}

	returned := struct {
		RoleName string `json:"role_name"`
		User     struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return false, nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return true, &data.Permission{
		Link:        link,
		Type:        data.Repository,
		Username:    username,
		GithubId:    returned.User.Id,
		AccessLevel: returned.RoleName,
	}, nil
}

func (g *github) CheckOrgCollaborator(link, username string) (bool, *data.Permission, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username), nil)
	if err != nil {
		return false, nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.Github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil, errors.Wrap(err, " error making http request")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return false, nil, errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.CheckOrgCollaborator(link, username)
	}

	if res.StatusCode == http.StatusNotFound {
		return false, nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return false, nil, errors.Errorf("unexpected status %s", res.Status)
	}

	returned := struct {
		Role string `json:"role"`
		User struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return false, nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return true, &data.Permission{
		Link:        link,
		Type:        data.Organization,
		Username:    username,
		GithubId:    returned.User.Id,
		AccessLevel: returned.Role,
	}, nil
}
