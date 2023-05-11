package processor

import (
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
	"gitlab.com/distributed_lab/acs/github-module/internal/pqueue"
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

func (p *processor) HandleAddUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateAddUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}
	msg.Link = strings.ToLower(msg.Link)
	userId, err := strconv.ParseInt(msg.UserId, 10, 64)
	if err != nil {
		p.log.WithError(err).Errorf("failed to parse user id `%s` for message action with id `%s`", msg.UserId, msg.RequestId)
		return errors.Wrap(err, "failed to parse user id")
	}

	permission, err := p.addUser(msg.Link, msg.Username, msg.AccessLevel)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while adding user from api")
	}

	permission.UserId = &userId
	permission.RequestId = msg.RequestId
	permission.CreatedAt = time.Now()

	user := data.User{
		Id:        &userId,
		Username:  permission.Username,
		GithubId:  permission.GithubId,
		CreatedAt: permission.CreatedAt,
		AvatarUrl: permission.AvatarUrl,
	}

	err = p.managerQ.Transaction(func() error {
		if err = p.usersQ.Upsert(user); err != nil {
			p.log.WithError(err).Errorf("failed to creat user in user db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to create user in user db")
		}

		if err = p.permissionsQ.Upsert(*permission); err != nil {
			p.log.WithError(err).Errorf("failed to upsert permission in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to upsert permission in permission db")
		}

		//in case if we have some rows without id from identity
		if err = p.permissionsQ.FilterByGithubIds(permission.GithubId).Update(data.PermissionToUpdate{UserId: permission.UserId}); err != nil {
			p.log.WithError(err).Errorf("failed to update user id in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to update user id in user db")
		}

		err = p.indexHasParentChild(permission.GithubId, permission.Link)
		if err != nil {
			p.log.WithError(err).Errorf("failed to check has parent/child for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to check parent level")
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make add user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make add user transaction")
	}

	err = p.SendDeleteUser(msg.RequestId, user)
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to publish users")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) addUser(link, username, accessLevel string) (*data.Permission, error) {
	typeTo, err := p.getLinkType(link, pqueue.NormalPriority)
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting link type api")
	}

	isHere, err := p.isUserInSubmodule(link, username, typeTo)
	if err != nil {
		return nil, errors.Wrap(err, "some error while checking user for link")
	}

	if isHere {
		return nil, errors.New("user is already in submodule")
	}

	permission, err := github.GetPermission(
		p.pqueues.SuperUserPQueue,
		any(p.githubClient.AddUserFromApi),
		[]any{any(typeTo), any(link), any(username), any(accessLevel)},
		pqueue.NormalPriority,
	)
	if err != nil {
		return nil, errors.Wrap(err, "some error while adding user from api")
	}

	if permission == nil {
		return nil, errors.New("something wrong with adding user")
	}

	return permission, nil
}
