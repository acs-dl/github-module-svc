package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
)

func (g *github) AddUserFromApi(link, username, permission string) (*data.Permission, error) {
	findType, err := g.FindType(link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get type")
	}

	if err = validation.Validate(findType, validation.In(data.Organization, data.Repository)); err != nil {
		return nil, errors.Wrap(err, "something wrong with link type")
	}

	switch findType {
	case data.Repository:
		isCollaborator, err := g.CheckRepoCollaborator(link, username)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if user in repo from api")
		}

		if isCollaborator {
			return nil, errors.Errorf("such user is already in repository")
		}

		return g.AddOrUpdateUserInRepoFromApi(link, username, permission)
	case data.Organization:
		isCollaborator, err := g.CheckOrgCollaborator(link, username)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if user in org from api")
		}

		if isCollaborator {
			return nil, errors.Errorf("such user is already in organisation")
		}

		return g.AddOrUpdateUserInOrgFromApi(link, username, permission)
	default:
		return nil, errors.Wrap(err, "unexpected type")
	}
}

func (g *github) UpdateUserFromApi(link, username, permission string) (*data.Permission, error) {
	findType, err := g.FindType(link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get type")
	}

	if err = validation.Validate(findType, validation.In(data.Organization, data.Repository)); err != nil {
		return nil, errors.Wrap(err, "something wrong with link type")
	}

	switch findType {
	case data.Repository:
		isCollaborator, err := g.CheckRepoCollaborator(link, username)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if user in repo from api")
		}

		if !isCollaborator {
			return nil, errors.Errorf("such user is not in repository")
		}

		owned, err := g.FindRepoOwner(link)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if repository owner")
		}

		if owned == data.UserOwned {
			permission = "write"
		}

		return g.AddOrUpdateUserInRepoFromApi(link, username, permission)
	case data.Organization:
		isCollaborator, err := g.CheckOrgCollaborator(link, username)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if user in org from api")
		}

		if !isCollaborator {
			return nil, errors.Errorf("`%s` is not in organisation `%s`", username, link)
		}

		return g.AddOrUpdateUserInOrgFromApi(link, username, permission)
	default:
		return nil, errors.Errorf("unexpected type `%s`", findType)
	}
}

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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res == nil {
		return nil, errors.New("failed to process request: response is nil")
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("failed to process request: bad status (%s)", res.Status))
	}

	//we updated permission
	if res.StatusCode == 204 {
		return &data.Permission{
			Link:       link,
			Username:   username,
			Permission: permission,
			Type:       data.Repository,
		}, nil
	}

	returned := struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Invitee struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"invitee"`
		Permissions string `json:"permissions"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&returned); err != nil {
		return nil, errors.Wrap(err, " failed to unmarshal body")
	}

	return &data.Permission{
		Link:       returned.Repository.FullName,
		Username:   returned.Invitee.Login,
		GithubId:   returned.Invitee.Id,
		Permission: returned.Permissions,
		Type:       data.Repository,
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, " error making http request")
	}

	if res == nil {
		return nil, errors.New("failed to process request: response is nil")
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
		Link:       returned.Repository.FullName,
		Username:   returned.Invitee.Login,
		GithubId:   returned.Invitee.Id,
		Permission: returned.Role,
		Type:       data.Organization,
	}, nil

}
