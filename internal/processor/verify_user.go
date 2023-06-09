package processor

import (
	"strconv"
	"time"

	"github.com/acs-dl/github-module-svc/internal/data"
	"github.com/acs-dl/github-module-svc/internal/github"
	"github.com/acs-dl/github-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateVerifyUser(msg data.ModulePayload) error {
	return validation.Errors{
		"user_id":  validation.Validate(msg.UserId, validation.Required),
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) HandleVerifyUserAction(msg data.ModulePayload) error {
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

	userApi, err := github.GetUser(p.pqueues.UserPQueue, any(p.githubClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting user from api")
	}

	if userApi == nil {
		p.log.WithError(err).Errorf("something wrong with user for message action with id `%s`", msg.RequestId)
		return errors.Errorf("something wrong with user from api")
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

		if err = p.permissionsQ.FilterByGithubIds(userApi.GithubId).Update(data.PermissionToUpdate{UserId: &userId}); err != nil {
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

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
