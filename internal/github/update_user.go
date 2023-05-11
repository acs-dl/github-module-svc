package github

import (
	"github.com/acs-dl/github-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) UpdateUserFromApi(typeTo, link, username, permission string) (*data.Permission, error) {
	switch typeTo {
	case data.Repository:
		owned, err := g.FindRepositoryOwner(link)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check if repository owner")
		}

		if owned == data.UserOwned {
			permission = "write"
		}

		return g.AddOrUpdateUserInRepositoryFromApi(link, username, permission)
	case data.Organization:
		return g.AddOrUpdateUserInOrganizationFromApi(link, username, permission)
	default:
		return nil, errors.Errorf("unexpected type `%s`", typeTo)
	}
}
