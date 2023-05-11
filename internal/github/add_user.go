package github

import (
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) AddUserFromApi(typeTo, link, username, permission string) (*data.Permission, error) {
	switch typeTo {
	case data.Repository:
		return g.AddOrUpdateUserInRepositoryFromApi(link, username, permission)
	case data.Organization:
		return g.AddOrUpdateUserInOrganizationFromApi(link, username, permission)
	default:
		return nil, errors.New("unexpected type")
	}
}
