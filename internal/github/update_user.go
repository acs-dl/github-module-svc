package github

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *github) UpdateUserFromApi(link, username, permission string) (*data.Permission, error) {
	findType, _, err := g.FindType(link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get type")
	}

	if err = validation.Validate(findType, validation.In(data.Organization, data.Repository)); err != nil {
		return nil, errors.Wrap(err, "something wrong with link type")
	}

	switch findType {
	case data.Repository:
		isCollaborator, _, err := g.CheckRepoCollaborator(link, username)
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
		isCollaborator, _, err := g.CheckOrgCollaborator(link, username)
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
