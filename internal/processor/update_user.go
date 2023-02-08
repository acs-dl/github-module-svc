package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
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

	dbUser, err := p.usersQ.GetByUsername(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from user db for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get user from user db")
	}

	if dbUser == nil {
		p.log.Errorf("no user with such username for message action with id `%s`", msg.RequestId)
		return errors.New("no user with such username")
	}

	_, githubId, err := p.githubClient.GetUserIdFromApi(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get github id from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting github id from api")
	}

	permission, err := p.githubClient.UpdateUserFromApi(msg.Link, msg.Username, msg.AccessLevel)
	if err != nil {
		p.log.WithError(err).Errorf("failed to update user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while updating user from api")
	}
	if permission == nil {
		p.log.Errorf("something wrong with updating user from api for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with updating user from api")
	}

	permission.GithubId = *githubId
	permission.RequestId = msg.RequestId
	permission.UserId = dbUser.Id

	err = p.managerQ.Transaction(func() error {
		permission.Link = msg.Link
		if err = p.permissionsQ.Update(*permission); err != nil {
			p.log.WithError(err).Errorf("failed to update user in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to update user in permission db")
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make update user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make update user transaction")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
