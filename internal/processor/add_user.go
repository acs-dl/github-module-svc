package processor

import (
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateAddUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":         validation.Validate(msg.Link, validation.Required),
		"username":     validation.Validate(msg.Username, validation.Required),
		"user_id":      validation.Validate(msg.UserId, validation.Required),
		"access_level": validation.Validate(msg.AccessLevel, validation.Required),
	}.Filter()
}

func (p *processor) handleAddUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateAddUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	userId, err := strconv.ParseInt(msg.UserId, 10, 64)
	if err != nil {
		p.log.WithError(err).Errorf("failed to parse user id `%s` for message action with id `%s`", msg.UserId, msg.RequestId)
		return errors.Wrap(err, "failed to parse user id")
	}

	permission, err := p.githubClient.AddUserFromApi(msg.Link, msg.Username, msg.AccessLevel)
	if err != nil {
		return err
	}
	if permission == nil {
		p.log.Errorf("something wrong with adding user from api for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with adding user from api")
	}

	permission.UserId = &userId
	permission.RequestId = msg.RequestId
	permission.CreatedAt = time.Now()

	dbUser := data.User{
		Id:        &userId,
		Username:  permission.Username,
		GithubId:  permission.GithubId,
		CreatedAt: permission.CreatedAt,
		AvatarUrl: permission.AvatarUrl,
	}

	err = p.managerQ.Transaction(func() error {
		if err = p.usersQ.Upsert(dbUser); err != nil {
			p.log.WithError(err).Errorf("failed to creat user in user db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to create user in user db")
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

	err = p.SendDeleteUser(msg.RequestId, dbUser)
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to publish users")
	}

	p.resetFilters()
	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) resetFilters() {
	p.usersQ = p.usersQ.New()
	p.permissionsQ = p.permissionsQ.New()
}
