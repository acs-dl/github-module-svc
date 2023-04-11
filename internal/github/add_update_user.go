package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) AddOrUpdateUserInRepoFromApi(link, username, permission string) (*data.Permission, error) {
	jsonBody, _ := json.Marshal(struct {
		Permission string `json:"permission"`
	}{
		Permission: permission,
	})

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", link, username), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.superToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res == nil {
		return nil, errors.New("failed to process request: response is nil")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.AddOrUpdateUserInRepoFromApi(link, username, permission)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("failed to process request: bad status (%s)", res.Status))
	}

	//we updated permission
	if res.StatusCode == 204 {
		return &data.Permission{
			Link:        link,
			Username:    username,
			AccessLevel: permission,
			Type:        data.Repository,
		}, nil
	}

	returned := struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Invitee struct {
			Login     string `json:"login"`
			Id        int64  `json:"id"`
			AvatarUrl string `json:"avatar_url"`
		} `json:"invitee"`
		Permissions string `json:"permissions"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return &data.Permission{
		Link:        returned.Repository.FullName,
		Username:    returned.Invitee.Login,
		GithubId:    returned.Invitee.Id,
		AccessLevel: returned.Permissions,
		Type:        data.Repository,
		AvatarUrl:   returned.Invitee.AvatarUrl,
	}, nil

}

func (g *github) AddOrUpdateUserInOrgFromApi(link, username, permission string) (*data.Permission, error) {
	jsonBody, _ := json.Marshal(struct {
		Permission string `json:"role"`
	}{
		Permission: permission,
	})

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://api.github.com/orgs/%s/memberships/%s", link, username), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, errors.Wrap(err, " couldn't create request")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.superToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res == nil {
		return nil, errors.New("failed to process request: response is nil")
	}

	if res.StatusCode == http.StatusForbidden {
		timeoutDuration, err := g.getDuration(res)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get time duration from response")
		}
		g.log.Warnf("we need to wait `%d`", timeoutDuration)
		time.Sleep(timeoutDuration)
		return g.AddOrUpdateUserInOrgFromApi(link, username, permission)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.Errorf("failed to process request: bad status `%s`", res.Status)
	}

	returned := struct {
		Repository struct {
			FullName string `json:"login"`
		} `json:"organization"`
		Invitee struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
		Role string `json:"role"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return &data.Permission{
		Link:        returned.Repository.FullName,
		Username:    returned.Invitee.Login,
		GithubId:    returned.Invitee.Id,
		AccessLevel: returned.Role,
		Type:        data.Organization,
	}, nil

}
