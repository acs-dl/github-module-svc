package processor

import (
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateVerifyUser(msg data.ModulePayload) error {
	return validation.Errors{
		"user_id":  validation.Validate(msg.UserId, validation.Required),
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) handleVerifyUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateVerifyUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	userId, err := strconv.ParseInt(msg.UserId, 10, 64)
	if err != nil {
		p.log.WithError(err).Errorf("failed to parse user id `%s` for message action with id `%s`", msg.UserId, msg.RequestId)
		return errors.Wrap(err, "failed to parse user id")
	}

	userApi, githubId, err := p.githubClient.GetUserIdFromApi(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user id from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting user id from api")
	}

	user := data.User{
		Id:        &userId,
		Username:  userApi.Username,
		GithubId:  userApi.GithubId,
		AvatarUrl: userApi.AvatarUrl,
		CreatedAt: time.Now(),
	}

	err = p.managerQ.Transaction(func() error {
		if err = p.usersQ.Upsert(user); err != nil {
			p.log.WithError(err).Errorf("failed to upsert user in user db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to upsert user in user db")
		}

		if err = p.permissionsQ.UpdateUserId(data.Permission{UserId: &userId, GithubId: *githubId}); err != nil {
			p.log.WithError(err).Errorf("failed to update user id in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to update user id in user db")
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make add user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make add user transaction")
	}

	err = p.SendDeleteUser(msg.RequestId, user)
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish delete user for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to publish delete user")
	}

	p.resetFilters()
	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
