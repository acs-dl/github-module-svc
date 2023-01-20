package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
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

	dbUser, err := p.usersQ.GetByUsername(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from user db for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get user from user db")
	}

	if dbUser == nil {
		p.log.Errorf("no user with such username for message action with id `%s`", msg.RequestId)
		return errors.New("no user with such username")
	}

	githubId, err := p.githubClient.GetUserIdFromApi(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get github id from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting github id from api")
	}

	msg.Type, err = p.githubClient.FindType(msg.Link)
	if err != nil {
		return errors.Wrap(err, "failed to get type")
	}

	if validation.Validate(msg.Type, validation.In(data.Organization, data.Repository)) != nil {
		return errors.Wrap(err, "something wrong with link type")
	}

	err = p.githubClient.RemoveUserFromApi(msg.Link, msg.Username, msg.Type)
	if err != nil {
		p.log.WithError(err).Errorf("failed to remove user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while removing user from api")
	}

	err = p.managerQ.Transaction(func() error {
		if err = p.permissionsQ.Delete(*githubId, msg.Type, msg.Link); err != nil {
			p.log.WithError(err).Errorf("failed to delete user from db for message action with id `%s`", msg.RequestId)
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make remove user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make remove user transaction")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
