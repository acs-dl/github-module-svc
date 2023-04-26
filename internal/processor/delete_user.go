package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
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

	userApi, err := github.GetUser(p.pqueues.UsualPQueue, any(p.githubClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting user from api")
	}

	if userApi == nil {
		p.log.WithError(err).Errorf("something wrong with user for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with user from api")
	}

	permissions, err := p.permissionsQ.FilterByGithubIds(userApi.GithubId).Select()
	if err != nil {
		p.log.WithError(err).Errorf("failed to get permissions by github id `%d` for message action with id `%s`", userApi.GithubId, msg.RequestId)
		return errors.Wrap(err, "failed to get permissions")
	}

	for _, permission := range permissions {
		checkSub, err := github.GetPermissionWithCheck(
			p.pqueues.SuperPQueue,
			any(p.githubClient.CheckUserFromApi),
			[]any{any(permission.Link), any(msg.Username), any(permission.Type)},
			pqueue.NormalPriority,
		)
		if err != nil {
			p.log.WithError(err).Errorf("failed to check user from API for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "some error while checking link type api")
		}

		if checkSub != nil {
			err = github.GetRequestError(
				p.pqueues.SuperPQueue,
				any(p.githubClient.RemoveUserFromApi),
				[]any{any(permission.Link), any(permission.Username), any(permission.Type)},
				pqueue.NormalPriority)
			if err != nil {
				p.log.WithError(err).Errorf("failed to remove user from API for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "some error while removing user from api")
			}
		}

		err = p.permissionsQ.FilterByGithubIds(userApi.GithubId).FilterByLinks(permission.Link).FilterByTypes(permission.Type).Delete()
		if err != nil {
			p.log.WithError(err).Errorf("failed to delete permission from db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to delete permission")
		}
	}

	var dbUser *data.User
	dbUser, err = p.usersQ.FilterByGithubIds(userApi.GithubId).Get()
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user by github id `%d` for message action with id `%s`", userApi.GithubId, msg.RequestId)
		return errors.Wrap(err, "failed to get user")
	}

	if dbUser == nil {
		p.log.WithError(err).Errorf("something wrong with db user for message action with id `%s`", userApi.GithubId, msg.RequestId)
		return errors.Wrap(err, "something wrong with db user")
	}

	err = p.usersQ.FilterByGithubIds(userApi.GithubId).Delete()
	if err != nil {
		p.log.WithError(err).Errorf("failed to delete user by github id `%d` for message action with id `%s`", userApi.GithubId, msg.RequestId)
		return errors.Wrap(err, "failed to delete user")
	}

	if dbUser.Id == nil {
		err = p.SendDeleteUser(msg.RequestId, *dbUser)
		if err != nil {
			p.log.WithError(err).Errorf("failed to publish delete user for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to publish delete user")
		}
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
