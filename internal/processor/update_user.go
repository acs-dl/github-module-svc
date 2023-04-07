package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/helpers"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateUpdateUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":         validation.Validate(msg.Link, validation.Required),
		"username":     validation.Validate(msg.Username, validation.Required),
		"access_level": validation.Validate(msg.AccessLevel, validation.Required),
	}.Filter()
}

func (p *processor) handleUpdateUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateUpdateUser(msg)
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

	item, err := helpers.AddFunctionInPqueue(p.pqueue, any(p.githubClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
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

	item, err = helpers.AddFunctionInPqueue(p.pqueue, any(p.githubClient.UpdateUserFromApi), []any{any(msg.Link), any(msg.Username), any(msg.AccessLevel)}, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add function in pqueue for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		p.log.WithError(err).Errorf("failed to update user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while updating user from api")
	}
	permission, ok := item.Response.Value.(*data.Permission)
	if !ok {
		return errors.Errorf("wrong response type")
	}

	if permission == nil {
		p.log.Errorf("something wrong with updating user from api for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with updating user from api")
	}

	permission.GithubId = userApi.GithubId
	permission.RequestId = msg.RequestId
	permission.UserId = dbUser.Id

	err = p.managerQ.Transaction(func() error {
		permission.Link = msg.Link
		if err = p.permissionsQ.FilterByGithubIds(permission.GithubId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{
			Username:    &permission.Username,
			AccessLevel: &permission.AccessLevel,
		}); err != nil {
			p.log.WithError(err).Errorf("failed to update user in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to update user in permission db")
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make update user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make update user transaction")
	}

	p.resetFilters()
	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
