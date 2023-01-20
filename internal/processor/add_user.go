package processor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateAddUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":       validation.Validate(msg.Link, validation.Required),
		"username":   validation.Validate(msg.Username, validation.Required),
		"user_id":    validation.Validate(msg.UserId, validation.Required),
		"permission": validation.Validate(msg.Permission, validation.Required),
	}.Filter()
}

func (p *processor) handleAddUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateAddUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	permission, err := p.githubClient.AddUserFromApi(msg.Link, msg.Username, msg.Permission)
	if err != nil {
		return err
	}
	if permission == nil {
		p.log.Errorf("something wrong with adding user from api for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with adding user from api")
	}

	permission.UserId = &msg.UserId
	permission.RequestId = msg.RequestId

	err = p.managerQ.Transaction(func() error {
		if err = p.usersQ.Upsert(data.User{Id: &msg.UserId, Username: permission.Username, GithubId: permission.GithubId}); err != nil {
			p.log.WithError(err).Errorf("failed to upsert user in user db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to upsert user in user db")
		}

		if err = p.permissionsQ.Upsert(*permission); err != nil {
			p.log.WithError(err).Errorf("failed to upsert permission in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to upsert permission in permission db")
		}

		//in case if we have some rows without id from identity
		if err = p.permissionsQ.UpdateUserId(*permission); err != nil {
			p.log.WithError(err).Errorf("failed to update user id in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to update user id in user db")
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make add user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make add user transaction")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
