package github

import (
	"encoding/json"

	"github.com/acs-dl/github-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func populateAddUserInRepositoryResponse(res *data.ResponseParams) (*data.Permission, error) {
	response := struct {
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

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Permission{
		Link:        response.Repository.FullName,
		Username:    response.Invitee.Login,
		GithubId:    response.Invitee.Id,
		AccessLevel: response.Permissions,
		Type:        data.Repository,
		AvatarUrl:   response.Invitee.AvatarUrl,
	}, nil
}

func populateAddUserInOrganizationResponse(res *data.ResponseParams) (*data.Permission, error) {
	response := struct {
		Repository struct {
			FullName string `json:"login"`
		} `json:"organization"`
		Invitee struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
		Role string `json:"role"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Permission{
		Link:        response.Repository.FullName,
		Username:    response.Invitee.Login,
		GithubId:    response.Invitee.Id,
		AccessLevel: response.Role,
		Type:        data.Organization,
	}, nil
}

func populateCheckRepositoryCollaboratorResponse(res *data.ResponseParams, link, username string) (*data.Permission, error) {
	response := struct {
		RoleName string `json:"role_name"`
		User     struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Permission{
		Link:        link,
		Type:        data.Repository,
		Username:    username,
		GithubId:    response.User.Id,
		AccessLevel: response.RoleName,
	}, nil
}

func populateCheckOrganizationCollaboratorResponse(res *data.ResponseParams, link, username string) (*data.Permission, error) {
	response := struct {
		Role string `json:"role"`
		User struct {
			Login string `json:"login"`
			Id    int64  `json:"id"`
		} `json:"user"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Permission{
		Link:        link,
		Type:        data.Organization,
		Username:    username,
		GithubId:    response.User.Id,
		AccessLevel: response.Role,
	}, nil
}

func populateGetOrganizationResponse(res *data.ResponseParams) (*data.Sub, error) {
	response := struct {
		Id    int64  `json:"id"`
		Login string `json:"login"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal body")
	}

	return &data.Sub{
		Id:   response.Id,
		Link: response.Login,
		Path: response.Login,
		Type: data.Organization,
	}, nil
}
