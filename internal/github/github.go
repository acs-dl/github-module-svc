package github

import "gitlab.com/distributed_lab/acs/github-module/internal/data"

type GithubClient interface {
	AddUserFromApi(link, username, permission string) (*data.Permission, error)
	UpdateUserFromApi(link, username, permission string) (*data.Permission, error)
	AddOrUpdateUserInRepoFromApi(link, username, permission string) (*data.Permission, error)
	AddOrUpdateUserInOrgFromApi(link, username, permission string) (*data.Permission, error)

	GetUsersFromApi(link, typeTo string) ([]data.Permission, error)
	GetUserIdFromApi(username string) (*int64, error)

	RemoveUserFromApi(link, username, typeTo string) error

	CheckOrgFromApi(link string) (bool, error)
	CheckRepoFromApi(link string) (bool, error)
	CheckRepoCollaborator(link, username string) (bool, error)
	CheckOrgCollaborator(link, username string) (bool, error)

	//GetRolesFromApi(link string) (bool, error)
	FindType(link string) (string, error)
	FindRepoOwner(link string) (string, error)
}

type github struct {
	token string
}

func NewGithub(token string) GithubClient {
	return &github{
		token: token,
	}
}
