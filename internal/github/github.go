package github

import (
	"context"

	"gitlab.com/distributed_lab/acs/github-module/internal/config"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/service/background"
	"gitlab.com/distributed_lab/logan/v3"
)

type GithubClient interface {
	AddUserFromApi(typeTo, link, username, permission string) (*data.Permission, error)
	UpdateUserFromApi(typeTo, link, username, permission string) (*data.Permission, error)
	AddOrUpdateUserInRepositoryFromApi(link, username, permission string) (*data.Permission, error)
	AddOrUpdateUserInOrganizationFromApi(link, username, permission string) (*data.Permission, error)

	GetUsersFromApi(link, typeTo string) ([]data.Permission, error)
	GetUserFromApi(username string) (*data.User, error)

	RemoveUserFromApi(link, username, typeTo string) error

	GetOrganizationFromApi(link string) (*data.Sub, error)
	GetRepositoryFromApi(link string) (*data.Sub, error)

	CheckUserFromApi(link, username, typeTo string) (*CheckPermission, error)
	CheckRepositoryCollaborator(link, username string) (*CheckPermission, error)
	CheckOrganizationCollaborator(link, username string) (*CheckPermission, error)

	FindType(link string) (*TypeSub, error)
	FindRepositoryOwner(link string) (string, error)

	SearchByFromApi(username string) ([]data.User, error)
	GetProjectsFromApi(link string) ([]data.Sub, error)
}

type CheckPermission struct {
	Ok         bool
	Permission data.Permission
}

type TypeSub struct {
	Type string
	Sub  data.Sub
}

type github struct {
	superToken string
	usualToken string
	log        *logan.Entry
}

func NewGithubAsInterface(cfg config.Config, _ context.Context) interface{} {
	return interface{}(&github{
		superToken: cfg.Github().SuperToken,
		usualToken: cfg.Github().UsualToken,
		log:        cfg.Log(),
	})
}

func GithubClientInstance(ctx context.Context) GithubClient {
	return ctx.Value(background.GithubClientCtxKey).(GithubClient)
}

func CtxGithubClientInstance(entry interface{}, ctx context.Context) context.Context {
	return context.WithValue(ctx, background.GithubClientCtxKey, entry)
}
