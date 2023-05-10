package processor

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/acs/github-module/internal/github"
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

func (p *processor) HandleUpdateUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateUpdateUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}
	msg.Link = strings.ToLower(msg.Link)

	user, err := p.checkUserExistence(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to check user existence for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to check user existence")
	}

	msg.Type, err = p.getLinkType(msg.Link, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get link type from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting link type api")
	}

	isHere, err := p.isUserInSubmodule(msg.Link, msg.Username, msg.Type)
	if err != nil {
		p.log.WithError(err).Errorf("failed to check user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while checking user from api")
	}
	if !isHere {
		p.log.Errorf("user is not in submodule from API for message action with id `%s`", msg.RequestId)
		return errors.New("user is not in submodule")
	}

	err = p.updateUser(data.Permission{
		RequestId:   msg.RequestId,
		UserId:      user.Id,
		GithubId:    user.GithubId,
		Username:    user.Username,
		Link:        msg.Link,
		Type:        msg.Type,
		AccessLevel: msg.AccessLevel,
	})
	if err != nil {
		p.log.Errorf("failed to update user for message action with id `%s`", msg.RequestId)
		return errors.New("failed to update user")
	}

	err = p.indexHasParentChild(user.GithubId, msg.Link)
	if err != nil {
		p.log.WithError(err).Errorf("failed to check has parent/child for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to check parent level")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) updateUser(info data.Permission) error {
	permission, err := github.GetPermission(
		p.pqueues.SuperUserPQueue,
		any(p.githubClient.UpdateUserFromApi),
		[]any{any(info.Type), any(info.Link), any(info.Username), any(info.AccessLevel)},
		pqueue.NormalPriority)
	if err != nil {
		return errors.Wrap(err, "some error while updating user from api")
	}

	if permission == nil {
		return errors.Errorf("something wrong with updating user from api")
	}

	permission.GithubId = info.GithubId
	permission.RequestId = info.RequestId
	permission.UserId = info.UserId
	permission.Link = info.Link

	err = p.permissionsQ.FilterByGithubIds(permission.GithubId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{
		Username:    &permission.Username,
		AccessLevel: &permission.AccessLevel,
	})
	if err != nil {
		return errors.Wrap(err, "failed to update user in permission db")
	}

	return nil
}

func (p *processor) checkUserExistence(username string) (*data.User, error) {
	dbUser, err := p.usersQ.FilterByUsernames(username).Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from user db")
	}

	if dbUser == nil {
		return nil, errors.New("no user with such username")
	}

	userApi, err := github.GetUser(p.pqueues.UserPQueue, any(p.githubClient.GetUserFromApi), []any{any(username)}, pqueue.NormalPriority)
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting user from api")
	}

	if userApi == nil {
		return nil, errors.Errorf("something wrong with user from api")
	}

	return dbUser, nil
}
