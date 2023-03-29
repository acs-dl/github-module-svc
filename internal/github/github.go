package github

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

type GithubClient interface {
	AddUserFromApi(link, username, permission string) (*data.Permission, error)
	UpdateUserFromApi(link, username, permission string) (*data.Permission, error)
	AddOrUpdateUserInRepoFromApi(link, username, permission string) (*data.Permission, error)
	AddOrUpdateUserInOrgFromApi(link, username, permission string) (*data.Permission, error)

	GetUsersFromApi(link, typeTo string) ([]data.Permission, error)
	GetUserIdFromApi(username string) (*data.User, *int64, error)

	RemoveUserFromApi(link, username, typeTo string) error

	GetOrgFromApi(link string) (*data.Sub, error)
	GetRepoFromApi(link string) (*data.Sub, error)

	CheckUserFromApi(link, username, typeTo string) (bool, *data.Permission, error)
	CheckRepoCollaborator(link, username string) (bool, *data.Permission, error)
	CheckOrgCollaborator(link, username string) (bool, *data.Permission, error)

	FindType(link string) (string, *data.Sub, error)
	FindRepoOwner(link string) (string, error)

	SearchByFromApi(username string) ([]data.User, error)
	GetProjectsFromApi(link string) ([]data.Sub, error)
}

type github struct {
	token string
	log   *logan.Entry
}

func NewGithub(token string, log *logan.Entry) GithubClient {
	return &github{
		token: token,
		log:   log,
	}
}
