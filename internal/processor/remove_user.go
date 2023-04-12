package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateRemoveUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":     validation.Validate(msg.Link, validation.Required),
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) handleRemoveUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateRemoveUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	dbUser, err := p.usersQ.FilterByUsernames(msg.Username).Get()
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from user db for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get user from user db")
	}

	if dbUser == nil {
		p.log.Errorf("no user with such username for message action with id `%s`", msg.RequestId)
		return errors.New("no user with such username")
	}

	item, err := helpers.AddFunctionInPQueue(p.pqueues.UsualPQueue, any(p.githubClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting user from api")
	}
	userApi, err := p.convertUserFromInterfaceAndCheck(item.Response.Value)
	if err != nil {
		p.log.WithError(err).Errorf("something wrong with user for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with user from api")
	}

	item, err = helpers.AddFunctionInPQueue(p.pqueues.SuperPQueue, any(p.githubClient.FindType), []any{any(msg.Link)}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		p.log.WithError(err).Errorf("failed to get type for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get type")
	}
	typeSub, ok := item.Response.Value.(*github.TypeSub)
	if !ok {
		return errors.Errorf("wrong response type")
	}

	if typeSub == nil {
		p.log.WithError(err).Errorf("type not found for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "type not found")
	}
	msg.Type = typeSub.Type

	if validation.Validate(msg.Type, validation.In(data.Organization, data.Repository)) != nil {
		return errors.Wrap(err, "something wrong with link type")
	}

	item, err = helpers.AddFunctionInPQueue(p.pqueues.SuperPQueue, any(p.githubClient.RemoveUserFromApi), []any{any(msg.Link), any(msg.Username), any(msg.Type)}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		p.log.WithError(err).Errorf("failed to remove user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while removing user from api")
	}

	err = p.managerQ.Transaction(func() error {
		err = p.permissionsQ.FilterByGithubIds(userApi.GithubId).FilterByLinks(msg.Link).FilterByTypes(msg.Type).Delete()
		if err != nil {
			p.log.WithError(err).Errorf("failed to delete user from db for message action with id `%s`", msg.RequestId)
		}

		permissions, err := p.permissionsQ.FilterByGithubIds(userApi.GithubId).Select()
		if err != nil {
			p.log.WithError(err).Errorf("failed to select permissions by github id `%d` for message action with id `%s`", userApi.GithubId, msg.RequestId)
			return errors.Wrap(err, "failed to select permissions")
		}

		if len(permissions) == 0 {
			err = p.usersQ.FilterByGithubIds(userApi.GithubId).Delete()
			if err != nil {
				p.log.WithError(err).Errorf("failed to delete user by telegram id `%d` for message action with id `%s`", userApi.GithubId, msg.RequestId)
				return errors.Wrap(err, "failed to delete user")
			}

			if dbUser.Id == nil {
				err = p.SendDeleteUser(msg.RequestId, *dbUser)
				if err != nil {
					p.log.WithError(err).Errorf("failed to publish delete user for message action with id `%s`", msg.RequestId)
					return errors.Wrap(err, "failed to publish delete user")
				}
			}
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make remove user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make remove user transaction")
	}

	p.resetFilters()
	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
