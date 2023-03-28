package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateDeleteUser(msg data.ModulePayload) error {
	return validation.Errors{
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) handleDeleteUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateDeleteUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	_, githubId, err := p.githubClient.GetUserIdFromApi(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user id from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting user id from api")
	}

	permissions, err := p.permissionsQ.FilterByGithubIds(*githubId).Select()
	if err != nil {
		p.log.WithError(err).Errorf("failed to get permissions by github id `%d` for message action with id `%s`", *githubId, msg.RequestId)
		return errors.Wrap(err, "failed to get permissions")
	}

	for _, permission := range permissions {
		isHere, _, err := p.githubClient.CheckUserFromApi(permission.Link, msg.Username, permission.Type)
		if err != nil {
			p.log.WithError(err).Errorf("failed to check user from API for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "some error while checking user from api")
		}

		if isHere == true {
			err = p.githubClient.RemoveUserFromApi(permission.Link, permission.Username, permission.Type)
			if err != nil {
				p.log.WithError(err).Errorf("failed to remove user from API for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "some error while removing user from api")
			}
		}

		if err = p.permissionsQ.Delete(*githubId, permission.Type, permission.Link); err != nil {
			p.log.WithError(err).Errorf("failed to delete permission from db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to delete permission")
		}
	}

	var dbUser *data.User
	dbUser, err = p.usersQ.GetByGithubId(*githubId)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user by github id `%d` for message action with id `%s`", *githubId, msg.RequestId)
		return errors.Wrap(err, "failed to get user")
	}

	if dbUser == nil {
		p.log.WithError(err).Errorf("something wrong with db user for message action with id `%s`", *githubId, msg.RequestId)
		return errors.Wrap(err, "something wrong with db user")
	}

	err = p.usersQ.Delete(*githubId)
	if err != nil {
		p.log.WithError(err).Errorf("failed to delete user by github id `%d` for message action with id `%s`", *githubId, msg.RequestId)
		return errors.Wrap(err, "failed to delete user")
	}

	if dbUser.Id == nil {
		err = p.sendDeleteUser(msg.RequestId, *dbUser)
		if err != nil {
			p.log.WithError(err).Errorf("failed to publish delete user for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to publish delete user")
		}
	}

	p.resetFilters()
	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
